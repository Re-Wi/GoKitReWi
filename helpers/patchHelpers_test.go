package helpers

import (
	"encoding/json"
	"testing"
)

const (
	OldFilePath   = "old_file.bin"
	NewFilePath   = "new_file.bin"
	UpdatedPath   = "updated_file.bin"
	TestBlockSize = 4 * 1024 // 测试块大小
)

// TestIncrementalUpdate 测试增量更新流程
func TestIncrementalUpdate(t *testing.T) {
	blockSize := TestBlockSize

	// 1. 计算新旧文件的块哈希
	oldHashes, err := ComputeFileHashes(OldFilePath, blockSize)
	if err != nil {
		t.Fatalf("Failed to compute old file hashes: %v", err)
	}

	newHashes, err := ComputeFileHashes(NewFilePath, blockSize)
	if err != nil {
		t.Fatalf("Failed to compute new file hashes: %v", err)
	}

	// 2. 比较块哈希，找出差异块
	diffBlocks := CompareHashes(oldHashes, newHashes)

	// 3. 生成补丁数据
	patchData, err := GeneratePatch(NewFilePath, diffBlocks, blockSize)
	if err != nil {
		t.Fatalf("Failed to generate patch data: %v", err)
	}

	// 4. 解析补丁数据
	var patch []PatchBlock

	err = json.Unmarshal(patchData, &patch)
	if err != nil {
		t.Fatalf("Failed to unmarshal patch data: %v", err)
	}

	// 5. 应用补丁
	err = ApplyPatch(OldFilePath, UpdatedPath, patch, blockSize)
	if err != nil {
		t.Fatalf("Failed to apply patch: %v", err)
	}

	// 6. 验证更新后的文件是否与新文件一致
	updatedHashes, err := ComputeFileHashes(UpdatedPath, blockSize)
	if err != nil {
		t.Fatalf("Failed to compute updated file hashes: %v", err)
	}

	for idx, hash := range newHashes {
		if updatedHash, ok := updatedHashes[idx]; !ok || updatedHash != hash {
			t.Errorf("Block %d mismatch: expected %v, got %v", idx, hash, updatedHash)
		}
	}

	t.Logf("Incremental update test passed successfully!")
}
