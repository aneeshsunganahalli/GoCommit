package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestGenerateCommitMessage(t *testing.T) {
	t.Parallel()

	t.Run("returns error for empty URL", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "", "model", nil)
		if err == nil {
			t.Fatal("expected error for empty URL")
		}
	})

	t.Run("returns error for empty changes", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCommitMessage(&types.Config{}, "", "http://localhost:11434/api/generate", "model", nil)
		if err == nil {
			t.Fatal("expected error for empty changes")
		}
	})

	t.Run("uses default model when none provided", func(t *testing.T) {
		t.Parallel()

		expectedResponse := OllamaResponse{
			Response: "feat: add new feature",
			Done:     true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST method, got %s", r.Method)
			}

			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected 'application/json', got %s", got)
			}

			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if req["model"] != "llama3:latest" {
				t.Fatalf("expected model 'llama3:latest', got %v", req["model"])
			}

			if req["stream"] != false {
				t.Fatalf("expected stream false, got %v", req["stream"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		t.Cleanup(server.Close)

		// Test with empty model to verify default is used
		_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("uses provided model", func(t *testing.T) {
		t.Parallel()

		expectedResponse := OllamaResponse{
			Response: "feat: add new feature",
			Done:     true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode request: %v", err)
			}

			if req["model"] != "custom-model" {
				t.Fatalf("expected model 'custom-model', got %v", req["model"])
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		t.Cleanup(server.Close)

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "custom-model", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestGenerateCommitMessageWithMockServer(t *testing.T) {
	t.Parallel()

	t.Run("successful response", func(t *testing.T) {
		t.Parallel()

		expectedResponse := OllamaResponse{
			Response: "feat: add new feature",
			Done:     true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		t.Cleanup(server.Close)

		result, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "llama3:latest", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != expectedResponse.Response {
			t.Fatalf("expected response '%s', got '%s'", expectedResponse.Response, result)
		}
	})

	t.Run("API error response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		t.Cleanup(server.Close)

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "llama3:latest", nil)
		if err == nil {
			t.Fatal("expected error for API error response")
		}

		if !strings.Contains(err.Error(), "status 500") {
			t.Fatalf("expected error to contain status 500, got: %v", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid json}`))
		}))
		t.Cleanup(server.Close)

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "llama3:latest", nil)
		if err == nil {
			t.Fatal("expected error for invalid JSON response")
		}

		if !strings.Contains(err.Error(), "failed to decode response") {
			t.Fatalf("expected error to contain 'failed to decode response', got: %v", err)
		}
	})

	t.Run("empty response content", func(t *testing.T) {
		t.Parallel()

		expectedResponse := OllamaResponse{
			Response: "",
			Done:     true,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		t.Cleanup(server.Close)

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "llama3:latest", nil)
		if err == nil {
			t.Fatal("expected error for empty response content")
		}

		if !strings.Contains(err.Error(), "received empty response") {
			t.Fatalf("expected error to contain 'received empty response', got: %v", err)
		}
	})

	t.Run("network error", func(t *testing.T) {
		t.Parallel()

		// Use an invalid URL to simulate network error
		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "http://localhost:99999/invalid", "llama3:latest", nil)
		if err == nil {
			t.Fatal("expected error for network error")
		}
	})
}

func TestGenerateCommitMessageIncludesStyleInstructions(t *testing.T) {
	t.Parallel()

	opts := &types.GenerationOptions{
		StyleInstruction: "Use a casual tone",
		Attempt:          2,
	}

	expectedResponse := OllamaResponse{
		Response: "feat: add new feature",
		Done:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		prompt, ok := req["prompt"].(string)
		if !ok {
			t.Fatal("expected prompt to be a string")
		}

		// Check that style instruction is included in the prompt
		if !strings.Contains(prompt, "Use a casual tone") {
			t.Fatal("expected style instruction to be included in prompt")
		}

		// Check that attempt context is included
		if !strings.Contains(prompt, "Regeneration context:") {
			t.Fatal("expected regeneration context to be included in prompt")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	t.Cleanup(server.Close)

	_, err := GenerateCommitMessage(&types.Config{}, "some changes", server.URL, "llama3:latest", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateCommitMessageWithLongChanges(t *testing.T) {
	t.Parallel()

	// Create a very long changes string
	longChanges := strings.Repeat("This is a test change. ", 1000)

	expectedResponse := OllamaResponse{
		Response: "feat: add new feature",
		Done:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		prompt, ok := req["prompt"].(string)
		if !ok {
			t.Fatal("expected prompt to be a string")
		}

		// Check that the long changes are included in the prompt
		if !strings.Contains(prompt, longChanges) {
			t.Fatal("expected long changes to be included in prompt")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	t.Cleanup(server.Close)

	_, err := GenerateCommitMessage(&types.Config{}, longChanges, server.URL, "llama3:latest", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateCommitMessageWithSpecialCharacters(t *testing.T) {
	t.Parallel()

	// Test with special characters in changes
	changes := `Added special characters: !@#$%^&*()_+-=[]{}|;':",./<>?
Also added unicode: ñáéíóú 🚀 🎉
And newlines:
Line 1
Line 2
Line 3`

	expectedResponse := OllamaResponse{
		Response: "feat: add new feature",
		Done:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		prompt, ok := req["prompt"].(string)
		if !ok {
			t.Fatal("expected prompt to be a string")
		}

		// Check that special characters are included in the prompt
		if !strings.Contains(prompt, "ñáéíóú 🚀 🎉") {
			t.Fatal("expected unicode characters to be included in prompt")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	t.Cleanup(server.Close)

	_, err := GenerateCommitMessage(&types.Config{}, changes, server.URL, "llama3:latest", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOllamaRequestSerialization(t *testing.T) {
	t.Parallel()

	req := OllamaRequest{
		Model:  "llama3:latest",
		Prompt: "test prompt",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var unmarshaled OllamaRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if unmarshaled.Model != req.Model {
		t.Fatalf("expected model %s, got %s", req.Model, unmarshaled.Model)
	}

	if unmarshaled.Prompt != req.Prompt {
		t.Fatalf("expected prompt %s, got %s", req.Prompt, unmarshaled.Prompt)
	}
}

func TestOllamaResponseDeserialization(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"response": "feat: add new feature",
		"done": true
	}`

	var resp OllamaResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Response != "feat: add new feature" {
		t.Fatalf("expected response 'feat: add new feature', got %s", resp.Response)
	}

	if resp.Done != true {
		t.Fatal("expected done to be true")
	}
}

func TestGenerateCommitMessageWithConfig(t *testing.T) {
	t.Parallel()

	config := &types.Config{
		// Add any config fields that might be relevant
	}

	expectedResponse := OllamaResponse{
		Response: "feat: add new feature",
		Done:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	t.Cleanup(server.Close)

	_, err := GenerateCommitMessage(config, "some changes", server.URL, "llama3:latest", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateCommitMessageWithNilConfig(t *testing.T) {
	t.Parallel()

	expectedResponse := OllamaResponse{
		Response: "feat: add new feature",
		Done:     true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	t.Cleanup(server.Close)

	_, err := GenerateCommitMessage(nil, "some changes", server.URL, "llama3:latest", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
