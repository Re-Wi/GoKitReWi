package producer

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateAccessToken(tokenLength int) (string, error) {
	// 随机生成字节数
	randomBytes := make([]byte, tokenLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// 使用 base64 编码生成令牌
	token := base64.URLEncoding.EncodeToString(randomBytes)

	return token, nil
}
