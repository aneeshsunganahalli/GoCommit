package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath handles both forward and backslashes
func NormalizePath(path string) string {
	// Replace backslashes with forward slashes
	normalized := strings.ReplaceAll(path, "\\", "/")
	// Remove any trailing slash
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized
}

// IsTextFile checks if a file is likely to be a text file
func IsTextFile(filename string) bool {
	// List of common text file extensions
	textExtensions := []string{
		".txt", ".md", ".go", ".js", ".py", ".java", ".c", ".cpp", ".h",
		".html", ".css", ".json", ".xml", ".yaml", ".yml", ".sh", ".bash",
		".ts", ".tsx", ".jsx", ".php", ".rb", ".rs", ".dart",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

// IsSmallFile checks if a file is small enough to include in context
func IsSmallFile(filename string) bool {
	const maxSize = 10 * 1024 // 10KB max

	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return info.Size() <= maxSize
}

// FilterEmpty removes empty strings from a slice
func FilterEmpty(slice []string) []string {
	filtered := []string{}
	for _, s := range slice {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
