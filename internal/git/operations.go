package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dfanso/commit-msg/internal/scrubber"
	"github.com/dfanso/commit-msg/internal/utils"
	"github.com/dfanso/commit-msg/pkg/types"
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

// parseGitStatusLine represents a parsed git status line
type parseGitStatusLine struct {
	status    string
	filenames []string
}

// parseGitNameStatus parses a single line from git diff --name-status output
// Handles various git status codes including rename (R) and copy (C) operations
func parseGitNameStatus(line string) parseGitStatusLine {
	if line == "" {
		return parseGitStatusLine{}
	}
	
	// Git uses tabs to separate fields in --name-status output
	parts := strings.Split(line, "\t")
	if len(parts) < 2 {
		return parseGitStatusLine{}
	}
	
	status := parts[0]
	
	// Handle rename/copy status codes (e.g., "R100", "C75")
	if len(status) > 1 && (status[0] == 'R' || status[0] == 'C') {
		// For rename/copy, we expect: "R100\toldname\tnewname" or "C75\toldname\tnewname"
		if len(parts) >= 3 {
			// For renames/copies, both old and new filenames need to be checked
			oldFile := parts[1]
			newFile := parts[2]
			return parseGitStatusLine{
				status:    status,
				filenames: []string{oldFile, newFile},
			}
		}
	}
	
	// Handle regular status codes (M, A, D, etc.)
	filename := parts[1]
	return parseGitStatusLine{
		status:    status,
		filenames: []string{filename},
	}
}

// processGitStatusOutput processes git diff --name-status output and returns filtered results
func processGitStatusOutput(nameStatusOutput string, returnFilenames bool) ([]string, []string) {
	if nameStatusOutput == "" {
		return nil, nil
	}
	
	lines := strings.Split(strings.TrimSpace(nameStatusOutput), "\n")
	var filteredLines []string
	var nonBinaryFiles []string
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		parsed := parseGitNameStatus(line)
		if len(parsed.filenames) == 0 {
			continue
		}
		
		// Check if any of the filenames are binary
		hasBinaryFile := false
		for _, filename := range parsed.filenames {
			if utils.IsBinaryFile(filename) {
				hasBinaryFile = true
				break
			}
		}
		
		// If no binary files found, include this line/files
		if !hasBinaryFile {
			filteredLines = append(filteredLines, line)
			if returnFilenames {
				nonBinaryFiles = append(nonBinaryFiles, parsed.filenames...)
			}
		}
	}
	
	return filteredLines, nonBinaryFiles
}

// filterBinaryFiles filters out binary files from git diff --name-status output
func filterBinaryFiles(nameStatusOutput string) string {
	filteredLines, _ := processGitStatusOutput(nameStatusOutput, false)
	
	if len(filteredLines) == 0 {
		return ""
	}
	
	return strings.Join(filteredLines, "\n")
}

// extractNonBinaryFiles extracts non-binary filenames from git diff --name-status output
func extractNonBinaryFiles(nameStatusOutput string) []string {
	_, nonBinaryFiles := processGitStatusOutput(nameStatusOutput, true)
	return nonBinaryFiles
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
		// Filter out binary files from the name-status output
		filteredOutput := filterBinaryFiles(string(output))
		
		if filteredOutput != "" {
			changes.WriteString("Unstaged changes:\n")
			changes.WriteString(filteredOutput)
			changes.WriteString("\n\n")

			// Get the content of these changes (only for non-binary files)
			nonBinaryFiles := extractNonBinaryFiles(string(output))
			if len(nonBinaryFiles) > 0 {
				diffCmd := exec.Command("git", "-C", config.Path, "diff", "--")
				diffCmd.Args = append(diffCmd.Args, nonBinaryFiles...)
				diffOutput, err := diffCmd.Output()
				if err != nil {
					return "", fmt.Errorf("git diff content failed: %v", err)
				}

				changes.WriteString("Unstaged diff content:\n")
				changes.WriteString(string(diffOutput))
				changes.WriteString("\n\n")
			}
		}
	}

	// 2. Check for staged changes
	stagedCmd := exec.Command("git", "-C", config.Path, "diff", "--name-status", "--cached")
	stagedOutput, err := stagedCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff --cached failed: %v", err)
	}

	if len(stagedOutput) > 0 {
		// Filter out binary files from the staged changes
		filteredStagedOutput := filterBinaryFiles(string(stagedOutput))
		
		if filteredStagedOutput != "" {
			changes.WriteString("Staged changes:\n")
			changes.WriteString(filteredStagedOutput)
			changes.WriteString("\n\n")

			// Get the content of these changes (only for non-binary files)
			nonBinaryStagedFiles := extractNonBinaryFiles(string(stagedOutput))
			if len(nonBinaryStagedFiles) > 0 {
				stagedDiffCmd := exec.Command("git", "-C", config.Path, "diff", "--cached", "--")
				stagedDiffCmd.Args = append(stagedDiffCmd.Args, nonBinaryStagedFiles...)
				stagedDiffOutput, err := stagedDiffCmd.Output()
				if err != nil {
					return "", fmt.Errorf("git diff --cached content failed: %v", err)
				}

				changes.WriteString("Staged diff content:\n")
				changes.WriteString(string(stagedDiffOutput))
				changes.WriteString("\n\n")
			}
		}
	}

	// 3. Check for untracked files
	untrackedCmd := exec.Command("git", "-C", config.Path, "ls-files", "--others", "--exclude-standard")
	untrackedOutput, err := untrackedCmd.Output()
	if err != nil {
		return "", fmt.Errorf("git ls-files failed: %v", err)
	}

	if len(untrackedOutput) > 0 {
		// Filter out binary files from untracked files
		untrackedFiles := strings.Split(strings.TrimSpace(string(untrackedOutput)), "\n")
		var nonBinaryUntrackedFiles []string
		
		for _, file := range untrackedFiles {
			if file == "" {
				continue
			}
			if !utils.IsBinaryFile(file) {
				nonBinaryUntrackedFiles = append(nonBinaryUntrackedFiles, file)
			}
		}
		
		if len(nonBinaryUntrackedFiles) > 0 {
			changes.WriteString("Untracked files:\n")
			changes.WriteString(strings.Join(nonBinaryUntrackedFiles, "\n"))
			changes.WriteString("\n\n")

			// Try to get content of untracked files (limited to text files and smaller size)
			for _, file := range nonBinaryUntrackedFiles {
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
