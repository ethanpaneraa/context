package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Analyzer struct {
	config *Config
	matcher *PatternMatcher
	tokenizer Tokenizer
}


func NewAnalyzer(cfg *Config) (*Analyzer, error) {
	tokenizer, err := NewTokenizer(cfg.TokenizerType, cfg.TokenizerModel, cfg.TokenLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	return &Analyzer{
		config:    cfg,
		matcher:   NewPatternMatcher(cfg.Include, cfg.Exclude),
		tokenizer: tokenizer,
	}, nil
}

func (a *Analyzer) ProcessDirectory() error {
	var entries []FileEntry
	var wg sync.WaitGroup
	entriesChan := make(chan FileEntry)
	errorsChan := make(chan error)
	done := make(chan bool)

	go func() {
		for entry := range entriesChan {
			entries = append(entries, entry)
		}
		done <- true
	}()

	err := filepath.Walk(a.config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if relPath, err := filepath.Rel(a.config.Path, path); err == nil {
			if strings.Count(relPath, string(os.PathSeparator)) > a.config.MaxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

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

	go func() {
		wg.Wait()
		close(entriesChan)
	}()

	<-done

	return generateOutput(entries, a.config.Output, a.config.UseClip)
}

func (a *Analyzer) shouldProcessFile(path string, info os.FileInfo) bool {
	if !a.config.Hidden && strings.HasPrefix(filepath.Base(path), ".") {
		return false
	}

	if info.Size() > a.config.MaxSize {
		return false
	}

	return a.matcher.ShouldProcess(path)
}

func (a *Analyzer) processFile(path string) (FileEntry, error) {
    // First check if it's a text file by reading the first few bytes
    f, err := os.Open(path)
    if err != nil {
        return FileEntry{}, err
    }
    defer f.Close()

    // Read first 512 bytes to check content type
    buffer := make([]byte, 512)
    n, err := f.Read(buffer)
    if err != nil && err != io.EOF {
        return FileEntry{}, err
    }
    buffer = buffer[:n]

    // Check if file appears to be binary
    if isBinary(buffer) {
        return FileEntry{}, nil // Skip binary files silently
    }

    // If we get here, file is probably text, read the whole thing
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

    // Count tokens if tokenizer is configured
    var tokenCount *TokenCount
    if a.tokenizer != nil {
        count, err := a.tokenizer.CountTokens(string(content))
        if err != nil {
            return FileEntry{}, fmt.Errorf("failed to count tokens: %w", err)
        }
        tokenCount = &count
    }

    return FileEntry{
        Path:       relPath,
        Content:    string(content),
        Size:       info.Size(),
        TokenCount: tokenCount,
    }, nil
}

func isBinary(buf []byte) bool {
    const binary_threshold = 0.3
    if len(buf) == 0 {
        return false
    }

    binaryCount := 0
    for _, b := range buf {
        if b == 0 || (b < 7 && b != 5 && b != 4) || (b > 14 && b < 32 && b != '\n' && b != '\r' && b != '\t') {
            binaryCount++
        }
    }

    return float64(binaryCount)/float64(len(buf)) > binary_threshold
}

func (a *Analyzer) CollectFiles() ([]FileEntry, error) {
    var entries []FileEntry
    var wg sync.WaitGroup
    entriesChan := make(chan FileEntry)
    errorsChan := make(chan error)
    done := make(chan bool)

    go func() {
        for entry := range entriesChan {
            entries = append(entries, entry)
        }
        done <- true
    }()

    err := filepath.Walk(a.config.Path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if relPath, err := filepath.Rel(a.config.Path, path); err == nil {
            if strings.Count(relPath, string(os.PathSeparator)) > a.config.MaxDepth {
                if info.IsDir() {
                    return filepath.SkipDir
                }
                return nil
            }
        }

        if info.IsDir() {
            return nil
        }

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
        return nil, err
    }

    go func() {
        wg.Wait()
        close(entriesChan)
    }()

    <-done

    return entries, nil
}