package helpers

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"path/filepath"

	"github.com/konsorten/go-xdelta"
	"github.com/spf13/cobra"
)

func FixBlockCreatePatchFile(oldFile, newFile, patchFile string, blockKB int) (float64, error) {

	// 输入参数验证
	if oldFile == "" || newFile == "" || patchFile == "" {
		return 0, fmt.Errorf("无效的输入参数：oldFile=%q, newFile=%q, patchFile=%q",
			oldFile, newFile, patchFile)
	}
	// ================== 文件校验 ==================
	if err := VerifyFileAccess(oldFile, os.O_RDONLY); err != nil {
		return 0, fmt.Errorf("旧文件不可访问: %w \n", err)
	}
	if err := VerifyFileAccess(newFile, os.O_RDONLY); err != nil {
		return 0, fmt.Errorf("新文件不可访问: %w \n", err)
	}

	// 1. 安全打开文件流
	oldFileIo, err := os.Open(oldFile)
	if err != nil {
		return 0, fmt.Errorf("打开旧文件失败: %w \n", err)
	}
	defer SafeClose(oldFileIo, &err)

	newFileIo, err := os.Open(newFile)
	if err != nil {
		return 0, fmt.Errorf("打开新文件失败: %w \n", err)
	}
	defer SafeClose(newFileIo, &err)

	// 确保补丁目录存在
	if err := os.MkdirAll(filepath.Dir(patchFile), 0755); err != nil {
		return 0, fmt.Errorf("创建补丁目录失败: %w \n", err)
	}

	patchFileIo, err := os.Create(patchFile)
	if err != nil {
		return 0, fmt.Errorf("创建补丁文件失败: %w \n", err)
	}
	defer SafeClose(patchFileIo, &err)

	// 2. 配置编码选项并验证
	options := xdelta.EncoderOptions{
		BlockSizeKB: blockKB,                         // 块大小 4 MB ( 4*1024 KB)
		FileID:      "v1.0.0",                        // 版本标识
		FromFile:    oldFileIo,                       // 旧文件
		ToFile:      newFileIo,                       // 新文件
		PatchFile:   patchFileIo,                     // 补丁输出
		Header:      []byte("metadata:checksum=123"), // 自定义头部
		EnableStats: true,                            // 启用统计
	}

	// 3. 创建编码器（带错误包装）
	enc, err := xdelta.NewEncoder(options)
	if err != nil {
		return 0, fmt.Errorf("编码器初始化失败: %w \n", err)
	}
	defer func() {
		if closeErr := enc.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("编码器关闭错误: %w \n", closeErr)
		}
	}()

	// 4. 处理过程（带上下文超时控制）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := enc.Process(ctx); err != nil {
		// return 0, fmt.Errorf("补丁生成过程失败: %w \n", err)
		return 0, nil
	}

	// 5. 输出统计信息（调试用）
	if options.EnableStats {
		enc.DumpStatsToStdout()
	}
	var sizeValue float64
	// 6. 最终校验
	if sizeValue, err = EnsureFileSize(patchFile, "byte"); err != nil {
		return 0, fmt.Errorf("补丁文件校验失败: %w", err)
	}

	return sizeValue, nil
}

// OptimizeBlockSize 自动寻找最优块大小
func OptimizeBlockSize(oldFile, newFile, patchFile string) (int, error) {
	fmt.Println("开始自动寻找最优块大小...", oldFile, newFile, patchFile)
	// 记录最佳结果
	type result struct {
		blockSizeKB int
		patchSize   int64
		err         error
	}
	// 初始化最佳结果
	best := result{blockSizeKB: 0, patchSize: math.MaxInt64}

	// 测试的块大小序列（KB），从1开始按2倍增长
	blockSizes := []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384}

	tempDir, err := os.MkdirTemp("", "xdelta-optimize-")
	if err != nil {
		return 0, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 遍历测试所有块大小
	for _, blockKB := range blockSizes {
		// 生成临时补丁文件路径
		tempPatch := filepath.Join(tempDir, fmt.Sprintf("patch-%dkb.xd", blockKB))

		// 执行补丁生成
		start := time.Now()
		fileSize, err := FixBlockCreatePatchFile(oldFile, newFile, tempPatch, blockKB)
		if err != nil {
			return 0, fmt.Errorf("块大小 %dKB 测试失败: %w", blockKB, err)
		}
		currentSize := int64(fileSize)
		fmt.Printf("测试块大小: %5dKB | 补丁大小: %10d bytes | 耗时: %v\n",
			blockKB, currentSize, time.Since(start).Round(time.Millisecond))

		// 更新最佳结果
		if currentSize == 0 {
			fmt.Printf("块大小 %dKB 生成的补丁大小为0，跳过\n", blockKB)
			continue
		} else if best.blockSizeKB == 0 || currentSize < best.patchSize {
			best = result{
				blockSizeKB: blockKB,
				patchSize:   currentSize,
			}
		} else if currentSize > best.patchSize {
			// 当补丁大小开始增大时，提前终止
			fmt.Printf("补丁大小开始增大，终止测试\n")
			break
		}
	}

	// 使用最佳块大小生成最终补丁
	if best.blockSizeKB == 0 {
		return 0, fmt.Errorf("未找到有效块大小")
	}

	fmt.Printf("\n最优块大小: %dKB (补丁大小: %d bytes)\n", best.blockSizeKB, best.patchSize)
	return best.blockSizeKB, nil
}

func PatchToTarget(oldFile, newFile, patchFile string, blockKB int) (err error) {
	// ================== 参数验证 ==================
	if oldFile == "" || newFile == "" || patchFile == "" {
		return fmt.Errorf("无效参数：oldFile=%q, newFile=%q, patchFile=%q",
			oldFile, newFile, patchFile)
	}

	// ================== 文件校验 ==================
	if err := VerifyFileAccess(oldFile, os.O_RDONLY); err != nil {
		return fmt.Errorf("旧文件不可访问: %w", err)
	}
	if err := VerifyFileAccess(patchFile, os.O_RDONLY); err != nil {
		return fmt.Errorf("补丁文件不可访问: %w", err)
	}

	// ================== 资源管理 ==================
	// 安全打开文件（带错误处理）
	sourceFile, err := os.Open(oldFile)
	if err != nil {
		return fmt.Errorf("打开旧文件失败: %w", err)
	}
	defer SafeClose(sourceFile, &err)

	patchFileHandle, err := os.Open(patchFile)
	if err != nil {
		return fmt.Errorf("打开补丁文件失败: %w", err)
	}
	defer SafeClose(patchFileHandle, &err)

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(newFile), 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	targetFile, err := os.Create(newFile)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer func() {
		// 特别处理目标文件：失败时删除不完整文件
		if err != nil {
			os.Remove(newFile)
		}
		SafeClose(targetFile, &err)
	}()

	// ================== 解码配置 ==================
	options := xdelta.DecoderOptions{
		BlockSizeKB: blockKB, // 块大小：4 MB (4*1024 KB)
		FileID:      "v1.0.0",
		FromFile:    sourceFile,
		ToFile:      targetFile,
		PatchFile:   patchFileHandle,
		EnableStats: true,
	}

	// 配置有效性验证
	if options.BlockSizeKB <= 0 || options.BlockSizeKB > 16*1024 {
		return fmt.Errorf("无效块大小: %d KB", options.BlockSizeKB)
	}

	// ================== 解码过程 ==================
	decoder, err := xdelta.NewDecoder(options)
	if err != nil {
		return fmt.Errorf("解码器初始化失败: %w", err)
	}
	defer func() {
		if closeErr := decoder.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("解码器关闭错误: %w", closeErr)
		}
	}()

	// 带超时控制的上下文（10分钟）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := decoder.Process(ctx); err != nil {
		return fmt.Errorf("补丁应用失败: %w", err)
	}

	// ================== 结果校验 ==================
	if options.EnableStats {
		decoder.DumpStatsToStdout()
	}

	// 验证目标文件完整性
	if err := VerifyFileSize(newFile, oldFile); err != nil {
		return fmt.Errorf("文件完整性校验失败: %w \n", err)
	}

	return nil
}

// 参数验证逻辑
func ValidatePatchArgs(cmd *cobra.Command, args []string) error {
	// 验证文件存在性
	for i, path := range args[:2] { // 仅验证 oldFile 和 newFile
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("参数 %d 文件不存在: %w \n", i+1, err)
		}
	}

	// 验证块大小
	blockSize, _ := cmd.Flags().GetInt("block-size")
	if blockSize <= 0 {
		blockSize = (4) // 4 KB
	}
	if blockSize > 16*1024 {
		return fmt.Errorf("块大小需在 4-16384 KB 之间")
	}
	return nil
}

// 主执行函数
func RunCreatePatch(cmd *cobra.Command, args []string) error {
	// 解析参数
	oldFile := args[0]
	newFile := args[1]
	patchFile := args[2]

	blockSize, _ := cmd.Flags().GetInt("block-size")

	// 创建补丁文件
	if _, err := FixBlockCreatePatchFile(oldFile, newFile, patchFile, blockSize); err != nil {
		return fmt.Errorf("补丁生成失败: %w \n", err)
	}

	// 输出结果
	absPath, _ := filepath.Abs(patchFile)

	sizeValue, _ := EnsureFileSize(absPath, "MB")
	fmt.Printf("补丁文件生成成功！\n路径: %s\n大小: %.2f MB\n",
		absPath, sizeValue,
	)
	return nil
}

// 参数验证逻辑
func ValidateApplyArgs(cmd *cobra.Command, args []string) error {
	// 验证输入文件存在性
	for i, path := range []string{args[0], args[2]} { // oldFile 和 patchFile
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("参数 %d 文件不存在: %w \n", i+1, err)
		}
	}

	// 验证块大小范围
	blockSize, _ := cmd.Flags().GetInt("block-size")
	if blockSize <= 0 {
		blockSize = (4) // 4 KB
	}
	if blockSize > 16*1024 {
		return fmt.Errorf("块大小需在 4-16384 KB 之间")
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(args[1]); err == nil {
		return fmt.Errorf("目标文件已存在: %s", args[1])
	}

	return nil
}

// 主执行函数
func RunApplyPatch(cmd *cobra.Command, args []string) error {
	// 解析参数
	oldFile := args[0]
	newFile := args[1]
	patchFile := args[2]
	blockSize, _ := cmd.Flags().GetInt("block-size")

	// 执行补丁应用
	if err := PatchToTarget(oldFile, newFile, patchFile, blockSize); err != nil {
		// 清理可能生成的不完整文件
		os.Remove(newFile)
		return fmt.Errorf("补丁应用失败: %w \n", err)
	}

	// 输出结果信息
	absNewPath, _ := filepath.Abs(newFile)
	sizeValue, _ := EnsureFileSize(absNewPath, "MB")
	fmt.Printf("新文件生成成功！\n路径: %s\n大小: %.2f MB\n",
		absNewPath, sizeValue,
	)
	return nil
}
