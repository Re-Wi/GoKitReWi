package helpers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Re-Wi/GoKitReWi/handlers"
	"github.com/icedream/go-bsdiff"
)

type PatchApp struct {
	TargetDir    string
	PatchTempDir string
	NewTempDir   string
}

func (pa *PatchApp) ParsePackageJSON(tempDir string) (*handlers.UpdatePackage, error) {
	path := filepath.Join(tempDir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read package.json: %w", err)
	}

	var pkg handlers.UpdatePackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	return &pkg, nil
}

func (pa *PatchApp) ProcessAdded(file handlers.FileEntry) error {
	src := filepath.Join(pa.PatchTempDir, file.Path)
	dest := filepath.Join(pa.NewTempDir, file.Path)

	_, fileExists, _, _ := PathInfo(dest)

	if fileExists {
		return fmt.Errorf("file %s already exists", dest)
	}

	if err := CopyFile(src, dest); err != nil {
		return err
	}
	sizeValue, err := EnsureFileSize(dest, "byte")
	if err != nil {
		return fmt.Errorf("ensure file size: %w", err) // EnsureFileSize 函数返回的是 error 类型，这里使用 fmt.Errorf 包装为 custo
	}

	if sizeValue != float64(file.Size) {
		return fmt.Errorf("file size mismatch: expected %d, got %.0f", file.Size, sizeValue)
	}

	return VerifyFileHash(dest, file.Hash, md5.New)
}

func (pa *PatchApp) ProcessModified(file handlers.FileEntry) error {
	oldPath := filepath.Join(pa.TargetDir, file.Path)
	patchPath := filepath.Join(pa.PatchTempDir, file.Patch.Path)
	newFilePath := filepath.Join(pa.NewTempDir, file.Path)

	_, fileExists, _, _ := PathInfo(oldPath)

	if !fileExists {
		return fmt.Errorf("file %s does not exist", oldPath)
	}

	sizeValue, err := EnsureFileSize(patchPath, "byte")
	if err != nil {
		return fmt.Errorf("ensure file size: %w", err)
	}

	if sizeValue != float64(file.Patch.Size) {
		return fmt.Errorf("patch file size mismatch: expected %d, got %.0f", file.Patch.Size, sizeValue)
	}

	if err := VerifyFileHash(patchPath, file.Patch.Hash, md5.New); err != nil {
		return fmt.Errorf("verify patch file hash: %w", err)
	}

	oldData, err := os.Open(oldPath)
	if err != nil {
		return err
	}

	patchData, err := os.Open(patchPath)
	if err != nil {
		return err
	}

	newFile, err := os.Create(newFilePath)
	defer newFile.Close()
	if err != nil {
		return err
	}
	err = bsdiff.Patch(oldData, newFile, patchData)
	if err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	_, fileExists, _, _ = PathInfo(newFilePath)

	if !fileExists {
		return fmt.Errorf("file %s does not exist", newFilePath)
	}

	sizeValue, err = EnsureFileSize(newFilePath, "byte")
	if err != nil {
		return fmt.Errorf("ensure file size: %w", err)
	}

	if sizeValue != float64(file.Size) {
		return fmt.Errorf("file size mismatch: expected %d, got %.0f", file.Size, sizeValue)
	}

	if err = VerifyFileHash(newFilePath, file.Hash, md5.New); err != nil {
		return fmt.Errorf("verify new file hash: %w", err)
	}

	return nil
}

func (pa *PatchApp) ProcessDeleted(file handlers.FileEntry) error {
	return nil
}
