package engine_test

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mayingjie198448/niuZip/core/engine"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func createTestFiles(t *testing.T) (string, []string) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "hello.txt")
	if err := os.WriteFile(file1, []byte("Hello, World!"), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(tempDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(subDir, "deep.log")
	if err := os.WriteFile(file2, []byte("log entry 12345"), 0644); err != nil {
		t.Fatal(err)
	}

	macFile := filepath.Join(tempDir, ".DS_Store")
	if err := os.WriteFile(macFile, []byte("mac junk"), 0644); err != nil {
		t.Fatal(err)
	}

	return tempDir, []string{file1, file2, macFile}
}

func TestPureZip_CreateAndList(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths:   files,
		DestPath:      zipPath,
		CleanMetadata: true, // Should filter .DS_Store
	}

	var progressReports int
	err := zipEng.Create(ctx, opts, func(percent float64, currentFile string) {
		progressReports++
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if progressReports == 0 {
		t.Error("expected at least one progress callback report")
	}

	entries, err := zipEng.List(ctx, zipPath)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Because CleanMetadata is true, .DS_Store should NOT be in the entries
	for _, entry := range entries {
		if entry.Name == ".DS_Store" {
			t.Errorf("expected .DS_Store to be filtered out, but found: %s", entry.Path)
		}
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (hello.txt, deep.log), got %d: %v", len(entries), entries)
	}
}

func TestPureZip_ExtractSingle(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:2],
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	previewDest := filepath.Join(tempDir, "preview.txt")
	err := zipEng.ExtractSingle(ctx, zipPath, "hello.txt", previewDest)
	if err != nil {
		t.Fatalf("ExtractSingle failed: %v", err)
	}

	content, err := os.ReadFile(previewDest)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("content mismatch, got %q, want 'Hello, World!'", string(content))
	}
}

func TestPureZip_ExtractAll(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:2],
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	extractDir := filepath.Join(tempDir, "extracted")
	err := zipEng.Extract(ctx, zipPath, extractDir, nil, nil)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("extracted content mismatch: %q", string(content))
	}
}

func TestPureZip_InPlaceDelete(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:2],
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	// Delete "hello.txt"
	err := zipEng.DeleteEntries(ctx, zipPath, []string{"hello.txt"})
	if err != nil {
		t.Fatalf("DeleteEntries failed: %v", err)
	}

	entries, err := zipEng.List(ctx, zipPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 remaining entry, got %d", len(entries))
	}
	if entries[0].Name != "deep.log" {
		t.Errorf("expected remaining entry 'deep.log', got %q", entries[0].Name)
	}
}

func TestPureZip_InPlaceRename(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:2],
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	// Rename "hello.txt" -> "greeting.txt"
	err := zipEng.RenameEntry(ctx, zipPath, "hello.txt", "greeting.txt")
	if err != nil {
		t.Fatalf("RenameEntry failed: %v", err)
	}

	entries, err := zipEng.List(ctx, zipPath)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, e := range entries {
		if e.Path == "greeting.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("renamed entry 'greeting.txt' not found in entries")
	}

	// Check content of renamed entry
	previewDest := filepath.Join(tempDir, "check_renamed.txt")
	if err := zipEng.ExtractSingle(ctx, zipPath, "greeting.txt", previewDest); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(previewDest)
	if !bytes.Equal(content, []byte("Hello, World!")) {
		t.Errorf("content changed after rename: %s", string(content))
	}
}

func TestPureZip_InPlaceAppend(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "test.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:1], // only hello.txt
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	// Append a new file "appended.txt"
	newFilePath := filepath.Join(tempDir, "new_file.txt")
	if err := os.WriteFile(newFilePath, []byte("Appended Content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := zipEng.AppendEntry(ctx, zipPath, newFilePath, "extra/appended.txt")
	if err != nil {
		t.Fatalf("AppendEntry failed: %v", err)
	}

	entries, err := zipEng.List(ctx, zipPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	previewDest := filepath.Join(tempDir, "check_appended.txt")
	if err := zipEng.ExtractSingle(ctx, zipPath, "extra/appended.txt", previewDest); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(previewDest)
	if !bytes.Equal(content, []byte("Appended Content")) {
		t.Errorf("appended content mismatch: %s", string(content))
	}
}

func TestPureZip_PasswordProtectedDetection(t *testing.T) {
	// A zip file with the encryption bit set (bit 0 of general purpose bit flag)
	buf := new(bytes.Buffer)
	zw := engine.NewMockEncryptedZipWriter(buf)
	zw.WriteEncryptedEntry("secret.txt", []byte("secret data"))
	zw.Close()

	tempFile := filepath.Join(t.TempDir(), "encrypted.zip")
	if err := os.WriteFile(tempFile, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	zipEng := engine.NewPureZipEngine()
	entries, err := zipEng.List(context.Background(), tempFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 || !entries[0].IsEncrypted {
		t.Fatalf("expected 1 encrypted entry, got: %+v", entries)
	}
}

func TestPureZip_GBKChineseFilenames(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "gbk_test.zip")

	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)

	// Create GBK encoded entry with nonUTF8: true (standard Chinese Windows behavior)
	gbkEncoder := simplifiedchinese.GB18030.NewEncoder()
	gbkName, err := gbkEncoder.String("立达自动化/设备清单_2026.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	h := &zip.FileHeader{
		Name:    gbkName,
		NonUTF8: true, // Bit 11 is 0
		Method:  zip.Deflate,
	}
	w, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("Excel content"))
	zw.Close()
	f.Close()

	// 1. Test listing decodes to proper UTF-8 Chinese
	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()
	entries, err := zipEng.List(ctx, zipPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "立达自动化/设备清单_2026.xlsx" {
		t.Errorf("expected decoded Chinese path, got: %q", entries[0].Path)
	}
	if entries[0].Name != "设备清单_2026.xlsx" {
		t.Errorf("expected decoded Chinese name, got: %q", entries[0].Name)
	}

	// 2. Test single extraction with decoded Chinese path
	extractDest := filepath.Join(tempDir, "extracted.xlsx")
	if err := zipEng.ExtractSingle(ctx, zipPath, "立达自动化/设备清单_2026.xlsx", extractDest); err != nil {
		t.Fatalf("ExtractSingle failed: %v", err)
	}
	data, _ := os.ReadFile(extractDest)
	if string(data) != "Excel content" {
		t.Errorf("extracted content mismatch: %s", string(data))
	}

	// 3. Test full extraction creates proper Chinese directory and file on disk
	extractAllDir := filepath.Join(tempDir, "all_out")
	if err := zipEng.Extract(ctx, zipPath, extractAllDir, nil, nil); err != nil {
		t.Fatalf("Extract all failed: %v", err)
	}
	diskPath := filepath.Join(extractAllDir, "立达自动化", "设备清单_2026.xlsx")
	if _, err := os.Stat(diskPath); err != nil {
		t.Fatalf("expected file to exist at %s, but err: %v", diskPath, err)
	}
}
