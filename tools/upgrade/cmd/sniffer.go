package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
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
				fmt.Println("Error: Non-Git environments must specify the following parameters:")
				fmt.Println("  --platform    Platform name")
				fmt.Println("  --dependency  Dependency name (depending on the largest hardware or software or system version)")
				fmt.Println("  --project     project name")
				_ = fmt.Errorf("platform: %s, dependency: %s, project: %s", platform, dependency, project)
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
			config := helpers.NetManager{
				BaseURL: baseURL,
			}
			latest, err = config.GetRemoteVersion(platform, dependency, project)
			if verbose {
				fmt.Printf("StatusCode: %v, ReqURL: %v \n", config.RespCode, config.ReqURL)
				if err != nil {
					fmt.Printf("StatusCode: %v, Error: %v \n", config.RespCode, err)
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

// versionCmd 获取最新版本号
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get latest version number",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getLatestVersion()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(version)
	},
}

// descCmd 获取描述文件
var descCmd = &cobra.Command{
	Use:   "fetch-description",
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
	baseURL         = "http://localhost:8888"
	versionEndpoint = "latest"
	descriptionURL  = baseURL + "description"
	configURL       = baseURL + "config"
)

func init() {
	rootCmd.AddCommand(snifferCmd)

	// 添加子命令
	snifferCmd.AddCommand(checkCmd, versionCmd, descCmd, configCmd)

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

// 实际业务逻辑函数
func checkForUpdates() (string, error) {
	resp, err := http.Get(baseURL + versionEndpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func getLatestVersion() (string, error) {
	return checkForUpdates() // 复用检查逻辑
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
