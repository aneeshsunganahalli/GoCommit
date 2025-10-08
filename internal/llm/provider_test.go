package llm

import (
	"context"
	"errors"
	"testing"

	"github.com/dfanso/commit-msg/pkg/types"
)

func TestNewProviderRequiresCredential(t *testing.T) {
	remoteProviders := []types.LLMProvider{
		types.ProviderOpenAI,
		types.ProviderClaude,
		types.ProviderGemini,
		types.ProviderGrok,
		types.ProviderGroq,
	}

	for _, provider := range remoteProviders {
		provider := provider
		t.Run(provider.String(), func(t *testing.T) {
			switch provider {
			case types.ProviderOpenAI:
				t.Setenv("OPENAI_API_KEY", "")
			case types.ProviderClaude:
				t.Setenv("CLAUDE_API_KEY", "")
			case types.ProviderGemini:
				t.Setenv("GEMINI_API_KEY", "")
			case types.ProviderGrok:
				t.Setenv("GROK_API_KEY", "")
			case types.ProviderGroq:
				t.Setenv("GROQ_API_KEY", "")
			}

			_, err := NewProvider(provider, ProviderOptions{})
			if !errors.Is(err, ErrMissingCredential) {
				t.Fatalf("expected ErrMissingCredential for %s, got %v", provider, err)
			}
		})
	}
}

func TestNewProviderUsesEnvFallback(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "env-key")
	provider, err := NewProvider(types.ProviderOpenAI, ProviderOptions{})
	if err != nil {
		t.Fatalf("expected no error using env fallback, got %v", err)
	}

	p, ok := provider.(*openAIProvider)
	if !ok {
		t.Fatalf("expected *openAIProvider, got %T", provider)
	}

	if p.apiKey != "env-key" {
		t.Fatalf("expected api key to come from env, got %q", p.apiKey)
	}
}

func TestNewProviderUnsupported(t *testing.T) {
	_, err := NewProvider(types.LLMProvider("unknown"), ProviderOptions{})
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
}

func TestNewProviderOllamaDefaults(t *testing.T) {
	t.Setenv("OLLAMA_URL", "")
	t.Setenv("OLLAMA_MODEL", "")

	provider, err := NewProvider(types.ProviderOllama, ProviderOptions{})
	if err != nil {
		t.Fatalf("expected no error for ollama provider, got %v", err)
	}

	p, ok := provider.(*ollamaProvider)
	if !ok {
		t.Fatalf("expected *ollamaProvider, got %T", provider)
	}

	if p.url == "" {
		t.Fatalf("expected default URL to be set")
	}

	if p.model == "" {
		t.Fatalf("expected default model to be set")
	}
}

func TestRegisterFactoryOverrides(t *testing.T) {
	factoryMu.Lock()
	original := factories[types.ProviderOpenAI]
	factoryMu.Unlock()

	t.Cleanup(func() {
		RegisterFactory(types.ProviderOpenAI, original)
	})

	called := 0
	RegisterFactory(types.ProviderOpenAI, func(opts ProviderOptions) (Provider, error) {
		called++
		return fakeProvider{name: types.ProviderOpenAI}, nil
	})

	provider, err := NewProvider(types.ProviderOpenAI, ProviderOptions{})
	if err != nil {
		t.Fatalf("expected no error from overridden factory, got %v", err)
	}

	if called != 1 {
		t.Fatalf("expected overridden factory to be called once, got %d", called)
	}

	if provider.Name() != types.ProviderOpenAI {
		t.Fatalf("expected provider name %s, got %s", types.ProviderOpenAI, provider.Name())
	}
}

type fakeProvider struct {
	name types.LLMProvider
}

func (f fakeProvider) Name() types.LLMProvider {
	return f.name
}

func (f fakeProvider) Generate(context.Context, string, *types.GenerationOptions) (string, error) {
	return "", nil
}
