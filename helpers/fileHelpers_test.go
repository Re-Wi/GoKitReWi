package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 准备测试目录结构
func prepareTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "filetest")
	require.NoError(t, err)

	createFiles(t, tmpDir, []string{
		"root.txt",
		".hidden",
		"docs/doc1.doc",
		"docs/.idea/workspace.xml",
		"src/main.go",
		"src/utils/helper.go",
		"excluded/rewi.log",
		"excluded/.vs/config.vs",
	})

	return tmpDir
}

// 批量创建文件和目录
func createFiles(t *testing.T, base string, paths []string) {
	for _, p := range paths {
		fullPath := filepath.Join(base, p)
		if strings.HasSuffix(p, "/") {
			os.MkdirAll(fullPath, 0755)
		} else {
			os.MkdirAll(filepath.Dir(fullPath), 0755)
			os.WriteFile(fullPath, []byte("test"), 0644)
		}
	}
}
func TestGetAllFiles(t *testing.T) {
	// 准备测试目录
	tmpDir := prepareTestDir(t)
	defer os.RemoveAll(tmpDir)

	t.Run("正常目录", func(t *testing.T) {
		result, err := GetAllFiles(tmpDir, nil)
		assert.NoError(t, err)
		expected := []string{
			filepath.Join(tmpDir, "root.txt"),
			filepath.Join(tmpDir, ".hidden"),
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("空目录", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		os.Mkdir(emptyDir, 0755)

		result, err := GetAllFiles(emptyDir, nil)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("无效路径", func(t *testing.T) {
		_, err := GetAllFiles(filepath.Join(tmpDir, "nonexist"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")
	})

	t.Run("混合内容", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "mixed")
		os.Mkdir(dir, 0755)
		createFiles(t, dir, []string{"a.txt", "b.jpg", "sub/"})

		result, err := GetAllFiles(dir, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 2) // 排除sub目录
	})
}

func TestGetPathFiles(t *testing.T) {
	tmpDir := prepareTestDir(t)
	defer os.RemoveAll(tmpDir)

	// 每次测试前重置全局变量
	t.Cleanup(func() { filePathList = nil })

	t.Run("默认过滤", func(t *testing.T) {
		result := GetPathFiles(tmpDir, "")
		expected := []string{
			filepath.Join(tmpDir, "root.txt"),
			filepath.Join(tmpDir, "docs/doc1.doc"),
			filepath.Join(tmpDir, "src/main.go"),
			filepath.Join(tmpDir, "src/utils/helper.go"),
		}
		assert.ElementsMatch(t, expected, result)
		assert.NotContains(t, result, "excluded/rewi.log") // 排除rewi
		assert.NotContains(t, result, ".vs/config.vs")     // 排除.vs
	})

	t.Run("后缀过滤", func(t *testing.T) {
		result := GetPathFiles(tmpDir, "go")
		expected := []string{
			filepath.Join(tmpDir, "src/main.go"),
			filepath.Join(tmpDir, "src/utils/helper.go"),
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("排除规则", func(t *testing.T) {
		GetPathFiles(tmpDir, "")
		for _, path := range filePathList {
			assert.False(t, strings.Contains(path, ".idea"),
				"应跳过.idea目录")
			assert.False(t, strings.Contains(path, "rewi"),
				"应跳过rewi文件")
		}
	})

	t.Run("符号链接", func(t *testing.T) {
		linkPath := filepath.Join(tmpDir, "link")
		os.Symlink(filepath.Join(tmpDir, "src"), linkPath)

		result := GetPathFiles(linkPath, "")
		assert.Len(t, result, 2) // main.go和helper.go
	})
}
func TestEdgeCases(t *testing.T) {
	t.Run("超大目录", func(t *testing.T) {
		tmpDir := prepareLargeDir(t, 5000) // 创建5000个文件
		defer os.RemoveAll(tmpDir)

		// GetAllFiles性能测试
		start := time.Now()
		_, err := GetAllFiles(tmpDir, nil)
		assert.NoError(t, err)
		assert.WithinDuration(t, time.Now(), start, 500*time.Millisecond)

		// GetPathFiles递归测试
		start = time.Now()
		GetPathFiles(tmpDir, "")
		assert.True(t, len(filePathList) >= 5000)
	})
}

func prepareLargeDir(t *testing.T, fileCount int) string {
	tmpDir, err := os.MkdirTemp("", "large")
	require.NoError(t, err)

	for i := 0; i < fileCount; i++ {
		name := fmt.Sprintf("file%d.txt", i)
		os.WriteFile(filepath.Join(tmpDir, name), []byte{}, 0644)
	}
	return tmpDir
}

func TestFileExists(t *testing.T) {
	// 准备测试文件
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(validFile, []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	tests := []struct {
		name       string
		dir        string
		filename   string
		wantExists bool
		wantIsFile bool
		wantErr    bool
	}{
		{"存在的文件", tmpDir, "test.txt", true, true, false},
		{"不存在的文件", tmpDir, "missing.txt", false, false, false},
		{"目录而非文件", tmpDir, "subdir", true, false, false},
		{"无效路径", "/nonexistent/path", "any.txt", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, isFile, _, err := PathInfo(filepath.Join(tt.dir, tt.filename))
			if (err != nil) != tt.wantErr {
				t.Errorf("错误预期: %v, 实际错误: %v", tt.wantErr, err)
			}
			if exists != tt.wantExists {
				t.Errorf("存在性预期: %v, 实际: %v", tt.wantExists, exists)
			}
			if isFile != tt.wantIsFile {
				t.Errorf("文件类型预期: %v, 实际: %v", tt.wantIsFile, isFile)
			}
		})
	}
}
