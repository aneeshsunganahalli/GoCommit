package llm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dfanso/commit-msg/internal/chatgpt"
	"github.com/dfanso/commit-msg/internal/claude"
	"github.com/dfanso/commit-msg/internal/gemini"
	"github.com/dfanso/commit-msg/internal/grok"
	"github.com/dfanso/commit-msg/internal/groq"
	"github.com/dfanso/commit-msg/internal/ollama"
	"github.com/dfanso/commit-msg/pkg/types"
)

// ErrMissingCredential signals that a provider requires a credential such as an API key or URL.
var ErrMissingCredential = errors.New("llm: missing credential")

// Provider declares the behaviour required by commit-msg to talk to an LLM backend.
type Provider interface {
	// Name returns the LLM provider identifier this instance represents.
	Name() types.LLMProvider
	// Generate requests a commit message for the supplied repository changes.
	Generate(ctx context.Context, changes string, opts *types.GenerationOptions) (string, error)
}

// ProviderOptions captures the data needed to construct a provider instance.
type ProviderOptions struct {
	Credential string
	Config     *types.Config
}

// Factory describes a function capable of building a Provider.
type Factory func(ProviderOptions) (Provider, error)

var (
	factoryMu sync.RWMutex
	factories = map[types.LLMProvider]Factory{
		types.ProviderOpenAI: newOpenAIProvider,
		types.ProviderClaude: newClaudeProvider,
		types.ProviderGemini: newGeminiProvider,
		types.ProviderGrok:   newGrokProvider,
		types.ProviderGroq:   newGroqProvider,
		types.ProviderOllama: newOllamaProvider,
	}
)

// RegisterFactory allows callers (primarily tests) to override or extend provider creation logic.
func RegisterFactory(name types.LLMProvider, factory Factory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	factories[name] = factory
}

// NewProvider returns a concrete Provider implementation for the requested name.
func NewProvider(name types.LLMProvider, opts ProviderOptions) (Provider, error) {
	factoryMu.RLock()
	factory, ok := factories[name]
	factoryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("llm: unsupported provider %s", name)
	}

	opts.Config = ensureConfig(opts.Config)
	return factory(opts)
}

type missingCredentialError struct {
	provider types.LLMProvider
}

func (e *missingCredentialError) Error() string {
	return fmt.Sprintf("%s credential is required", e.provider.String())
}

func (e *missingCredentialError) Unwrap() error {
	return ErrMissingCredential
}

func newMissingCredentialError(provider types.LLMProvider) error {
	return &missingCredentialError{provider: provider}
}

func ensureConfig(cfg *types.Config) *types.Config {
	if cfg != nil {
		return cfg
	}
	return &types.Config{}
}

// --- Provider implementations ------------------------------------------------

type openAIProvider struct {
	apiKey string
	config *types.Config
}

func newOpenAIProvider(opts ProviderOptions) (Provider, error) {
	key := strings.TrimSpace(opts.Credential)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	}
	if key == "" {
		return nil, newMissingCredentialError(types.ProviderOpenAI)
	}
	return &openAIProvider{apiKey: key, config: opts.Config}, nil
}

func (p *openAIProvider) Name() types.LLMProvider {
	return types.ProviderOpenAI
}

func (p *openAIProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return chatgpt.GenerateCommitMessage(p.config, changes, p.apiKey, opts)
}

type claudeProvider struct {
	apiKey string
	config *types.Config
}

func newClaudeProvider(opts ProviderOptions) (Provider, error) {
	key := strings.TrimSpace(opts.Credential)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("CLAUDE_API_KEY"))
	}
	if key == "" {
		return nil, newMissingCredentialError(types.ProviderClaude)
	}
	return &claudeProvider{apiKey: key, config: opts.Config}, nil
}

func (p *claudeProvider) Name() types.LLMProvider {
	return types.ProviderClaude
}

func (p *claudeProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return claude.GenerateCommitMessage(p.config, changes, p.apiKey, opts)
}

type geminiProvider struct {
	apiKey string
	config *types.Config
}

func newGeminiProvider(opts ProviderOptions) (Provider, error) {
	key := strings.TrimSpace(opts.Credential)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	}
	if key == "" {
		return nil, newMissingCredentialError(types.ProviderGemini)
	}
	return &geminiProvider{apiKey: key, config: opts.Config}, nil
}

func (p *geminiProvider) Name() types.LLMProvider {
	return types.ProviderGemini
}

func (p *geminiProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return gemini.GenerateCommitMessage(p.config, changes, p.apiKey, opts)
}

type grokProvider struct {
	apiKey string
	config *types.Config
}

func newGrokProvider(opts ProviderOptions) (Provider, error) {
	key := strings.TrimSpace(opts.Credential)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GROK_API_KEY"))
	}
	if key == "" {
		return nil, newMissingCredentialError(types.ProviderGrok)
	}
	return &grokProvider{apiKey: key, config: opts.Config}, nil
}

func (p *grokProvider) Name() types.LLMProvider {
	return types.ProviderGrok
}

func (p *grokProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return grok.GenerateCommitMessage(p.config, changes, p.apiKey, opts)
}

type groqProvider struct {
	apiKey string
	config *types.Config
}

func newGroqProvider(opts ProviderOptions) (Provider, error) {
	key := strings.TrimSpace(opts.Credential)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("GROQ_API_KEY"))
	}
	if key == "" {
		return nil, newMissingCredentialError(types.ProviderGroq)
	}
	return &groqProvider{apiKey: key, config: opts.Config}, nil
}

func (p *groqProvider) Name() types.LLMProvider {
	return types.ProviderGroq
}

func (p *groqProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return groq.GenerateCommitMessage(p.config, changes, p.apiKey, opts)
}

type ollamaProvider struct {
	url    string
	model  string
	config *types.Config
}

func newOllamaProvider(opts ProviderOptions) (Provider, error) {
	url := strings.TrimSpace(opts.Credential)
	if url == "" {
		url = strings.TrimSpace(os.Getenv("OLLAMA_URL"))
		if url == "" {
			url = "http://localhost:11434/api/generate"
		}
	}

	model := strings.TrimSpace(os.Getenv("OLLAMA_MODEL"))
	if model == "" {
		model = "llama3.1"
	}

	return &ollamaProvider{url: url, model: model, config: opts.Config}, nil
}

func (p *ollamaProvider) Name() types.LLMProvider {
	return types.ProviderOllama
}

func (p *ollamaProvider) Generate(_ context.Context, changes string, opts *types.GenerationOptions) (string, error) {
	return ollama.GenerateCommitMessage(p.config, changes, p.url, p.model, opts)
}
