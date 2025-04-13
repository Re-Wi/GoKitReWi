package cmd

import (
	"fmt"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// 解压命令
var decompressCmd = &cobra.Command{
	Use:   "decompress [zip_file] [target_dir]",
	Short: "Extract a zip archive to a target directory",
	Long:  `Extracts the contents of a zip archive to the specified target directory.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		zipFile := args[0]
		targetDir := args[1]
		err := helpers.ExtractTarGz(zipFile, targetDir)
		if err != nil {
			fmt.Printf("Error extracting zip: %v\n", err)
		} else {
			fmt.Printf("Successfully extracted zip to: %s\n", targetDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(decompressCmd)
}
