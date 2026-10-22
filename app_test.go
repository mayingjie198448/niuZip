package main_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	niuzip "github.com/mayingjie198448/niuZip"
	"github.com/mayingjie198448/niuZip/core/engine"
)

func TestAppService_OpenAndSearchAndPreview(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "sample.zip")

	// Create test zip
	f1 := filepath.Join(tempDir, "notes.txt")
	os.WriteFile(f1, []byte("Important Notes 2026"), 0644)
	subDir := filepath.Join(tempDir, "images")
	os.MkdirAll(subDir, 0755)
	f2 := filepath.Join(subDir, "pic.png")
	os.WriteFile(f2, []byte("fake png bytes"), 0644)

	zipEng := engine.NewPureZipEngine()
	err := zipEng.Create(context.Background(), engine.CompressOptions{
		SourcePaths: []string{f1, f2},
		DestPath:    zipPath,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	app := niuzip.NewApp()

	// 1. OpenArchive
	summary, err := app.OpenArchive(zipPath)
	if err != nil {
		t.Fatalf("OpenArchive failed: %v", err)
	}

	if summary.Format != "zip" {
		t.Errorf("expected format 'zip', got %q", summary.Format)
	}
	if summary.TotalFiles != 2 {
		t.Errorf("expected 2 total files, got %d", summary.TotalFiles)
	}
	if summary.Tree == nil {
		t.Fatal("expected non-nil Tree")
	}

	// 2. SearchEntries
	searchResults := app.SearchEntries("notes")
	if len(searchResults) != 1 || searchResults[0].Name != "notes.txt" {
		t.Errorf("unexpected search results: %+v", searchResults)
	}

	// 3. PreviewEntry
	previewPath, err := app.PreviewEntry("notes.txt")
	if err != nil {
		t.Fatalf("PreviewEntry failed: %v", err)
	}
	data, err := os.ReadFile(previewPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Important Notes 2026" {
		t.Errorf("preview content mismatch: %q", string(data))
	}
}

func TestAppService_SmartFlow_AppDetection(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "app_package.zip")

	appBundleDir := filepath.Join(tempDir, "Sample.app", "Contents")
	os.MkdirAll(appBundleDir, 0755)
	os.WriteFile(filepath.Join(appBundleDir, "Info.plist"), []byte("<xml/>"), 0644)

	zipEng := engine.NewPureZipEngine()
	err := zipEng.Create(context.Background(), engine.CompressOptions{
		SourcePaths: []string{filepath.Join(tempDir, "Sample.app")},
		DestPath:    zipPath,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	app := niuzip.NewApp()
	summary, err := app.OpenArchive(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	flows := app.DetectFlows(summary)
	if len(flows) == 0 {
		t.Fatal("expected at least one flow action for .app, got 0")
	}

	if flows[0].Type != "install_app" {
		t.Errorf("expected flow type 'install_app', got %q", flows[0].Type)
	}
}

func TestAppService_SmartFlow_FontDetection(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "fonts.zip")

	font1 := filepath.Join(tempDir, "FiraCode.ttf")
	os.WriteFile(font1, []byte("font 1"), 0644)
	font2 := filepath.Join(tempDir, "Inter.otf")
	os.WriteFile(font2, []byte("font 2"), 0644)

	zipEng := engine.NewPureZipEngine()
	err := zipEng.Create(context.Background(), engine.CompressOptions{
		SourcePaths: []string{font1, font2},
		DestPath:    zipPath,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	app := niuzip.NewApp()
	summary, err := app.OpenArchive(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	flows := app.DetectFlows(summary)
	foundFontFlow := false
	for _, f := range flows {
		if f.Type == "install_font" {
			foundFontFlow = true
			break
		}
	}
	if !foundFontFlow {
		t.Error("expected 'install_font' flow action, but none found")
	}
}

func TestAppService_GetPathInfos(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "doc.txt")
	os.WriteFile(file1, []byte("content"), 0644)

	dir1 := filepath.Join(tempDir, "subfolder")
	os.MkdirAll(dir1, 0755)

	app := niuzip.NewApp()
	infos := app.GetPathInfos([]string{file1, dir1, filepath.Join(tempDir, "non_existent.file")})

	if len(infos) != 2 {
		t.Fatalf("expected 2 valid path infos, got %d", len(infos))
	}

	if infos[0].Name != "doc.txt" || infos[0].IsDir {
		t.Errorf("expected file doc.txt with IsDir=false, got %+v", infos[0])
	}

	if infos[1].Name != "subfolder" || !infos[1].IsDir {
		t.Errorf("expected dir subfolder with IsDir=true, got %+v", infos[1])
	}
}
