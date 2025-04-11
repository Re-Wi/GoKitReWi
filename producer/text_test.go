package producer

import (
	"fmt"
	"testing"
)

func TestGenerateRandomTextFile(t *testing.T) {
	// 定义文件名和大小范围
	filename := "random.txt"
	minSizeKB := 300 // 最小 300 KB
	maxSizeKB := 600 // 最大 600 KB

	// 调用函数生成随机文件
	err := GenerateRandomTextFile(filename, minSizeKB, maxSizeKB)
	if err != nil {
		fmt.Println("生成文件时发生错误:", err)
	}
}
