package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	// "path/filepath"
	"strings"
)

type Config struct {
	Path      string
	Include   []string
	Exclude   []string
	MaxSize   int64
	MaxDepth  int
	Output    string
	Threads   int
	Hidden    bool
	UseClip   bool
	Interactive bool
}

func parseFlags() (*Config, error) {
    cfg := &Config{}

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

    return cfg, nil
}

func main() {
    cfg, err := parseFlags()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    analyzer := NewAnalyzer(cfg)
    files, err := analyzer.CollectFiles()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if cfg.Interactive {
        // Create a channel to receive selected files
        selectedChan := make(chan []FileEntry, 1)
        
        picker := NewFilePicker(files, func(selected []FileEntry) {
            // Process selected files and send them through the channel
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
        
        if err := picker.Run(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        selectedFiles := <-selectedChan

		if len(selectedFiles) == 0 {
			fmt.Println("No files selected.")
			os.Exit(0)
		}

        if len(selectedFiles) > 0 {
            // Generate output for selected files only
            if err := generateOutput(selectedFiles, cfg.Output, cfg.UseClip); err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
            }
        }
    } else {
        // Non-interactive mode - process all files
        if err := generateOutput(files, cfg.Output, cfg.UseClip); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    }
}