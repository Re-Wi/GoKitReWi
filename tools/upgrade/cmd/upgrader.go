package cmd

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/Re-Wi/GoKitReWi/helpers"
	"github.com/spf13/cobra"
)

var upgraderCmd = &cobra.Command{
	Use:   "upgrader",
	Short: "System upgrade tool",
	Long:  "Validate and apply system upgrade package",
	Run:   upgradeMain,
}

func verifyFileExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	}
	return nil
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func upgradeMain(cmd *cobra.Command, args []string) {

	tarPath, _ := cmd.Flags().GetString("input")
	targetDir, _ := cmd.Flags().GetString("output")
	if tarPath == "" || targetDir == "" {
		fatal("Input and output are required")
	}

	hashPath := tarPath + ".md5"

	// Step 1: Validate tar.gz hash
	expectedHashData, err := os.ReadFile(hashPath)
	if err != nil {
		fatal("read hash file: %v", err)
	}

	err = helpers.VerifyFileHash(tarPath, string(expectedHashData), md5.New)

	if err != nil {
		fatal("Tar validation failed: %v", err)
	}

	// Step 2: Extract tar.gz to temp dir
	patchTempDir, err := os.MkdirTemp("", "patch-")
	if err != nil {
		fatal("Create temp dir failed: %v", err)
	}
	defer os.RemoveAll(patchTempDir)

	newTempDir, err := os.MkdirTemp("", "new-")
	if err != nil {
		fatal("Create temp dir failed: %v", err)
	}
	defer os.RemoveAll(newTempDir)

	err = helpers.ExtractTarGz(tarPath, patchTempDir)
	if err != nil {
		fmt.Printf("Error extracting zip: %v\n", err)
	}
	fmt.Printf("Successfully extracted zip to: %s\n", patchTempDir)

	// Step 3: Parse package.json
	config := helpers.PatchApp{
		TargetDir:    targetDir,
		PatchTempDir: patchTempDir,
		NewTempDir:   newTempDir,
	}

	pkg, err := config.ParsePackageJSON(patchTempDir)
	if err != nil {
		fatal("Package.json error: %v", err)
	}

	// Step 4: Process files
	for _, file := range pkg.Files {
		switch file.Status {
		case "added":
			if err := config.ProcessAdded(file); err != nil {
				fatal("Process added failed: %v", err)
			}
		case "modified":
			if err := config.ProcessModified(file); err != nil {
				fatal("Process modified failed: %v", err)
			}
		case "deleted":
			if err := config.ProcessDeleted(file); err != nil {
				fatal("Process deleted failed: %v", err)
			}
		default:
			fatal("Unknown status: %s", file.Status)
		}
	}

	// 升级完成，并保存到临时文件夹，删除目标文件夹所有文件，将临时文件夹所有文件复制到目标文件夹
	err = os.RemoveAll(targetDir)
	if err != nil {
		fatal("Remove target dir failed: %v", err)
	}
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		fatal("Create target dir failed: %v", err)
	}

	err = helpers.CopyDir(newTempDir, targetDir)
	if err != nil {
		fatal("Copy new dir failed: %v", err)
	}

	fmt.Println("Upgrade completed successfully")
}

func init() {
	rootCmd.AddCommand(upgraderCmd)
	// 输入升级包，输出指定目录
	upgraderCmd.Flags().StringP("input", "i", "", "Input tar.gz file")
	upgraderCmd.Flags().StringP("output", "o", "", "Output directory")
}
