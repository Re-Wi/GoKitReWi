package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// 请求详细信息的模拟数据结构
type InfoDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// 模拟的键值对数据
var mockData = map[string]string{
	//作者
	"author": "ReWi",
	// 邮箱
	"email": "RejoiceWindow@yeah.net",
	//版本
	"version": "v0.0.0",
	// 描述
	"description": "增量升级工具",
	// 文档
	"doc": "up.rewi.xyz",
	// 仓库
	"repository": "https://github.com/Re-Wi/GoKitReWi.git",
}

// 获取请求详细信息
func getInfoDetail(key string) (*InfoDetail, error) {
	value, exists := mockData[key]
	if !exists {
		return nil, fmt.Errorf("key '%s' not found", key)
	}
	return &InfoDetail{Key: key, Value: value}, nil
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

func init() {
	rootCmd.AddCommand(infoCmd)
}
