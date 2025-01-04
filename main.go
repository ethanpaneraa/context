package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func parseFlags() (*Config, error) {
    cfg := &Config{}
	cfg.TokenizerType = TiktokenGPT35
    flag.StringVar(&cfg.Path, "path", ".", "Directory to analyze")
    includeStr := flag.String("include", "", "Patterns to include (comma-separated)")
    excludeStr := flag.String("exclude", "", "Patterns to exclude (comma-separated)")
    flag.Int64Var(&cfg.MaxSize, "max-size", 10*1024*1024, "Maximum file size in bytes")
    flag.IntVar(&cfg.MaxDepth, "max-depth", 20, "Maximum directory depth")
    flag.StringVar(&cfg.Output, "output", "both", "Output format (tree, files, or both)")
    flag.IntVar(&cfg.Threads, "threads", 0, "Number of threads for parallel processing")
    flag.BoolVar(&cfg.Hidden, "hidden", false, "Show hidden files and directories")
    flag.BoolVar(&cfg.UseClip, "c", false, "Copy output to clipboard")
    flag.BoolVar(&cfg.Interactive, "i", false, "Interactive mode")
    
    // New tokenizer flags
    tokenizerType := flag.String("tokenizer", "", "Tokenizer type (gpt-3.5-turbo, gpt-4, claude, huggingface)")
    flag.StringVar(&cfg.TokenizerModel, "tokenizer-model", "", "Path to HuggingFace tokenizer model")
    flag.IntVar(&cfg.TokenLimit, "token-limit", 4096, "Maximum token limit")

    flag.Parse()

    // Validate path
    if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
        return nil, fmt.Errorf("path '%s' does not exist", cfg.Path)
    }

    // Parse include/exclude patterns
    if *includeStr != "" {
        cfg.Include = strings.Split(*includeStr, ",")
    }
    if *excludeStr != "" {
        cfg.Exclude = strings.Split(*excludeStr, ",")
    }

    // Set tokenizer type if specified
    if *tokenizerType != "" {
        switch *tokenizerType {
        case "gpt-3.5-turbo":
            cfg.TokenizerType = TiktokenGPT35
        case "gpt-4":
            cfg.TokenizerType = TiktokenGPT4
        case "claude":
            cfg.TokenizerType = TiktokenClaude
        case "huggingface":
            if cfg.TokenizerModel == "" {
                return nil, fmt.Errorf("must specify --tokenizer-model for HuggingFace tokenizer")
            }
            cfg.TokenizerType = HuggingFace
        default:
            return nil, fmt.Errorf("unsupported tokenizer type: %s", *tokenizerType)
        }
    }

    return cfg, nil
}

func main() {
    cfg, err := parseFlags()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    analyzer, err := NewAnalyzer(cfg)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    files, err := analyzer.CollectFiles()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if cfg.Interactive {
        // Create a channel to receive selected files
        selectedChan := make(chan []FileEntry, 1)
        
        picker := NewFilePicker(files, func(selected []FileEntry) {
            defer close(selectedChan) // Make sure we close the channel
            
            // If no files were selected (user cancelled), send empty slice
            if len(selected) == 0 {
                selectedChan <- nil
                return
            }
            
            // Process selected files
            var processedFiles []FileEntry
            for _, file := range selected {
                if file.Content == "" {
                    entry, err := analyzer.processFile(filepath.Join(cfg.Path, file.Path))
                    if err != nil {
                        fmt.Fprintf(os.Stderr, "Error loading file %s: %v\n", file.Path, err)
                        continue
                    }
                    processedFiles = append(processedFiles, entry)
                } else {
                    processedFiles = append(processedFiles, file)
                }
            }
            selectedChan <- processedFiles
        })

        if picker == nil {
            fmt.Fprintf(os.Stderr, "Error: failed to create file picker\n")
            os.Exit(1)
        }
        
        done := make(chan error, 1)
        go func() {
            done <- picker.Run()
        }()

        select {
        case err := <-done:
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
        case <-time.After(100 * time.Millisecond): 
        }

        // Wait for selected files
        selectedFiles := <-selectedChan
        if selectedFiles == nil {
            fmt.Println("Operation cancelled.")
            os.Exit(0)
        }

        if len(selectedFiles) == 0 {
            fmt.Println("No files selected.")
            os.Exit(0)
        }

        // Generate output for selected files
        if err := generateOutput(selectedFiles, cfg.Output, cfg.UseClip); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    } else {
        if err := generateOutput(files, cfg.Output, cfg.UseClip); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    }
}