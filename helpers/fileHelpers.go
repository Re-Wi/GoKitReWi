package helpers

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var filePathList []string

// 获取根目录下直属所有文件（不包括文件夹及其中的文件）
func GetAllFiles(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}

	for _, fi := range rd {
		if !fi.IsDir() {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

// 获取当前项目根目录下所有文件（包括文件夹中的文件）
func GetPathFiles(folder string, suffix string) []string {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if strings.Contains(file.Name(), ".idea") || strings.Contains(file.Name(), "rewi") || strings.Contains(file.Name(), ".vs") || strings.Contains(file.Name(), ".git") {
			fmt.Println("Skip :", file.Name())
			continue
		}
		if file.IsDir() {
			GetPathFiles(folder+"/"+file.Name(), suffix)
		} else {
			var filename = file.Name()
			if strings.Contains(strings.ToLower(path.Ext(filename)), suffix) {
				filePathList = append(filePathList, folder+"/"+filename)
			}
			//fmt.Println(folder + "/" + file.Name())
		}
	}
	return filePathList
}

// PathInfo 检查指定路径是否存在及其类型
// 参数：
//
//	path: 文件或目录路径（支持绝对/相对路径）
//
// 返回值：
//
//	exists: 路径是否存在
//	isFile: 是文件(true)还是目录(false)
//	size: 文件大小(字节)，目录返回0
//	err: 错误信息
func PathInfo(path string) (exists bool, isFile bool, size int64, err error) {
	// 规范化路径，避免多余符号（如双斜杠、点符号）
	path = filepath.Clean(path)

	// 获取绝对路径并标准化
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, false, 0, fmt.Errorf("路径解析失败: %w", err)
	}

	// 获取文件信息
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, 0, nil
		}
		return false, false, 0, fmt.Errorf("访问路径失败: %w", err)
	}

	// 返回检查结果
	return true, !info.IsDir(), info.Size(), nil
}

// CheckPaths 批量检查路径状态
// 返回映射表：path -> exists
func CheckPaths(paths []string) map[string]bool {
	results := make(map[string]bool, len(paths))

	// 遍历所有路径并检查
	for _, p := range paths {
		exists, _, _, err := PathInfo(p)
		results[p] = exists
		if err != nil {
			// 记录错误但继续处理其他路径
			fmt.Printf("警告: 检查路径 %s 失败: %v\n", p, err)
		}
	}
	return results
}

type FileChangeStatus int

const (
	FileUnknown FileChangeStatus = iota
	FileAdded
	FileDeleted
	FileModified
)

// getFileChangeStatus 检测文件变化状态
func GetFileChangeStatus(basePath, targetPath string) (FileChangeStatus, error) {
	_, baseExists, _, _ := PathInfo(basePath)
	_, targetExists, _, _ := PathInfo(targetPath)
	switch {
	case !baseExists && targetExists:
		return FileAdded, fmt.Errorf("文件新增 \n")
	case baseExists && !targetExists:
		return FileDeleted, fmt.Errorf("文件删除 \n")
	case baseExists && targetExists:
		if same, err := FilesEqual(basePath, targetPath); err != nil {
			return FileUnknown, err
		} else if !same {
			return FileModified, nil
		}
		return FileUnknown, fmt.Errorf("文件内容相同 \n")
	default:
		return FileUnknown, fmt.Errorf("文件在两边都不存在 \n")
	}
}

// ------------------ 辅助函数 ------------------
// 安全关闭资源（处理关闭错误）
func SafeClose(closer io.Closer, err *error) {
	if closeErr := closer.Close(); closeErr != nil && *err == nil {
		*err = fmt.Errorf("资源关闭错误: %w", closeErr)
	}
}

// 文件访问权限验证
func VerifyFileAccess(path string, mode int) error {
	file, err := os.OpenFile(path, mode, 0)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

// ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
// │ 文件基本信息  │ →  │ 元数据对比   │ →  │ 分块哈希对比 │
// └──────────────┘    └──────────────┘    └──────────────┘
//
//	快速排除           中等可靠性          最高准确性
//
// FilesEqual 增强版文件比较函数
// FilesEqual 高效准确的文件比较函数
// 高精度场景（配置文件、二进制文件）
// changed, err := FilesEqual(v1Path, v2Path)
func FilesEqual(path1, path2 string) (bool, error) {
	// 1. 快速元数据检查
	same, quickResult, err := FastFilesEqual(path1, path2)
	if quickResult != checkUnknown {
		return same, err
	}

	// 2. 完整内容比较
	return deepCompare(path1, path2)
}

// 检查结果枚举
type checkResult int

const (
	checkDifferent checkResult = iota // 确定不同
	checkSame                         // 确定相同
	checkUnknown                      // 需要进一步检查
)

// quickCheck 快速检查（返回是否可确定结果）
// FastFilesEqual 快速比较版本（适用于性能敏感场景）
// 高性能场景（日志文件、临时文件）
// changed, err := FastFilesEqual(v1Path, v2Path)
func FastFilesEqual(path1, path2 string) (bool, checkResult, error) {
	info1, err := os.Stat(path1)
	if err != nil {
		return false, checkUnknown, fmt.Errorf("stat %s: %w", path1, err)
	}

	info2, err := os.Stat(path2)
	if err != nil {
		return false, checkUnknown, fmt.Errorf("stat %s: %w", path2, err)
	}

	// 大小不同 → 肯定不同
	if info1.Size() != info2.Size() {
		return false, checkDifferent, nil
	}

	// 零字节文件 → 肯定相同
	if info1.Size() == 0 {
		return true, checkSame, nil
	}

	// 相同inode → 肯定相同（硬链接或相同文件）
	if os.SameFile(info1, info2) {
		return true, checkSame, nil
	}

	// 修改时间和模式相同 → 可能相同
	if info1.ModTime().Equal(info2.ModTime()) && info1.Mode() == info2.Mode() {
		return true, checkUnknown, nil
	}

	return false, checkUnknown, nil
}

// deepCompare 深度内容比较
func deepCompare(path1, path2 string) (bool, error) {
	// 智能选择块大小（根据文件大小动态调整）
	chunkSize, err := DetermineChunkSize(path1)
	if err != nil {
		return false, err
	}

	// 并行计算哈希
	resultChan := make(chan hashResult, 2)
	go calculateHash(path1, chunkSize, resultChan)
	go calculateHash(path2, chunkSize, resultChan)

	// 收集结果
	result1 := <-resultChan
	result2 := <-resultChan

	if result1.err != nil {
		return false, fmt.Errorf("hash %s: %w", path1, result1.err)
	}
	if result2.err != nil {
		return false, fmt.Errorf("hash %s: %w", path2, result2.err)
	}

	return result1.hash == result2.hash, nil
}

// 哈希计算结果结构
type hashResult struct {
	hash string
	err  error
}

// DetermineChunkSize 智能确定块大小
func DetermineChunkSize(path1 string) (int64, error) {
	info1, err := os.Stat(path1)
	if err != nil {
		return 0, err
	}

	size := info1.Size()
	switch {
	case size <= 4*1024: // <4KB → 一次性读取
		return size, nil
	case size <= 64*1024: // <64KB → 4KB块
		return 4 * 1024, nil
	case size <= 1*1024*1024: // <1MB → 16KB块
		return 16 * 1024, nil
	case size <= 10*1024*1024: // <10MB → 64KB块
		return 64 * 1024, nil
	default: // >10MB → 1MB块
		return 1024 * 1024, nil
	}
}

// calculateHash 计算文件哈希
func calculateHash(path string, chunkSize int64, result chan<- hashResult) {
	file, err := os.Open(path)
	if err != nil {
		result <- hashResult{err: err}
		return
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, chunkSize)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			result <- hashResult{err: err}
			return
		}
		if n == 0 {
			break
		}

		if _, err := hash.Write(buf[:n]); err != nil {
			result <- hashResult{err: err}
			return
		}
	}

	result <- hashResult{hash: string(hash.Sum(nil))}
}

// calculateFileHash 分块计算文件哈希（内存高效）
func calculateFileHash(path string, chunkSize int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	buf := make([]byte, chunkSize)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}

		if _, err := hash.Write(buf[:n]); err != nil {
			return "", err
		}
	}

	return string(hash.Sum(nil)), nil
}

// EnsureFileSize 检查文件是否存在并返回其大小（单位：MB）
// 如果文件不存在或无法访问，返回 -1 和错误
func EnsureFileSize(path string, unit string) (float64, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("文件不存在: %s", path)
		}
		return 0, fmt.Errorf("无法获取文件信息: %w", err)
	}

	if info.IsDir() {
		return 0, fmt.Errorf("路径是目录，不是文件: %s", path)
	}

	sizeVaule := float64(info.Size())
	switch unit {
	case "KB":
		sizeVaule = float64(sizeVaule) / 1024
	case "MB":
		sizeVaule = float64(sizeVaule) / (1024 * 1024)
	case "GB":
		sizeVaule = float64(sizeVaule) / (1024 * 1024 * 1024)
	default:
		sizeVaule = float64(sizeVaule)
	}

	return sizeVaule, nil
}

// 文件大小校验
func VerifyFileSize(targetPath, sourcePath string) error {
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("获取目标文件信息失败: %w", err)
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	if sourceInfo.Size() == 0 {
		return fmt.Errorf("源文件大小为0")
	}

	if float64(targetInfo.Size()) != float64(sourceInfo.Size()) {
		return fmt.Errorf("目标文件过小（可能不完整），目标大小: %d，源大小: %d",
			targetInfo.Size(), sourceInfo.Size())
	}

	return nil
}

// 创建 tar.gz 压缩包（支持多个文件和文件夹）
func CreateTarGz(sources []string, target string) error {
	// 创建目标文件
	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer file.Close()

	// 创建 gzip 写入器，设置最高压缩级别
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// 创建 tar 写入器
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// 遍历每个输入路径
	for _, source := range sources {
		// 获取绝对路径并验证
		absSource, err := filepath.Abs(source)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for '%s': %w", source, err)
		}

		// 获取文件信息
		info, err := os.Stat(absSource)
		if err != nil {
			return fmt.Errorf("failed to stat source '%s': %w", absSource, err)
		}

		// 计算基准路径（源文件/目录的父目录）
		basePath := filepath.Dir(absSource)

		// 定义递归添加文件到 tar 的函数
		var addToTar func(path string, info os.FileInfo) error
		addToTar = func(path string, info os.FileInfo) error {
			// 创建 tar 文件头
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header for '%s': %w", path, err)
			}

			// 计算相对于基准路径的相对路径
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return fmt.Errorf("failed to get relative path for '%s': %w", path, err)
			}
			header.Name = relPath

			// 将文件头写入 tar
			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header for '%s': %w", path, err)
			}

			// 如果是普通文件，写入文件内容
			if !info.IsDir() {
				data, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open file '%s': %w", path, err)
				}
				defer data.Close()
				if _, err := io.Copy(tarWriter, data); err != nil {
					return fmt.Errorf("failed to copy file content for '%s': %w", path, err)
				}
			}

			// 如果是目录，递归处理子文件
			if info.IsDir() {
				files, err := os.ReadDir(path)
				if err != nil {
					return fmt.Errorf("failed to read directory '%s': %w", path, err)
				}
				for _, f := range files {
					subPath := filepath.Join(path, f.Name())
					subInfo, err := f.Info()
					if err != nil {
						return fmt.Errorf("failed to get subfile info for '%s': %w", subPath, err)
					}
					if err := addToTar(subPath, subInfo); err != nil {
						return err
					}
				}
			}

			return nil
		}

		// 调用递归函数，开始打包
		if err := addToTar(absSource, info); err != nil {
			return err
		}
	}

	return nil
}

func ExtractTarGz(source, target string) error {
	// 打开 tar.gz 文件
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	// 创建 gzip 读取器
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// 创建 tar 读取器
	tarReader := tar.NewReader(gzipReader)

	// 确保目标目录存在
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 遍历 tar 文件中的每个条目
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 文件结束
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// 安全地构建目标路径
		targetPath := filepath.Join(target, header.Name)

		// 安全检查：防止路径遍历攻击
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		// 根据文件类型处理
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录（确保权限正确）
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			// 确保父目录存在
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
			}

			// 创建文件并写入内容
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			// 复制文件内容
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write to file %s: %w", targetPath, err)
			}
			outFile.Close()
		case tar.TypeSymlink:
			// 处理符号链接
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", targetPath, header.Linkname, err)
			}
		default:
			fmt.Printf("Unsupported file type: %v in %s\n", header.Typeflag, header.Name)
		}

		// 设置文件修改时间（如果支持）
		if err := os.Chtimes(targetPath, time.Now(), header.ModTime); err != nil {
			fmt.Printf("Warning: failed to set modification time for %s: %v\n", targetPath, err)
		}
	}

	return nil
}
