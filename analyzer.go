package main

import (
	// "fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileEntry struct {
	Path    string
	Content string
	Size    int64
}

type Analyzer struct {
	config *Config
	matcher *PatternMatcher
}

func NewAnalyzer(cfg *Config) *Analyzer {
	return &Analyzer{
		config:  cfg,
		matcher: NewPatternMatcher(cfg.Include, cfg.Exclude),
	}
}

func (a *Analyzer) ProcessDirectory() error {
	var entries []FileEntry
	var wg sync.WaitGroup
	entriesChan := make(chan FileEntry)
	errorsChan := make(chan error)
	done := make(chan bool)

	// Start worker to collect entries
	go func() {
		for entry := range entriesChan {
			entries = append(entries, entry)
		}
		done <- true
	}()

	// Walk directory
	err := filepath.Walk(a.config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check depth
		if relPath, err := filepath.Rel(a.config.Path, path); err == nil {
			if strings.Count(relPath, string(os.PathSeparator)) > a.config.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip if directory
		if info.IsDir() {
			return nil
		}

		// Check if file should be processed
		if !a.shouldProcessFile(path, info) {
			return nil
		}

		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			if entry, err := a.processFile(p); err != nil {
				errorsChan <- err
			} else {
				entriesChan <- entry
			}
		}(path)

		return nil
	})

	if err != nil {
		return err
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(entriesChan)
	}()

	// Wait for collection to finish
	<-done

	// Generate output
	return generateOutput(entries, a.config.Output)
}

func (a *Analyzer) shouldProcessFile(path string, info os.FileInfo) bool {
	// Check hidden files
	if !a.config.Hidden && strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	// Check file size
	if info.Size() > a.config.MaxSize {
		return false
	}

	// Check patterns
	return a.matcher.ShouldProcess(path)
}

func (a *Analyzer) processFile(path string) (FileEntry, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return FileEntry{}, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return FileEntry{}, err
	}

	relPath, err := filepath.Rel(a.config.Path, path)
	if err != nil {
		return FileEntry{}, err
	}

	return FileEntry{
		Path:    relPath,
		Content: string(content),
		Size:    info.Size(),
	}, nil
}