package groq

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

type capturedRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

func withTestServer(t *testing.T, handler http.HandlerFunc, fn func()) {
	t.Helper()

	t.Setenv("GROQ_API_URL", "")
	t.Setenv("GROQ_MODEL", "")

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	prevURL := baseURL
	prevClient := httpClient

	baseURL = srv.URL
	httpClient = srv.Client()

	t.Cleanup(func() {
		baseURL = prevURL
		httpClient = prevClient
	})

	fn()
}

func TestGenerateCommitMessageSuccess(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %s", got)
		}

		var payload capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if payload.Model != "llama-3.3-70b-versatile" {
			t.Fatalf("unexpected model: %s", payload.Model)
		}

		if len(payload.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(payload.Messages))
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "Feat: add groq provider"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}, func() {
		msg, err := GenerateCommitMessage(&types.Config{}, "diff", "test-key", nil)
		if err != nil {
			t.Fatalf("GenerateCommitMessage returned error: %v", err)
		}

		expected := "Feat: add groq provider"
		if msg != expected {
			t.Fatalf("expected %q, got %q", expected, msg)
		}
	})
}

func TestGenerateCommitMessageNonOK(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bad things"}`, http.StatusBadGateway)
	}, func() {
		_, err := GenerateCommitMessage(&types.Config{}, "changes", "key", nil)
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})
}

func TestGenerateCommitMessageEmptyChanges(t *testing.T) {
	t.Setenv("GROQ_MODEL", "")
	t.Setenv("GROQ_API_URL", "")

	if _, err := GenerateCommitMessage(&types.Config{}, "", "key", nil); err == nil {
		t.Fatal("expected error for empty changes")
	}
}

func TestGenerateCommitMessageIncludesStyleInstruction(t *testing.T) {
	recorded := ""

	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var payload capturedRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		recorded = payload.Messages[1].Content

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "feat: add style support"}},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}, func() {
		opts := &types.GenerationOptions{StyleInstruction: "Use a casual tone.", Attempt: 2}
		if _, err := GenerateCommitMessage(&types.Config{}, "diff", "key", opts); err != nil {
			t.Fatalf("GenerateCommitMessage returned error: %v", err)
		}
	})

	if !strings.Contains(recorded, "Additional instructions:") {
		t.Fatalf("expected request payload to contain additional instructions, got: %q", recorded)
	}
	if !strings.Contains(recorded, "Use a casual tone.") {
		t.Fatalf("expected request payload to contain custom instruction, got: %q", recorded)
	}
	if !strings.Contains(recorded, "Regeneration context:") {
		t.Fatalf("expected request payload to contain regeneration context, got: %q", recorded)
	}
}
