package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mayingjie198448/niuZip/core/vfs"
)

// Broker coordinates archive engine selection based on file extensions and magic headers.
type Broker struct {
	zipEngine      ArchiveEngine
	tarEngine      ArchiveEngine
	sevenZipEngine ArchiveEngine
}

func NewBroker() *Broker {
	return &Broker{
		zipEngine:      NewPureZipEngine(),
		tarEngine:      NewTarEngine(),
		sevenZipEngine: NewSevenZipEngine(),
	}
}

// DetectFormat detects archive format from filename or file extension.
func (b *Broker) DetectFormat(filename string) string {
	lower := strings.ToLower(filename)

	// Compound extensions
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return "tar.gz"
	}
	if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz") || strings.HasSuffix(lower, ".tbz2") {
		return "tar.bz2"
	}
	if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return "tar.xz"
	}

	ext := strings.TrimPrefix(filepath.Ext(lower), ".")
	switch ext {
	case "zip", "zipx", "jar", "apk":
		return "zip"
	case "tar":
		return "tar"
	case "7z":
		return "7z"
	case "rar", "cbr":
		return "rar"
	case "iso":
		return "iso"
	case "dmg":
		return "dmg"
	case "cab":
		return "cab"
	case "gz":
		return "gz"
	case "bz2":
		return "bz2"
	case "xz":
		return "xz"
	default:
		return ""
	}
}

// DetectFormatWithFile attempts extension detection, and falls back to magic byte inspection.
func (b *Broker) DetectFormatWithFile(filePath string) string {
	fmtName := b.DetectFormat(filePath)
	if fmtName != "" {
		return fmtName
	}

	// Read magic bytes
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	header := make([]byte, 16)
	n, err := io.ReadFull(f, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return ""
	}
	header = header[:n]

	// Check magic signatures
	if bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) {
		return "zip"
	}
	if bytes.HasPrefix(header, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}) {
		return "7z"
	}
	if bytes.HasPrefix(header, []byte{0x52, 0x61, 0x72, 0x21}) {
		return "rar"
	}
	if bytes.HasPrefix(header, []byte{0x1F, 0x8B}) {
		return "gz"
	}
	if bytes.HasPrefix(header, []byte{0x42, 0x5A, 0x68}) {
		return "bz2"
	}
	if bytes.HasPrefix(header, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}) {
		return "xz"
	}

	return ""
}

// GetEngineForFile returns the suitable ArchiveEngine for the given file.
func (b *Broker) GetEngineForFile(filePath string) (ArchiveEngine, error) {
	fmtName := b.DetectFormatWithFile(filePath)
	switch fmtName {
	case "zip":
		return b.zipEngine, nil
	case "tar", "tar.gz", "tar.bz2", "tar.xz":
		return b.tarEngine, nil
	case "7z", "rar", "iso", "dmg", "cab", "gz", "bz2", "xz":
		return b.sevenZipEngine, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, filePath)
	}
}

// List delegates to the resolved engine.
func (b *Broker) List(ctx context.Context, archivePath string) ([]vfs.FileEntry, error) {
	eng, err := b.GetEngineForFile(archivePath)
	if err != nil {
		return nil, err
	}
	return eng.List(ctx, archivePath)
}

// Extract delegates to the resolved engine.
func (b *Broker) Extract(ctx context.Context, archivePath string, destDir string, selectedFiles []string, onProgress ProgressCallback) error {
	eng, err := b.GetEngineForFile(archivePath)
	if err != nil {
		return err
	}
	return eng.Extract(ctx, archivePath, destDir, selectedFiles, onProgress)
}

// ExtractSingle delegates to the resolved engine.
func (b *Broker) ExtractSingle(ctx context.Context, archivePath string, entryPath string, destFile string) error {
	eng, err := b.GetEngineForFile(archivePath)
	if err != nil {
		return err
	}
	return eng.ExtractSingle(ctx, archivePath, entryPath, destFile)
}
