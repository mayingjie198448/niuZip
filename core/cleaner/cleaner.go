package cleaner

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ErrInsecurePath indicates that the file path violates containment (Zip Slip).
	ErrInsecurePath = errors.New("insecure file path: possible directory traversal attack")

	// Regex for Windows forbidden path characters: \/:*?"<>|
	// Note: We handle directory separators '/' and '\' explicitly.
	windowsForbiddenChars = regexp.MustCompile(`[:*?"<>|]`)
)

// IsMacMetadata checks if a relative path inside an archive belongs to macOS
// metadata artifacts that typically pollute Windows or Linux archives:
// - .DS_Store
// - AppleDouble resource forks (prefixed with "._")
// - __MACOSX resource directories
func IsMacMetadata(relPath string) bool {
	normalized := filepath.ToSlash(filepath.Clean(relPath))

	// Check for __MACOSX directory
	parts := strings.Split(normalized, "/")
	for _, part := range parts {
		if part == "__MACOSX" {
			return true
		}
		if part == ".DS_Store" {
			return true
		}
		if strings.HasPrefix(part, "._") {
			return true
		}
	}

	return false
}

// SanitizeForWindows cleans forbidden characters in relative file paths so they
// can be safely extracted or read on Windows file systems.
func SanitizeForWindows(relPath string) string {
	// Standardize to forward slash first
	normalized := strings.ReplaceAll(relPath, `\`, "/")
	parts := strings.Split(normalized, "/")

	sanitizedParts := make([]string, len(parts))
	for i, part := range parts {
		sanitizedParts[i] = windowsForbiddenChars.ReplaceAllString(part, "_")
	}

	return strings.Join(sanitizedParts, "/")
}

// ValidateSafePath verifies that extracting relPath under destDir will not
// escape destDir (Zip Slip vulnerability protection).
func ValidateSafePath(destDir, relPath string) (string, error) {
	if filepath.IsAbs(relPath) || strings.HasPrefix(relPath, "/") || strings.HasPrefix(relPath, `\`) {
		return "", ErrInsecurePath
	}

	cleanDest := filepath.Clean(destDir)
	targetPath := filepath.Join(cleanDest, relPath)

	// Ensure targetPath is strictly inside cleanDest
	rel, err := filepath.Rel(cleanDest, targetPath)
	if err != nil {
		return "", err
	}

	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrInsecurePath
	}

	return targetPath, nil
}
