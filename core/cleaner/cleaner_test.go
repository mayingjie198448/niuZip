package cleaner_test

import (
	"testing"

	"github.com/mayingjie198448/niuZip/core/cleaner"
)

func TestIsMacMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{"DS_Store root", ".DS_Store", true},
		{"DS_Store in dir", "folder/.DS_Store", true},
		{"DS_Store deeply nested", "a/b/c/.DS_Store", true},
		{"AppleDouble dot-underscore", "._test.txt", true},
		{"AppleDouble in dir", "folder/._photo.png", true},
		{"MACOSX directory prefix", "__MACOSX/image.png", true},
		{"MACOSX deep", "a/__MACOSX/sub/file", true},
		{"Normal hidden file", ".gitignore", false},
		{"Normal dot file", ".env", false},
		{"Normal file", "document.pdf", false},
		{"File containing DS_Store in name", "my_DS_Store_notes.txt", false},
		{"Normal path with underscore", "foo_bar/test.txt", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := cleaner.IsMacMetadata(tc.path)
			if actual != tc.expected {
				t.Errorf("IsMacMetadata(%q) = %v; want %v", tc.path, actual, tc.expected)
			}
		})
	}
}

func TestSanitizeForWindows(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal path", "folder/subfolder/file.txt", "folder/subfolder/file.txt"},
		{"Contains colon", "report:2026.pdf", "report_2026.pdf"},
		{"Contains asterisk and question", "what*is?this.txt", "what_is_this.txt"},
		{"Contains quotes and pipe", `a|"b"|c.txt`, "a__b__c.txt"},
		{"Contains less and greater", "less<greater>.txt", "less_greater_.txt"},
		{"Mixed deep path", "dir:name/bad*file?.log", "dir_name/bad_file_.log"},
		{"Clean backslashes to forward slash", `dir\sub\file.txt`, "dir/sub/file.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := cleaner.SanitizeForWindows(tc.input)
			if actual != tc.expected {
				t.Errorf("SanitizeForWindows(%q) = %q; want %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestValidateSafePath(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		destDir     string
		expectError bool
	}{
		{"Safe relative path", "file.txt", "/tmp/dest", false},
		{"Safe nested path", "sub/folder/file.txt", "/tmp/dest", false},
		{"Dot slash path", "./sub/file.txt", "/tmp/dest", false},
		{"ZipSlip parent traversal", "../evil.sh", "/tmp/dest", true},
		{"ZipSlip deep traversal", "sub/../../evil.sh", "/tmp/dest", true},
		{"ZipSlip absolute path escape", "/etc/passwd", "/tmp/dest", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			safePath, err := cleaner.ValidateSafePath(tc.destDir, tc.path)
			if tc.expectError {
				if err == nil {
					t.Errorf("ValidateSafePath(%q, %q) expected error, got nil (result: %q)", tc.destDir, tc.path, safePath)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSafePath(%q, %q) unexpected error: %v", tc.destDir, tc.path, err)
				}
			}
		})
	}
}
