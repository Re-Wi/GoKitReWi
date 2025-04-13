package helpers

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// mustGetString 安全获取字符串类型命令行参数
func MustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("致命错误: 获取参数 %s 失败 - %v", flagName, err))
	}
	return val
}

// mustGetBool 安全获取布尔类型命令行参数
func MustGetBool(cmd *cobra.Command, flagName string) bool {
	val, err := cmd.Flags().GetBool(flagName)
	if err != nil {
		panic(fmt.Sprintf("致命错误: 获取参数 %s 失败 - %v", flagName, err))
	}
	return val
}

// mustGetInt 安全获取整数类型命令行参数
func MustGetInt(cmd *cobra.Command, flagName string) int {
	val, err := cmd.Flags().GetInt(flagName)
	if err != nil {
		panic(fmt.Sprintf("致命错误: 获取参数 %s 失败 - %v", flagName, err))
	}
	return val
}

func MustDo(err error, msg string) {
	if err != nil {
		log.Fatalf("%s : %v", msg, err)
	}
}

func MightDo(err error, msg string) {
	if err != nil {
		log.Printf("%s : %v", msg, err)
	}
}
