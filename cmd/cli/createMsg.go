package cmd

import (
	"os"

	"github.com/atotto/clipboard"
	"github.com/dfanso/commit-msg/cmd/cli/store"
	"github.com/dfanso/commit-msg/internal/chatgpt"
	"github.com/dfanso/commit-msg/internal/claude"
	"github.com/dfanso/commit-msg/internal/display"
	"github.com/dfanso/commit-msg/internal/gemini"
	"github.com/dfanso/commit-msg/internal/git"
	"github.com/dfanso/commit-msg/internal/grok"
	"github.com/dfanso/commit-msg/internal/ollama"
	"github.com/dfanso/commit-msg/internal/stats"
	"github.com/dfanso/commit-msg/pkg/types"
	"github.com/pterm/pterm"
)


func CreateCommitMsg () {
	
    // Validate COMMIT_LLM and required API keys
	useLLM,err := store.DefaultLLMKey()
	if err != nil {
		pterm.Error.Printf("No LLM configured. Run: commit llm setup\n")
		os.Exit(1)
	}

	commitLLM := useLLM.LLM
	apiKey := useLLM.APIKey


		// Get current directory
		currentDir, err := os.Getwd()
		if err != nil {
			pterm.Error.Printf("Failed to get current directory: %v\n", err)
			os.Exit(1)
		}

		// Check if current directory is a git repository
		if !git.IsRepository(currentDir) {
			pterm.Error.Printf("Current directory is not a Git repository: %s\n", currentDir)
			os.Exit(1)
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
			pterm.Error.Printf("Failed to get file statistics: %v\n", err)
			os.Exit(1)
		}

		// Display header
		pterm.DefaultHeader.WithFullWidth().
			WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
			WithTextStyle(pterm.NewStyle(pterm.FgLightWhite)).
			Println("Commit Message Generator")

		pterm.Println()

		// Display file statistics with icons
		display.ShowFileStatistics(fileStats)

		if fileStats.TotalFiles == 0 {
			pterm.Warning.Println("No changes detected in the Git repository.")
			pterm.Info.Println("Tips:")
			pterm.Info.Println("  - Stage your changes with: git add .")
			pterm.Info.Println("  - Check repository status with: git status")
			pterm.Info.Println("  - Make sure you're in the correct Git repository")
			return
		}

		// Get the changes
		changes, err := git.GetChanges(&repoConfig)
		if err != nil {
			pterm.Error.Printf("Failed to get Git changes: %v\n", err)
			os.Exit(1)
		}

		if len(changes) == 0 {
			pterm.Warning.Println("No changes detected in the Git repository.")
			pterm.Info.Println("Tips:")
			pterm.Info.Println("  - Stage your changes with: git add .")
			pterm.Info.Println("  - Check repository status with: git status")
			pterm.Info.Println("  - Make sure you're in the correct Git repository")
			return
		}

		pterm.Println()

		// Show generating spinner
		spinnerGenerating, err := pterm.DefaultSpinner.
			WithSequence("⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏").
			Start("Generating commit message with " + commitLLM + "...")
		if err != nil {
			pterm.Error.Printf("Failed to start spinner: %v\n", err)
			os.Exit(1)
		}

		var commitMsg string

		switch commitLLM {
			
		case "Gemini":
			commitMsg, err = gemini.GenerateCommitMessage(config, changes, apiKey)

		case "OpenAI":
			commitMsg, err = chatgpt.GenerateCommitMessage(config, changes, apiKey)
		
		case "Claude":
			commitMsg, err = claude.GenerateCommitMessage(config, changes, apiKey)
		case "Ollama":
			model := os.Getenv("OLLAMA_MODEL")
			if model == "" {
				model = "llama3:latest"
			}
			commitMsg, err = ollama.GenerateCommitMessage(config, changes, apiKey, model)
		default:
			commitMsg, err = grok.GenerateCommitMessage(config, changes, apiKey)
		}

		
		if err != nil {
			spinnerGenerating.Fail("Failed to generate commit message")
			switch commitLLM {
			case "Gemini":
				pterm.Error.Printf("Gemini API error. Check your GEMINI_API_KEY environment variable or run: commit llm setup\n")
			case "OpenAI":
				pterm.Error.Printf("OpenAI API error. Check your OPENAI_API_KEY environment variable or run: commit llm setup\n")
			case "Claude":
				pterm.Error.Printf("Claude API error. Check your CLAUDE_API_KEY environment variable or run: commit llm setup\n")
			case "Grok":
				pterm.Error.Printf("Grok API error. Check your GROK_API_KEY environment variable or run: commit llm setup\n")
			default:
				pterm.Error.Printf("LLM API error: %v\n", err)
			}
			os.Exit(1)
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