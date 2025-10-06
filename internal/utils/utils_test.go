package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "windows style", input: "foo\\bar\\", expected: "foo/bar"},
		{name: "already normalized", input: "foo/bar", expected: "foo/bar"},
		{name: "no trailing slash", input: "foo", expected: "foo"},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizePath(tc.input); got != tc.expected {
				t.Fatalf("NormalizePath(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsTextFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "go source", filename: "main.go", want: true},
		{name: "markdown upper", filename: "README.MD", want: true},
		{name: "binary extension", filename: "image.png", want: false},
		{name: "no extension", filename: "LICENSE", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsTextFile(tt.filename); got != tt.want {
				t.Fatalf("IsTextFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsSmallFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	smallPath := filepath.Join(dir, "small.txt")
	if err := os.WriteFile(smallPath, bytes.Repeat([]byte("x"), 1024), 0o644); err != nil {
		t.Fatalf("failed to write small file: %v", err)
	}

	largePath := filepath.Join(dir, "large.txt")
	if err := os.WriteFile(largePath, bytes.Repeat([]byte("y"), 11*1024), 0o644); err != nil {
		t.Fatalf("failed to write large file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "small file", path: smallPath, want: true},
		{name: "large file", path: largePath, want: false},
		{name: "missing file", path: filepath.Join(dir, "missing.txt"), want: false},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsSmallFile(tc.path); got != tc.want {
				t.Fatalf("IsSmallFile(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestFilterEmpty(t *testing.T) {
	t.Parallel()

	input := []string{"feat", "", "test", " "}
	want := []string{"feat", "test", " "}

	got := FilterEmpty(input)

	if len(got) != len(want) {
		t.Fatalf("FilterEmpty returned %d items, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FilterEmpty mismatch at index %d: got %q want %q", i, got[i], want[i])
		}
	}
}
