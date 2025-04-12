package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// ================== 辅助函数 ==================

type DiffGenerator struct {
	RepoURL    string
	BaseRef    string
	TargetRef  string
	OutputDir  string
	Workers    int
	IncludeBin bool
}

// 本地模式差异检测
func (dg *DiffGenerator) getLocalDiffList(basePath string, targetPath string) ([]string, error) {
	// 实现本地文件系统差异检测
	cmd := exec.Command("diff", "-qr", basePath, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil && cmd.ProcessState.ExitCode() != 1 { // diff返回1表示有差异
		return nil, fmt.Errorf("本地差异检测失败: %w\n%s", err, output)
	}

	return dg.parseLocalDiffOutput(string(output)), nil
}

// 解析本地diff输出
func (dg *DiffGenerator) parseLocalDiffOutput(output string) []string {
	var files []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理 "Files ... differ" 格式
		if strings.HasPrefix(line, "Files ") && strings.Contains(line, " differ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 4 {
				relativePath, err := filepath.Rel(dg.BaseRef, parts[1])
				if err != nil {
					fmt.Println("无法计算相对路径:", err)
				}

				// 提取第一个文件路径（去掉前面的 ./v1.0.0/ 或 ./v2.0.0/）
				// filePath := strings.TrimPrefix(parts[1], dg.BaseRef)
				// fmt.Println("filePath:", filePath)
				// filePath = strings.TrimPrefix(filePath, dg.TargetRef)
				// fmt.Println("filePath:", filePath)
				files = append(files, relativePath)
			}
		} else if strings.HasPrefix(line, "Only in ") {
			// 处理 "Only in ..." 格式
			parts := strings.Split(line, ": ")
			if len(parts) == 2 {
				var relativePath string
				var err error
				// 构建完整路径（目录+文件名）
				dir := strings.TrimPrefix(parts[0], "Only in ")
				cleanDir := filepath.Clean(dir)
				if strings.Contains(dir, dg.BaseRef) {
					relativePath, err = filepath.Rel(filepath.Clean(dg.BaseRef), cleanDir)
					if err != nil {
						fmt.Println("无法计算相对路径:", err)
					}
				} else {
					relativePath, err = filepath.Rel(filepath.Clean(dg.TargetRef), cleanDir)
					if err != nil {
						fmt.Println("无法计算相对路径:", err)
					}
				}

				files = append(files, filepath.Join(relativePath, parts[1]))
			}
		}
	}

	// 去重（同一文件可能在两种情况下都被报告）
	return UniqueStringArry(files)
}

func (dg *DiffGenerator) Generate() error {
	// 统一变量声明（避免重复声明err）
	var (
		basePath   string
		targetPath string
		diffList   []string
		err        error
		cleanup    func() // 资源清理函数
	)
	// ================== 1. 准备版本代码 ==================
	if dg.RepoURL == "LOCALHOST" {
		// 本地模式处理
		basePath = dg.BaseRef
		targetPath = dg.TargetRef

		// 本地模式不需要特殊清理
		cleanup = func() {}

		// 获取本地差异列表
		var err error
		diffList, err = dg.getLocalDiffList(basePath, targetPath)
		if err != nil {
			return fmt.Errorf("获取本地差异列表失败: %w", err)
		}
	} else {
		// 远程仓库模式
		var err error
		// 创建临时工作目录
		tmpDir, err := os.MkdirTemp("", "diff-tool-")
		if err != nil {
			return fmt.Errorf("创建临时目录失败: %w", err)
		}
		cleanup = func() { os.RemoveAll(tmpDir) }

		// 克隆仓库
		repoPath := filepath.Join(tmpDir, "repo")
		if err := gitClone(dg.RepoURL, repoPath); err != nil {
			cleanup() // 清理临时目录
			return fmt.Errorf("克隆仓库失败: %w", err)
		}

		// 获取版本代码
		basePath, targetPath, err = dg.prepareVersions(repoPath, tmpDir)
		if err != nil {
			cleanup()
			return fmt.Errorf("准备版本代码失败: %w", err)
		}

		// 生成差异文件列表
		diffList, err = dg.getDiffList(repoPath)
		if err != nil {
			cleanup()
			return fmt.Errorf("生成差异列表失败: %w", err)
		}
	}
	defer cleanup() // 确保资源释放

	// ================== 2. 校验差异列表 ==================
	if len(diffList) == 0 {
		return fmt.Errorf("未检测到有效差异")
	}

	// ================== 3. 创建输出目录 ==================
	outputPath := dg.getOutputPath()
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// ================== 4. 生成差异文件 ==================
	sem := make(chan struct{}, dg.Workers)     // 限制并发数为 4
	errChan := make(chan error, len(diffList)) // 错误通道，带缓冲
	var wg sync.WaitGroup                      // 等待所有 goroutine 完成

	fmt.Printf("diffList: %v \n", diffList)
	fmt.Printf("basePath: %s \n", basePath)
	fmt.Printf("targetPath: %s \n", targetPath)
	fmt.Printf("sem: %v \n", sem)

	for _, file := range diffList {
		wg.Add(1) // 增加 WaitGroup 计数
		go func(f string) {
			defer wg.Done() // 确保完成时减少计数

			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			// 调用生成差异文件的函数
			if err := dg.generateFileDiff(basePath, targetPath, f, outputPath); err != nil {
				errChan <- err // 发送错误到通道
			}
		}(file)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(errChan) // 所有 goroutine 完成后关闭 errChan
	}()

	// 错误收集
	for e := range errChan {
		if err == nil {
			err = e
		} else {
			err = fmt.Errorf("%v\n%w", err, e)
		}
	}

	if err != nil {
		fmt.Println("发生错误:", err)
	} else {
		fmt.Println("所有任务完成，无错误")
	}

	return err
}

// 辅助方法实现（部分关键函数）
func gitClone(repoURL, path string) error {
	cmd := exec.Command("git", "clone", "--bare", repoURL, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("克隆仓库失败: %s → %w", string(output), err)
	}
	return nil
}

func (dg *DiffGenerator) prepareVersions(repoPath, tmpDir string) (string, string, error) {
	// 创建基准版本工作树
	baseWorktree := filepath.Join(tmpDir, "base")
	if err := gitWorktree(repoPath, dg.BaseRef, baseWorktree); err != nil {
		return "", "", err
	}

	// 创建目标版本工作树
	targetWorktree := filepath.Join(tmpDir, "target")
	if dg.TargetRef == "WORKDIR" {
		return "", ".", nil // 使用当前工作目录
	}
	if err := gitWorktree(repoPath, dg.TargetRef, targetWorktree); err != nil {
		return "", "", err
	}

	return baseWorktree, targetWorktree, nil
}

func gitWorktree(repoPath, ref, worktreePath string) error {
	cmd := exec.Command("git",
		"--git-dir", repoPath,
		"worktree", "add",
		"--detach",
		"--force",
		worktreePath,
		ref)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("检出版本失败(%s): %s → %w", ref, string(output), err)
	}
	return nil
}

// getDiffList 获取两个版本间的差异文件列表
func (dg *DiffGenerator) getDiffList(repoPath string) ([]string, error) {
	// 构造git命令
	cmd := exec.Command("git",
		"--git-dir", repoPath,
		"diff",
		"--name-status", // 显示状态和路径
		"--no-renames",  // 禁用重命名检测
		"--no-ext-diff", // 禁用外部差异工具
		dg.BaseRef,
		dg.TargetRef,
	)

	// 捕获输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff失败: %s → %w", stderr.String(), err)
	}

	// 解析输出
	return parseDiffOutput(stdout.String(), dg.IncludeBin)
}

// parseDiffOutput 解析git diff输出
func parseDiffOutput(output string, includeBin bool) ([]string, error) {
	var files []string
	seen := make(map[string]struct{}) // 防止重复项

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析状态码和路径
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue // 忽略非法行
		}

		status := parts[0]
		filePath := parts[1]

		// 处理重命名（RXXX 状态）
		if status[0] == 'R' && len(parts) == 3 {
			filePath = parts[2] // 使用新文件名
		}

		// 过滤二进制文件
		if !includeBin && isBinaryFile(filePath) {
			continue
		}

		// 去重
		if _, exists := seen[filePath]; !exists {
			seen[filePath] = struct{}{}
			files = append(files, filePath)
		}
	}

	return files, nil
}

// isBinaryFile 判断是否二进制文件（根据扩展名）
func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip", ".exe":
		return true
	}
	return false
}

// getOutputPath 生成输出路径（带版本和时间戳）
func (dg *DiffGenerator) getOutputPath() string {
	// 清理版本字符串
	cleanVersion := func(s string) string {
		s = strings.ReplaceAll(s, "/", "_")
		s = strings.ReplaceAll(s, " ", "-")
		return s
	}

	// 生成目录名模板
	dirName := fmt.Sprintf("diff_%s_to_%s_%d",
		cleanVersion(dg.BaseRef),
		cleanVersion(dg.TargetRef),
		time.Now().Unix(),
	)

	// 构造最终路径
	finalPath := filepath.Join(
		sanitizePath(dg.OutputDir), // 路径消毒
		dirName,
	)

	return finalPath
}

func (dg *DiffGenerator) generateFileDiff(basePath string, targetPath string, relFilePath string, outputPath string) (err error) {
	// ================== 1. 路径安全处理 ==================
	// 标准化输入路径（防止路径遍历攻击）
	safeRelPath := filepath.Clean(relFilePath)
	if strings.HasPrefix(safeRelPath, "../") || strings.Contains(safeRelPath, "/../") {
		return fmt.Errorf("非法路径: %s  \n", relFilePath)
	}

	// 构造绝对路径
	absBasePath := filepath.Join(basePath, safeRelPath)
	absTargetPath := filepath.Join(targetPath, safeRelPath)
	// fmt.Printf("absBasePath: %s\n", absBasePath)
	// fmt.Printf("absTargetPath: %s\n", absTargetPath)

	// // ================== 2. 文件存在性检查 ==================
	fileStatus, err := GetFileChangeStatus(absBasePath, absTargetPath)
	if err != nil {
		fmt.Printf("检查文件[%v][%s]: %v  \n", fileStatus, safeRelPath, err)
		return nil
	}

	// // ================== 3. 准备输出目录 ==================
	outputDir := filepath.Join(outputPath, "files", filepath.Dir(safeRelPath))
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败[%s]: %w  \n", outputDir, err)
	}

	// ================== 4. 生成差异内容 ==================
	outputFile := filepath.Join(outputDir, filepath.Base(safeRelPath)+".diff")
	switch fileStatus {
	case FileAdded:
		return dg.generateAdditionDiff(absTargetPath, outputFile)
	case FileDeleted:
		return dg.generateDeletionDiff(absBasePath, outputFile)
	case FileModified:
		return dg.generateModificationDiff(absBasePath, absTargetPath, outputFile)
	default:
		return fmt.Errorf("未知文件状态: %s  \n", safeRelPath)
	}
	// return nil
}

// 生成不同类型差异的详细实现
func (dg *DiffGenerator) generateAdditionDiff(targetFile, outputFile string) error {
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	diffContent := fmt.Sprintf("+++ Added File: %s\n%s", filepath.Base(targetFile), content)
	return os.WriteFile(outputFile, []byte(diffContent), 0644)
}

func (dg *DiffGenerator) generateModificationDiff(baseFile, targetFile, outputFile string) error {
	// 自动优化块大小
	optimalBlockKB, err := OptimizeBlockSize(baseFile, targetFile, outputFile)
	if err != nil {
		fmt.Printf("优化失败: %v\n", err)
		os.Exit(1)
	}
	// 使用最优块生成正式补丁
	fmt.Printf("正在生成最终补丁文件...\n")
	if _, err := FixBlockCreatePatchFile(baseFile, targetFile, outputFile, optimalBlockKB); err != nil {
		fmt.Printf("补丁生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("补丁生成成功！最优块大小: %dKB\n", optimalBlockKB)

	return nil
}

func (dg *DiffGenerator) generateDeletionDiff(baseFile, outputFile string) error {
	content, err := os.ReadFile(baseFile)
	if err != nil {
		return err
	}
	_ = fmt.Sprintf("--- Deleted File: %s\n%s", filepath.Base(baseFile), content)
	// return os.WriteFile(outputFile, []byte(diffContent), 0644)

	fmt.Printf("文件已删除: %s\n", baseFile)

	return nil
}

// sanitizePath 路径消毒防止遍历攻击
func sanitizePath(path string) string {
	// 转换为绝对路径
	if !filepath.IsAbs(path) {
		absPath, err := filepath.Abs(path)
		if err == nil {
			path = absPath
		}
	}

	// 清理路径中的特殊字符
	return filepath.Clean(path)
}

func RunGenerate(cmd *cobra.Command, args []string) error {
	config := DiffGenerator{
		RepoURL:    MustGetString(cmd, "repo"),
		BaseRef:    MustGetString(cmd, "base"),
		TargetRef:  MustGetString(cmd, "target"),
		OutputDir:  MustGetString(cmd, "output"),
		IncludeBin: MustGetBool(cmd, "bin"),
		Workers:    MustGetInt(cmd, "workers"),
	}

	if err := config.Generate(); err != nil {
		return fmt.Errorf("\n❌ 差异生成失败: %w", err)
	}

	absPath, _ := filepath.Abs(config.OutputDir)
	fmt.Printf("\n✅ 差异生成成功！\n输出目录: %s\n", absPath)
	return nil
}
