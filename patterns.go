package main

import (
	"path/filepath"
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
	// If include patterns exist, path must match at least one
	if len(pm.includePatterns) > 0 {
		matches := false
		for _, pattern := range pm.includePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	// If exclude patterns exist, path must not match any
	for _, pattern := range pm.excludePatterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return false
		}
	}

	return true
}

func defaultExcludes() []string {
	return []string{
		// Version control
		"**/.git/**",
		"**/.svn/**",
		"**/.hg/**",
		// Build artifacts and dependencies
		"**/target/**",
		"**/node_modules/**",
		"**/dist/**",
		"**/build/**",
		// Binaries and objects
		"**/*.exe",
		"**/*.dll",
		"**/*.so",
		"**/*.dylib",
		"**/*.o",
		"**/*.obj",
		// Cache directories
		"**/__pycache__/**",
		"**/.mypy_cache/**",
		"**/.pytest_cache/**",
	}
}