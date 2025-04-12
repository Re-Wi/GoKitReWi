package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
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

		if err := ioutil.WriteFile(patchFile, patch, 0644); err != nil {
			log.Fatalf("Error writing patch file: %v", err)
		}

		fmt.Printf("Successfully generated patch file: %s\n", patchFile)
	},
}

var patchCmd = &cobra.Command{
	Use:   "patch <oldfile> <patchfile> <newfile>",
	Short: "Apply binary patch to create new file",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		oldFile := args[0]
		newFile := args[1]
		patchFile := args[2]

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

		if err := ioutil.WriteFile(newFile, newData, 0644); err != nil {
			log.Fatalf("Error writing new file: %v", err)
		}

		fmt.Printf("Successfully created new file: %s\n", newFile)
	},
}

func main() {

	var rootCmd = &cobra.Command{
		Use:   "bsdiff-tool",
		Short: "A CLI tool for binary diff/patch operations",
	}

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(patchCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
