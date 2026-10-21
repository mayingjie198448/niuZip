package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mayingjie198448/niuZip/core/engine"
)

func TestSevenZipEngine_CreateListAndExtract(t *testing.T) {
	tempDir, files := createTestFiles(t)
	archivePath := filepath.Join(tempDir, "test_archive.7z")

	sevenEng := engine.NewSevenZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:2],
		DestPath:    archivePath,
	}

	// 1. Create 7z
	err := sevenEng.Create(ctx, opts, nil)
	if err != nil {
		t.Fatalf("7z Create failed: %v", err)
	}

	// 2. List 7z
	entries, err := sevenEng.List(ctx, archivePath)
	if err != nil {
		t.Fatalf("7z List failed: %v", err)
	}

	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d: %+v", len(entries), entries)
	}

	foundHello := false
	for _, e := range entries {
		if e.Name == "hello.txt" {
			foundHello = true
			if e.Size != int64(len("Hello, World!")) {
				t.Errorf("expected size %d, got %d", len("Hello, World!"), e.Size)
			}
		}
	}
	if !foundHello {
		t.Error("hello.txt not found in 7z entries")
	}

	// 3. Extract single file
	previewDest := filepath.Join(tempDir, "preview_7z.txt")
	err = sevenEng.ExtractSingle(ctx, archivePath, "hello.txt", previewDest)
	if err != nil {
		t.Fatalf("7z ExtractSingle failed: %v", err)
	}
	content, err := os.ReadFile(previewDest)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("preview content mismatch: %q", string(content))
	}

	// 4. Extract all files
	extractDir := filepath.Join(tempDir, "extracted_7z")
	err = sevenEng.Extract(ctx, archivePath, extractDir, nil, nil)
	if err != nil {
		t.Fatalf("7z Extract failed: %v", err)
	}
	extractedContent, err := os.ReadFile(filepath.Join(extractDir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(extractedContent) != "Hello, World!" {
		t.Errorf("extracted content mismatch: %q", string(extractedContent))
	}
}

func TestSevenZipEngine_EncryptedArchive(t *testing.T) {
	tempDir, files := createTestFiles(t)
	encArchive := filepath.Join(tempDir, "encrypted.7z")

	sevenEng := engine.NewSevenZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths:   files[:1],
		DestPath:      encArchive,
		Password:      "SecretKey123",
		EncryptHeader: true,
	}

	err := sevenEng.Create(ctx, opts, nil)
	if err != nil {
		t.Fatalf("Encrypted 7z Create failed: %v", err)
	}

	// Listing with encrypted header should indicate encryption
	entries, err := sevenEng.List(ctx, encArchive)
	if err != nil {
		// Encrypted header without password may return password prompt or error, which is valid
		t.Logf("List without password returned: %v", err)
	} else if len(entries) > 0 {
		if !entries[0].IsEncrypted {
			t.Error("expected entry to be marked IsEncrypted")
		}
	}
}
