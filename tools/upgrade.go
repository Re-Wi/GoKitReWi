package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// 请求详细信息的模拟数据结构
type RequestDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 模拟的键值对数据
var mockData = map[string]string{
	"example_key_1": "This is the value for example_key_1",
	"example_key_2": "This is the value for example_key_2",
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "A CLI tool for compression, decompression, and request details",
		Long:  `A simple CLI tool that supports creating zip files, extracting zip files, and fetching detailed information by key.`,
	}

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
				fmt.Printf("Error creating tar.gz: %v\n", err)
			} else {
				fmt.Printf("Successfully created tar.gz file: %s (with highest compression)\n", output)
			}
		},
	}

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

	// 请求详细信息命令
	var requestCmd = &cobra.Command{
		Use:   "request [key]",
		Short: "Get detailed information by key",
		Long:  `Fetches detailed information for the specified key from a mock data store.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			detail, err := getRequestDetail(key)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				jsonDetail, _ := json.MarshalIndent(detail, "", "  ")
				fmt.Printf("Request Detail:\n%s\n", string(jsonDetail))
			}
		},
	}

	// 将子命令添加到根命令
	rootCmd.AddCommand(compressCmd)
	rootCmd.AddCommand(decompressCmd)
	rootCmd.AddCommand(requestCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// 获取请求详细信息
func getRequestDetail(key string) (*RequestDetail, error) {
	value, exists := mockData[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", key)
	}
	return &RequestDetail{Key: key, Value: value}, nil
}
