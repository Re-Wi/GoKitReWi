package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	baseURL        = "http://localhost:8888"
	descriptionURL = baseURL + "description"
	configURL      = baseURL + "config"
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

			err = netM.BuildReqURL(platform, dependency, project, "version.txt")
			if err != nil {
				fmt.Printf("BuildReqURL failed: %v \n", err)
			} else {
				latest, err = netM.GetRemoteVersion(ctx)
			}
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
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download sniffer description",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		platform, _ := cmd.Flags().GetString("platform")
		dependency, _ := cmd.Flags().GetString("dependency")
		project, _ := cmd.Flags().GetString("project")

		if platform == "" || dependency == "" || project == "" || output == "" {
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
			if output == "" {
				output = "output.json"
				fmt.Println("  --output")
			}
			fmt.Printf("platform: %s, dependency: %s, project: %s", platform, dependency, project)
		}

		// 初始化网络管理器
		nm := helpers.NewNetManager()
		nm.BaseURL = baseURL
		nm.ReqURL = "/download"
		nm.ReqMethod = "GET"
		nm.ReqHeaders = []string{
			"Authorization: Bearer token",
			"X-Custom-Header: value",
		}
		nm.AllowInsecure = false
		nm.FollowRedirects = true
		nm.Timeout = 30 * time.Second
		nm.Retries = 5

		err := nm.BuildReqURL(platform, dependency, project, output)
		if err != nil {
			fmt.Printf("BuildReqURL failed: %v \n", err)
		} else {
			// 执行下载
			err = nm.DownloadFile(output)
			if err != nil {
				fmt.Printf("Download failed: %v \n", err)
			} else {
				fmt.Printf("file: %s\n", output)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(snifferCmd)

	// 添加子命令
	snifferCmd.AddCommand(checkCmd, fetchCmd)

	// 添加参数
	fetchCmd.Flags().StringP("output", "o", "", "Output file path")
	fetchCmd.Flags().StringP("platform", "p", "", "平台名称")
	fetchCmd.Flags().StringP("dependency", "d", "", "依赖组件名称")
	fetchCmd.Flags().StringP("project", "j", "", "项目名称")

	checkCmd.Flags().StringP("platform", "p", "", "平台名称")
	checkCmd.Flags().StringP("dependency", "d", "", "依赖组件名称")
	checkCmd.Flags().StringP("project", "j", "", "项目名称")
	checkCmd.Flags().BoolP("verbose", "v", false, "显示详细输出")
	checkCmd.Flags().BoolP("server", "s", false, "请求服务器而跳过 git仓库 检查")
}
