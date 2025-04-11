package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// 请求详细信息的模拟数据结构
type InfoDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 模拟的键值对数据
var mockData = map[string]string{
	"example_key_1": "This is the value for example_key_1",
	"example_key_2": "This is the value for example_key_2",
}

// 获取请求详细信息
func getInfoDetail(key string) (*InfoDetail, error) {
	value, exists := mockData[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", key)
	}
	return &InfoDetail{Key: key, Value: value}, nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "A CLI tool for compression, decompression, and info details",
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
	var infoCmd = &cobra.Command{
		Use:   "info [key]",
		Short: "Get detailed information by key",
		Long:  `Fetches detailed information for the specified key from a mock data store.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			detail, err := getInfoDetail(key)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				jsonDetail, _ := json.MarshalIndent(detail, "", "  ")
				fmt.Printf("Info Detail:\n%s\n", string(jsonDetail))
			}
		},
	}

	// 网络请求
	var requestCmd = &cobra.Command{
		Use:   "request",
		Short: "Send HTTP request with custom parameters",
		RunE:  helpers.SendRequest,
	}
	requestCmd.Flags().StringVarP(&helpers.ReqURL, "url", "u", "", "Target URL (required)")
	requestCmd.Flags().StringVarP(&helpers.ReqMethod, "method", "m", "GET", "HTTP method")
	requestCmd.Flags().StringArrayVarP(&helpers.ReqHeaders, "header", "H", []string{}, "Request headers (key:value)")
	requestCmd.Flags().StringVarP(&helpers.ReqBody, "body", "b", "", "Request body")
	requestCmd.Flags().DurationVar(&helpers.Timeout, "timeout", 30*time.Second, "Request timeout")
	_ = requestCmd.MarkFlagRequired("url")

	// patchCmd 表示生成补丁文件的子命令
	var patchCmd = &cobra.Command{
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

	// 添加命令行标志
	patchCmd.Flags().IntP(
		"block-size",
		"b",
		4, // 默认值 4 KB
		"差异计算块大小（单位：KB）",
	)

	// 将子命令添加到根命令
	rootCmd.AddCommand(compressCmd)
	rootCmd.AddCommand(decompressCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(requestCmd)
	rootCmd.AddCommand(patchCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
