package cmd

import (
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

var config = helpers.NetManager{
	BaseURL: baseURL,
}

// 网络请求
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Send HTTP request with custom parameters",
	RunE:  config.SendRequest,
}

func init() {
	rootCmd.AddCommand(requestCmd)
	requestCmd.Flags().StringVarP(&config.ReqURL, "url", "u", "", "Target URL (required)")
	requestCmd.Flags().StringVarP(&config.ReqMethod, "method", "m", "GET", "HTTP method")
	requestCmd.Flags().StringArrayVarP(&config.ReqHeaders, "header", "H", []string{}, "Request headers (key:value)")
	requestCmd.Flags().StringVarP(&config.ReqBody, "body", "b", "", "Request body")
	requestCmd.Flags().DurationVar(&config.Timeout, "timeout", 30*time.Second, "Request timeout")
	_ = requestCmd.MarkFlagRequired("url")
}
