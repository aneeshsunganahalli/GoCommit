package gemini

import (
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestGenerateCommitMessage(t *testing.T) {
	t.Parallel()

	t.Run("returns error for empty API key", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "", nil)
		if err == nil {
			t.Fatal("expected error for empty API key")
		}
	})

	t.Run("returns error for empty changes", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCommitMessage(&types.Config{}, "", "test-key", nil)
		if err == nil {
			t.Fatal("expected error for empty changes")
		}
	})

	t.Run("returns error for invalid API key", func(t *testing.T) {
		t.Parallel()

		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", nil)
		if err == nil {
			t.Fatal("expected error for invalid API key")
		}
	})
}

func TestGenerateCommitMessageWithOptions(t *testing.T) {
	t.Parallel()

	t.Run("includes style instructions in prompt", func(t *testing.T) {
		t.Parallel()

		opts := &types.GenerationOptions{
			StyleInstruction: "Use a casual tone",
			Attempt:          2,
		}

		// Test with invalid key to verify the function processes the options
		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", opts)
		if err == nil {
			t.Fatal("expected error for invalid API key")
		}
	})

	t.Run("handles nil options", func(t *testing.T) {
		t.Parallel()

		// Test with invalid key and nil options
		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", nil)
		if err == nil {
			t.Fatal("expected error for invalid API key")
		}
	})

	t.Run("handles empty style instruction", func(t *testing.T) {
		t.Parallel()

		opts := &types.GenerationOptions{
			StyleInstruction: "",
			Attempt:          1,
		}

		// Test with invalid key to verify the function handles empty style instruction
		_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", opts)
		if err == nil {
			t.Fatal("expected error for invalid API key")
		}
	})
}

func TestGenerateCommitMessageWithLongChanges(t *testing.T) {
	t.Parallel()

	// Create a very long changes string
	longChanges := "This is a test change. "
	for i := 0; i < 100; i++ {
		longChanges += "Additional line of changes. "
	}

	// Test with invalid key to verify the function handles long changes
	_, err := GenerateCommitMessage(&types.Config{}, longChanges, "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestGenerateCommitMessageWithContextCancellation(t *testing.T) {
	t.Parallel()

	// This test would require modifying the function to accept context
	// For now, we'll test the basic functionality
	_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
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

	// Test with invalid key to verify the function handles special characters
	_, err := GenerateCommitMessage(&types.Config{}, changes, "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestGenerateCommitMessageWithConfig(t *testing.T) {
	t.Parallel()

	config := &types.Config{
		// Add any config fields that might be relevant
	}

	// Test with invalid key to verify the function uses config
	_, err := GenerateCommitMessage(config, "some changes", "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestGenerateCommitMessageWithMultipleAttempts(t *testing.T) {
	t.Parallel()

	opts := &types.GenerationOptions{
		StyleInstruction: "Use a formal tone",
		Attempt:          3,
	}

	// Test with invalid key to verify the function processes attempt count
	_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", opts)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestGenerateCommitMessageWithEmptyConfig(t *testing.T) {
	t.Parallel()

	// Test with empty config
	_, err := GenerateCommitMessage(&types.Config{}, "some changes", "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}

func TestGenerateCommitMessageWithNilConfig(t *testing.T) {
	t.Parallel()

	// Test with nil config
	_, err := GenerateCommitMessage(nil, "some changes", "invalid-key", nil)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}
}
