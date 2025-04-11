package helpers

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
)

// 统一定义类型
type PatchBlock struct {
	Index uint64 `json:"index"`
	Data  []byte `json:"data"`
}

type Patch []PatchBlock

// ComputeFileHashes 计算文件的块哈希列表
func ComputeFileHashes(filePath string, blockSize int) (map[uint64]uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	blockHashes := make(map[uint64]uint64)
	buffer := make([]byte, blockSize)
	blockIndex := uint64(0)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// 计算当前块的哈希
		h := fnv.New64a()
		h.Write(buffer[:n])
		blockHash := h.Sum64()

		blockHashes[blockIndex] = blockHash
		blockIndex++
	}

	return blockHashes, nil
}

// CompareHashes 比较两个块哈希列表，返回差异块索引
func CompareHashes(oldHashes, newHashes map[uint64]uint64) []uint64 {
	allIndexes := make(map[uint64]bool)
	for idx := range oldHashes {
		allIndexes[idx] = true
	}
	for idx := range newHashes {
		allIndexes[idx] = true
	}

	var diffBlocks []uint64
	for idx := range allIndexes {
		oldHash := oldHashes[idx]
		newHash := newHashes[idx]
		if oldHash != newHash {
			diffBlocks = append(diffBlocks, idx)
		}
	}
	return diffBlocks
}

// GeneratePatch 生成补丁数据
func GeneratePatch(newFilePath string, diffBlocks []uint64, blockSize int) ([]byte, error) {
	file, err := os.Open(newFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	patch := make(Patch, 0, len(diffBlocks))

	buffer := make([]byte, blockSize)
	for _, idx := range diffBlocks {
		offset := int64(idx) * int64(blockSize)
		_, err := file.Seek(offset, 0)
		if err != nil {
			return nil, err
		}

		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// 每次创建新切片
		data := make([]byte, n)
		copy(data, buffer[:n])
		patch = append(patch, PatchBlock{
			Index: idx,
			Data:  data,
		})
	}

	// 序列化为二进制或JSON
	// 这里以JSON为例
	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return patchJSON, nil
}

// ApplyPatch 应用补丁到旧文件生成新文件
func ApplyPatch(oldFilePath, newFilePath string, patch []PatchBlock, blockSize int) error {
	// 打开旧文件
	oldFile, err := os.Open(oldFilePath)
	if err != nil {
		return fmt.Errorf("打开旧文件失败: %w", err)
	}
	defer oldFile.Close()

	// 创建新文件
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return fmt.Errorf("创建新文件失败: %w", err)
	}
	defer newFile.Close()

	// 计算新文件总长度
	var maxIndex uint64
	for _, block := range patch {
		if block.Index > maxIndex {
			maxIndex = block.Index
		}
	}
	newFileSize := int64(maxIndex+1) * int64(blockSize)

	// 设置新文件初始长度
	if err := newFile.Truncate(newFileSize); err != nil {
		return fmt.Errorf("设置文件长度失败: %w", err)
	}

	// 复制旧文件内容到新文件（自动处理长度差异）
	oldInfo, err := oldFile.Stat()
	if err != nil {
		return fmt.Errorf("获取旧文件信息失败: %w", err)
	}
	copySize := oldInfo.Size()
	if newFileSize < copySize {
		copySize = newFileSize
	}
	if _, err := oldFile.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.CopyN(newFile, oldFile, copySize); err != nil {
		return fmt.Errorf("复制旧文件内容失败: %w", err)
	}

	// 应用补丁块
	for _, block := range patch {
		offset := int64(block.Index) * int64(blockSize)
		if _, err := newFile.WriteAt(block.Data, offset); err != nil {
			return fmt.Errorf("写入补丁块 %d 失败: %w", block.Index, err)
		}
	}

	return nil
}
