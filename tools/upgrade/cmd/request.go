package cmd

import (
	"time"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

// 网络请求
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Send HTTP request with custom parameters",
	RunE:  helpers.SendRequest,
}

func init() {
	rootCmd.AddCommand(requestCmd)
	requestCmd.Flags().StringVarP(&helpers.ReqURL, "url", "u", "", "Target URL (required)")
	requestCmd.Flags().StringVarP(&helpers.ReqMethod, "method", "m", "GET", "HTTP method")
	requestCmd.Flags().StringArrayVarP(&helpers.ReqHeaders, "header", "H", []string{}, "Request headers (key:value)")
	requestCmd.Flags().StringVarP(&helpers.ReqBody, "body", "b", "", "Request body")
	requestCmd.Flags().DurationVar(&helpers.Timeout, "timeout", 30*time.Second, "Request timeout")
	_ = requestCmd.MarkFlagRequired("url")
}
