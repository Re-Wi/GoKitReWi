package cmd

import (
	"fmt"
	"os"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/icedream/go-bsdiff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <oldfile> <newfile> <patchfile>",
	Short: "Generate binary diff between two files",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		oldFilePath := args[0]
		newFilePath := args[1]
		patchFilePath := args[2]

		oldFile, err := os.Open(oldFilePath)
		defer oldFile.Close()
		helpers.MustDo("Error opening old file", err)

		newFile, err := os.Open(newFilePath)
		defer newFile.Close()
		helpers.MustDo("Error opening new file", err)

		patchFile, err := os.Create(patchFilePath)
		defer patchFile.Close()
		helpers.MustDo("Error creating patch file", err)

		err = bsdiff.Diff(oldFile, newFile, patchFile)
		helpers.MightDo("Error generating diff", err)

		fmt.Println("Successfully generated patch file !")
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
