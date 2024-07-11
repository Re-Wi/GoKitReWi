package logger

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

type emailHook struct{}

// emailHook 实现了一个简单的发送邮件 Hook，其 Levels 方法返回 logrus.AllLevels，代表所有日志级别。Fire 方法接收一个 *logrus.Entry 参数，它包含了一条日志相关的所有信息，日志内容保存在 entry.Data Map 中，通过 entry.String() 能够获取日志的字符串格式内容。这里使用 fakeSendEmail 来模拟发送邮件。

// 作者：江湖十年
// 链接：https://juejin.cn/post/7212650952532148282
// 来源：稀土掘金
// 著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。

func (hook *emailHook) Levels() []logrus.Level {
	// return logrus.AllLevels // 所有日志级别都会执行 Fire 方法
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel}
}

func (hook *emailHook) Fire(entry *logrus.Entry) error {
	// 修改日志内容
	entry.Data["app"] = "email"
	// 发送邮件
	msg, _ := entry.String()
	fakeSendEmail(msg)
	return nil
}

func fakeSendEmail(msg string) {
	fmt.Printf("fakeSendEmail: %s", msg)
}

// newRotateHook() 利用第三方库 lumberjackrus 实现了日志切割和归档功能，并且能够将不同级别的日志输出到不同文件。lumberjackrus 是专门为 Logrus 而打造的文件日志 Hooks，其官方介绍为 local filesystem hook for Logrus。

// 作者：江湖十年
// 链接：https://juejin.cn/post/7212650952532148282
// 来源：稀土掘金
// 著作权归作者所有。商业转载请联系作者获得授权，非商业转载请注明出处。

func newRotateHook(fileName string, mylevel logrus.Level) logrus.Hook {
	hook, _ := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{ // 通用日志配置
			Filename:   fileName + "_general.log",
			MaxSize:    100,  // 日志文件在轮转之前的最大大小，默认 100 MB
			MaxBackups: 100,  // 保留旧日志文件的最大数量
			MaxAge:     100,  // 保留旧日志文件的最大天数
			Compress:   true, // 是否使用 gzip 对日志文件进行压缩归档
			LocalTime:  true, // 是否使用本地时间，默认 UTC 时间
		},
		mylevel,
		&logrus.TextFormatter{DisableColors: true},
		&lumberjackrus.LogFileOpts{ // 针对不同日志级别的配置
			logrus.TraceLevel: &lumberjackrus.LogFile{
				Filename:   fileName + "_trace.log",
				MaxSize:    100,  // 日志文件在轮转之前的最大大小，默认 100 MB
				MaxBackups: 100,  // 保留旧日志文件的最大数量
				MaxAge:     100,  // 保留旧日志文件的最大天数
				Compress:   true, // 是否使用 gzip 对日志文件进行压缩归档
				LocalTime:  true, // 是否使用本地时间，默认 UTC 时间
			},
			logrus.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   fileName + "_error.log",
				MaxSize:    100,  // 日志文件在轮转之前的最大大小，默认 100 MB
				MaxBackups: 100,  // 保留旧日志文件的最大数量
				MaxAge:     100,  // 保留旧日志文件的最大天数
				Compress:   true, // 是否使用 gzip 对日志文件进行压缩归档
				LocalTime:  true, // 是否使用本地时间，默认 UTC 时间
			},
		},
	)
	return hook
}

func InitLog(fileName, level string) *logrus.Logger {
	Log = logrus.New()
	Log.SetFormatter(new(LogFormatter)) // 设置日志输出格式，默认值为 logrus.TextFormatter
	Log.SetReportCaller(true)           // 设置日志是否记录被调用的位置，默认值为 false

	// 根据配置文件设置日志级别
	myLevel := logrus.DebugLevel
	var err error
	if level != "" {
		myLevel, err = logrus.ParseLevel(level)
		if err != nil {
			panic(any(fmt.Sprintf("日志级别不存在: %s", level)))
		}
		// Log.SetLevel(myLevel)
	}
	Log.SetLevel(myLevel) // 设置日志级别，默认值为 logrus.InfoLevel
	// if fileName != "" {
	// 	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|0666)
	// 	if err != nil {
	// 		panic(any(fmt.Sprintf("创建日志文件失败: %s", err.Error())))
	// 	}
	// 	Log.Out = file
	// }
	Log.AddHook(&emailHook{})
	Log.AddHook(newRotateHook(fileName, myLevel))

	return Log
}

type LogFormatter struct{}

func (l *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05.000")
	level := entry.Level
	logMsg := fmt.Sprintf("%s [%s]", timestamp, strings.ToUpper(level.String()))
	// 如果存在调用信息，且为error级别以上记录文件及行号
	if caller := entry.Caller; caller != nil {
		// 全路径切割，只获取项目相关路径，
		fp := filepath.Base(caller.File)
		logMsg = logMsg + fmt.Sprintf(" [%s:%d]", fp, caller.Line)
	}
	for k, v := range entry.Data {
		logMsg = logMsg + fmt.Sprintf(" [%s=%v]", k, v)
	}
	logMsg = logMsg + fmt.Sprintf(" : %s\n", entry.Message)
	return []byte(logMsg), nil
}
