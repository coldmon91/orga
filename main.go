package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"orga/modules/lister"
)

// Configuration
const (
	OllamaURL = "http://localhost:11434/api/generate"
)

var (
	targetDir  string
	outputDir  string
	modelName  string
	listMode   bool
	showHidden bool
	useTree    bool
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

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func main() {
	// Parse arguments
	flag.StringVar(&targetDir, "target", ".", "Target directory to scan")
	flag.StringVar(&outputDir, "output", "", "Output directory for organized files (default: {target}/organized)")
	flag.StringVar(&modelName, "model", "gemma3n:e4b", "Ollama model name")
	flag.BoolVar(&listMode, "list", false, "List files in target directory sorted by size")
	flag.BoolVar(&showHidden, "H", false, "Show hidden files in list mode")
	flag.BoolVar(&useTree, "T", false, "Show file list as a tree structure")
	flag.Parse()

	// Validate directories
	var err error
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		fmt.Printf("Error resolving target directory: %v\n", err)
		os.Exit(1)
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
		outputDir = filepath.Join(targetDir, "organized")
	} else {
		outputDir, err = filepath.Abs(outputDir)
		if err != nil {
			fmt.Printf("Error resolving output directory: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Target Directory: %s\n", targetDir)
	fmt.Printf("Output Directory: %s\n", outputDir)
	fmt.Printf("Model: %s\n", modelName)

	// Create output directories
	if err := createDirectories(); err != nil {
		fmt.Printf("Error creating directories: %v\n", err)
		os.Exit(1)
	}

	// Find files using fd
	files, err := findFiles()
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d files.\n", len(files))

	// Process files
	for _, file := range files {
		// Skip files inside the output directory to avoid re-processing
		if strings.HasPrefix(file, outputDir) {
			continue
		}
		
		// Skip the executable itself if it's in the list (unlikely with fd defaults but possible)
		if filepath.Base(file) == os.Args[0] {
			continue
		}

		processFile(file)
	}
}

func createDirectories() error {
	for _, cat := range categories {
		path := filepath.Join(outputDir, cat)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}

func findFiles() ([]string, error) {
	// fd command to find files. 
	// -t f: files only
	// --absolute-path: return absolute paths
	// We scan the targetDir
	
	// Check if fd is installed
	_, err := exec.LookPath("fd")
	if err != nil {
		return nil, fmt.Errorf("fd command not found. Please install fd-find")
	}

	cmd := exec.Command("fd", "--type", "f", "--absolute-path", ".", targetDir)
	// Exclude the output directory to prevent loops
	// fd accepts patterns to exclude. We exclude the relative path of outputDir from targetDir if possible, 
	// or just the full path name if valid. 
	// Simpler is to use --exclude pattern.
	
	relOutput, err := filepath.Rel(targetDir, outputDir)
	if err == nil && !strings.HasPrefix(relOutput, "..") {
		// outputDir is inside targetDir
		cmd.Args = append(cmd.Args, "--exclude", relOutput)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func processFile(path string) {
	filename := filepath.Base(path)
	fmt.Printf("Processing: %s ... ", filename)

	category, err := classifyFile(filename)
	if err != nil {
		fmt.Printf("Error classifying: %v\n", err)
		return
	}

	// Validate category
	valid := false
	for _, c := range categories {
		if c == category {
			valid = true
			break
		}
	}
	if !valid {
		category = "others"
	}

	// Move file
	destDir := filepath.Join(outputDir, category)
	destPath := filepath.Join(destDir, filename)

	// Handle duplicate filenames
	destPath = ensureUniquePath(destPath)

	err = os.Rename(path, destPath)
	if err != nil {
		fmt.Printf("Error moving file: %v\n", err)
	} else {
		fmt.Printf("-> %s/%s\n", category, filepath.Base(destPath))
	}
}

func classifyFile(filename string) (string, error) {
	prompt := fmt.Sprintf(`You are a file organization assistant.
Classify the following file into exactly one of these categories:
- documents
- images
- videos
- audio
- archives
- executables
- installers
- others

Filename: "%s"

Rules:
1. Reply ONLY with the category name in lowercase.
2. Do not add any punctuation or explanation.
3. If unsure, use "others".

	Category:`, filename)

	reqBody := OllamaRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: false,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(OllamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API returned status: %s", resp.Status)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.ToLower(ollamaResp.Response)), nil
}

func ensureUniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
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
