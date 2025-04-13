package helpers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var ()

type NetManager struct {
	BaseURL    string
	ReqURL     string
	ReqMethod  string
	ReqHeaders []string
	ReqBody    string
	Timeout    time.Duration
	RespCode   int
}

func (netM *NetManager) SendRequest(cmd *cobra.Command, args []string) error {
	// 创建请求体
	var reqBodyReader io.Reader
	if netM.ReqBody != "" {
		reqBodyReader = strings.NewReader(netM.ReqBody)
	}

	// 创建请求对象
	req, err := http.NewRequest(
		strings.ToUpper(netM.ReqMethod),
		netM.ReqURL,
		reqBodyReader,
	)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加请求头
	for _, h := range netM.ReqHeaders {
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

// 获取远程版本信息
func (netM *NetManager) GetRemoteVersion(platform, dependency, project string) (string, error) {
	// 解析基础 URL
	parsedURL, err := url.Parse(netM.BaseURL)
	if err != nil {
		panic(err)
	}
	// 设置路径
	parsedURL.Path = fmt.Sprintf("/%s/%s/%s/version.txt",
		url.PathEscape(platform),
		url.PathEscape(dependency),
		url.PathEscape(project))

	// 获取完整的 URL 字符串
	netM.ReqURL = parsedURL.String()

	// 带超时的 HTTP 客户端
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(netM.ReqURL)
	if err != nil {
		return "", fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	netM.RespCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("The server returns an exception status code: %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read response failure: %v", err)
	}

	return strings.TrimSpace(string(content)), nil
}
