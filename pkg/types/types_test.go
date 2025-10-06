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
