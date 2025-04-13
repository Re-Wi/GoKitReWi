package cmd

import (
	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// syncCmd 表示代码同步命令
var syncCmd = &cobra.Command{
	Use:   "sync-code",
	Short: "同步代码仓库更新",
	Long: `自动检测并同步Git代码仓库更新
	
示例:
  # 同步当前目录仓库
  diff-tool sync-code
  
  # 同步指定目录仓库
  diff-tool sync-code --path /projects/my-repo`,
	Args:    cobra.NoArgs, // 不接受位置参数
	PreRunE: helpers.ValidateSyncArgs,
	RunE:    helpers.RunCodeSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	// 添加命令行参数
	syncCmd.Flags().StringP("path", "p", ".", "Git仓库路径")
	syncCmd.Flags().BoolP("force", "f", false, "强制同步（忽略检测结果）")
	syncCmd.Flags().StringP("branch", "b", "", "指定同步分支")
}
