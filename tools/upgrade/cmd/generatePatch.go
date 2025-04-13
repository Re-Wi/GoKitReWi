package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <oldfile> <newfile> <patchfile>",
	Short: "Generate binary diff between two files",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		oldFile := args[0]
		newFile := args[1]
		patchFile := args[2]

		oldData, err := os.ReadFile(oldFile)
		if err != nil {
			log.Fatalf("Error reading old file: %v", err)
		}

		newData, err := os.ReadFile(newFile)
		if err != nil {
			log.Fatalf("Error reading new file: %v", err)
		}

		patch, err := bsdiff.Bytes(oldData, newData)
		if err != nil {
			log.Fatalf("Error generating diff: %v", err)
		}

		if err := os.WriteFile(patchFile, patch, 0644); err != nil {
			log.Fatalf("Error writing patch file: %v", err)
		}

		fmt.Printf("Successfully generated patch file: %s\n", patchFile)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
