package vfs_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mayingjie198448/niuZip/core/vfs"
)

func TestBuildTree(t *testing.T) {
	entries := []vfs.FileEntry{
		{
			Path:    "root.txt",
			Name:    "root.txt",
			IsDir:   false,
			Size:    100,
			ModTime: time.Now(),
		},
		{
			Path:    "docs/readme.md",
			Name:    "readme.md",
			IsDir:   false,
			Size:    200,
			ModTime: time.Now(),
		},
		{
			Path:    "docs/images/logo.png",
			Name:    "logo.png",
			IsDir:   false,
			Size:    3000,
			ModTime: time.Now(),
		},
		{
			Path:    "empty_folder/",
			Name:    "empty_folder",
			IsDir:   true,
			Size:    0,
			ModTime: time.Now(),
		},
	}

	root := vfs.BuildTree(entries)
	if root == nil {
		t.Fatal("BuildTree returned nil")
	}

	if !root.IsDir {
		t.Error("root node must be a directory")
	}

	// Root should have 3 direct children: docs, empty_folder, root.txt
	if len(root.Children) != 3 {
		t.Fatalf("expected 3 children in root, got %d", len(root.Children))
	}

	// Verify sorting: directories first, then alphabetical files
	// Expected order: docs, empty_folder, root.txt
	if root.Children[0].Name != "docs" || !root.Children[0].IsDir {
		t.Errorf("expected first child 'docs' (dir), got %q", root.Children[0].Name)
	}
	if root.Children[1].Name != "empty_folder" || !root.Children[1].IsDir {
		t.Errorf("expected second child 'empty_folder' (dir), got %q", root.Children[1].Name)
	}
	if root.Children[2].Name != "root.txt" || root.Children[2].IsDir {
		t.Errorf("expected third child 'root.txt' (file), got %q", root.Children[2].Name)
	}

	// Verify docs children: images (dir), readme.md (file)
	docsNode := root.Children[0]
	if len(docsNode.Children) != 2 {
		t.Fatalf("expected 2 children in docs, got %d", len(docsNode.Children))
	}
	if docsNode.Children[0].Name != "images" || !docsNode.Children[0].IsDir {
		t.Errorf("expected docs child 'images' (dir), got %q", docsNode.Children[0].Name)
	}
	if docsNode.Children[1].Name != "readme.md" || docsNode.Children[1].IsDir {
		t.Errorf("expected docs child 'readme.md' (file), got %q", docsNode.Children[1].Name)
	}
}

func TestSearchEntries(t *testing.T) {
	entries := []vfs.FileEntry{
		{Path: "src/main.go", Name: "main.go", IsDir: false},
		{Path: "src/utils/helper.go", Name: "helper.go", IsDir: false},
		{Path: "README.md", Name: "README.md", IsDir: false},
		{Path: "docs/设计方案.pdf", Name: "设计方案.pdf", IsDir: false},
		{Path: "docs/架构图.PNG", Name: "架构图.PNG", IsDir: false},
	}

	t.Run("Match extension case-insensitive", func(t *testing.T) {
		res := vfs.Search(entries, "png")
		if len(res) != 1 || res[0].Name != "架构图.PNG" {
			t.Errorf("expected 1 match for 'png', got %v", res)
		}
	})

	t.Run("Match chinese filename", func(t *testing.T) {
		res := vfs.Search(entries, "设计")
		if len(res) != 1 || res[0].Name != "设计方案.pdf" {
			t.Errorf("expected 1 match for '设计', got %v", res)
		}
	})

	t.Run("Match path fragment", func(t *testing.T) {
		res := vfs.Search(entries, "utils")
		if len(res) != 1 || res[0].Path != "src/utils/helper.go" {
			t.Errorf("expected 1 match for 'utils', got %v", res)
		}
	})

	t.Run("Empty query", func(t *testing.T) {
		res := vfs.Search(entries, "")
		if len(res) != len(entries) {
			t.Errorf("empty query should return all %d entries, got %d", len(entries), len(res))
		}
	})
}

func BenchmarkBuildTree_10K(b *testing.B) {
	entries := make([]vfs.FileEntry, 10000)
	for i := 0; i < 10000; i++ {
		entries[i] = vfs.FileEntry{
			Path:    fmt.Sprintf("folder_%d/sub_%d/file_%d.txt", i%100, i%10, i),
			Name:    fmt.Sprintf("file_%d.txt", i),
			IsDir:   false,
			Size:    int64(i * 10),
			ModTime: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vfs.BuildTree(entries)
	}
}

func BenchmarkSearch_10K(b *testing.B) {
	entries := make([]vfs.FileEntry, 10000)
	for i := 0; i < 10000; i++ {
		entries[i] = vfs.FileEntry{
			Path: fmt.Sprintf("folder_%d/sub_%d/file_%d.txt", i%100, i%10, i),
			Name: fmt.Sprintf("file_%d.txt", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vfs.Search(entries, "file_999")
	}
}
