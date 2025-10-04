package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dfanso/commit-msg/src/internal/scrubber"
	"github.com/dfanso/commit-msg/src/internal/utils"
	"github.com/dfanso/commit-msg/src/types"
)

// IsRepository checks if a directory is a git repository
func IsRepository(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// GetChanges retrieves all Git changes including staged, unstaged, and untracked files
func GetChanges(config *types.RepoConfig) (string, error) {
	var changes strings.Builder

	// 1. Check for unstaged changes
	cmd := exec.Command("git", "-C", config.Path, "diff", "--name-status")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %v", err)
	}

	if len(output) > 0 {
		changes.WriteString("Unstaged changes:\n")
		changes.WriteString(string(output))
		changes.WriteString("\n\n")

		// Get the content of these changes
		diffCmd := exec.Command("git", "-C", config.Path, "diff")
		diffOutput, err := diffCmd.Output()
		if err != nil {
			return "", fmt.Errorf("git diff content failed: %v", err)
		}

		changes.WriteString("Unstaged diff content:\n")
		changes.WriteString(string(diffOutput))
		changes.WriteString("\n\n")
	}

	// 2. Check for staged changes
	stagedCmd := exec.Command("git", "-C", config.Path, "diff", "--name-status", "--cached")
	stagedOutput, err := stagedCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached failed: %v", err)
	}

	if len(stagedOutput) > 0 {
		changes.WriteString("Staged changes:\n")
		changes.WriteString(string(stagedOutput))
		changes.WriteString("\n\n")

		// Get the content of these changes
		stagedDiffCmd := exec.Command("git", "-C", config.Path, "diff", "--cached")
		stagedDiffOutput, err := stagedDiffCmd.Output()
		if err != nil {
			return "", fmt.Errorf("git diff --cached content failed: %v", err)
		}

		changes.WriteString("Staged diff content:\n")
		changes.WriteString(string(stagedDiffOutput))
		changes.WriteString("\n\n")
	}

	// 3. Check for untracked files
	untrackedCmd := exec.Command("git", "-C", config.Path, "ls-files", "--others", "--exclude-standard")
	untrackedOutput, err := untrackedCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git ls-files failed: %v", err)
	}

	if len(untrackedOutput) > 0 {
		changes.WriteString("Untracked files:\n")
		changes.WriteString(string(untrackedOutput))
		changes.WriteString("\n\n")

		// Try to get content of untracked files (limited to text files and smaller size)
        untrackedFiles := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
        for _, file := range untrackedFiles {
            if file == "" {
                continue
            }

            fullPath := filepath.Join(config.Path, file)
            if utils.IsTextFile(fullPath) && utils.IsSmallFile(fullPath) {
                fileContent, err := os.ReadFile(fullPath)
                if err != nil {
                    // Log but don't fail - untracked file may have been deleted or is inaccessible
                    continue
                }
                changes.WriteString(fmt.Sprintf("Content of new file %s:\n", file))
                
                // Use special scrubbing for .env files
                if strings.HasSuffix(strings.ToLower(file), ".env") || 
                   strings.Contains(strings.ToLower(file), ".env.") {
                    changes.WriteString(scrubber.ScrubEnvFile(string(fileContent)))
                } else {
                    changes.WriteString(string(fileContent))
                }
                changes.WriteString("\n\n")
            }
        }
	}

	// 4. Get recent commits for context
	recentCommitsCmd := exec.Command("git", "-C", config.Path, "log", "--oneline", "-n", "3")
	recentCommitsOutput, err := recentCommitsCmd.Output()
	if err == nil && len(recentCommitsOutput) > 0 {
		changes.WriteString("Recent commits for context:\n")
		changes.WriteString(string(recentCommitsOutput))
		changes.WriteString("\n")
	}

	// Scrub sensitive data before returning
	scrubbedChanges := scrubber.ScrubDiff(changes.String())
	
	return scrubbedChanges, nil
}
