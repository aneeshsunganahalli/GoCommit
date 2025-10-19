package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/dfanso/commit-msg/pkg/types"
)

// DiffHasher generates consistent hash keys for git diffs to enable caching.
type DiffHasher struct{}

// NewDiffHasher creates a new DiffHasher instance.
func NewDiffHasher() *DiffHasher {
	return &DiffHasher{}
}

// GenerateHash creates a hash key for a given diff and generation options.
// This hash is used as the cache key to identify similar diffs.
func (h *DiffHasher) GenerateHash(diff string, opts *types.GenerationOptions) string {
	// Normalize the diff by removing timestamps, file paths, and other variable content
	normalizedDiff := h.normalizeDiff(diff)

	// Create a hash input that includes the normalized diff and relevant options
	hashInput := h.buildHashInput(normalizedDiff, opts)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// normalizeDiff removes variable content from git diff to create consistent hashes
// for similar changes across different repositories or time periods.
func (h *DiffHasher) normalizeDiff(diff string) string {
	lines := strings.Split(diff, "\n")
	var normalizedLines []string

	for _, line := range lines {
		// Skip lines that contain variable information
		if h.shouldSkipLine(line) {
			continue
		}

		// Normalize the line content
		normalizedLine := h.normalizeLine(line)
		if normalizedLine != "" {
			normalizedLines = append(normalizedLines, normalizedLine)
		}
	}

	// Sort lines to ensure consistent ordering regardless of git output order
	sort.Strings(normalizedLines)

	return strings.Join(normalizedLines, "\n")
}

// shouldSkipLine determines if a line should be excluded from the hash calculation.
func (h *DiffHasher) shouldSkipLine(line string) bool {
	// Skip empty lines
	if strings.TrimSpace(line) == "" {
		return true
	}

	// Skip git diff headers that contain timestamps or file paths
	if strings.HasPrefix(line, "diff --git") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "+++") ||
		strings.HasPrefix(line, "---") ||
		strings.HasPrefix(line, "@@") {
		return true
	}

	// Skip lines with file paths (but keep the actual changes)
	if strings.Contains(line, "/") && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) {
		// This is a file path line, skip it
		return true
	}

	return false
}

// normalizeLine normalizes a single line of diff content.
func (h *DiffHasher) normalizeLine(line string) string {
	// Remove leading/trailing whitespace
	line = strings.TrimSpace(line)

	// If it's a diff line (+ or -), normalize the content
	if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
		// Keep the + or - prefix but normalize the rest
		prefix := line[:1]
		content := line[1:]

		// Remove line numbers and other variable content
		// This is a simple approach - in practice, you might want more sophisticated normalization
		content = strings.TrimSpace(content)

		return prefix + content
	}

	return line
}

// buildHashInput creates the input string for hashing by combining
// the normalized diff with relevant generation options.
func (h *DiffHasher) buildHashInput(normalizedDiff string, opts *types.GenerationOptions) string {
	var parts []string

	// Add the normalized diff
	parts = append(parts, normalizedDiff)

	// Add style instruction if present
	if opts != nil && opts.StyleInstruction != "" {
		parts = append(parts, "style:"+strings.TrimSpace(opts.StyleInstruction))
	}

	// Add attempt number (but only if it's the first attempt, as we want to cache
	// the base generation, not regenerations)
	if opts == nil || opts.Attempt <= 1 {
		parts = append(parts, "attempt:1")
	}

	return strings.Join(parts, "|")
}

// GenerateCacheKey creates a complete cache key including provider information.
func (h *DiffHasher) GenerateCacheKey(provider types.LLMProvider, diff string, opts *types.GenerationOptions) string {
	diffHash := h.GenerateHash(diff, opts)
	return fmt.Sprintf("%s:%s", provider.String(), diffHash)
}
