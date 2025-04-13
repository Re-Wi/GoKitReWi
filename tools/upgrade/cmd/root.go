package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "upgradeReWi",
	Short: "A CLI tool for compression, decompression, and info details",
	Long:  `A simple CLI tool that supports creating zip files, extracting zip files, and fetching detailed information by key.`,
}

func Execute() {
	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
