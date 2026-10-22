package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mayingjie198448/niuZip/core/engine"
)

// handleCLI processes command-line shortcuts like --extract-here, --extract-to-dir, --add-to-zip.
// It returns true if a CLI command was handled and the application should exit without launching the GUI.
func handleCLI(args []string) bool {
	if len(args) == 0 {
		return false
	}

	cmd := args[0]
	switch cmd {
	case "-v", "--version", "version":
		fmt.Printf("niuZip %s\n", CurrentAppVersion)
		return true

	case "-h", "--help", "help":
		printHelp()
		return true

	case "--extract-here", "-x":
		if len(args) < 2 {
			showNativeError("niuZip 错误", "请提供要解压的压缩包路径。\n用法: niuZip --extract-here <压缩包路径>")
			return true
		}
		archivePath := args[1]
		destDir := filepath.Dir(archivePath)
		if err := runExtract(archivePath, destDir); err != nil {
			showNativeError("niuZip 解压失败", fmt.Sprintf("解压文件 %s 到当前文件夹失败:\n%v", filepath.Base(archivePath), err))
		}
		return true

	case "--extract-to-dir":
		if len(args) < 2 {
			showNativeError("niuZip 错误", "请提供要解压的压缩包路径。\n用法: niuZip --extract-to-dir <压缩包路径>")
			return true
		}
		archivePath := args[1]
		baseName := getArchiveBaseName(archivePath)
		destDir := filepath.Join(filepath.Dir(archivePath), baseName)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			showNativeError("niuZip 创建目录失败", fmt.Sprintf("无法创建目标文件夹 %s:\n%v", destDir, err))
			return true
		}
		if err := runExtract(archivePath, destDir); err != nil {
			showNativeError("niuZip 解压失败", fmt.Sprintf("解压文件 %s 到文件夹 %s 失败:\n%v", filepath.Base(archivePath), baseName, err))
		}
		return true

	case "--add-to-zip":
		if len(args) < 2 {
			showNativeError("niuZip 错误", "请提供要压缩的文件或文件夹路径。\n用法: niuZip --add-to-zip <源文件或文件夹路径>")
			return true
		}
		sourcePath := args[1]
		if err := runAddToZip(sourcePath); err != nil {
			showNativeError("niuZip 压缩失败", fmt.Sprintf("压缩 %s 失败:\n%v", filepath.Base(sourcePath), err))
		}
		return true
	}

	return false
}

func getArchiveBaseName(filePath string) string {
	cleanPath := strings.ReplaceAll(filePath, "\\", "/")
	base := filepath.Base(cleanPath)
	lower := strings.ToLower(base)
	for _, compoundExt := range []string{".tar.gz", ".tar.bz2", ".tar.xz", ".tar.zst", ".tar.lz4", ".tar.sz"} {
		if strings.HasSuffix(lower, compoundExt) {
			return base[:len(base)-len(compoundExt)]
		}
	}
	ext := filepath.Ext(base)
	if ext != "" {
		return strings.TrimSuffix(base, ext)
	}
	return base
}

func getSourceBaseName(sourcePath string) string {
	clean := filepath.Clean(sourcePath)
	base := filepath.Base(clean)
	fi, err := os.Stat(clean)
	if err == nil && fi.IsDir() {
		return base
	}
	ext := filepath.Ext(base)
	if ext != "" {
		return strings.TrimSuffix(base, ext)
	}
	return base
}

func runExtract(archivePath, destDir string) error {
	fi, err := os.Stat(archivePath)
	if err != nil {
		return fmt.Errorf("找不到文件: %w", err)
	}
	if fi.IsDir() {
		return fmt.Errorf("指定路径是文件夹而非压缩文件")
	}

	broker := engine.NewBroker()
	eng, err := broker.GetEngineForFile(archivePath)
	if err != nil {
		return fmt.Errorf("不支持的压缩格式: %w", err)
	}

	return eng.Extract(context.Background(), archivePath, destDir, nil, nil)
}

func runAddToZip(sourcePath string) error {
	cleanSource := filepath.Clean(sourcePath)
	if _, err := os.Stat(cleanSource); err != nil {
		return fmt.Errorf("找不到源文件或目录: %w", err)
	}

	baseName := getSourceBaseName(cleanSource)
	destZip := filepath.Join(filepath.Dir(cleanSource), baseName+".zip")
	if strings.EqualFold(cleanSource, destZip) {
		destZip = filepath.Join(filepath.Dir(cleanSource), baseName+".niu.zip")
	}

	zipEng := engine.NewPureZipEngine()
	return zipEng.Create(context.Background(), engine.CompressOptions{
		SourcePaths:   []string{cleanSource},
		DestPath:      destZip,
		Format:        "zip",
		CleanMetadata: true,
	}, nil)
}

func printHelp() {
	fmt.Printf(`niuZip - 跨平台新一代高效现代化解压缩管理工具

用法:
  niuZip [选项] [文件/目录]

快捷操作命令行参数 (供资源管理器右键及命令行脚本调用):
  --extract-here <压缩包路径>   解压归档到当前文件夹 (同目录)
  --extract-to-dir <压缩包路径> 解压归档到同名子文件夹中
  --add-to-zip <源文件/文件夹>  快速将文件或文件夹打包为同名 .zip 压缩包
  --add-dialog <源文件/文件夹>  打开主界面并自动预载该路径进入新建压缩向导
  -v, --version                 查看当前版本
  -h, --help                    查看帮助信息
`)
}
