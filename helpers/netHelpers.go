package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	ReqURL     string
	ReqMethod  string
	ReqHeaders []string
	ReqBody    string
	Timeout    time.Duration
)

func SendRequest(cmd *cobra.Command, args []string) error {
	// 创建请求体
	var reqBodyReader io.Reader
	if ReqBody != "" {
		reqBodyReader = strings.NewReader(ReqBody)
	}

	// 创建请求对象
	req, err := http.NewRequest(
		strings.ToUpper(ReqMethod),
		ReqURL,
		reqBodyReader,
	)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加请求头
	for _, h := range ReqHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("无效的请求头格式: %q，应使用 key:value 格式", h)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		req.Header.Add(key, value)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 输出结果
	fmt.Printf("状态码: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Println("\n响应头:")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println("\n响应体:")
	fmt.Println(string(respBody))
	return nil
}
