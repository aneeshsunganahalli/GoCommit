package types

// LLMProvider identifies the large language model backend used to author
// commit messages.
type LLMProvider string

const (
	ProviderOpenAI LLMProvider = "OpenAI"
	ProviderClaude LLMProvider = "Claude"
	ProviderGemini LLMProvider = "Gemini"
	ProviderGrok   LLMProvider = "Grok"
	ProviderGroq   LLMProvider = "Groq"
	ProviderOllama LLMProvider = "Ollama"
)

// String returns the provider identifier as a plain string.
func (p LLMProvider) String() string {
	return string(p)
}

// IsValid reports whether the provider is part of the supported set.
func (p LLMProvider) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderClaude, ProviderGemini, ProviderGrok, ProviderGroq, ProviderOllama:
		return true
	default:
		return false
	}
}

// GetSupportedProviders returns all available provider enums.
func GetSupportedProviders() []LLMProvider {
	return []LLMProvider{
		ProviderOpenAI,
		ProviderClaude,
		ProviderGemini,
		ProviderGrok,
		ProviderGroq,
		ProviderOllama,
	}
}

// GetSupportedProviderStrings returns the human-friendly names for providers.
func GetSupportedProviderStrings() []string {
	providers := GetSupportedProviders()
	strings := make([]string, len(providers))
	for i, provider := range providers {
		strings[i] = provider.String()
	}
	return strings
}

// ParseLLMProvider converts a string into an LLMProvider enum when supported.
func ParseLLMProvider(s string) (LLMProvider, bool) {
	provider := LLMProvider(s)
	return provider, provider.IsValid()
}

// Config stores CLI-level configuration including named repositories.
type Config struct {
	GrokAPI string                `json:"grok_api"`
	Repos   map[string]RepoConfig `json:"repos"`
}

// RepoConfig tracks metadata for a configured Git repository.
type RepoConfig struct {
	Path    string `json:"path"`
	LastRun string `json:"last_run"`
}

// GrokRequest represents a chat completion request sent to X.AI's API.
type GrokRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature"`
}

// Message captures the role/content pairs exchanged with Grok.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GrokResponse contains the relevant fields parsed from X.AI responses.
type GrokResponse struct {
	Message Message   `json:"message,omitempty"`
	Choices []Choice  `json:"choices,omitempty"`
	Id      string    `json:"id,omitempty"`
	Object  string    `json:"object,omitempty"`
	Created int64     `json:"created,omitempty"`
	Model   string    `json:"model,omitempty"`
	Usage   UsageInfo `json:"usage,omitempty"`
}

// Choice details a single response option returned by Grok.
type Choice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
}

// UsageInfo reports token usage statistics from Grok responses.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
