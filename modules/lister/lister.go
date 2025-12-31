package lister

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type FileInfo struct {
	Path string
	Size int64
}

// ListFilesBySize lists files. If useTree is true, it prints a tree structure.
// Otherwise, it prints a flat list with absolute paths.
// Files are sorted by size (descending).
func ListFilesBySize(root string, showHidden bool, useTree bool) error {
	if useTree {
		fmt.Println(root)
		return visitDir(root, "", showHidden)
	}
	return listFlat(root, showHidden)
}

func listFlat(root string, showHidden bool) error {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		
		if !showHidden && strings.HasPrefix(d.Name(), ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		var size int64
		if d.IsDir() {
			s, err := getDirSize(path)
			if err != nil {
				size = 0
			} else {
				size = s
			}
		} else {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			size = info.Size()
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		files = append(files, FileInfo{Path: absPath, Size: size})
		return nil
	})
	if err != nil {
		return err
	}

	// Sort descending
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	// Print
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SIZE\tPATH")
	for _, f := range files {
		fmt.Fprintf(w, "%s\t%s\n", formatSize(f.Size), f.Path)
	}
	w.Flush()
	return nil
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
		
		var size int64
		fullPath := filepath.Join(path, e.Name())

		if e.IsDir() {
			s, err := getDirSize(fullPath)
			if err != nil {
				size = 0
			} else {
				size = s
			}
		} else {
			info, err := e.Info()
			if err != nil {
				continue
			}
			size = info.Size()
		}
		
		visible = append(visible, entryInfo{entry: e, size: size})
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

// getDirSize returns the size of a directory in bytes using du -sk
func getDirSize(path string) (int64, error) {
	cmd := exec.Command("du", "-sk", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	// Output format: "size_in_kbytes\tpath\n"
	fields := strings.Fields(string(out))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected output from du")
	}
	
	kb, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, err
	}
	
	return kb * 1024, nil
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