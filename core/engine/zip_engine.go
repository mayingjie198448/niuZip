package engine

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mayingjie198448/niuZip/core/cleaner"
	"github.com/mayingjie198448/niuZip/core/vfs"
)

// PureZipEngine implements ArchiveEngine and ArchiveEditor for ZIP files using pure Go.
type PureZipEngine struct{}

func NewPureZipEngine() *PureZipEngine {
	return &PureZipEngine{}
}

// List reads entries directly from the ZIP Central Directory.
func (e *PureZipEngine) List(ctx context.Context, archivePath string) ([]vfs.FileEntry, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	entries := make([]vfs.FileEntry, 0, len(r.File))
	for _, f := range r.File {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		decodedName := cleaner.DecodeZipFilename(f.Name, f.NonUTF8)
		info := f.FileInfo()
		isEncrypted := (f.Flags & 1) != 0

		entries = append(entries, vfs.FileEntry{
			Path:        filepath.ToSlash(decodedName),
			Name:        filepath.Base(decodedName),
			IsDir:       info.IsDir() || strings.HasSuffix(f.Name, "/") || strings.HasSuffix(decodedName, "/"),
			Size:        int64(f.UncompressedSize64),
			PackedSize:  int64(f.CompressedSize64),
			ModTime:     f.Modified,
			IsEncrypted: isEncrypted,
			CRC32:       f.CRC32,
		})
	}

	return entries, nil
}

// ExtractSingle extracts a single entry to destFile for preview.
func (e *PureZipEngine) ExtractSingle(ctx context.Context, archivePath string, entryPath string, destFile string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	normTarget := filepath.ToSlash(entryPath)
	var targetFile *zip.File
	for _, f := range r.File {
		decoded := filepath.ToSlash(cleaner.DecodeZipFilename(f.Name, f.NonUTF8))
		if decoded == normTarget || filepath.ToSlash(f.Name) == normTarget {
			targetFile = f
			break
		}
	}

	if targetFile == nil {
		return ErrEntryNotFound
	}

	if (targetFile.Flags & 1) != 0 {
		return ErrPasswordRequired
	}

	rc, err := targetFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// Extract extracts selected or all files to destDir with Zip Slip protection.
func (e *PureZipEngine) Extract(ctx context.Context, archivePath string, destDir string, selectedFiles []string, onProgress ProgressCallback) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	filterMap := make(map[string]bool)
	for _, file := range selectedFiles {
		filterMap[filepath.ToSlash(file)] = true
	}

	total := len(r.File)
	for i, f := range r.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		decodedName := cleaner.DecodeZipFilename(f.Name, f.NonUTF8)
		cleanName := filepath.ToSlash(decodedName)
		rawCleanName := filepath.ToSlash(f.Name)
		if len(filterMap) > 0 && !filterMap[cleanName] && !filterMap[rawCleanName] {
			continue
		}

		if (f.Flags & 1) != 0 {
			return ErrPasswordRequired
		}

		// Security: Zip Slip Defense
		targetPath, err := cleaner.ValidateSafePath(destDir, cleanName)
		if err != nil {
			return fmt.Errorf("security violation for %s: %w", cleanName, err)
		}

		if f.FileInfo().IsDir() || strings.HasSuffix(cleanName, "/") || strings.HasSuffix(f.Name, "/") {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		if err := extractFile(f, targetPath); err != nil {
			return err
		}

		if onProgress != nil {
			percent := float64(i+1) / float64(total) * 100.0
			onProgress(percent, cleanName)
		}
	}

	return nil
}

func extractFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// Create packages sources into a new ZIP archive.
func (e *PureZipEngine) Create(ctx context.Context, opts CompressOptions, onProgress ProgressCallback) error {
	if err := os.MkdirAll(filepath.Dir(opts.DestPath), 0755); err != nil {
		return err
	}

	zipFile, err := os.Create(opts.DestPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	type fileToZip struct {
		diskPath string
		zipPath  string
		isDir    bool
	}

	var items []fileToZip
	for _, src := range opts.SourcePaths {
		srcBase := filepath.Base(src)
		info, err := os.Stat(src)
		if err != nil {
			return err
		}

		if info.IsDir() {
			err = filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				rel, err := filepath.Rel(src, p)
				if err != nil {
					return err
				}
				entryZipPath := filepath.ToSlash(filepath.Join(srcBase, rel))
				if opts.CleanMetadata && cleaner.IsMacMetadata(entryZipPath) {
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				items = append(items, fileToZip{
					diskPath: p,
					zipPath:  entryZipPath,
					isDir:    fi.IsDir(),
				})
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			if opts.CleanMetadata && cleaner.IsMacMetadata(srcBase) {
				continue
			}
			items = append(items, fileToZip{
				diskPath: src,
				zipPath:  srcBase,
				isDir:    false,
			})
		}
	}

	total := len(items)
	for i, item := range items {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fi, err := os.Lstat(item.diskPath)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		header.Name = item.zipPath
		if item.isDir && !strings.HasSuffix(header.Name, "/") {
			header.Name += "/"
		}
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if !item.isDir {
			fileContent, err := os.Open(item.diskPath)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, fileContent)
			fileContent.Close()
			if err != nil {
				return err
			}
		}

		if onProgress != nil {
			percent := float64(i+1) / float64(total) * 100.0
			onProgress(percent, item.zipPath)
		}
	}

	return nil
}

// DeleteEntries performs transactional removal of target entries without recompressing unchanged files.
func (e *PureZipEngine) DeleteEntries(ctx context.Context, archivePath string, targetPaths []string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	deleteSet := make(map[string]bool)
	for _, p := range targetPaths {
		deleteSet[filepath.ToSlash(p)] = true
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(archivePath), ".niuzip_tmp_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up on failure

	w := zip.NewWriter(tmpFile)

	for _, f := range r.File {
		select {
		case <-ctx.Done():
			_ = w.Close()
			_ = tmpFile.Close()
			return ctx.Err()
		default:
		}

		decodedName := filepath.ToSlash(cleaner.DecodeZipFilename(f.Name, f.NonUTF8))
		if deleteSet[decodedName] || deleteSet[filepath.ToSlash(f.Name)] {
			continue
		}

		fw, err := w.CreateRaw(&f.FileHeader)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		rc, err := f.OpenRaw()
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		_, err = io.Copy(fw, rc)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}
	}

	if err := w.Close(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()
	_ = r.Close()

	// Atomic replace
	return os.Rename(tmpFile.Name(), archivePath)
}

// RenameEntry performs transactional renaming of an entry without recompressing file contents.
func (e *PureZipEngine) RenameEntry(ctx context.Context, archivePath string, oldPath, newPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	normOld := filepath.ToSlash(oldPath)
	normNew := filepath.ToSlash(newPath)

	tmpFile, err := os.CreateTemp(filepath.Dir(archivePath), ".niuzip_tmp_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	w := zip.NewWriter(tmpFile)

	for _, f := range r.File {
		select {
		case <-ctx.Done():
			_ = w.Close()
			_ = tmpFile.Close()
			return ctx.Err()
		default:
		}

		header := f.FileHeader
		decodedName := filepath.ToSlash(cleaner.DecodeZipFilename(f.Name, f.NonUTF8))
		if decodedName == normOld || filepath.ToSlash(f.Name) == normOld {
			header.Name = normNew
			header.NonUTF8 = false
			header.Flags |= 0x800
		}

		fw, err := w.CreateRaw(&header)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		rc, err := f.OpenRaw()
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		_, err = io.Copy(fw, rc)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}
	}

	if err := w.Close(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()
	_ = r.Close()

	return os.Rename(tmpFile.Name(), archivePath)
}

// AppendEntry appends a new file into an existing ZIP archive using transactional raw copy.
func (e *PureZipEngine) AppendEntry(ctx context.Context, archivePath string, sourceFilePath, targetEntryPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	tmpFile, err := os.CreateTemp(filepath.Dir(archivePath), ".niuzip_tmp_*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	w := zip.NewWriter(tmpFile)

	// Copy existing files raw
	for _, f := range r.File {
		select {
		case <-ctx.Done():
			_ = w.Close()
			_ = tmpFile.Close()
			return ctx.Err()
		default:
		}

		fw, err := w.CreateRaw(&f.FileHeader)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		rc, err := f.OpenRaw()
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}

		_, err = io.Copy(fw, rc)
		if err != nil {
			_ = w.Close()
			_ = tmpFile.Close()
			return err
		}
	}

	// Append the new file
	fi, err := os.Stat(sourceFilePath)
	if err != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return err
	}

	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return err
	}
	header.Name = filepath.ToSlash(targetEntryPath)
	header.Method = zip.Deflate

	fw, err := w.CreateHeader(header)
	if err != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return err
	}

	srcFile, err := os.Open(sourceFilePath)
	if err != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return err
	}
	_, err = io.Copy(fw, srcFile)
	srcFile.Close()
	if err != nil {
		_ = w.Close()
		_ = tmpFile.Close()
		return err
	}

	if err := w.Close(); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()
	_ = r.Close()

	return os.Rename(tmpFile.Name(), archivePath)
}

// MockEncryptedZipWriter helps unit tests generate a ZIP with encryption flag set.
type MockEncryptedZipWriter struct {
	w *zip.Writer
}

func NewMockEncryptedZipWriter(w io.Writer) *MockEncryptedZipWriter {
	return &MockEncryptedZipWriter{w: zip.NewWriter(w)}
}

func (m *MockEncryptedZipWriter) WriteEncryptedEntry(name string, data []byte) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
		Flags:  1, // Bit 0 = Encrypted
	}
	fw, err := m.w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

func (m *MockEncryptedZipWriter) Close() error {
	return m.w.Close()
}
