package cmd

import (
	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成版本差异文件",
	Long:  `比较Git仓库两个版本之间的差异并生成差异文件`,
	RunE:  helpers.RunGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("repo", "r", "", "Git仓库URL (必填)")
	generateCmd.Flags().StringP("base", "b", "", "基准版本 (必填)")
	generateCmd.Flags().StringP("target", "t", "HEAD", "目标版本 (默认HEAD)")
	generateCmd.Flags().StringP("output", "o", "./diffs", "输出目录")
	generateCmd.Flags().Bool("bin", false, "包含二进制文件")
	generateCmd.Flags().IntP("workers", "w", 4, "并行工作数")

	generateCmd.MarkFlagRequired("repo")
	generateCmd.MarkFlagRequired("base")

}
