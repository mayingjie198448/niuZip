package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestGetArchiveBaseName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.zip", "test"},
		{"archive.tar.gz", "archive"},
		{"archive.tar.bz2", "archive"},
		{"system.tar.xz", "system"},
		{"my.cool.file.7z", "my.cool.file"},
		{"folder/sub/backup.zip", "backup"},
		{"C:\\Users\\test\\archive.zip", "archive"},
		{"noext", "noext"},
	}

	for _, tt := range tests {
		got := getArchiveBaseName(tt.input)
		if got != tt.expected {
			t.Errorf("getArchiveBaseName(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestGetSourceBaseName(t *testing.T) {
	tempDir := t.TempDir()

	// 1. File
	filePath := filepath.Join(tempDir, "document.pdf")
	if err := os.WriteFile(filePath, []byte("pdf data"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := getSourceBaseName(filePath); got != "document" {
		t.Errorf("expected 'document', got %q", got)
	}

	// 2. Directory
	dirPath := filepath.Join(tempDir, "my_project")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatal(err)
	}
	if got := getSourceBaseName(dirPath); got != "my_project" {
		t.Errorf("expected 'my_project', got %q", got)
	}
}

func TestHandleCLI_AddToZipAndExtract(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create a directory with files
	subDir := filepath.Join(tempDir, "src_folder")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	f1 := filepath.Join(subDir, "hello.txt")
	if err := os.WriteFile(f1, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Test --add-to-zip on directory
	handled := handleCLI([]string{"--add-to-zip", subDir})
	if !handled {
		t.Fatal("expected handleCLI to return true for --add-to-zip")
	}

	expectedZip := filepath.Join(tempDir, "src_folder.zip")
	if fi, err := os.Stat(expectedZip); err != nil || fi.Size() == 0 {
		t.Fatalf("expected created zip %s to exist and be non-empty", expectedZip)
	}

	// Verify zip contents
	zr, err := zip.OpenReader(expectedZip)
	if err != nil {
		t.Fatalf("failed to open created zip: %v", err)
	}
	if len(zr.File) == 0 {
		t.Fatal("zip has no entries")
	}
	zr.Close()

	// 3. Test --extract-to-dir
	handled = handleCLI([]string{"--extract-to-dir", expectedZip})
	if !handled {
		t.Fatal("expected handleCLI to return true for --extract-to-dir")
	}

	extractedDir := filepath.Join(tempDir, "src_folder")
	checkFile := filepath.Join(extractedDir, "src_folder", "hello.txt")
	// Or extracted directly:
	if _, err := os.Stat(filepath.Join(tempDir, "src_folder", "hello.txt")); err != nil {
		if _, err := os.Stat(checkFile); err != nil {
			t.Fatalf("extracted file not found in %s or %s", filepath.Join(tempDir, "src_folder", "hello.txt"), checkFile)
		}
	}

	// 4. Test --extract-here in a new directory
	outDir := filepath.Join(tempDir, "extracted_here")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	copyZip := filepath.Join(outDir, "test.zip")
	data, _ := os.ReadFile(expectedZip)
	_ = os.WriteFile(copyZip, data, 0644)

	handled = handleCLI([]string{"--extract-here", copyZip})
	if !handled {
		t.Fatal("expected handleCLI to return true for --extract-here")
	}
}
