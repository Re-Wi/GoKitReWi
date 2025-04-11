package main

import (
	"context"
	"os"
	"runtime"

	"github.com/konsorten/go-xdelta"
)

func main() {
	// 1. 打开文件流
	oldFile, _ := os.Open("./tools/setting.txt")
	defer oldFile.Close()

	newFile, _ := os.Open("./tools/hhh.txt")
	defer newFile.Close()

	patchFile, _ := os.Create("./tools/patch.xd")
	defer patchFile.Close()

	// 2. 配置编码选项
	options := xdelta.EncoderOptions{
		BlockSizeKB: 4,                               // 块大小 4 MB ( 4*1024 )
		FileID:      "v1.0.0",                        // 版本标识
		FromFile:    oldFile,                         // 旧文件
		ToFile:      newFile,                         // 新文件
		PatchFile:   patchFile,                       // 补丁输出
		Header:      []byte("metadata:checksum=123"), // 自定义头部
		EnableStats: true,                            // 启用统计
	}

	// 3. 创建编码器
	enc, err := xdelta.NewEncoder(options)
	if err != nil {
		panic("编码器初始化失败: " + err.Error())
		// return err
	}

	defer enc.Close()

	// create the patch
	err = enc.Process(context.TODO())
	enc.DumpStatsToStdout()
	if err != nil {
		// return err
		panic("Process失败: " + err.Error())
	}

	MyDecoder()
}

func MyDecoder() {
	// 打开源文件、目标文件和补丁文件
	sourceFile, err := os.Open("./tools/hhh.txt")
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	targetFile, err := os.Create("./tools/target.txt")
	if err != nil {
		panic(err)
	}
	defer targetFile.Close()

	patchFile, err := os.Open("./tools/patch.xd")
	if err != nil {
		panic(err)
	}
	defer patchFile.Close()

	// 配置 DecoderOptions
	options := xdelta.DecoderOptions{
		BlockSizeKB: 4, // 设置块大小为 4 MB ( 4*1024 )
		FileID:      "v1.0.0",
		FromFile:    sourceFile,
		ToFile:      targetFile,
		PatchFile:   patchFile,
		EnableStats: true, // 启用统计功能
	}

	// 创建解码器
	decoder, err := xdelta.NewDecoder(options)
	if err != nil {
		panic(err)
	}
	defer func() {
		// 确保解码器资源被释放
		runtime.SetFinalizer(decoder, nil)
	}()
	defer decoder.Close()
	// 使用解码器进行解码操作
	// （假设这里有一个方法可以触发解码过程）
	err = decoder.Process(context.TODO())
	decoder.DumpStatsToStdout()
	if err != nil {
		// return err
		panic("Process失败: " + err.Error())
	}
}
