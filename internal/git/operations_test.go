package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestIsRepository(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	t.Run("returns true for initialized repo", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()

		cmd := exec.Command("git", "init", dir)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to init git repo: %v: %s", err, string(output))
		}

		if got := IsRepository(dir); !got {
			t.Fatalf("IsRepository(%q) = false, want true", dir)
		}
	})

	t.Run("returns false for non repo", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		if got := IsRepository(dir); got {
			t.Fatalf("IsRepository(%q) = true, want false", dir)
		}
	})

	t.Run("returns false for missing path", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		missing := filepath.Join(dir, "does-not-exist")
		if got := IsRepository(missing); got {
			t.Fatalf("IsRepository(%q) = true, want false", missing)
		}
	})
}

func TestGetChangesIncludesSections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration-style test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git executable not available")
	}

	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	tracked := filepath.Join(dir, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("first version\n"), 0o644); err != nil {
		t.Fatalf("failed to write tracked file: %v", err)
	}
	runGit(t, dir, "add", "tracked.txt")
	runGit(t, dir, "commit", "-m", "initial commit")

	if err := os.WriteFile(tracked, []byte("updated version\n"), 0o644); err != nil {
		t.Fatalf("failed to modify tracked file: %v", err)
	}

	staged := filepath.Join(dir, "staged.txt")
	if err := os.WriteFile(staged, []byte("staged content\n"), 0o644); err != nil {
		t.Fatalf("failed to write staged file: %v", err)
	}
	runGit(t, dir, "add", "staged.txt")

	untracked := filepath.Join(dir, "new.txt")
	if err := os.WriteFile(untracked, []byte("brand new file\n"), 0o644); err != nil {
		t.Fatalf("failed to write untracked file: %v", err)
	}

	output, err := GetChanges(&types.RepoConfig{Path: dir})
	if err != nil {
		t.Fatalf("GetChanges returned error: %v", err)
	}

	fragments := []string{
		"Unstaged changes:",
		"Staged changes:",
		"Untracked files:",
		"Content of new file new.txt:",
		"Recent commits for context:",
		"brand new file",
	}

	for _, fragment := range fragments {
		if !strings.Contains(output, fragment) {
			t.Fatalf("output missing fragment %q\noutput: %s", fragment, output)
		}
	}
}

func TestGetChangesErrorsOutsideRepo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if _, err := GetChanges(&types.RepoConfig{Path: dir}); err == nil {
		t.Fatal("expected error for directory without git repository")
	}
}

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
