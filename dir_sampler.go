package main

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

const (
	dirSampleLimit    = 20
	dirSampleMaxDepth = 2
)

var errSampleFull = errors.New("sample limit reached")

// sampleDirFiles returns up to limit visible file paths under dir, relative to dir,
// walking at most dirSampleMaxDepth levels so large trees stay cheap to inspect.
func sampleDirFiles(dir string, limit int) ([]string, error) {
	var samples []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == dir {
				return err
			}
			return nil // Skip unreadable entries
		}
		if path == dir {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if strings.Count(rel, string(filepath.Separator))+1 >= dirSampleMaxDepth {
				return filepath.SkipDir
			}
			return nil
		}

		samples = append(samples, rel)
		if len(samples) >= limit {
			return errSampleFull
		}
		return nil
	})
	if err != nil && !errors.Is(err, errSampleFull) {
		return nil, fmt.Errorf("sample directory %s: %w", dir, err)
	}
	return samples, nil
}
