package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mayingjie198448/niuZip/core/engine"
)

func TestTarEngine_CreateAndListAndExtract(t *testing.T) {
	tempDir, files := createTestFiles(t)
	tarPath := filepath.Join(tempDir, "archive.tar.gz")

	tarEng := engine.NewTarEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths:   files,
		DestPath:      tarPath,
		CleanMetadata: true,
	}

	err := tarEng.Create(ctx, opts, nil)
	if err != nil {
		t.Fatalf("Tar Create failed: %v", err)
	}

	entries, err := tarEng.List(ctx, tarPath)
	if err != nil {
		t.Fatalf("Tar List failed: %v", err)
	}

	// .DS_Store filtered out by CleanMetadata
	for _, e := range entries {
		if e.Name == ".DS_Store" {
			t.Errorf("expected .DS_Store to be filtered, found %s", e.Path)
		}
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Single extract
	previewDest := filepath.Join(tempDir, "preview_tar.txt")
	err = tarEng.ExtractSingle(ctx, tarPath, "hello.txt", previewDest)
	if err != nil {
		t.Fatalf("Tar ExtractSingle failed: %v", err)
	}

	content, err := os.ReadFile(previewDest)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("content mismatch, got %q", string(content))
	}

	// Full extract
	extractDir := filepath.Join(tempDir, "extracted_tar")
	err = tarEng.Extract(ctx, tarPath, extractDir, nil, nil)
	if err != nil {
		t.Fatalf("Tar Extract failed: %v", err)
	}

	extractedFile := filepath.Join(extractDir, "hello.txt")
	c, _ := os.ReadFile(extractedFile)
	if string(c) != "Hello, World!" {
		t.Errorf("extracted content mismatch: %s", string(c))
	}
}
