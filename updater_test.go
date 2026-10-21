package main

import (
	"runtime"
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current  string
		remote   string
		expected bool
	}{
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "v1.0.1", true},
		{"1.0.0", "1.1.0", true},
		{"1.0.0", "2.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.1", "1.0.0", false},
		{"2.0.0", "1.9.9", false},
		{"1.0.0", "1.0.0-beta", false},
	}

	for _, tc := range tests {
		got := isNewerVersion(tc.current, tc.remote)
		if got != tc.expected {
			t.Errorf("isNewerVersion(%q, %q) = %v; want %v", tc.current, tc.remote, got, tc.expected)
		}
	}
}

func TestFindMatchingAsset(t *testing.T) {
	assets := []GitHubAsset{
		{Name: "niuZip-Setup-x64.exe", BrowserDownloadURL: "https://example.com/setup.exe"},
		{Name: "niuZip-arm64.dmg", BrowserDownloadURL: "https://example.com/arm64.dmg"},
		{Name: "niuZip-darwin-arm64", BrowserDownloadURL: "https://example.com/bin"},
	}

	asset := findMatchingAsset(assets)
	if asset == nil {
		t.Fatal("expected matching asset, got nil")
	}
	if runtime.GOOS == "darwin" && asset.Name != "niuZip-arm64.dmg" {
		t.Fatalf("expected niuZip-arm64.dmg on darwin, got %s", asset.Name)
	}
}
