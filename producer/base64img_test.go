package producer

import (
	"fmt"
	"testing"
)

func TestUsername(t *testing.T) {
	// 创建一个字符串数组
	strArray := []string{"ReWi", "ABC", "BCA", "CCA", "DCA", "ECA", "FCA", "GCA", "HCA", "ICA"}
	// 替换为您的用户名,暂不能中文

	// 遍历字符串数组
	for _, value := range strArray {
		base64Image, err := GenerateInitialImage(value)
		if err != nil {
			fmt.Println("生成图像出错:", err)
			return
		}
		fmt.Println(base64Image)
	}
}
