package engine

import (
	"context"
	"errors"

	"github.com/mayingjie198448/niuZip/core/vfs"
)

var (
	ErrPasswordRequired = errors.New("archive is password protected")
	ErrEntryNotFound    = errors.New("entry not found in archive")
	ErrUnsupportedFormat = errors.New("unsupported archive format")
)

// ProgressCallback reports operation percentage (0.0 to 100.0) and current filename.
type ProgressCallback func(percent float64, currentFile string)

// CompressOptions defines parameters for creating an archive.
type CompressOptions struct {
	SourcePaths   []string // List of files/folders to compress
	DestPath      string   // Target archive path
	Format        string   // "zip", "7z", "tar.gz"
	Password      string   // Password for AES encryption
	EncryptHeader bool     // Encrypt filenames (7z specific)
	SplitSizeMB   int      // Split size in MB (0 = no split)
	CleanMetadata bool     // Filter macOS .DS_Store, __MACOSX, ._*
}

// ArchiveEngine defines the contract for archive format handlers.
type ArchiveEngine interface {
	// List reads metadata from the archive header/central directory without extracting content.
	List(ctx context.Context, archivePath string) ([]vfs.FileEntry, error)

	// Extract extracts selected files (or all files if selectedFiles is empty) to destDir.
	Extract(ctx context.Context, archivePath string, destDir string, selectedFiles []string, onProgress ProgressCallback) error

	// ExtractSingle extracts a single file to destFile for preview purposes.
	ExtractSingle(ctx context.Context, archivePath string, entryPath string, destFile string) error

	// Create creates a new archive with the specified options.
	Create(ctx context.Context, opts CompressOptions, onProgress ProgressCallback) error
}

// ArchiveEditor defines capabilities for transactional in-place editing.
type ArchiveEditor interface {
	DeleteEntries(ctx context.Context, archivePath string, targetPaths []string) error
	RenameEntry(ctx context.Context, archivePath string, oldPath, newPath string) error
	AppendEntry(ctx context.Context, archivePath string, sourceFilePath, targetEntryPath string) error
}
