package cmd

import (
	"crypto/md5"
	"log"
	"os"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// 压缩命令
var compressCmd = &cobra.Command{
	Use:   "compress [output] [source1] [source2] ...",
	Short: "Create a tar.gz archive from multiple files and directories while preserving folder structure",
	Long: `Compresses the specified files and directories into a single tar.gz archive.
Supports multiple input files and directories, preserving the original folder structure.`,
	Args: cobra.MinimumNArgs(2), // 至少需要一个输出文件名和一个输入源
	Run: func(cmd *cobra.Command, args []string) {
		output := args[0]
		sources := args[1:]

		err := helpers.CreateTarGz(sources, output)
		if err != nil {
			log.Fatalf("Error creating tar.gz: %v\n", err)
		}
		log.Printf("Successfully created tar.gz file: %s (with highest compression)\n", output)

		hashPath := output + ".md5"
		// 计算 MD5 哈希值
		md5Hash, err := helpers.CalculateFileHash(output, md5.New)
		if err != nil {
			log.Fatalf("Error calculating hash: %v", err)
		}
		log.Printf("MD5 Hash: %s\n", md5Hash)
		// 写入哈希文件
		err = os.WriteFile(hashPath, []byte(md5Hash), 0644)
		if err != nil {
			log.Fatalf("Error writing hash file: %v", err)
		}
		log.Printf("Successfully created hash file: %s\n", hashPath)
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)
}
