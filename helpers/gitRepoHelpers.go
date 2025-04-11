package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// CheckGitRepo 检查当前目录是否为Git仓库
func CheckGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}

// GetCurrentBranch 获取当前分支名称
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取分支失败: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// 解析 Git 错误信息
func ParseGitError(output string, originalErr error) error {
	switch {
	case strings.Contains(output, "Permission denied (publickey)"):
		return fmt.Errorf(`SSH认证失败 (%w)
请检查以下配置:
1. SSH密钥是否存在: ls -al ~/.ssh
2. 公钥是否已添加到代码平台
3. 测试连接: ssh -T git@gitee.com`, originalErr)

	case strings.Contains(output, "Could not resolve hostname"):
		return fmt.Errorf("无法解析仓库地址 (%w)\n请检查网络连接或仓库URL", originalErr)

	case strings.Contains(output, "Repository not found"):
		return fmt.Errorf("仓库不存在或无访问权限 (%w)", originalErr)

	default:
		return fmt.Errorf("%s → %w", output, originalErr)
	}
}

// CheckRemoteUpdates 检查远程是否有更新
func CheckRemoteUpdates() (bool, error) {
	// 执行 git fetch 并捕获错误输出
	fetchCmd := exec.Command("git", "fetch", "--all")
	var stderr bytes.Buffer
	fetchCmd.Stderr = &stderr

	if err := fetchCmd.Run(); err != nil {
		errOutput := strings.TrimSpace(stderr.String())
		errorMsg := ParseGitError(errOutput, err)
		return false, fmt.Errorf("远程仓库访问失败: %w", errorMsg)
	}

	// 比较本地和远程差异
	branch, err := GetCurrentBranch()
	if err != nil {
		return false, err
	}

	diffCmd := exec.Command("git", "log", "HEAD..origin/"+branch, "--oneline")
	output, err := diffCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("检查差异失败: %w", err)
	}

	return len(bytes.TrimSpace(output)) > 0, nil
}

// PullUpdates 执行代码更新
func PullUpdates() error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AutoSyncGitRepo 自动同步流程
func AutoSyncGitRepo() error {
	// 步骤1：确认是Git仓库
	if !CheckGitRepo() {
		return fmt.Errorf("当前目录不是Git仓库")
	}

	// 步骤2：检查更新
	needUpdate, err := CheckRemoteUpdates()
	if err != nil {
		return err
	}

	if !needUpdate {
		fmt.Println("当前代码已是最新版本")
		return nil
	}

	// 步骤3：执行更新
	fmt.Println("发现新版本，开始更新代码...")
	if err := PullUpdates(); err != nil {
		return fmt.Errorf("代码更新失败: %w", err)
	}

	fmt.Println("代码更新成功！")
	return nil
}

// validateSyncArgs 参数验证
func ValidateSyncArgs(cmd *cobra.Command, _ []string) error {
	repoPath, _ := cmd.Flags().GetString("path")

	// 验证路径有效性
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return fmt.Errorf("仓库路径不存在: %s", repoPath)
	}

	// 验证是否为Git仓库
	if !IsGitRepo(repoPath) {
		return fmt.Errorf("路径不是Git仓库: %s", repoPath)
	}

	return nil
}

// runCodeSync 执行同步逻辑
func RunCodeSync(cmd *cobra.Command, _ []string) error {
	// 解析参数
	repoPath, _ := cmd.Flags().GetString("path")
	forceSync, _ := cmd.Flags().GetBool("force")
	targetBranch, _ := cmd.Flags().GetString("branch")

	// 获取绝对路径
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("路径解析失败: %w", err)
	}

	fmt.Printf("正在检查代码更新 [仓库: %s]\n", filepath.Base(absPath))

	// 执行同步逻辑
	if err := SyncGitRepo(absPath, targetBranch, forceSync); err != nil {
		return fmt.Errorf("同步失败: %w", err)
	}

	fmt.Println("✅ 代码仓库状态已同步")
	return nil
}

// syncGitRepo 核心同步逻辑
func SyncGitRepo(repoPath, branch string, force bool) error {
	// 切换工作目录
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(repoPath); err != nil {
		return err
	}

	// 分支处理
	if branch != "" {
		if err := SwitchBranch(branch); err != nil {
			return err
		}
	}

	// 强制同步模式
	if force {
		return exec.Command("git", "reset", "--hard", "HEAD").Run()
	}

	// 常规同步流程
	return AutoSyncGitRepo()
}

// 辅助函数
func IsGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	_, err := os.Stat(gitPath)
	return !os.IsNotExist(err)
}

func SwitchBranch(branch string) error {
	if out, err := exec.Command("git", "checkout", branch).CombinedOutput(); err != nil {
		return fmt.Errorf("切换分支失败: %s → %v", string(out), err)
	}
	return nil
}

// 平台配置信息
var PlatformConfig = map[string]struct {
	TestHost  string
	HelpURL   string
	SSHConfig string
}{
	"github": {
		TestHost:  "github.com",
		HelpURL:   "https://docs.github.com/zh/authentication/connecting-to-github-with-ssh",
		SSHConfig: "Host github.com\n  IdentityFile ~/.ssh/github_key",
	},
	"gitee": {
		TestHost:  "gitee.com",
		HelpURL:   "https://gitee.com/help/articles/4181",
		SSHConfig: "Host gitee.com\n  IdentityFile ~/.ssh/gitee_key",
	},
	"gitlab": {
		TestHost:  "gitlab.com",
		HelpURL:   "https://docs.gitlab.com/ee/user/ssh.html",
		SSHConfig: "Host gitlab.com\n  IdentityFile ~/.ssh/gitlab_key",
	},
}

// detectGitRemote 检测远程仓库信息
func DetectGitRemote() (string, string, error) {
	// 检查是否是Git仓库
	if !CheckGitRepo() {
		return "", "", fmt.Errorf("当前目录不是Git仓库")
	}

	// 获取远程URL
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("获取远程信息失败: %w", err)
	}

	// 解析第一个远程URL
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "origin") && strings.Contains(line, "(push)") {
			// 提取URL
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			url := parts[1]

			// 识别平台
			if platform := ParseGitPlatform(url); platform != "" {
				return url, platform, nil
			}
		}
	}

	return "", "", fmt.Errorf("无法识别代码平台")
}

// parseGitPlatform 解析代码平台
func ParseGitPlatform(url string) string {
	// 匹配常见平台
	patterns := map[string]*regexp.Regexp{
		"github": regexp.MustCompile(`(?i)github\.com[:/]`),
		"gitee":  regexp.MustCompile(`(?i)gitee\.com[:/]`),
		"gitlab": regexp.MustCompile(`(?i)gitlab\.com[:/]`),
	}

	for platform, re := range patterns {
		if re.MatchString(url) {
			return platform
		}
	}

	// 自定义域名检测
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		return "custom"
	}

	return ""
}
