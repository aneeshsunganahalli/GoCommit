package cache

import (
	"strings"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestDiffHasher_GenerateHash(t *testing.T) {
	hasher := NewDiffHasher()

	tests := []struct {
		name     string
		diff     string
		opts     *types.GenerationOptions
		expected string
	}{
		{
			name: "simple diff without options",
			diff: `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`,
			opts:     nil,
			expected: "", // We'll check it's not empty
		},
		{
			name: "same diff with style instruction",
			diff: `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`,
			opts: &types.GenerationOptions{
				StyleInstruction: "Write in a casual tone",
			},
			expected: "", // We'll check it's not empty
		},
		{
			name: "different diff",
			diff: `diff --git a/other.txt b/other.txt
index 1111111..2222222 100644
--- a/other.txt
+++ b/other.txt
@@ -1,2 +1,2 @@
 hello
-world
+universe`,
			opts:     nil,
			expected: "", // We'll check it's not empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hasher.GenerateHash(tt.diff, tt.opts)
			hash2 := hasher.GenerateHash(tt.diff, tt.opts)

			// Hash should be consistent
			if hash1 != hash2 {
				t.Errorf("GenerateHash() inconsistent: %s != %s", hash1, hash2)
			}

			// Hash should not be empty
			if hash1 == "" {
				t.Errorf("GenerateHash() returned empty hash")
			}

			// Hash should be 64 characters (SHA256 hex)
			if len(hash1) != 64 {
				t.Errorf("GenerateHash() returned hash of length %d, expected 64", len(hash1))
			}
		})
	}
}

func TestDiffHasher_GenerateHash_Consistency(t *testing.T) {
	hasher := NewDiffHasher()

	diff := `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`

	opts := &types.GenerationOptions{
		StyleInstruction: "Write in a casual tone",
		Attempt:          1,
	}

	// Generate hash multiple times
	hash1 := hasher.GenerateHash(diff, opts)
	hash2 := hasher.GenerateHash(diff, opts)
	hash3 := hasher.GenerateHash(diff, opts)

	// All hashes should be identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("GenerateHash() not consistent across calls")
	}
}

func TestDiffHasher_GenerateHash_DifferentInputs(t *testing.T) {
	hasher := NewDiffHasher()

	diff1 := `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`

	diff2 := `diff --git a/other.txt b/other.txt
index 1111111..2222222 100644
--- a/other.txt
+++ b/other.txt
@@ -1,2 +1,2 @@
 hello
-world
+universe`

	opts1 := &types.GenerationOptions{
		StyleInstruction: "Write in a casual tone",
		Attempt:          1,
	}

	opts2 := &types.GenerationOptions{
		StyleInstruction: "Write in a formal tone",
		Attempt:          1,
	}

	hash1 := hasher.GenerateHash(diff1, opts1)
	hash2 := hasher.GenerateHash(diff2, opts1)
	hash3 := hasher.GenerateHash(diff1, opts2)

	// Different diffs should produce different hashes
	if hash1 == hash2 {
		t.Errorf("GenerateHash() returned same hash for different diffs")
	}

	// Same diff with different style instructions should produce different hashes
	if hash1 == hash3 {
		t.Errorf("GenerateHash() returned same hash for different style instructions")
	}
}

func TestDiffHasher_GenerateCacheKey(t *testing.T) {
	hasher := NewDiffHasher()

	diff := `diff --git a/file.txt b/file.txt
index 1234567..abcdefg 100644
--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 line1
-line2
+line2 updated
 line3`

	opts := &types.GenerationOptions{
		StyleInstruction: "Write in a casual tone",
		Attempt:          1,
	}

	key1 := hasher.GenerateCacheKey(types.ProviderOpenAI, diff, opts)
	key2 := hasher.GenerateCacheKey(types.ProviderClaude, diff, opts)
	key3 := hasher.GenerateCacheKey(types.ProviderOpenAI, diff, opts)

	// Same provider and diff should produce same key
	if key1 != key3 {
		t.Errorf("GenerateCacheKey() not consistent for same provider")
	}

	// Different providers should produce different keys
	if key1 == key2 {
		t.Errorf("GenerateCacheKey() returned same key for different providers")
	}

	// Keys should contain provider name
	if !containsString(key1, "OpenAI") {
		t.Errorf("GenerateCacheKey() does not contain provider name")
	}

	if !containsString(key2, "Claude") {
		t.Errorf("GenerateCacheKey() does not contain provider name")
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
