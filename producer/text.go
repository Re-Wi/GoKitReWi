package producer

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// GenerateRandomTextFile 生成一个随机大小的文本文件（可打印ASCII字符）
func GenerateRandomTextFile(filename string, minSizeKB, maxSizeKB int) error {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 随机生成文件大小（单位：字节）
	fileSize := (rand.Intn(maxSizeKB-minSizeKB+1) + minSizeKB) * 1024

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 生成随机可打印ASCII内容（范围32-126）
	randomText := make([]byte, fileSize)
	for i := 0; i < fileSize; i++ {
		randomText[i] = byte(rand.Intn(95) + 32) // 95个可打印字符: 32(空格)~126(~)
	}

	// 加入换行符增加可读性（每100字符加一个换行）
	for i := 100; i < fileSize; i += 100 + rand.Intn(20) {
		if i >= fileSize {
			break
		}
		randomText[i] = '\n'
	}

	// 写入文件
	if _, err := file.Write(randomText); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Printf("成功生成文本文件 %s，大小 %.2f KB\n", filename, float64(fileSize)/1024)
	return nil
}
