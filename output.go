package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
)

const (
    PIPE        = "│"
    BRANCH      = "├─"    // Single character
    LAST_BRANCH = "└─"    // Single character
    INDENT      = "  "    // Two spaces is standard for tree views
    INDENT_PIPE = "│ "    // Single space after pipe
)

func generateOutput(entries []FileEntry, format string, useClip bool) error {
	var buf bytes.Buffer

	switch format {
	case "tree":
		if err := printTree(entries, &buf); err != nil {
			return err
		}
	case "files":
		if err := printFiles(entries, &buf); err != nil {
			return err
		}
	case "both":
		if err := printTree(entries, &buf); err != nil {
			return err
		}
		buf.WriteString("\nFile Contents:\n")
		if err := printFiles(entries, &buf); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format specified")
	}

	fmt.Print(buf.String())

	if useClip {
		if err := clipboard.WriteAll(buf.String()); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Println("\nOutput copied to clipboard!")
	}

	return nil
}

func printTree(entries []FileEntry, buf *bytes.Buffer) error {
	if len(entries) == 0 {
		return nil
	}

	buf.WriteString(".\n")

	// First, organize entries into a tree structure
	root := make(map[string][]string)
	for _, entry := range entries {
		dir := filepath.Dir(entry.Path)
		if dir == "." {
			root[dir] = append(root[dir], entry.Path)
		} else {
			parts := strings.Split(dir, string(filepath.Separator))
			current := ""
			for i, part := range parts {
				if i == 0 {
					current = part
				} else {
					current = filepath.Join(current, part)
				}
				if _, exists := root[current]; !exists {
					parent := "."
					if i > 0 {
						parent = filepath.Join(parts[:i]...)
					}
					root[parent] = append(root[parent], current)
				}
			}
			root[dir] = append(root[dir], entry.Path)
		}
	}

	// Then print the tree
	printNode(root, ".", "", "", buf)

	// Print summary
	dirs := len(root)
	files := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Path, "/") {
			files++
		}
	}
	fmt.Fprintf(buf, "\n%d directories, %d files\n", dirs, files)

	return nil
}

func printNode(tree map[string][]string, node, prefix, childPrefix string, buf *bytes.Buffer) {
	children := tree[node]
	if len(children) == 0 {
		return
	}

	for i, child := range children {
		isLast := i == len(children)-1
		// Skip if child is a parent of another node we've already seen
		if _, exists := tree[child]; exists {
			// This is a directory
			if isLast {
				fmt.Fprintf(buf, "%s%s %s\n", prefix, LAST_BRANCH, filepath.Base(child))
				printNode(tree, child, childPrefix+INDENT, childPrefix+INDENT, buf)
			} else {
				fmt.Fprintf(buf, "%s%s %s\n", prefix, BRANCH, filepath.Base(child))
				printNode(tree, child, childPrefix+INDENT_PIPE, childPrefix+INDENT_PIPE, buf)
			}
		} else {
			// This is a file
			if isLast {
				fmt.Fprintf(buf, "%s%s %s\n", prefix, LAST_BRANCH, filepath.Base(child))
			} else {
				fmt.Fprintf(buf, "%s%s %s\n", prefix, BRANCH, filepath.Base(child))
			}
		}
	}
}

func printFiles(entries []FileEntry, buf *bytes.Buffer) error {
	for _, entry := range entries {
		fmt.Fprintf(buf, "\nFile: %s\n", entry.Path)
		fmt.Fprintf(buf, "%s\n", strings.Repeat("=", 48))
		fmt.Fprintf(buf, "%s\n", entry.Content)
	}
	
	return nil
}