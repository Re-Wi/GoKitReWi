package helpers

import (
	"archive/tar"
	"compress/gzip"
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

// FileExists 检查指定文件夹中是否存在某个文件
// 参数：
//
//	dirPath: 目标文件夹路径（绝对路径或相对路径）
//	filename: 需要检查的文件名（含扩展名）
//
// 返回值：
//
//	exists: 文件是否存在
//	isFile: 是否是普通文件（而非目录）
//	err: 错误信息（权限问题等）
func FileExists(dirPath string, filename string) (exists bool, isFile bool, err error) {
	// 拼接完整路径（自动处理路径分隔符）
	fullPath := filepath.Join(dirPath, filename)

	// 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在不是错误，正常返回
			return false, false, nil
		}
		// 其他错误（如权限问题）
		return false, false, err
	}

	// 检查是否为普通文件（排除目录）
	return true, !fileInfo.IsDir(), nil
}

func FilesExist(dir string, filenames []string) map[string]bool {
	results := make(map[string]bool)
	for _, f := range filenames {
		exists, _, _ := FileExists(dir, f)
		results[f] = exists
	}
	return results
}

// ------------------ 辅助函数 ------------------
// 安全关闭资源（处理关闭错误）
func SafeClose(closer io.Closer, err *error) {
	if closeErr := closer.Close(); closeErr != nil && *err == nil {
		*err = fmt.Errorf("资源关闭错误: %w", closeErr)
	}
}

// 文件存在性验证
func VerifyFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%q 是目录而非文件", path)
	}
	return nil
}

// 补丁文件基础校验
func VerifyPatchFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return fmt.Errorf("生成的补丁文件为空")
	}
	return nil
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

// 文件大小校验（示例逻辑）
func VerifyFileSize(target, source string) error {
	tInfo, err := os.Stat(target)
	if err != nil {
		return err
	}

	sInfo, _ := os.Stat(source)
	if sInfo != nil && tInfo.Size() < sInfo.Size()/2 {
		return fmt.Errorf("目标文件过小（可能不完整）")
	}
	return nil
}

// 获取文件大小（MB）
func GetFileSizeMB(path string) float64 {
	info, _ := os.Stat(path)
	return float64(info.Size()) / (1024 * 1024)
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
