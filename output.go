// output.go
package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
)

const (
    PIPE        = "│"
    BRANCH      = "├─"
    LAST_BRANCH = "└─"
    INDENT      = "  "
    INDENT_PIPE = "│ "
)

func generateOutput(entries []FileEntry, format string, useClip bool) error {
    var contentBuf, treeBuf, tokenBuf bytes.Buffer

    switch format {
    case "tree":
        if err := printTree(entries, &treeBuf); err != nil {
            return err
        }
        if err := printTokenSummary(entries, &tokenBuf); err != nil {
            return err
        }
    case "files":
        if err := printFiles(entries, &contentBuf); err != nil {
            return err
        }
        if err := printTokenSummary(entries, &tokenBuf); err != nil {
            return err
        }
    case "both":
        if err := printFiles(entries, &contentBuf); err != nil {
            return err
        }
        if err := printTree(entries, &treeBuf); err != nil {
            return err
        }
        if err := printTokenSummary(entries, &tokenBuf); err != nil {
            return err
        }
    default:
        return fmt.Errorf("invalid output format specified")
    }
    
    if contentBuf.Len() > 0 {
        fmt.Print(contentBuf.String())
    }
    
    if treeBuf.Len() > 0 {
        if contentBuf.Len() > 0 {
            fmt.Print("\nDirectory Structure:\n")
        }
        fmt.Print(treeBuf.String())
    }
    
    if tokenBuf.Len() > 0 {
        fmt.Print("\nToken Summary:\n")
        fmt.Print(tokenBuf.String())
    }

    if useClip {
        var clipBuf bytes.Buffer
        clipBuf.Write(contentBuf.Bytes())
        if contentBuf.Len() > 0 && treeBuf.Len() > 0 {
            clipBuf.WriteString("\nDirectory Structure:\n")
        }
        clipBuf.Write(treeBuf.Bytes())
        if tokenBuf.Len() > 0 {
            clipBuf.WriteString("\nToken Summary:\n")
            clipBuf.Write(tokenBuf.Bytes())
        }
        
        if err := clipboard.WriteAll(clipBuf.String()); err != nil {
            return fmt.Errorf("failed to copy to clipboard: %w", err)
        }
        fmt.Println("\nOutput copied to clipboard!")
    }

    return nil
}

func printTokenSummary(entries []FileEntry, buf *bytes.Buffer) error {
    var totalTokens int
    var maxTokenLimit int

    for _, entry := range entries {
        if entry.TokenCount != nil {
            totalTokens += entry.TokenCount.Count
            if entry.TokenCount.TokenLimit > maxTokenLimit {
                maxTokenLimit = entry.TokenCount.TokenLimit
            }
        }
    }

    if maxTokenLimit > 0 {
        fmt.Fprintf(buf, "Total Tokens: %d\n", totalTokens)
        fmt.Fprintf(buf, "Token Limit: %d\n", maxTokenLimit)
        fmt.Fprintf(buf, "Usage: %.1f%%\n", float64(totalTokens)/float64(maxTokenLimit)*100)
    }

    return nil
}

func printFiles(entries []FileEntry, buf *bytes.Buffer) error {
    for _, entry := range entries {
        fmt.Fprintf(buf, "\nFile: %s\n", entry.Path)
        fmt.Fprintf(buf, "%s\n", strings.Repeat("=", 48))

        if entry.TokenCount != nil && entry.TokenCount.TokensPerc >= 80 {
            fmt.Fprintf(buf, "⚠️ Token usage: %d (%.1f%% of limit)\n",
                entry.TokenCount.Count, entry.TokenCount.TokensPerc)
        }

        formattedContent := formatFileContent(entry.Content)
        fmt.Fprintf(buf, "%s\n", formattedContent)
    }

    return nil
}

func formatFileContent(content string) string {
    lines := strings.Split(content, "\n")
    var formatted []string
    indent := 0

    for _, line := range lines {
        trimmed := strings.TrimRightFunc(line, unicode.IsSpace)
        if trimmed == "" {
            formatted = append(formatted, "")
            continue
        }

        leadingSpace := countLeadingSpace(trimmed)
        trimmed = strings.TrimSpace(trimmed)

        if strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[") {
            formatted = append(formatted, strings.Repeat("  ", indent)+trimmed)
            indent++
        } else if strings.HasPrefix(trimmed, "}") || strings.HasPrefix(trimmed, "]") {
            indent = max(0, indent-1)
            formatted = append(formatted, strings.Repeat("  ", indent)+trimmed)
        } else {
            if leadingSpace > 0 {
                indent = leadingSpace / 2
            }
            formatted = append(formatted, strings.Repeat("  ", indent)+trimmed)
        }
    }

    return strings.Join(formatted, "\n")
}

func countLeadingSpace(s string) int {
    count := 0
    for _, r := range s {
        if unicode.IsSpace(r) {
            if r == '\t' {
                count += 4 
            } else {
                count++
            }
        } else {
            break
        }
    }
    return count
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func printTree(entries []FileEntry, buf *bytes.Buffer) error {
    if len(entries) == 0 {
        return nil
    }

    buf.WriteString(".\n")

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

    printNode(root, ".", "", "", buf)

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
        if _, exists := tree[child]; exists {
            if isLast {
                fmt.Fprintf(buf, "%s%s %s\n", prefix, LAST_BRANCH, filepath.Base(child))
                printNode(tree, child, childPrefix+INDENT, childPrefix+INDENT, buf)
            } else {
                fmt.Fprintf(buf, "%s%s %s\n", prefix, BRANCH, filepath.Base(child))
                printNode(tree, child, childPrefix+INDENT_PIPE, childPrefix+INDENT_PIPE, buf)
            }
        } else {
            if isLast {
                fmt.Fprintf(buf, "%s%s %s\n", prefix, LAST_BRANCH, filepath.Base(child))
            } else {
                fmt.Fprintf(buf, "%s%s %s\n", prefix, BRANCH, filepath.Base(child))
            }
        }
    }
}