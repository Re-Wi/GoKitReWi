package producer

import (
	"fmt"
	"testing"
)

func TestGenerateRandomTextFile(t *testing.T) {
	// 定义文件名和大小范围
	filename := "random.txt"
	minSizeKB := 100 // 最小 100 KB
	maxSizeKB := 500 // 最大 500 KB

	// 调用函数生成随机文件
	err := GenerateRandomTextFile(filename, minSizeKB, maxSizeKB)
	if err != nil {
		fmt.Println("生成文件时发生错误:", err)
	}
}
