package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// sshCheckCmd 动态平台检测版
var sshCheckCmd = &cobra.Command{
	Use:   "ssh-check",
	Short: "智能诊断SSH连接问题",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. 检测Git仓库信息
		remoteURL, platform, err := helpers.DetectGitRemote()
		if err != nil {
			return fmt.Errorf("仓库检测失败: %w", err)
		}

		// 2. 获取平台配置
		config, ok := helpers.PlatformConfig[platform]
		if !ok {
			return fmt.Errorf("不支持的代码平台: %s", platform)
		}

		// 3. 执行连接测试
		fmt.Printf("测试连接至 [%s] 平台...\n", platform)
		testCmd := exec.Command("ssh", "-T", fmt.Sprintf("git@%s", config.TestHost))
		output, _ := testCmd.CombinedOutput()

		// 4. 输出结果
		if strings.Contains(string(output), "successfully authenticated") {
			fmt.Printf("✅ SSH认证正常 (%s)\n", remoteURL)
			return nil
		}

		// 5. 错误处理
		fmt.Printf(`🔴 [%s] SSH连接失败

=== 错误信息 ===
%s

=== 解决方案 ===
1. 生成专用密钥:
ssh-keygen -t ed25519 -f ~/.ssh/%s_key -C "your_email@example.com"

2. 添加SSH配置到 ~/.ssh/config:
%s

3. 查看公钥并添加到平台:
cat ~/.ssh/%s_key.pub

4. 测试连接:
ssh -T git@%s

官方指南: %s
`, platform, output, platform, config.SSHConfig, platform, config.TestHost, config.HelpURL)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(sshCheckCmd)
}
