# Orga

[한국어 버전 (Korean)](README.kr.md)

## Overview

- A file management tool using local LLMs.
- Classifies files based on their names to determine their type.
- Moves classified files to designated directories.

## Variables

- `target_dir`: The directory path to scan for files (Default: current directory).
- `output_dir`: The directory path where organized files will be moved (Default: `{target_dir}/organized`).

## LLM Configuration

- File Search Tool: Uses the `fd` command.
- String Search Tool: Uses the `rg` command.
- Default Model: `gemma3n:e4b` (via Ollama).

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

```bash
# Build the project
go build -o orga
```

## Usage

```bash
# Organize current directory using default model
./orga

# Specify target directory and model
./orga -target ~/Downloads -model gemma3:4b
```

## Arguments

- `-target`: Directory to scan (default ".")
- `-output`: Output directory (default "{target}/organized")
- `-model`: Ollama model name (default "gemma3n:e4b")
