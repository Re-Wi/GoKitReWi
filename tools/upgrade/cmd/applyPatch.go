package cmd

import (
	"fmt"
	"os"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/icedream/go-bsdiff"
	"github.com/spf13/cobra"
)

var patchCmd = &cobra.Command{
	Use:   "patch <oldfile> <patchfile> <newfile>",
	Short: "Apply binary patch to create new file",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		oldFilePath := args[0]
		patchFilePath := args[1]
		newFilePath := args[2]

		oldFile, err := os.Open(oldFilePath)
		defer oldFile.Close()
		helpers.MustDo("Error opening old file", nil)

		newFile, err := os.Create(newFilePath)
		defer newFile.Close()
		helpers.MustDo("Error creating new file", err)

		patchFile, err := os.Open(patchFilePath)
		defer patchFile.Close()
		helpers.MustDo("Error opening patch file", err)

		err = bsdiff.Patch(oldFile, newFile, patchFile)
		helpers.MightDo("Error applying patch", err)

		fmt.Println("Successfully created new file !")
	},
}

func init() {
	rootCmd.AddCommand(patchCmd)
}
