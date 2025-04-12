package helpers

import (
	"os"
	"testing"
)

func TestOptimizeBlockSize(t *testing.T) {
	oldFile := "../tools/V3.00.bin"
	newFile := "../tools/V3.60.bin"
	patchFile := "../tools/patch.xd"

	// 自动优化块大小
	optimalBlockKB, err := OptimizeBlockSize(oldFile, newFile, patchFile)
	if err != nil {
		t.Logf("优化失败: %v \n", err)
		os.Exit(1)
	}

	// 使用最优块生成正式补丁
	t.Logf("正在生成最终补丁文件... \n")
	if _, err := FixBlockCreatePatchFile(oldFile, newFile, patchFile, optimalBlockKB); err != nil {
		t.Logf("补丁生成失败: %v \n", err)
		os.Exit(1)
	}

	t.Logf("补丁生成成功！最优块大小: %dKB\n", optimalBlockKB)
}
