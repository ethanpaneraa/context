package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
    // Create test directory
    testDir := "test_files"
    if err := os.MkdirAll(testDir, 0755); err != nil {
        fmt.Printf("Error creating directory: %v\n", err)
        return
    }

    // Generate some large files
    sizes := []struct {
        name string
        size int64
    }{
        {"small.txt", 500 * 1024},        
        {"medium.txt", 2 * 1024 * 1024},  // 2MB
        {"large.txt", 5 * 1024 * 1024},   // 5MB
        {"huge.txt", 10 * 1024 * 1024},   // 10MB
    }

    // Create some nested directories
    dirs := []string{
        "dir1/subdir1",
        "dir1/subdir2",
        "dir2/subdir1",
        "dir2/subdir2/subsubdir",
    }

    for _, dir := range dirs {
        path := filepath.Join(testDir, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            fmt.Printf("Error creating directory %s: %v\n", dir, err)
            continue
        }
    }

    // Generate files in each directory
    for _, dir := range append(dirs, "") {
        dirPath := filepath.Join(testDir, dir)
        for _, size := range sizes {
            filePath := filepath.Join(dirPath, size.name)
            if err := generateLargeFile(filePath, size.size); err != nil {
                fmt.Printf("Error generating file %s: %v\n", filePath, err)
                continue
            }
        }
    }

    fmt.Println("Test files generated successfully!")
    fmt.Println("\nTo test the progress bars, run:")
    fmt.Printf("go run . -path %s\n", testDir)
}

func generateLargeFile(path string, size int64) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    // Create random text content
    words := []string{
        "Lorem", "ipsum", "dolor", "sit", "amet", "consectetur",
        "adipiscing", "elit", "sed", "do", "eiusmod", "tempor",
        "incididunt", "ut", "labore", "et", "dolore", "magna",
        "aliqua", "\n",
    }

    // Write in chunks to avoid memory issues
    chunkSize := int64(1024 * 1024) // 1MB chunks
    var written int64
    
    for written < size {
        // Generate a chunk of text
        var chunk strings.Builder
        for chunk.Len() < int(chunkSize) && written+int64(chunk.Len()) < size {
            word := words[rand.Intn(len(words))]
            chunk.WriteString(word)
            chunk.WriteString(" ")
        }
        
        n, err := f.WriteString(chunk.String())
        if err != nil {
            return err
        }
        written += int64(n)
        
        // Add a small delay to simulate slower disk operations
        time.Sleep(50 * time.Millisecond)
    }

    return nil
}