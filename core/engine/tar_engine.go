package engine

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mayingjie198448/niuZip/core/cleaner"
	"github.com/mayingjie198448/niuZip/core/vfs"
)

// TarEngine implements ArchiveEngine for TAR and TAR.GZ / TAR.BZ2 using pure Go.
type TarEngine struct{}

func NewTarEngine() *TarEngine {
	return &TarEngine{}
}

func (e *TarEngine) openReader(filePath string) (io.ReadCloser, *tar.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}

	lower := strings.ToLower(filePath)
	var stream io.Reader = file

	if strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".tgz") {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			file.Close()
			return nil, nil, err
		}
		stream = gzReader
	} else if strings.HasSuffix(lower, ".bz2") || strings.HasSuffix(lower, ".tbz") || strings.HasSuffix(lower, ".tbz2") {
		stream = bzip2.NewReader(file)
	}

	return file, tar.NewReader(stream), nil
}

func (e *TarEngine) List(ctx context.Context, archivePath string) ([]vfs.FileEntry, error) {
	rawFile, tr, err := e.openReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer rawFile.Close()

	var entries []vfs.FileEntry
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		decodedName := cleaner.DecodeZipFilename(hdr.Name, false)
		cleanPath := filepath.ToSlash(decodedName)
		entries = append(entries, vfs.FileEntry{
			Path:        cleanPath,
			Name:        filepath.Base(cleanPath),
			IsDir:       hdr.Typeflag == tar.TypeDir || strings.HasSuffix(cleanPath, "/"),
			Size:        hdr.Size,
			ModTime:     hdr.ModTime,
			IsEncrypted: false,
		})
	}

	return entries, nil
}

func (e *TarEngine) ExtractSingle(ctx context.Context, archivePath string, entryPath string, destFile string) error {
	rawFile, tr, err := e.openReader(archivePath)
	if err != nil {
		return err
	}
	defer rawFile.Close()

	normTarget := filepath.ToSlash(entryPath)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		decoded := filepath.ToSlash(cleaner.DecodeZipFilename(hdr.Name, false))
		if decoded == normTarget || filepath.ToSlash(hdr.Name) == normTarget {
			if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
				return err
			}
			out, err := os.Create(destFile)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, tr)
			return err
		}
	}

	return ErrEntryNotFound
}

func (e *TarEngine) Extract(ctx context.Context, archivePath string, destDir string, selectedFiles []string, onProgress ProgressCallback) error {
	rawFile, tr, err := e.openReader(archivePath)
	if err != nil {
		return err
	}
	defer rawFile.Close()

	filterMap := make(map[string]bool)
	for _, file := range selectedFiles {
		filterMap[filepath.ToSlash(file)] = true
	}

	count := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		decodedName := cleaner.DecodeZipFilename(hdr.Name, false)
		cleanName := filepath.ToSlash(decodedName)
		rawCleanName := filepath.ToSlash(hdr.Name)
		if len(filterMap) > 0 && !filterMap[cleanName] && !filterMap[rawCleanName] {
			continue
		}

		targetPath, err := cleaner.ValidateSafePath(destDir, cleanName)
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeDir || strings.HasSuffix(cleanName, "/") {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()

		count++
		if onProgress != nil {
			onProgress(0.0, hdr.Name)
		}
	}

	return nil
}

func (e *TarEngine) Create(ctx context.Context, opts CompressOptions, onProgress ProgressCallback) error {
	if err := os.MkdirAll(filepath.Dir(opts.DestPath), 0755); err != nil {
		return err
	}

	tarFile, err := os.Create(opts.DestPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	var writer io.Writer = tarFile
	var gzWriter *gzip.Writer

	if strings.HasSuffix(strings.ToLower(opts.DestPath), ".gz") || strings.HasSuffix(strings.ToLower(opts.DestPath), ".tgz") {
		gzWriter = gzip.NewWriter(tarFile)
		defer gzWriter.Close()
		writer = gzWriter
	}

	tw := tar.NewWriter(writer)
	defer tw.Close()

	for _, src := range opts.SourcePaths {
		base := filepath.Base(src)
		err := filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}
			entryPath := filepath.ToSlash(filepath.Join(base, rel))
			if opts.CleanMetadata && cleaner.IsMacMetadata(entryPath) {
				if fi.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			hdr, err := tar.FileInfoHeader(fi, "")
			if err != nil {
				return err
			}
			hdr.Name = entryPath

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			if !fi.IsDir() {
				file, err := os.Open(p)
				if err != nil {
					return err
				}
				defer file.Close()
				if _, err := io.Copy(tw, file); err != nil {
					return err
				}
			}

			if onProgress != nil {
				onProgress(0.0, entryPath)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
