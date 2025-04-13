package cmd

import (
	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// applyCmd 表示应用补丁的子命令
var applyCmd = &cobra.Command{
	Use:   "apply-patch",
	Short: "应用补丁文件生成新版本",
	Long: `使用旧文件和补丁文件生成新版本文件
	
示例：
  diff-tool apply-patch old.bin new.bin patch.xd
  diff-tool apply-patch old.bin new.bin patch.xd --block-size 8192`,
	Args:    cobra.ExactArgs(3),        // 强制三个位置参数
	PreRunE: helpers.ValidateApplyArgs, // 参数预校验
	RunE:    helpers.RunApplyPatch,     // 主执行逻辑
}

func init() {
	rootCmd.AddCommand(applyCmd)
	// 添加命令行标志
	applyCmd.Flags().IntP(
		"block-size",
		"b",
		4, // 默认值 4 KB
		"补丁解码块大小（单位：KB）",
	)
}
