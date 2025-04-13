package cmd

import (
	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// patchCmd 表示生成补丁文件的子命令
var cratePatchCmd = &cobra.Command{
	Use:   "create-patch",
	Short: "生成二进制差异补丁文件",
	Long: `根据旧文件和新文件生成二进制差异补丁文件，支持自定义块大小
	
示例：
  diff-tool create-patch old.bin new.bin patch.xd
  diff-tool create-patch old.bin new.bin patch.xd --block-size 8192`,
	Args:    cobra.ExactArgs(3),        // 必须包含三个位置参数
	PreRunE: helpers.ValidatePatchArgs, // 参数预校验
	RunE:    helpers.RunCreatePatch,    // 主执行函数
}

func init() {
	rootCmd.AddCommand(cratePatchCmd)
	// 添加命令行标志
	cratePatchCmd.Flags().IntP(
		"block-size",
		"b",
		0, // 默认值 0 KB
		"差异计算块大小（单位：KB）",
	)

}
