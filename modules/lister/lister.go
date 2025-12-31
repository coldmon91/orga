package lister

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListFilesBySize walks the directory tree rooted at root,
// and prints them in a tree structure, sorted by size (descending) at each level.
func ListFilesBySize(root string, showHidden bool) error {
	// Print root
	fmt.Println(root)
	return visitDir(root, "", showHidden)
}

func visitDir(path string, prefix string, showHidden bool) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil // Skip unreadable directories
	}

	// Filter files
	var visible []entryInfo
	for _, e := range entries {
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		visible = append(visible, entryInfo{entry: e, size: info.Size()})
	}

	// Sort by size descending
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].size > visible[j].size
	})

	for i, item := range visible {
		isLast := i == len(visible)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		// Print current item
		fmt.Printf("%s%s[%s] %s\n", prefix, connector, formatSize(item.size), item.entry.Name())

		// Recurse if directory
		if item.entry.IsDir() {
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			visitDir(filepath.Join(path, item.entry.Name()), newPrefix, showHidden)
		}
	}
	return nil
}

type entryInfo struct {
	entry fs.DirEntry
	size  int64
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
