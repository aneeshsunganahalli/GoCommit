package stats

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestGetFileStatistics(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Run("handles non-existent directory", func(t *testing.T) {
		t.Parallel()

		config := &types.RepoConfig{
			Path: "/non/existent/path",
		}

		// The function should not panic and should return some stats
		stats, err := GetFileStatistics(config)
		// It may not return an error, but should handle it gracefully
		_ = err
		if stats == nil {
			t.Fatal("expected stats to be returned even for non-existent directory")
		}
	})

	t.Run("handles non-git directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		config := &types.RepoConfig{
			Path: dir,
		}

		// The function should not panic and should return some stats
		stats, err := GetFileStatistics(config)
		// It may not return an error, but should handle it gracefully
		_ = err
		if stats == nil {
			t.Fatal("expected stats to be returned even for non-git directory")
		}
	})

	t.Run("handles empty repository", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		setupGitRepo(t, dir)

		config := &types.RepoConfig{
			Path: dir,
		}

		stats, err := GetFileStatistics(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(stats.StagedFiles) != 0 {
			t.Fatalf("expected 0 staged files, got %d", len(stats.StagedFiles))
		}

		if len(stats.UnstagedFiles) != 0 {
			t.Fatalf("expected 0 unstaged files, got %d", len(stats.UnstagedFiles))
		}

		if len(stats.UntrackedFiles) != 0 {
			t.Fatalf("expected 0 untracked files, got %d", len(stats.UntrackedFiles))
		}

		if stats.TotalFiles != 0 {
			t.Fatalf("expected 0 total files, got %d", stats.TotalFiles)
		}

		if stats.LinesAdded != 0 {
			t.Fatalf("expected 0 lines added, got %d", stats.LinesAdded)
		}

		if stats.LinesDeleted != 0 {
			t.Fatalf("expected 0 lines deleted, got %d", stats.LinesDeleted)
		}
	})
}

func TestGetFileStatisticsWithChanges(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	// Create a tracked file and commit it
	trackedFile := filepath.Join(dir, "tracked.go")
	if err := os.WriteFile(trackedFile, []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"), 0o644); err != nil {
		t.Fatalf("failed to write tracked file: %v", err)
	}
	runGit(t, dir, "add", "tracked.go")
	runGit(t, dir, "commit", "-m", "initial commit")

	// Modify the tracked file (unstaged changes)
	if err := os.WriteFile(trackedFile, []byte("package main\n\nfunc main() {\n\tprintln(\"hello world\")\n}"), 0o644); err != nil {
		t.Fatalf("failed to modify tracked file: %v", err)
	}

	// Create a new staged file
	stagedFile := filepath.Join(dir, "staged.js")
	if err := os.WriteFile(stagedFile, []byte("console.log('hello');"), 0o644); err != nil {
		t.Fatalf("failed to write staged file: %v", err)
	}
	runGit(t, dir, "add", "staged.js")

	// Create an untracked file
	untrackedFile := filepath.Join(dir, "untracked.txt")
	if err := os.WriteFile(untrackedFile, []byte("untracked content"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check unstaged files
	if len(stats.UnstagedFiles) != 1 {
		t.Fatalf("expected 1 unstaged file, got %d", len(stats.UnstagedFiles))
	}
	if !strings.Contains(stats.UnstagedFiles[0], "tracked.go") {
		t.Fatalf("expected tracked.go in unstaged files, got %s", stats.UnstagedFiles[0])
	}

	// Check staged files
	if len(stats.StagedFiles) != 1 {
		t.Fatalf("expected 1 staged file, got %d", len(stats.StagedFiles))
	}
	if !strings.Contains(stats.StagedFiles[0], "staged.js") {
		t.Fatalf("expected staged.js in staged files, got %s", stats.StagedFiles[0])
	}

	// Check untracked files
	if len(stats.UntrackedFiles) != 1 {
		t.Fatalf("expected 1 untracked file, got %d", len(stats.UntrackedFiles))
	}
	if !strings.Contains(stats.UntrackedFiles[0], "untracked.txt") {
		t.Fatalf("expected untracked.txt in untracked files, got %s", stats.UntrackedFiles[0])
	}

	// Check total files
	expectedTotal := 3
	if stats.TotalFiles != expectedTotal {
		t.Fatalf("expected %d total files, got %d", expectedTotal, stats.TotalFiles)
	}

	// Check line statistics (should only include staged files)
	if stats.LinesAdded == 0 {
		t.Fatal("expected lines added > 0 for staged files")
	}
}

func TestGetFileStatisticsWithLineNumbers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	// Create and commit initial file
	initialFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(initialFile, []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}
	runGit(t, dir, "add", "test.txt")
	runGit(t, dir, "commit", "-m", "initial commit")

	// Modify file with added and deleted lines
	modifiedContent := []byte("line1\nline2 modified\nline3\nline4 added\nline5 added\n")
	if err := os.WriteFile(initialFile, modifiedContent, 0o644); err != nil {
		t.Fatalf("failed to modify file: %v", err)
	}
	runGit(t, dir, "add", "test.txt")

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have added at least 2 lines (line4 and line5)
	if stats.LinesAdded < 2 {
		t.Fatalf("expected at least 2 lines added, got %d", stats.LinesAdded)
	}

	// Should have deleted at least 1 line (original line2)
	if stats.LinesDeleted < 1 {
		t.Fatalf("expected at least 1 line deleted, got %d", stats.LinesDeleted)
	}
}

func TestGetFileStatisticsWithBinaryFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	// Create a binary file (staged)
	binaryFile := filepath.Join(dir, "binary.bin")
	binaryContent := make([]byte, 100)
	for i := range binaryContent {
		binaryContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(binaryFile, binaryContent, 0o644); err != nil {
		t.Fatalf("failed to write binary file: %v", err)
	}
	runGit(t, dir, "add", "binary.bin")

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Binary files should be counted but may not have line statistics
	if len(stats.StagedFiles) != 1 {
		t.Fatalf("expected 1 staged file, got %d", len(stats.StagedFiles))
	}
	if !strings.Contains(stats.StagedFiles[0], "binary.bin") {
		t.Fatalf("expected binary.bin in staged files, got %s", stats.StagedFiles[0])
	}
}

func TestGetFileStatisticsWithEmptyFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	// Create an empty file (staged)
	emptyFile := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte{}, 0o644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}
	runGit(t, dir, "add", "empty.txt")

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty files should be counted
	if len(stats.StagedFiles) != 1 {
		t.Fatalf("expected 1 staged file, got %d", len(stats.StagedFiles))
	}
	if !strings.Contains(stats.StagedFiles[0], "empty.txt") {
		t.Fatalf("expected empty.txt in staged files, got %s", stats.StagedFiles[0])
	}
}

func TestGetFileStatisticsWithSubdirectories(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	// Create subdirectories
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	nestedDir := filepath.Join(dir, "subdir", "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	// Create files in subdirectories
	subdirFile := filepath.Join(subdir, "sub.js")
	if err := os.WriteFile(subdirFile, []byte("console.log('sub');"), 0o644); err != nil {
		t.Fatalf("failed to write subdir file: %v", err)
	}
	runGit(t, dir, "add", "subdir/sub.js")

	nestedFile := filepath.Join(nestedDir, "nested.py")
	if err := os.WriteFile(nestedFile, []byte("print('nested')"), 0o644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}
	runGit(t, dir, "add", "subdir/nested/nested.py")

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find files in subdirectories
	if len(stats.StagedFiles) != 2 {
		t.Fatalf("expected 2 staged files, got %d", len(stats.StagedFiles))
	}

	// Check that paths are correct
	var foundSubdir, foundNested bool
	for _, file := range stats.StagedFiles {
		if strings.Contains(file, "sub.js") {
			foundSubdir = true
		}
		if strings.Contains(file, "nested.py") {
			foundNested = true
		}
	}

	if !foundSubdir {
		t.Fatal("expected to find sub.js in staged files")
	}
	if !foundNested {
		t.Fatal("expected to find nested.py in staged files")
	}
}

func TestGetFileStatisticsReturnsCorrectType(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Parallel()

	dir := t.TempDir()
	setupGitRepo(t, dir)

	config := &types.RepoConfig{
		Path: dir,
	}

	stats, err := GetFileStatistics(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the returned type has the expected fields
	if stats.StagedFiles == nil {
		t.Fatal("expected StagedFiles to be initialized")
	}

	if stats.UnstagedFiles == nil {
		t.Fatal("expected UnstagedFiles to be initialized")
	}

	if stats.UntrackedFiles == nil {
		t.Fatal("expected UntrackedFiles to be initialized")
	}
}

// setupGitRepo initializes a git repository in the given directory
func setupGitRepo(t *testing.T, dir string) {
	t.Helper()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")
}

// runGit executes a git command in the given directory
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
