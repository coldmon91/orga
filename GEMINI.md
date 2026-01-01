# Orga Project Context

## Project Overview

**Orga** is a CLI-based file management tool that leverages local Large Language Models (LLMs) via Ollama to intelligently organize files. It scans a target directory, uses an LLM to classify files based on their filenames, and moves them into categorized subdirectories (e.g., documents, images, videos).

### Key Features
*   **AI-Powered Classification:** Uses Ollama (defaulting to `gemma3n:e4b`) to determine file categories.
*   **File Organization:** Moves files into structured folders (`organized/documents`, `organized/images`, etc.).
*   **File Listing:** Provides a utility to list files sorted by size, either as a flat list or a tree structure.
*   **External Tools Integration:** Utilizes `fd` for fast file finding and `du` for directory size calculation (in list mode).

### Architecture
*   **Entry Point:** `main.go` parses CLI arguments, handles the main application loop, interacts with the Ollama API, and performs file operations (finding, classifying, moving).
*   **Lister Module:** `modules/lister/lister.go` implements the file listing logic (flat and tree view), using `du` for directory sizes.
*   **Dependencies:**
    *   **Go:** Main programming language (v1.24+).
    *   **Ollama:** Local LLM server (required for classification).
    *   **fd:** Fast file finding tool (required for finding files).
    *   **du:** Disk usage utility (required for list mode directory sizes).

## Building and Running

### Prerequisites
1.  **Go:** Install Go (1.24 or later).
2.  **Ollama:** Install and run Ollama. Pull the default model: `ollama pull gemma3n:e4b`
3.  **fd:** Install `fd` (or `fd-find` on Linux).
4.  **du:** Ensure `du` is available (standard on Unix-like systems; may need Git Bash or similar on Windows).

### Build Command
```bash
go build -o orga.exe .
```

### Usage Examples
```bash
# Organize current directory
./orga.exe

# Organize specific directory with a custom model
./orga.exe -target "C:\Downloads" -model "llama3"

# List files in target directory sorted by size
./orga.exe -list -target "C:\Work"

# List files as a tree structure including hidden files
./orga.exe -list -T -H -target "."
```

## Development Conventions

*   **Code Structure:**
    *   `main.go`: Contains the core business logic, including CLI flag parsing, HTTP client for Ollama, and file manipulation.
    *   `modules/`: Contains reusable sub-packages. Currently, `lister` is the only module.
*   **External Commands:** The application relies on `exec.Command` to interface with system tools like `fd` and `du`. Ensure these are handled gracefully (though currently `lister` assumes `du` availability).
*   **Error Handling:** Errors are generally printed to `stdout` with `fmt.Printf` followed by `os.Exit(1)` for critical failures, or logged as warnings for individual file processing errors.
*   **Configuration:** Configuration is primarily done via CLI flags.
*   **API Interaction:** Ollama interaction is done via raw HTTP POST requests using `net/http` and standard `encoding/json` for serialization.
