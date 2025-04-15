package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

func RunGenerate(cmd *cobra.Command, args []string) error {
	config := helpers.DiffGenerator{
		RepoURL:   helpers.MustGetString(cmd, "repo"),
		BaseRef:   helpers.MustGetString(cmd, "base"),
		TargetRef: helpers.MustGetString(cmd, "target"),
		OutputDir: helpers.MustGetString(cmd, "output"),
		Workers:   helpers.MustGetInt(cmd, "workers"),
	}

	if err := config.Generate(); err != nil {
		return fmt.Errorf("\n❌ 差异生成失败: %w", err)
	}

	absPath, _ := filepath.Abs(config.OutputDir)
	fmt.Printf("\n✅ 差异生成成功！\n输出目录: %s\n", absPath)
	return nil
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成版本差异文件",
	Long:  `通过 git 工具自动生成所有差异文件`,
	RunE:  RunGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("repo", "r", "", "Git仓库URL (必填)")
	generateCmd.Flags().StringP("base", "b", "", "基准版本 (必填)")
	generateCmd.Flags().StringP("target", "t", "HEAD", "目标版本 (默认HEAD)")
	generateCmd.Flags().StringP("output", "o", "./vX.X.X", "输出目录")
	generateCmd.Flags().IntP("workers", "w", 4, "并行工作数")

	generateCmd.MarkFlagRequired("repo")
	generateCmd.MarkFlagRequired("base")
}
