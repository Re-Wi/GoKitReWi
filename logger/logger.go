package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Log *logrus.Logger

func InitLog(fileName, level string) *logrus.Logger {
	Log = logrus.New()
	Log.SetFormatter(new(LogFormatter))
	Log.SetReportCaller(true)

	// 根据配置文件设置日志级别
	if level != "" {
		l, err := logrus.ParseLevel(level)
		if err != nil {
			panic(any(fmt.Sprintf("日志级别不存在: %s", level)))
		}
		Log.SetLevel(l)
	} else {
		Log.SetLevel(logrus.DebugLevel)
	}
	if fileName != "" {
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|0666)
		if err != nil {
			panic(any(fmt.Sprintf("创建日志文件失败: %s", err.Error())))
		}
		Log.Out = file
	}

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
