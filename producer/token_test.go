package producer

import (
	"fmt"
	"testing"
)

func TestGenerateAccessToken(t *testing.T) {
	tokenLength := 32 // 定义令牌长度
	access_token, err := GenerateAccessToken(tokenLength)
	if err != nil {
		fmt.Println("生成令牌时发生错误:", err)
		return
	}

	fmt.Println("生成的访问令牌:", access_token)
}
