# Peeker

A powerful CLI tool for analyzing code files and calculating token usage across multiple tokenizer models. Perfect for preparing code for LLM context windows or analyzing repository token usage patterns.

## Features

- **Multiple Tokenizer Support**
  - GPT-3.5 Turbo
  - GPT-4
  - Claude
  - Custom HuggingFace tokenizers
- **Interactive File Selection**
  - Visual file picker interface
  - Search and filter capabilities
  - Multiple file selection support
- **Smart File Filtering**
  - Customizable include/exclude patterns
  - Default exclusions for common binary and build files
  - Binary file detection
  - Hidden file handling
- **Flexible Output Options**
  - Tree view of file structure
  - Detailed file contents with token counts
  - Clipboard integration
  - Token usage warnings and summaries

## Installation

```bash
go get github.com/yourusername/peeker
```

## Usage

Basic usage:

```bash
peeker --path /path/to/project
```

### Common Options

```bash
# Analyze with specific tokenizer
peeker --path . --tokenizer gpt-3.5-turbo

# Interactive mode with file picker
peeker --path . -i

# Include/exclude specific patterns
peeker --path . --include "*.go,*.py" --exclude "test/*"

# Copy output to clipboard
peeker --path . -c

# Specify token limit
peeker --path . --token-limit 8192
```

### Advanced Options

```bash
  --path string          Directory to analyze (default ".")
  --include string       Patterns to include (comma-separated)
  --exclude string       Patterns to exclude (comma-separated)
  --max-size int        Maximum file size in bytes (default 10MB)
  --max-depth int       Maximum directory depth (default 20)
  --output string       Output format: tree, files, or both (default "both")
  --threads int         Number of threads for parallel processing
  --hidden              Show hidden files and directories
  -c                    Copy output to clipboard
  -i                    Interactive mode
  --tokenizer string    Tokenizer type (gpt-3.5-turbo, gpt-4, claude, huggingface)
  --tokenizer-model     Path to HuggingFace tokenizer model
  --token-limit int     Maximum token limit (default 4096)
```

## Interactive Mode Controls

When using the `-i` flag:

- ↑/↓: Navigate files
- Space: Select/deselect file
- Tab: Switch between search and list
- /: Focus search
- Enter: Confirm selection
- Esc: Cancel

## Default Exclusions

The tool automatically excludes common patterns:

- Version control: `.git/`, `.svn/`, `.hg/`
- Build directories: `target/`, `node_modules/`, `dist/`, `build/`
- Binary files: `*.exe`, `*.dll`, `*.so`, etc.
- Archives: `*.gz`, `*.zip`, `*.tar`, etc.
- Images: `*.jpg`, `*.png`, `*.gif`, etc.
- Cache directories: `__pycache__/`, `.mypy_cache/`, etc.

## Output Example

```
.
├─ src
│  ├─ main.go
│  └─ utils
│     ├─ parser.go
│     └─ helpers.go
└─ tests
   └─ main_test.go

Token Summary:
Total Tokens: 2048
Token Limit: 4096
Usage: 50.0%
```

## Dependencies

- github.com/gdamore/tcell/v2 - Terminal UI
- github.com/rivo/tview - Interactive TUI components
- github.com/pkoukk/tiktoken-go - GPT tokenizer
- github.com/sugarme/tokenizer - HuggingFace tokenizers
- github.com/atotto/clipboard - Clipboard integration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
