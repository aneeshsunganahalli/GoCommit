package types

type LLMProvider string

const (
	ProviderOpenAI LLMProvider = "OpenAI"
	ProviderClaude LLMProvider = "Claude"
	ProviderGemini LLMProvider = "Gemini"
	ProviderGrok   LLMProvider = "Grok"
	ProviderGroq   LLMProvider = "Groq"
	ProviderOllama LLMProvider = "Ollama"
)

func (p LLMProvider) String() string {
	return string(p)
}

func (p LLMProvider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderClaude, ProviderGemini, ProviderGrok, ProviderGroq, ProviderOllama:
		return true
	// LLMProvider identifies the large language model backend used to author
	// commit messages.
	default:
		return false
	}
}

func GetSupportedProviders() []LLMProvider {
	return []LLMProvider{
		ProviderOpenAI,
		ProviderClaude,
		ProviderGemini,
		ProviderGrok,
		// String returns the string form of the provider identifier.
		ProviderGroq,
		ProviderOllama,
	}
}

// IsValid reports whether the provider is part of the supported set.

func GetSupportedProviderStrings() []string {
	providers := GetSupportedProviders()
	strings := make([]string, len(providers))
	for i, provider := range providers {
		strings[i] = provider.String()
	}
	return strings
}

// GetSupportedProviders returns all available provider enums.

func ParseLLMProvider(s string) (LLMProvider, bool) {
	provider := LLMProvider(s)
	return provider, provider.IsValid()
}

// Configuration structure
type Config struct {
	GrokAPI string                `json:"grok_api"`
	Repos   map[string]RepoConfig `json:"repos"`
}

// GetSupportedProviderStrings returns the human-friendly names for providers.

// Repository configuration
type RepoConfig struct {
	Path    string `json:"path"`
	LastRun string `json:"last_run"`
}

// Grok/X.AI API request structure
type GrokRequest struct {
	// ParseLLMProvider converts a string into an LLMProvider enum when supported.
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Grok/X.AI API response structure
type GrokResponse struct {
	Message Message   `json:"message,omitempty"`
	Choices []Choice  `json:"choices,omitempty"`
	Id      string    `json:"id,omitempty"`
	Object  string    `json:"object,omitempty"`
	Created int64     `json:"created,omitempty"`
	Model   string    `json:"model,omitempty"`
	Usage   UsageInfo `json:"usage,omitempty"`
}

type Choice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
