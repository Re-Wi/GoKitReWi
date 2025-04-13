package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

func (nm *NetManager) DownloadFile(filePath string) error {

	// 准备HTTP客户端配置
	client := nm.prepareHTTPClient()

	var lastErr error
	var resp *http.Response

	// 重试逻辑
	for attempt := 0; attempt <= nm.Retries; attempt++ {
		// 创建新的请求（每次重试都需要新请求）
		req, err := nm.createRequest(nm.ReqURL)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// 发送请求
		resp, err = client.Do(req)
		if err != nil {
			lastErr = err
			if nm.shoulRetry(attempt, nil, err) {
				nm.logRetry(attempt, err)
				continue
			}
			break
		}

		// 处理响应
		nm.RespCode = resp.StatusCode
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			if nm.shoulRetry(attempt, resp, nil) {
				nm.logRetry(attempt, lastErr)
				continue
			}
			break
		}

		// 成功响应，处理文件下载
		if err := nm.saveResponseToFile(resp, filePath); err != nil {
			_ = resp.Body.Close()
			return fmt.Errorf("failed to save file: %w", err)
		}

		_ = resp.Body.Close()
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("request failed after %d attempts: %w", nm.Retries+1, lastErr)
	}
	return fmt.Errorf("unexpected error occurred")
}

func (nm *NetManager) prepareHTTPClient() *http.Client {
	// 复制基础客户端配置
	client := *nm.HTTPClient

	// 配置TLS
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport = transport.Clone()
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = nm.AllowInsecure
		client.Transport = transport
	}

	// 配置重定向
	if !nm.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &client
}

func (nm *NetManager) createRequest(fullURL string) (*http.Request, error) {
	var body io.Reader
	if nm.ReqBody != "" {
		body = strings.NewReader(nm.ReqBody)
	}

	req, err := http.NewRequest(nm.ReqMethod, fullURL, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for _, h := range nm.ReqHeaders {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			nm.Logger.Warn("Invalid header format", zap.String("header", h))
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		req.Header.Add(key, value)
	}

	return req, nil
}

func (nm *NetManager) shoulRetry(attempt int, resp *http.Response, err error) bool {
	if attempt >= nm.Retries {
		return false
	}

	// 网络错误自动重试
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return true
		}
		return true
	}

	// 5xx状态码重试
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}

	return false
}

func (nm *NetManager) logRetry(attempt int, err error) {
	nm.Logger.Info("Retrying request",
		zap.Int("attempt", attempt+1),
		zap.Int("max_retries", nm.Retries),
		zap.Error(err),
	)
	time.Sleep(nm.calculateBackoff(attempt))
}

func (nm *NetManager) calculateBackoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

func (nm *NetManager) saveResponseToFile(resp *http.Response, filePath string) error {
	// 处理相对路径，转换为绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// 创建目标目录（确保所有父目录都存在）
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// 创建目标文件（使用绝对路径）
	file, err := os.Create(absPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			nm.Logger.Error("Failed to close file", zap.Error(closeErr))
		}
	}()

	// 限制读取大小
	limitedReader := &io.LimitedReader{
		R: resp.Body,
		N: nm.MaxBodySize,
	}

	// 复制数据（使用缓冲写入提高性能）
	if _, err := io.CopyBuffer(file, limitedReader, make([]byte, 32*1024)); err != nil {
		// 删除可能已写入的部分文件
		_ = os.Remove(absPath)
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 检查是否超出最大限制
	if limitedReader.N <= 0 {
		_ = os.Remove(absPath)
		return fmt.Errorf("response body exceeds maximum allowed size of %d bytes", nm.MaxBodySize)
	}

	// 确保数据写入磁盘
	if err := file.Sync(); err != nil {
		nm.Logger.Warn("Failed to sync file to disk", zap.Error(err))
	}

	return nil
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

// 构建完整请求URL
func (netM *NetManager) BuildReqURL(platform, dependency, project string, filePath string) error {
	parsedURL, err := url.Parse(netM.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	parsedURL.Path = fmt.Sprintf("/%s/%s/%s/%s",
		url.PathEscape(platform),
		url.PathEscape(dependency),
		url.PathEscape(project),
		filePath)

	netM.ReqURL = parsedURL.String()
	fmt.Printf("ReqURL: %v \n", netM.ReqURL)
	return nil
}

// 获取远程版本信息
func (netM *NetManager) GetRemoteVersion(ctx context.Context) (string, error) {
	var (
		lastError error
	)
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
				zap.String("path", netM.ReqURL))
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
