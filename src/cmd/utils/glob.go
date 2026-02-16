// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsGlobPattern checks if a path contains glob pattern characters
func IsGlobPattern(path string) bool {
	// Check for glob pattern characters: *, ?, [, ]
	return strings.ContainsAny(path, "*?[")
}

// ExpandGlobPattern expands a glob pattern to a list of matching files
func ExpandGlobPattern(pattern string) ([]string, error) {
	// Use filepath.Glob for simple patterns
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	// Filter out directories, only return files
	var result []string
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue // Skip files we can't stat
		}
		if !info.IsDir() {
			result = append(result, file)
		}
	}

	return result, nil
}
