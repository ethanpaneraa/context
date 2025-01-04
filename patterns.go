package main

import (
	"path/filepath"
	"strings"
)

type PatternMatcher struct {
	includePatterns []string
	excludePatterns []string
}

func NewPatternMatcher(include, exclude []string) *PatternMatcher {
	if exclude == nil {
		exclude = defaultExcludes()
	}
	
	return &PatternMatcher{
		includePatterns: include,
		excludePatterns: exclude,
	}
}

func (pm *PatternMatcher) ShouldProcess(path string) bool {
	if len(pm.includePatterns) > 0 {
		matches := false
		for _, pattern := range pm.includePatterns {
			if matched := matchPattern(pattern, path); matched {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	for _, pattern := range pm.excludePatterns {
		if matched := matchPattern(pattern, path); matched {
			return false
		}
	}

	return true
}

func matchPattern(pattern, path string) bool {
    path = filepath.ToSlash(path)
    pattern = filepath.ToSlash(pattern)

	
    if !strings.Contains(pattern, "*") {
        return strings.Contains(path, "/"+pattern+"/") || strings.HasPrefix(path, pattern+"/")
    }

    if strings.HasPrefix(pattern, "**/") {
        pattern = pattern[3:] 
        for {
            if matched, _ := filepath.Match(pattern, path); matched {
                return true
            }
            lastSlash := strings.LastIndex(path, "/")
            if lastSlash == -1 {
                break
            }
            path = path[lastSlash+1:]
        }
        return false
    }

    matched, _ := filepath.Match(pattern, filepath.Base(path))
    return matched
}

func defaultExcludes() []string {
    return []string{
        ".git", ".git/**", ".gitignore", ".gitattributes", ".gitmodules",
        ".svn", ".hg",
        
        "target", "node_modules", "dist", "build",

        "*.exe", "*.dll", "*.so", "*.dylib", "*.o", "*.obj",
        "*.bin", "*.dat", "*.db", "*.sqlite", "*.sqlite3",
        
        "*.gz", "*.zip", "*.tar", "*.rar", "*.7z", "*.bz2",
        
        "*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp",
        "*.ico", "*.svg", "*.webp", "*.tiff", "*.raw", "*.heic",
        
        "__pycache__", ".mypy_cache", ".pytest_cache",
    }
}