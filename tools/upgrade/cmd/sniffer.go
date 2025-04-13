package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// snifferCmd 父命令
var snifferCmd = &cobra.Command{
	Use:   "sniffer",
	Short: "Sniffer management",
	Long:  ` Sniffer, analyzer. You can check for updates, get the latest version number, get files, get configurations.`,
}

// checkCmd 智能环境检测
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for available sniffers",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		server, _ := cmd.Flags().GetBool("server")
		platform, _ := cmd.Flags().GetString("platform")
		dependency, _ := cmd.Flags().GetString("dependency")
		project, _ := cmd.Flags().GetString("project")

		isGitRepo := false
		// 环境检测逻辑
		if !server {
			isGitRepo = helpers.CheckGitRepository()
		}
		if verbose {
			if server {
				fmt.Println("Request the server and skip the git repository check")
			} else if isGitRepo {
				fmt.Println("Currently in a Git repository environment")
			} else {
				fmt.Println("Currently in a non-Git repository environment")
			}
		}

		// 参数验证逻辑
		if !isGitRepo {
			if platform == "" || dependency == "" || project == "" {
				fmt.Println("Error: Parameters required in non-Git mode")
				fmt.Println("Missing parameters:")
				if platform == "" {
					fmt.Println("  --platform")
				}
				if dependency == "" {
					fmt.Println("  --dependency")
				}
				if project == "" {
					fmt.Println("  --project")
				}
				fmt.Printf("platform: %s, dependency: %s, project: %s", platform, dependency, project)
			}
		}

		// 版本检查逻辑
		var (
			latest string
			err    error
		)

		if isGitRepo {
			latest, err = helpers.GetGitVersion()
		} else {
			// 创建 NetManager 实例
			netM := helpers.NewNetManager()
			// 设置基础配置
			netM.BaseURL = baseURL
			netM.ReqMethod = "GET"
			netM.ReqHeaders = []string{
				"User-Agent: MyClient/1.0",
				"Authorization: Bearer your_token",
			}
			netM.Timeout = 15 * time.Second
			netM.Retries = 3
			netM.AllowInsecure = false // 生产环境建议为 false

			// 替换默认日志记录器
			logger, _ := zap.NewDevelopment()
			defer logger.Sync()
			netM.Logger = logger
			// netM.Logger = zap.NewExample() // 替换为实际日志配置

			// 调用 GetRemoteVersion 方法
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			latest, err = netM.GetRemoteVersion(ctx, platform, dependency, project)

			if verbose {
				fmt.Printf("StatusCode: %v, ReqURL: %v \n", netM.RespCode, netM.ReqURL)
				if err != nil {
					fmt.Printf("StatusCode: %v, Error: %v \n", netM.RespCode, err)
				}
			}
		}
		if verbose {
			fmt.Printf("latest: %v, err: %v \n", latest, err)
		}
		if err != nil {
			fmt.Printf("Check failure: %v \n", err)
		} else {
			fmt.Printf("latest version: %s\n", latest)
		}
	},
}

// descCmd 获取描述文件
var descCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download sniffer description",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "sniffer-description.yaml"
		}

		if err := downloadFile(descriptionURL, output); err != nil {
			fmt.Printf("Download failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Description saved to %s\n", output)
	},
}

// configCmd 获取配置
var configCmd = &cobra.Command{
	Use:   "fetch-config",
	Short: "Download sniffer config",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "sniffer-config.yaml"
		}

		if err := downloadFile(configURL, output); err != nil {
			fmt.Printf("Download failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Config saved to %s\n", output)
	},
}

const (
	baseURL        = "http://localhost:8888"
	descriptionURL = baseURL + "description"
	configURL      = baseURL + "config"
)

func init() {
	rootCmd.AddCommand(snifferCmd)

	// 添加子命令
	snifferCmd.AddCommand(checkCmd, descCmd, configCmd)

	// 公共参数
	descCmd.Flags().StringP("output", "o", "", "Output file path")
	configCmd.Flags().StringP("output", "o", "", "Output file path")
	// 添加参数
	checkCmd.Flags().StringP("platform", "p", "", "目标平台名称 (非 Git 环境必填)")
	checkCmd.Flags().StringP("dependency", "d", "", "依赖组件名称 (非 Git 环境必填)")
	checkCmd.Flags().StringP("project", "j", "", "项目标识名称 (非 Git 环境必填)")
	checkCmd.Flags().BoolP("verbose", "v", false, "显示详细输出")
	checkCmd.Flags().BoolP("server", "s", false, "请求服务器而跳过 git仓库 检查")

	// 可以添加以下增强功能：
	// 1. 添加超时控制
	var timeout time.Duration
	checkCmd.Flags().DurationVar(&timeout, "timeout", 10*time.Second, "Request timeout")

	// 2. 添加重试机制
	var retries int
	descCmd.Flags().IntVar(&retries, "retries", 3, "Download retry attempts")

	// 3. 添加校验功能
	var checksum string
	descCmd.Flags().StringVar(&checksum, "checksum", "", "File checksum verification")
}

func downloadFile(url string, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(output, data, 0644)
}
