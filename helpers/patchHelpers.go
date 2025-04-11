package helpers

import (
	"encoding/json"
	"hash/fnv"
	"io"
	"os"
)

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
	var diffBlocks []uint64
	maxIndex := uint64(0)

	// 找出最大索引
	for idx := range newHashes {
		if idx > maxIndex {
			maxIndex = idx
		}
	}

	for blockIndex := uint64(0); blockIndex <= maxIndex; blockIndex++ {
		oldHash, okOld := oldHashes[blockIndex]
		newHash, okNew := newHashes[blockIndex]

		if !okOld || !okNew || oldHash != newHash {
			diffBlocks = append(diffBlocks, blockIndex)
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

	patch := make([]struct {
		Index uint64
		Data  []byte
	}, 0, len(diffBlocks))

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

		patch = append(patch, struct {
			Index uint64
			Data  []byte
		}{
			Index: idx,
			Data:  buffer[:n],
		})
	}

	// 序列化为二进制或JSON
	// 这里以JSON为例
	patchJSON, _ := json.Marshal(patch)
	return patchJSON, nil
}

// ApplyPatch 应用补丁到旧文件，生成新文件
func ApplyPatch(oldFilePath, newFilePath string, patch []struct {
	Index uint64
	Data  []byte
}, blockSize int) error {
	oldFile, err := os.Open(oldFilePath)
	if err != nil {
		return err
	}
	defer oldFile.Close()

	newFile, err := os.Create(newFilePath)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// 读取旧文件内容并应用补丁
	buffer := make([]byte, blockSize)
	for {
		_, err := oldFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 默认写入旧数据
		newFile.Write(buffer)
	}

	// 覆盖差异块
	for _, block := range patch {
		offset := int64(block.Index) * int64(blockSize)
		_, err := newFile.Seek(offset, 0)
		if err != nil {
			return err
		}
		newFile.Write(block.Data)
	}

	return nil
}
