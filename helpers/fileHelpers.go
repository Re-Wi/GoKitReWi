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
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("failed to stat source '%s': %w", source, err)
		}

		// 定义递归添加文件到 tar 的函数
		var addToTar func(path string, info os.FileInfo) error
		addToTar = func(path string, info os.FileInfo) error {
			// 创建 tar 文件头
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header for '%s': %w", path, err)
			}

			// 设置文件头的名称（相对路径）
			header.Name = filepath.Base(path)

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
		if err := addToTar(source, info); err != nil {
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

	// 遍历 tar 文件中的每个条目
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 文件结束
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// 构建目标路径
		targetPath := filepath.Join(target, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// 创建文件并写入内容
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}
		default:
			fmt.Printf("Unsupported file type: %v in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
