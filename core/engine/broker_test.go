package engine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mayingjie198448/niuZip/core/engine"
)

func TestEngineBroker_RouteByExtension(t *testing.T) {
	broker := engine.NewBroker()

	testCases := []struct {
		filename    string
		expectedFmt string
	}{
		{"archive.zip", "zip"},
		{"archive.ZIP", "zip"},
		{"files.tar", "tar"},
		{"files.tar.gz", "tar.gz"},
		{"files.tgz", "tar.gz"},
		{"files.tar.bz2", "tar.bz2"},
		{"files.tbz", "tar.bz2"},
		{"package.7z", "7z"},
		{"package.rar", "rar"},
		{"disk.iso", "iso"},
		{"installer.dmg", "dmg"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			format := broker.DetectFormat(tc.filename)
			if format != tc.expectedFmt {
				t.Errorf("DetectFormat(%q) = %q; want %q", tc.filename, format, tc.expectedFmt)
			}
		})
	}
}

func TestEngineBroker_RouteByMagicBytes(t *testing.T) {
	tempDir := t.TempDir()
	broker := engine.NewBroker()

	// 1. Create file with PK.. (ZIP magic) but named without extension
	zipMagicFile := filepath.Join(tempDir, "unknown_zip")
	zipMagicData := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(zipMagicFile, zipMagicData, 0644); err != nil {
		t.Fatal(err)
	}

	fmt := broker.DetectFormatWithFile(zipMagicFile)
	if fmt != "zip" {
		t.Errorf("expected 'zip' from magic bytes, got %q", fmt)
	}

	// 2. Create file with 7z magic (37 7A BC AF 27 1C)
	sevenZipFile := filepath.Join(tempDir, "unknown_7z")
	sevenZipData := []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}
	if err := os.WriteFile(sevenZipFile, sevenZipData, 0644); err != nil {
		t.Fatal(err)
	}

	fmt7z := broker.DetectFormatWithFile(sevenZipFile)
	if fmt7z != "7z" {
		t.Errorf("expected '7z' from magic bytes, got %q", fmt7z)
	}

	// 3. Create file with RAR magic (52 61 72 21 1A 07)
	rarFile := filepath.Join(tempDir, "unknown_rar")
	rarData := []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}
	if err := os.WriteFile(rarFile, rarData, 0644); err != nil {
		t.Fatal(err)
	}

	fmtRar := broker.DetectFormatWithFile(rarFile)
	if fmtRar != "rar" {
		t.Errorf("expected 'rar' from magic bytes, got %q", fmtRar)
	}
}

func TestEngineBroker_Dispatch(t *testing.T) {
	broker := engine.NewBroker()
	eng, err := broker.GetEngineForFile("test.zip")
	if err != nil {
		t.Fatalf("expected engine for zip, got error: %v", err)
	}
	if eng == nil {
		t.Fatal("expected non-nil engine for zip")
	}

	_, err = broker.GetEngineForFile("unsupported.xyz123")
	if err == nil {
		t.Error("expected error for unsupported extension, got nil")
	}
}

func TestEngineBroker_ListDelegation(t *testing.T) {
	tempDir, files := createTestFiles(t)
	zipPath := filepath.Join(tempDir, "delegated.zip")

	zipEng := engine.NewPureZipEngine()
	ctx := context.Background()

	opts := engine.CompressOptions{
		SourcePaths: files[:1],
		DestPath:    zipPath,
	}
	if err := zipEng.Create(ctx, opts, nil); err != nil {
		t.Fatal(err)
	}

	broker := engine.NewBroker()
	entries, err := broker.List(ctx, zipPath)
	if err != nil {
		t.Fatalf("Broker.List failed: %v", err)
	}

	if len(entries) != 1 || entries[0].Name != "hello.txt" {
		t.Errorf("Broker.List unexpected entries: %+v", entries)
	}
}
