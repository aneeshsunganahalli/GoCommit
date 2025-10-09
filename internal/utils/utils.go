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
		".ts", ".tsx", ".jsx", ".php", ".rb", ".rs", ".dart", ".sql", ".r",
		".scala", ".kt", ".swift", ".m", ".pl", ".lua", ".vim", ".csv", 
		".log", ".cfg", ".conf", ".ini", ".toml", ".lock", ".gitignore",
		".dockerfile", ".makefile", ".cmake", ".pro", ".pri", ".svg",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	// Common extensionless files that are typically text
	if ext == "" {
		baseName := strings.ToLower(filepath.Base(filename))
		commonTextFiles := []string{
			"readme", "dockerfile", "makefile", "rakefile", "gemfile", 
			"procfile", "jenkinsfile", "vagrantfile", "changelog", "authors",
			"contributors", "copying", "install", "news", "todo",
		}
		
		for _, textFile := range commonTextFiles {
			if baseName == textFile {
				return true
			}
		}
	}

	return false
}

// IsBinaryFile checks if a file is likely to be a binary file that should be excluded from diffs
func IsBinaryFile(filename string) bool {
	// List of common binary file extensions
	binaryExtensions := []string{
		// Images (excluding SVG which is XML text)
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".ico", ".webp",
		// Audio/Video
		".mp3", ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".wav", ".ogg", ".m4a",
		// Archives/Compressed
		".zip", ".tar", ".gz", ".7z", ".rar", ".bz2", ".xz", ".lz", ".lzma",
		// Executables/Libraries
		".exe", ".dll", ".so", ".dylib", ".a", ".lib", ".bin", ".deb", ".rpm", ".dmg", ".msi",
		// Documents
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp",
		// Fonts
		".ttf", ".otf", ".woff", ".woff2", ".eot",
		// Other binary formats
		".db", ".sqlite", ".sqlite3", ".mdb", ".accdb", ".pickle", ".pkl", ".pyc", ".pyo",
		".class", ".jar", ".war", ".ear", ".apk", ".ipa",
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, binExt := range binaryExtensions {
		if ext == binExt {
			return true
		}
	}

	// Note: Files with unknown extensions are not considered binary by default
	// This allows them to be processed as text files for diff analysis
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
