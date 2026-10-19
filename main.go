package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"orga/modules/lister"
)

var (
	targetDir  string
	outputDir  string
	backend    string
	modelName  string
	listMode   bool
	showHidden bool
	useTree    bool
	recursive  bool
	verbose    bool
)

func main() {
	// Parse arguments
	flag.StringVar(&targetDir, "target", ".", "Target directory to scan")
	flag.StringVar(&outputDir, "output", "", "Output directory for organized files (default: {target}/orga)")
	flag.StringVar(&backend, "backend", "ollama", "Classification backend: ollama (local) or jev (TypeSafe API)")
	flag.StringVar(&modelName, "model", "", "Model name (default: "+ollamaDefaultModel+" for ollama, "+jevDefaultModel+" for jev)")
	flag.BoolVar(&listMode, "list", false, "List files in target directory sorted by size")
	flag.BoolVar(&showHidden, "H", false, "Show hidden files in list mode")
	flag.BoolVar(&useTree, "T", false, "Show file list as a tree structure")
	flag.BoolVar(&recursive, "R", false, "Recursively scan subdirectories")
	flag.BoolVar(&verbose, "v", false, "Show verbose debugging logs")
	flag.Parse()

	// Validate directories
	var err error
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		fmt.Printf("Error resolving target directory: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("[DEBUG] Target directory resolved to: %s\n", targetDir)
	}

	if listMode {
		fmt.Printf("Listing files in %s sorted by size...\n", targetDir)
		if err := lister.ListFilesBySize(targetDir, showHidden, useTree); err != nil {
			fmt.Printf("Error listing files: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if outputDir == "" {
		outputDir = filepath.Join(targetDir, "orga")
	} else {
		outputDir, err = filepath.Abs(outputDir)
		if err != nil {
			fmt.Printf("Error resolving output directory: %v\n", err)
			os.Exit(1)
		}
	}

	if verbose {
		fmt.Printf("[DEBUG] Output directory resolved to: %s\n", outputDir)
	}

	classifier, modelName, err = newClassifier(backend, modelName)
	if err != nil {
		fmt.Printf("Error setting up classifier: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Target Directory: %s\n", targetDir)
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Printf("Backend: %s\n", backend)
	fmt.Printf("Model: %s\n", modelName)

	// Category directories are created on demand, so only check for a conflicting path here
	if err := validateOutputDir(); err != nil {
		fmt.Printf("Error checking output directory: %v\n", err)
		os.Exit(1)
	}

	// Find files using fd
	files, err := findEntries("f")
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d files.\n", len(files))

	// Recursive mode organizes files individually, so subdirectories are moved only in flat mode
	var dirs []string
	if !recursive {
		dirs, err = findEntries("d")
		if err != nil {
			fmt.Printf("Error finding directories: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d directories.\n", len(dirs))
	}

	// Process files
	for _, file := range files {
		// Skip files that are already inside one of our category directories
		isAlreadyOrganized := false
		for _, cat := range categories {
			catDir := filepath.Join(outputDir, cat)
			rel, err := filepath.Rel(catDir, file)
			if err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
				isAlreadyOrganized = true
				break
			}
		}

		if isAlreadyOrganized {
			if verbose {
				fmt.Printf("[DEBUG] Skipping already organized file: %s\n", file)
			}
			continue
		}
		
		// Skip the executable itself
		if filepath.Base(file) == os.Args[0] {
			if verbose {
				fmt.Printf("[DEBUG] Skipping current executable: %s\n", file)
			}
			continue
		}

		processEntry(file, false)
	}

	for _, dir := range dirs {
		// Moving a directory that holds the output directory would relocate the output itself
		if rel, err := filepath.Rel(dir, outputDir); err == nil && !strings.HasPrefix(rel, "..") {
			if verbose {
				fmt.Printf("[DEBUG] Skipping directory containing output: %s\n", dir)
			}
			continue
		}

		processEntry(dir, true)
	}
}

// validateOutputDir fails early when outputDir exists but is not a directory.
func validateOutputDir() error {
	info, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", outputDir)
	}
	return nil
}

// findEntries lists entries of entryType ("f" for files, "d" for directories) using fd.
// Directories are always limited to the top level since they are moved as a whole.
func findEntries(entryType string) ([]string, error) {
	// Check if fd is installed
	_, err := exec.LookPath("fd")
	if err != nil {
		return nil, fmt.Errorf("fd command not found. Please install fd-find")
	}

	cmd := exec.Command("fd", "--type", entryType, "--absolute-path", ".", targetDir)

	if !recursive || entryType == "d" {
		cmd.Args = append(cmd.Args, "--max-depth", "1")
	}

	// Exclude the output directory to prevent loops
	relOutput, err := filepath.Rel(targetDir, outputDir)
	if err == nil && !strings.HasPrefix(relOutput, "..") {
		// outputDir is inside targetDir
		cmd.Args = append(cmd.Args, "--exclude", relOutput)
	}

	if verbose {
		fmt.Printf("[DEBUG] Running command: %s %s\n", cmd.Path, strings.Join(cmd.Args[1:], " "))
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			// fd appends a trailing separator to directories
			result = append(result, filepath.Clean(line))
		}
	}
	return result, nil
}

func processEntry(path string, isDir bool) {
	filename := filepath.Base(path)
	fmt.Printf("Processing: %s ... ", filename)

	var category string
	var err error
	if isDir {
		category, err = classifyDir(path)
	} else {
		category, err = classifyFile(filename)
	}
	if err != nil {
		fmt.Printf("Error classifying: %v\n", err)
		return
	}

	if !isValidCategory(category) {
		if verbose {
			fmt.Printf("[DEBUG] Invalid category received: %q. Defaulting to 'others'.\n", category)
		}
		category = "others"
	}

	// Move file
	destDir := filepath.Join(outputDir, category)
	destPath := filepath.Join(destDir, filename)

	if verbose {
		fmt.Printf("[DEBUG] Ensuring directory exists: %s\n", destDir)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Handle duplicate filenames
	destPath = ensureUniquePath(destPath)

	if verbose && destPath != filepath.Join(destDir, filename) {
		fmt.Printf("[DEBUG] Filename conflict. Renamed to: %s\n", filepath.Base(destPath))
	}

	err = os.Rename(path, destPath)
	if err != nil {
		fmt.Printf("Error moving file: %v\n", err)
	} else {
		fmt.Printf("-> %s/%s\n", category, filepath.Base(destPath))
	}
}

func ensureUniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	if verbose {
		fmt.Printf("[DEBUG] File already exists at destination: %s\n", path)
	}

	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	for i := 1; ; i++ {
		newPath := filepath.Join(dir, fmt.Sprintf("%s_%d%s", name, i, ext))
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
