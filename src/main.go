package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/dfanso/commit-msg/src/chatgpt"
	"github.com/dfanso/commit-msg/src/gemini"
	"github.com/dfanso/commit-msg/src/grok"
	"github.com/dfanso/commit-msg/src/types"
	"github.com/pterm/pterm"
)

// Normalize path to handle both forward and backslashes
func normalizePath(path string) string {
	// Replace backslashes with forward slashes
	normalized := strings.ReplaceAll(path, "\\", "/")
	// Remove any trailing slash
	normalized = strings.TrimSuffix(normalized, "/")
	return normalized
}

// Main function
func main() {
	// Get API key from environment variables
	var apiKey string
	if os.Getenv("COMMIT_LLM") == "google" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			log.Fatalf("GOOGLE_API_KEY is not set")
		}
	} else if os.Getenv("COMMIT_LLM") == "grok" {
		apiKey = os.Getenv("GROK_API_KEY")
		if apiKey == "" {
			log.Fatalf("GROK_API_KEY is not set")
		}
	} else if os.Getenv("COMMIT_LLM") == "chatgpt" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatalf("OPENAI_API_KEY is not set")
		}
	} else {
		log.Fatalf("Invalid COMMIT_LLM value: %s", os.Getenv("COMMIT_LLM"))
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Check if current directory is a git repository
	if !isGitRepository(currentDir) {
		log.Fatalf("Current directory is not a Git repository: %s", currentDir)
	}

	// Create a minimal config for the API
	config := &types.Config{
		GrokAPI: "https://api.x.ai/v1/chat/completions",
	}

	// Create a repo config for the current directory
	repoConfig := types.RepoConfig{
		Path: currentDir,
	}

	// Get file statistics before fetching changes
	fileStats, err := getFileStatistics(&repoConfig)
	if err != nil {
		log.Fatalf("Failed to get file statistics: %v", err)
	}

	// Display header
	pterm.DefaultHeader.WithFullWidth().
		WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
		Println("🚀 Commit Message Generator")

	pterm.Println()

	// Display file statistics with icons
	displayFileStatistics(fileStats)

	if fileStats.TotalFiles == 0 {
		pterm.Warning.Println("No changes detected in the Git repository.")
		return
	}

	// Get the changes
	changes, err := getGitChanges(&repoConfig)
	if err != nil {
		log.Fatalf("Failed to get Git changes: %v", err)
	}

	if len(changes) == 0 {
		pterm.Warning.Println("No changes detected in the Git repository.")
		return
	}

	pterm.Println()

	// Show generating spinner
	spinnerGenerating, err := pterm.DefaultSpinner.
		WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
		Start("🤖 Generating commit message...")
	if err != nil {
		log.Fatalf("Failed to start spinner: %v", err)
	}

	var commitMsg string
	if os.Getenv("COMMIT_LLM") == "google" {
		commitMsg, err = gemini.GenerateCommitMessage(config, changes, apiKey)
	} else if os.Getenv("COMMIT_LLM") == "chatgpt" {
		commitMsg, err = chatgpt.GenerateCommitMessage(config, changes, apiKey)
	} else {
		commitMsg, err = grok.GenerateCommitMessage(config, changes, apiKey)
	}

	if err != nil {
		spinnerGenerating.Fail("Failed to generate commit message")
		log.Fatalf("Error: %v", err)
	}

	spinnerGenerating.Success("✅ Commit message generated successfully!")

	pterm.Println()

	// Display the commit message in a styled panel
	displayCommitMessage(commitMsg)

	// Copy to clipboard
	err = clipboard.WriteAll(commitMsg)
	if err != nil {
		pterm.Warning.Printf("⚠️  Could not copy to clipboard: %v\n", err)
	} else {
		pterm.Success.Println("📋 Commit message copied to clipboard!")
	}

	pterm.Println()

	// Display changes preview
	displayChangesPreview(fileStats)
}

// Check if directory is a git repository
func isGitRepository(path string) bool {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// Get changes using Git
func getGitChanges(config *types.RepoConfig) (string, error) {
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
			if isTextFile(fullPath) && isSmallFile(fullPath) {
				fileContent, err := os.ReadFile(fullPath)
				if err == nil {
					changes.WriteString(fmt.Sprintf("Content of new file %s:\n", file))
					changes.WriteString(string(fileContent))
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

	return changes.String(), nil
}

// Check if a file is likely to be a text file
func isTextFile(filename string) bool {
	// List of common text file extensions
	textExtensions := []string{".txt", ".md", ".go", ".js", ".py", ".java", ".c", ".cpp", ".h", ".html", ".css", ".json", ".xml", ".yaml", ".yml", ".sh", ".bash", ".ts", ".tsx", ".jsx", ".php", ".rb", ".rs", ".dart"}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, textExt := range textExtensions {
		if ext == textExt {
			return true
		}
	}

	return false
}

// Check if a file is small enough to include in context
func isSmallFile(filename string) bool {
	const maxSize = 10 * 1024 // 10KB max

	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return info.Size() <= maxSize
}

// FileStatistics holds statistics about changed files
type FileStatistics struct {
	StagedFiles    []string
	UnstagedFiles  []string
	UntrackedFiles []string
	TotalFiles     int
	LinesAdded     int
	LinesDeleted   int
}

// Get file statistics for display
func getFileStatistics(config *types.RepoConfig) (*FileStatistics, error) {
	stats := &FileStatistics{
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
	stats.StagedFiles = filterEmpty(stats.StagedFiles)
	stats.UnstagedFiles = filterEmpty(stats.UnstagedFiles)
	stats.UntrackedFiles = filterEmpty(stats.UntrackedFiles)

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

// Filter empty strings from slice
func filterEmpty(slice []string) []string {
	filtered := []string{}
	for _, s := range slice {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// Display file statistics with colored output
func displayFileStatistics(stats *FileStatistics) {
	pterm.DefaultSection.Println("📊 Changes Summary")

	// Create bullet list items
	bulletItems := []pterm.BulletListItem{}

	if len(stats.StagedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Green(fmt.Sprintf("✅ Staged files: %d", len(stats.StagedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgGreen),
			BulletStyle: pterm.NewStyle(pterm.FgGreen),
		})
		for i, file := range stats.StagedFiles {
			if i < 5 { // Show first 5 files
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.StagedFiles) > 5 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.StagedFiles)-5)),
			})
		}
	}

	if len(stats.UnstagedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Yellow(fmt.Sprintf("⚠️  Unstaged files: %d", len(stats.UnstagedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgYellow),
			BulletStyle: pterm.NewStyle(pterm.FgYellow),
		})
		for i, file := range stats.UnstagedFiles {
			if i < 3 {
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.UnstagedFiles) > 3 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.UnstagedFiles)-3)),
			})
		}
	}

	if len(stats.UntrackedFiles) > 0 {
		bulletItems = append(bulletItems, pterm.BulletListItem{
			Level:       0,
			Text:        pterm.Cyan(fmt.Sprintf("📝 Untracked files: %d", len(stats.UntrackedFiles))),
			TextStyle:   pterm.NewStyle(pterm.FgCyan),
			BulletStyle: pterm.NewStyle(pterm.FgCyan),
		})
		for i, file := range stats.UntrackedFiles {
			if i < 3 {
				bulletItems = append(bulletItems, pterm.BulletListItem{
					Level: 1,
					Text:  file,
				})
			}
		}
		if len(stats.UntrackedFiles) > 3 {
			bulletItems = append(bulletItems, pterm.BulletListItem{
				Level: 1,
				Text:  pterm.Gray(fmt.Sprintf("... and %d more", len(stats.UntrackedFiles)-3)),
			})
		}
	}

	pterm.DefaultBulletList.WithItems(bulletItems).Render()
}

// Display commit message in a styled panel
func displayCommitMessage(message string) {
	pterm.DefaultSection.Println("📝 Generated Commit Message")

	// Create a panel with the commit message
	panel := pterm.DefaultBox.
		WithTitle("Commit Message").
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgLightGreen)).
		WithHorizontalString("─").
		WithVerticalString("│").
		WithTopLeftCornerString("┌").
		WithTopRightCornerString("┐").
		WithBottomLeftCornerString("└").
		WithBottomRightCornerString("┘")

	panel.Println(pterm.LightGreen(message))
}

// Display changes preview
func displayChangesPreview(stats *FileStatistics) {
	pterm.DefaultSection.Println("🔍 Changes Preview")

	// Create info boxes
	if stats.LinesAdded > 0 || stats.LinesDeleted > 0 {
		infoData := [][]string{
			{"Lines Added", pterm.Green(fmt.Sprintf("+%d", stats.LinesAdded))},
			{"Lines Deleted", pterm.Red(fmt.Sprintf("-%d", stats.LinesDeleted))},
			{"Total Files", pterm.Cyan(fmt.Sprintf("%d", stats.TotalFiles))},
		}

		pterm.DefaultTable.WithHasHeader(false).WithData(infoData).Render()
	} else {
		pterm.Info.Println("No line statistics available for unstaged changes")
	}
}
