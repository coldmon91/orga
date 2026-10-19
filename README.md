# Orga

[한국어 (Korean)](README.kr.md)

## Overview

- A file management tool using local LLMs.
- Classifies files based on their names to determine their type.
- Moves classified files to designated directories.
- Top-level subdirectories are classified by name and moved as a whole; when the name is inconclusive, a sample of up to 20 files inside (2 levels deep) is used. With `-R`, files are organized individually instead.

## Variables

- `target_dir`: The directory path to scan for files (Default: current directory).
- `output_dir`: The directory path where organized files will be moved (Default: `{target_dir}/orga`).

## LLM Configuration

- File Search Tool: Uses the `fd` command.
- String Search Tool: Uses the `rg` command.
- Backends: `ollama` (default, local) or `jev` (TypeSafe API, sends file and directory names to an external service).
- Default Model: `gemma3n:e4b` for Ollama, `jev-latest` for Jev.
- Jev requires `TYPESAFE_API_KEY`, read from the environment or a `.env` file in the working directory or next to the executable.

## File Categories

- Documents (e.g., .docx, .pdf, .txt): `{output_dir}/documents`
- Images: `{output_dir}/images`
- Videos: `{output_dir}/videos`
- Audio: `{output_dir}/audio`
- Archives: `{output_dir}/archives`
- Executables: `{output_dir}/executables`
- Installers: `{output_dir}/installers`
- Others: `{output_dir}/others`

## Basic Operation

1. Scans all files within the `{target_dir}`.
2. Sends each file name to the LLM to identify the appropriate category.
3. Moves the file to the corresponding folder under `{output_dir}`.

## Installation

Ensure you have [Go](https://go.dev/), [fd](https://github.com/sharkdp/fd), and [Ollama](https://ollama.com/) installed.

### Prerequisites

#### Go
Download and install from [go.dev/doc/install](https://go.dev/doc/install).

#### fd (fd-find)
- **macOS**: `brew install fd`
- **Ubuntu/Debian**: `sudo apt install fd-find` (Note: executable may be named `fdfind`, the tool expects `fd`)
- **Windows**: `winget install sharkdp.fd`

#### Ollama
Download and install from [ollama.com](https://ollama.com/). After installation, pull the required model:
```bash
ollama pull gemma3n:e4b
```

### Build the project

```bash
go build -o orga
```

# Usage

```bash
# Organize current directory using default model
./orga

# Specify target directory and model
./orga -target ~/Downloads -model gemma3:4b

# Classify with TypeSafe Jev (requires TYPESAFE_API_KEY)
./orga -backend jev -target ~/Downloads

# List files in target directory sorted by size (absolute paths)
./orga -list -target ~/Downloads

# List files in a tree structure
./orga -list -T -target ~/Downloads

# List files including hidden files
./orga -list -H
```

## Arguments

- `-target`: Directory to scan (default ".")
- `-output`: Output directory (default "{target}/orga")
- `-backend`: Classification backend, `ollama` or `jev` (default "ollama")
- `-model`: Model name (default "gemma3n:e4b" for ollama, "jev-latest" for jev)
- `-R`: Recursively scan subdirectories (default false)
- `-list`: List files in target directory sorted by size
- `-T`: Show file list as a tree structure (use with -list)
- `-H`: Show hidden files in list mode
