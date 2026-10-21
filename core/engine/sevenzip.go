package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mayingjie198448/niuZip/core/cleaner"
	"github.com/mayingjie198448/niuZip/core/vfs"
)

var (
	progressPercentRegex = regexp.MustCompile(`([0-9]+)%`)
)

// SevenZipEngine delegates to the embedded or system 7-Zip CLI for 7z/RAR/encrypted/split archives.
type SevenZipEngine struct {
	customBinPath string
}

func NewSevenZipEngine() *SevenZipEngine {
	return &SevenZipEngine{}
}

func (e *SevenZipEngine) getBinaryPath() (string, error) {
	if e.customBinPath != "" {
		return e.customBinPath, nil
	}

	// 1. Try local cache extraction from EmbeddedSevenZip
	if len(EmbeddedSevenZip) > 0 {
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			binDir := filepath.Join(cacheDir, "niuzip", "bin")
			_ = os.MkdirAll(binDir, 0755)
			targetBin := filepath.Join(binDir, SevenZipBinName)

			fi, err := os.Stat(targetBin)
			if err != nil || fi.Size() != int64(len(EmbeddedSevenZip)) {
				if writeErr := os.WriteFile(targetBin, EmbeddedSevenZip, 0755); writeErr == nil {
					_ = os.Chmod(targetBin, 0755)
					return targetBin, nil
				}
			} else {
				return targetBin, nil
			}
		}
	}

	// 2. Check local repo dev binary
	repoBinPath := filepath.Join("core", "engine", "bin", "darwin_arm64", SevenZipBinName)
	if fi, err := os.Stat(repoBinPath); err == nil && !fi.IsDir() {
		return repoBinPath, nil
	}

	// 3. Check system PATH
	for _, name := range []string{"7zz", "7z", "7za"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("7-Zip binary not found (looked for %s)", SevenZipBinName)
}

// List runs '7zz l -slt <archive>' and parses the technical list output.
func (e *SevenZipEngine) List(ctx context.Context, archivePath string) ([]vfs.FileEntry, error) {
	bin, err := e.getBinaryPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, bin, "l", "-slt", "-ba", "-sccUTF-8", archivePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var entries []vfs.FileEntry
	scanner := bufio.NewScanner(stdout)

	var current vfs.FileEntry
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if inBlock && current.Path != "" {
				decodedPath := cleaner.DecodeZipFilename(current.Path, false)
				current.Path = filepath.ToSlash(decodedPath)
				current.Name = filepath.Base(decodedPath)
				entries = append(entries, current)
				current = vfs.FileEntry{}
				inBlock = false
			}
			continue
		}

		parts := strings.SplitN(trimmed, " = ", 2)
		if len(parts) != 2 {
			continue
		}

		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		inBlock = true

		switch key {
		case "Path":
			current.Path = val
		case "Size":
			if sz, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.Size = sz
			}
		case "Packed Size":
			if ps, err := strconv.ParseInt(val, 10, 64); err == nil {
				current.PackedSize = ps
			}
		case "Modified":
			// Typical 7-Zip format: 2026-10-17 22:54:10
			if t, err := time.Parse("2006-01-02 15:04:05", val); err == nil {
				current.ModTime = t
			}
		case "Folder":
			current.IsDir = (val == "+")
		case "Encrypted":
			current.IsEncrypted = (val == "+")
		case "CRC":
			if crc, err := strconv.ParseUint(val, 16, 32); err == nil {
				current.CRC32 = uint32(crc)
			}
		}
	}

	if inBlock && current.Path != "" {
		decodedPath := cleaner.DecodeZipFilename(current.Path, false)
		current.Path = filepath.ToSlash(decodedPath)
		current.Name = filepath.Base(decodedPath)
		entries = append(entries, current)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("7-Zip listing failed: %w", err)
	}

	return entries, nil
}

// Extract extracts archives via '7zz x <archive> -o<destDir> -bsp1 -y'.
func (e *SevenZipEngine) Extract(ctx context.Context, archivePath string, destDir string, selectedFiles []string, onProgress ProgressCallback) error {
	bin, err := e.getBinaryPath()
	if err != nil {
		return err
	}

	args := []string{"x", archivePath, "-o" + destDir, "-bsp1", "-y", "-sccUTF-8"}
	if len(selectedFiles) > 0 {
		args = append(args, selectedFiles...)
	}

	cmd := exec.CommandContext(ctx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := progressPercentRegex.FindStringSubmatch(line); len(matches) > 1 {
			if p, err := strconv.ParseFloat(matches[1], 64); err == nil && onProgress != nil {
				onProgress(p, "")
			}
		}
	}

	return cmd.Wait()
}

// ExtractSingle extracts one entry to destFile.
func (e *SevenZipEngine) ExtractSingle(ctx context.Context, archivePath string, entryPath string, destFile string) error {
	bin, err := e.getBinaryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cmd := exec.CommandContext(ctx, bin, "e", archivePath, entryPath, "-so", "-sccUTF-8")
	cmd.Stdout = outFile

	return cmd.Run()
}

// Create compresses files with optional AES-256 encryption and multi-volume splits.
func (e *SevenZipEngine) Create(ctx context.Context, opts CompressOptions, onProgress ProgressCallback) error {
	bin, err := e.getBinaryPath()
	if err != nil {
		return err
	}

	args := []string{"a", opts.DestPath, "-bsp1", "-y"}

	if opts.Password != "" {
		args = append(args, "-p"+opts.Password)
		if opts.EncryptHeader {
			args = append(args, "-mhe=on")
		}
	}

	if opts.SplitSizeMB > 0 {
		args = append(args, fmt.Sprintf("-v%dm", opts.SplitSizeMB))
	}

	args = append(args, opts.SourcePaths...)

	cmd := exec.CommandContext(ctx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := progressPercentRegex.FindStringSubmatch(line); len(matches) > 1 {
			if p, err := strconv.ParseFloat(matches[1], 64); err == nil && onProgress != nil {
				onProgress(p, "")
			}
		}
	}

	return cmd.Wait()
}
