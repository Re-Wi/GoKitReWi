package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type NetManager struct {
	BaseURL         string        // 基础 URL，用于拼接请求路径
	ReqURL          string        // 实际请求的完整 URL
	ReqMethod       string        // HTTP 请求方法（如 GET、POST）
	ReqHeaders      []string      // 自定义请求头
	ReqBody         string        // 请求体内容
	Timeout         time.Duration // 请求超时时间
	Retries         int           // 最大重试次数
	HTTPClient      *http.Client  // 自定义 HTTP 客户端
	MaxBodySize     int64         // 响应体的最大大小限制
	AllowInsecure   bool          // 是否允许不安全的 TLS 连接
	FollowRedirects bool          // 是否跟随重定向
	Logger          *zap.Logger   // 日志记录器
	RespCode        int
}

func NewNetManager() *NetManager {
	return &NetManager{
		Timeout:     10 * time.Second,
		Retries:     3,
		MaxBodySize: 1 << 30, // 改为 1GB
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
				MaxIdleConnsPerHost: 20,
			},
		},
		Logger: zap.NewNop(),
	}
}

// 获取远程版本信息
func (netM *NetManager) GetRemoteVersion(ctx context.Context, platform, dependency, project string) (string, error) {
	parsedURL, err := url.Parse(netM.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	parsedURL.Path = fmt.Sprintf("/%s/%s/%s/version.txt",
		url.PathEscape(platform),
		url.PathEscape(dependency),
		url.PathEscape(project))

	var (
		lastError error
	)
	netM.ReqURL = parsedURL.String()
	for i := 0; i < netM.Retries; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", netM.ReqURL, nil)
		for _, header := range netM.ReqHeaders {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				req.Header.Add(key, value)
			}
		}
		// start := time.Now()

		resp, err := netM.HTTPClient.Do(req)
		if err != nil {
			lastError = fmt.Errorf("network error: %w", err)
			netM.Logger.Warn("Request failed",
				zap.Error(err),
				zap.Int("attempt", i+1))
			continue
		}

		netM.RespCode = resp.StatusCode
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			content, err := io.ReadAll(io.LimitReader(resp.Body, netM.MaxBodySize))
			if err != nil {
				return "", fmt.Errorf("read body: %w", err)
			}
			return strings.TrimSpace(string(content)), nil
		}
		// fmt.Printf("resp: %v, err: %v \n", resp, err)

		// 记录服务端错误
		if resp.StatusCode >= 500 {
			netM.Logger.Error("Server error",
				zap.Int("status", resp.StatusCode),
				zap.String("path", parsedURL.Path))
		}

		// 非重试场景直接返回
		if !netM.shouldRetry(resp) {
			break
		}

		// 指数退避
		backoff := time.Duration(math.Pow(2, float64(i))) * 100 * time.Millisecond
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", fmt.Errorf("after %d attempts, last status: %d, error: %w",
		netM.Retries, netM.RespCode, lastError)
}

func (netM *NetManager) shouldRetry(resp *http.Response) bool {
	if resp == nil {
		return true // 网络错误重试
	}
	// 5xx 状态码重试（除 501）
	return (resp.StatusCode >= 500 && resp.StatusCode != 501) ||
		resp.StatusCode == 429 || resp.StatusCode == 408
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
