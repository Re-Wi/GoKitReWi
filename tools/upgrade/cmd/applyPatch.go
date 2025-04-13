package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bspatch"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch <oldfile> <patchfile> <newfile>",
	Short: "Apply binary patch to create new file",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		oldFile := args[0]
		patchFile := args[1]
		newFile := args[2]

		oldData, err := os.ReadFile(oldFile)
		if err != nil {
			log.Fatalf("Error reading old file: %v", err)
		}

		patchData, err := os.ReadFile(patchFile)
		if err != nil {
			log.Fatalf("Error reading patch file: %v", err)
		}

		newData, err := bspatch.Bytes(oldData, patchData)
		if err != nil {
			log.Fatalf("Error applying patch: %v", err)
		}

		if err := os.WriteFile(newFile, newData, 0644); err != nil {
			log.Fatalf("Error writing new file: %v", err)
		}

		fmt.Printf("Successfully created new file: %s\n", newFile)
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
}
