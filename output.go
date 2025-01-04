package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func generateOutput(entries []FileEntry, format string) error {
	switch format {
	case "tree":
		return printTree(entries)
	case "files":
		return printFiles(entries)
	case "both":
		if err := printTree(entries); err != nil {
			return err
		}
		fmt.Printf("\nFile Contents:\n")
		return printFiles(entries)
	default:
		return fmt.Errorf("invalid output format specified")
	}
}

func printTree(entries []FileEntry) error {
	fmt.Println("Directory Structure:")
	
	// Sort entries by path
	var currentPath []string
	
	for _, entry := range entries {
		components := strings.Split(entry.Path, string(filepath.Separator))
		
		// Print directory structure
		for i, component := range components {
			if i >= len(currentPath) || component != currentPath[i] {
				prefix := strings.Repeat("  ", i)
				if i == len(components)-1 {
					fmt.Printf("%s└── %s\n", prefix, component)
				} else {
					fmt.Printf("%s├── %s/\n", prefix, component)
				}
			}
		}
		currentPath = components
	}
	
	return nil
}

func printFiles(entries []FileEntry) error {
	for _, entry := range entries {
		fmt.Printf("\nFile: %s\n", entry.Path)
		fmt.Println(strings.Repeat("=", 48))
		fmt.Println(entry.Content)
	}
	
	// Print summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("Total files: %d\n", len(entries))
	
	var totalSize int64
	for _, entry := range entries {
		totalSize += entry.Size
	}
	fmt.Printf("Total size: %d bytes\n", totalSize)
	
	return nil
}