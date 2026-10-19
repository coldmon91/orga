package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Categories
var categories = []string{
	"documents",
	"images",
	"videos",
	"audio",
	"archives",
	"executables",
	"installers",
	"others",
}

// classifyItem describes one entry to classify. SampleFiles is set only when a
// directory name alone was inconclusive.
type classifyItem struct {
	Kind        string   `json:"kind"`
	Name        string   `json:"name"`
	SampleFiles []string `json:"sample_files,omitempty"`
}

// categoryClassifier maps an item to a category. conclusive reports whether the
// answer is reliable enough to skip looking inside a directory.
type categoryClassifier interface {
	classify(item classifyItem) (category string, conclusive bool, err error)
}

var classifier categoryClassifier

// newClassifier builds the classifier for backend; an empty model selects the backend default.
func newClassifier(backend, model string) (categoryClassifier, string, error) {
	switch backend {
	case "ollama":
		if model == "" {
			model = ollamaDefaultModel
		}
		return ollamaClassifier{model: model}, model, nil
	case "jev":
		if model == "" {
			model = jevDefaultModel
		}
		if err := loadAPIKeyFromDotEnv(); err != nil {
			return nil, "", err
		}
		c, err := newJevClassifier(model)
		return c, model, err
	default:
		return nil, "", fmt.Errorf("unknown backend %q (use ollama or jev)", backend)
	}
}

// loadAPIKeyFromDotEnv reads .env from the working directory and next to the executable.
func loadAPIKeyFromDotEnv() error {
	paths := []string{".env"}
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}
	for _, path := range paths {
		if err := loadDotEnv(path); err != nil {
			return err
		}
	}
	return nil
}

func isValidCategory(category string) bool {
	return slices.Contains(categories, category)
}

func classifyFile(filename string) (string, error) {
	category, _, err := classifier.classify(classifyItem{Kind: "file", Name: filename})
	return category, err
}

// classifyDir classifies a directory by its name first, and peeks at a sample of
// its files only when the name alone is not conclusive.
func classifyDir(path string) (string, error) {
	name := filepath.Base(path)
	category, conclusive, err := classifier.classify(classifyItem{Kind: "directory", Name: name})
	if err != nil || conclusive {
		return category, err
	}

	samples, err := sampleDirFiles(path, dirSampleLimit)
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Printf("[DEBUG] Name of %q is inconclusive; sampled %d files inside\n", name, len(samples))
	}
	if len(samples) == 0 {
		return "others", nil
	}

	category, _, err = classifier.classify(classifyItem{Kind: "directory", Name: name, SampleFiles: samples})
	return category, err
}
