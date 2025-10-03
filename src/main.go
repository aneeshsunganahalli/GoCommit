package main

import (
	"log"
	"os"

	"github.com/atotto/clipboard"
	"github.com/dfanso/commit-msg/src/chatgpt"
	"github.com/dfanso/commit-msg/src/gemini"
	"github.com/dfanso/commit-msg/src/grok"
	"github.com/dfanso/commit-msg/src/internal/display"
	"github.com/dfanso/commit-msg/src/internal/git"
	"github.com/dfanso/commit-msg/src/internal/stats"
	"github.com/dfanso/commit-msg/src/types"
	"github.com/pterm/pterm"
)

// main is the entry point of the commit message generator
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
	if !git.IsRepository(currentDir) {
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
	fileStats, err := stats.GetFileStatistics(&repoConfig)
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
	display.ShowFileStatistics(fileStats)

	if fileStats.TotalFiles == 0 {
		pterm.Warning.Println("No changes detected in the Git repository.")
		return
	}

	// Get the changes
	changes, err := git.GetChanges(&repoConfig)
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
	display.ShowCommitMessage(commitMsg)

	// Copy to clipboard
	err = clipboard.WriteAll(commitMsg)
	if err != nil {
		pterm.Warning.Printf("⚠️  Could not copy to clipboard: %v\n", err)
	} else {
		pterm.Success.Println("📋 Commit message copied to clipboard!")
	}

	pterm.Println()

	// Display changes preview
	display.ShowChangesPreview(fileStats)
}
