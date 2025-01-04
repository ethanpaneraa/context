package main

import (
	"flag"
	"fmt"
	"os"

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
	if err := analyzer.ProcessDirectory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}