package types

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConfigJSONMarshalling(t *testing.T) {
	t.Parallel()

	cfg := Config{
		GrokAPI: "https://api.x.ai/v1/chat/completions",
		Repos: map[string]RepoConfig{
			"repo-a": {
				Path:    "/tmp/project",
				LastRun: "2024-06-01T12:00:00Z",
			},
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, "\"grok_api\"") {
		t.Fatalf("expected grok_api key in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, "\"repos\"") {
		t.Fatalf("expected repos key in JSON: %s", jsonStr)
	}
}

func TestCommitPromptContent(t *testing.T) {
	t.Parallel()

	requiredFragments := []string{
		"Starts with a verb",
		"Is clear and descriptive",
		"Here are the changes",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(CommitPrompt, fragment) {
			t.Fatalf("CommitPrompt missing fragment %q", fragment)
		}
	}
}

func TestBuildCommitPromptDefault(t *testing.T) {
	t.Parallel()

	changes := "diff --git a/main.go b/main.go"
	prompt := BuildCommitPrompt(changes, nil)

	if !strings.HasSuffix(prompt, changes) {
		t.Fatalf("expected prompt to end with changes, got %q", prompt)
	}

	if strings.Contains(prompt, "Additional instructions:") {
		t.Fatalf("expected no additional instructions block in default prompt")
	}
}

func TestBuildCommitPromptWithInstructions(t *testing.T) {
	t.Parallel()

	changes := "diff --git a/main.go b/main.go"
	options := &GenerationOptions{StyleInstruction: "Use a playful tone."}
	prompt := BuildCommitPrompt(changes, options)

	if !strings.Contains(prompt, "Additional instructions:") {
		t.Fatalf("expected prompt to contain additional instructions block")
	}

	if !strings.Contains(prompt, options.StyleInstruction) {
		t.Fatalf("expected prompt to include style instruction %q", options.StyleInstruction)
	}

	if !strings.HasSuffix(prompt, changes) {
		t.Fatalf("expected prompt to end with changes, got %q", prompt)
	}
}

func TestBuildCommitPromptWithAttempt(t *testing.T) {
	t.Parallel()

	changes := "diff --git a/main.go b/main.go"
	options := &GenerationOptions{Attempt: 3}
	prompt := BuildCommitPrompt(changes, options)

	if !strings.Contains(prompt, "Regeneration context:") {
		t.Fatalf("expected regeneration context section, got %q", prompt)
	}

	if !strings.Contains(prompt, "attempt #3") {
		t.Fatalf("expected attempt number to be mentioned, got %q", prompt)
	}

	if !strings.HasSuffix(prompt, changes) {
		t.Fatalf("expected prompt to end with changes, got %q", prompt)
	}
}
