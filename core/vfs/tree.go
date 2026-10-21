package vfs

import (
	"path"
	"sort"
	"strings"
	"time"
)

// FileEntry represents a single archived item (file or directory).
type FileEntry struct {
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	IsDir       bool      `json:"is_dir"`
	Size        int64     `json:"size"`
	PackedSize  int64     `json:"packed_size"`
	ModTime     time.Time `json:"mod_time"`
	IsEncrypted bool      `json:"is_encrypted"`
	CRC32       uint32    `json:"crc32"`
}

// TreeNode represents a node in the hierarchical virtual file tree.
type TreeNode struct {
	FileEntry
	Children []*TreeNode `json:"children,omitempty"`
}

// BuildTree converts a flat list of FileEntries into a nested hierarchical tree.
// Root node is always a directory with Path="" and Name="".
func BuildTree(entries []FileEntry) *TreeNode {
	root := &TreeNode{
		FileEntry: FileEntry{
			Path:  "",
			Name:  "",
			IsDir: true,
		},
		Children: make([]*TreeNode, 0),
	}

	// nodeMap maps directory relative path to *TreeNode
	nodeMap := make(map[string]*TreeNode)
	nodeMap[""] = root

	// Helper to ensure a directory node exists
	var getOrCreateDir func(dirPath string) *TreeNode
	getOrCreateDir = func(dirPath string) *TreeNode {
		dirPath = strings.Trim(dirPath, "/")
		if dirPath == "" {
			return root
		}
		if node, exists := nodeMap[dirPath]; exists {
			return node
		}

		parentDir := path.Dir(dirPath)
		if parentDir == "." {
			parentDir = ""
		}
		parentNode := getOrCreateDir(parentDir)

		dirName := path.Base(dirPath)
		newNode := &TreeNode{
			FileEntry: FileEntry{
				Path:  dirPath + "/",
				Name:  dirName,
				IsDir: true,
			},
			Children: make([]*TreeNode, 0),
		}

		parentNode.Children = append(parentNode.Children, newNode)
		nodeMap[dirPath] = newNode
		return newNode
	}

	for _, entry := range entries {
		cleanPath := strings.Trim(entry.Path, "/")
		if cleanPath == "" {
			continue
		}

		if entry.IsDir {
			getOrCreateDir(cleanPath)
			continue
		}

		dirPart := path.Dir(cleanPath)
		if dirPart == "." {
			dirPart = ""
		}
		parent := getOrCreateDir(dirPart)

		fileNode := &TreeNode{
			FileEntry: entry,
		}
		// Ensure name is clean base name
		if fileNode.Name == "" {
			fileNode.Name = path.Base(cleanPath)
		}
		parent.Children = append(parent.Children, fileNode)
	}

	// Recursively sort all children: directories first, then alphabetical
	sortTree(root)

	return root
}

func sortTree(node *TreeNode) {
	if len(node.Children) == 0 {
		return
	}

	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]
		if a.IsDir != b.IsDir {
			return a.IsDir // directories first
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	for _, child := range node.Children {
		if child.IsDir {
			sortTree(child)
		}
	}
}

// Search performs case-insensitive fuzzy filtering on a flat list of FileEntries.
func Search(entries []FileEntry, query string) []FileEntry {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return entries
	}

	lowerQuery := strings.ToLower(trimmed)
	var matches []FileEntry

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerQuery) ||
			strings.Contains(strings.ToLower(entry.Path), lowerQuery) {
			matches = append(matches, entry)
		}
	}

	return matches
}
