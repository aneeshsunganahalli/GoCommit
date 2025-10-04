package stats

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dfanso/commit-msg/internal/display"
	"github.com/dfanso/commit-msg/internal/utils"
	"github.com/dfanso/commit-msg/pkg/types"
)

// GetFileStatistics collects comprehensive file statistics from Git
func GetFileStatistics(config *types.RepoConfig) (*display.FileStatistics, error) {
	stats := &display.FileStatistics{
		StagedFiles:    []string{},
		UnstagedFiles:  []string{},
		UntrackedFiles: []string{},
	}

	// Get staged files
	stagedCmd := exec.Command("git", "-C", config.Path, "diff", "--name-only", "--cached")
	stagedOutput, err := stagedCmd.Output()
	if err == nil && len(stagedOutput) > 0 {
		stats.StagedFiles = strings.Split(strings.TrimSpace(string(stagedOutput)), "\n")
	}

	// Get unstaged files
	unstagedCmd := exec.Command("git", "-C", config.Path, "diff", "--name-only")
	unstagedOutput, err := unstagedCmd.Output()
	if err == nil && len(unstagedOutput) > 0 {
		stats.UnstagedFiles = strings.Split(strings.TrimSpace(string(unstagedOutput)), "\n")
	}

	// Get untracked files
	untrackedCmd := exec.Command("git", "-C", config.Path, "ls-files", "--others", "--exclude-standard")
	untrackedOutput, err := untrackedCmd.Output()
	if err == nil && len(untrackedOutput) > 0 {
		stats.UntrackedFiles = strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
	}

	// Filter empty strings
	stats.StagedFiles = utils.FilterEmpty(stats.StagedFiles)
	stats.UnstagedFiles = utils.FilterEmpty(stats.UnstagedFiles)
	stats.UntrackedFiles = utils.FilterEmpty(stats.UntrackedFiles)

	stats.TotalFiles = len(stats.StagedFiles) + len(stats.UnstagedFiles) + len(stats.UntrackedFiles)

	// Get line statistics from staged changes
	if len(stats.StagedFiles) > 0 {
		statCmd := exec.Command("git", "-C", config.Path, "diff", "--cached", "--numstat")
		statOutput, err := statCmd.Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(statOutput)), "\n")
			for _, line := range lines {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					if added := parts[0]; added != "-" {
						var addedNum int
						fmt.Sscanf(added, "%d", &addedNum)
						stats.LinesAdded += addedNum
					}
					if deleted := parts[1]; deleted != "-" {
						var deletedNum int
						fmt.Sscanf(deleted, "%d", &deletedNum)
						stats.LinesDeleted += deletedNum
					}
				}
			}
		}
	}

	return stats, nil
}
