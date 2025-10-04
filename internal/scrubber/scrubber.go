package scrubber

import (
	"regexp"
	"strings"
)

// SensitivePattern represents a regex pattern to detect sensitive data
type SensitivePattern struct {
	Name    string
	Pattern *regexp.Regexp
	Redact  string
}

var (
	// Common patterns for sensitive data
	sensitivePatterns = []SensitivePattern{
		// API Keys and Tokens
		{
			Name:    "Generic API Key",
			Pattern: regexp.MustCompile(`(?i)(api[_-]?key|apikey|api[_-]?token)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_API_KEY]\"",
		},
		{
			Name:    "Bearer Token",
			Pattern: regexp.MustCompile(`(?i)(bearer\s+)([a-zA-Z0-9_\-\.]{20,})`),
			Redact:  "${1}[REDACTED_BEARER_TOKEN]",
		},
		{
			Name:    "Authorization Header",
			Pattern: regexp.MustCompile(`(?i)(authorization\s*[=:]\s*["\']?)([a-zA-Z0-9_\-\.]{20,})["\']?`),
			Redact:  "${1}[REDACTED_AUTH_TOKEN]\"",
		},
		
		// AWS Credentials
		{
			Name:    "AWS Access Key",
			Pattern: regexp.MustCompile(`(?i)(aws[_-]?access[_-]?key[_-]?id|AWS_ACCESS_KEY_ID)\s*[=:]\s*["\']?(AKIA[0-9A-Z]{16})["\']?`),
			Redact:  "${1}=\"[REDACTED_AWS_KEY]\"",
		},
		{
			Name:    "AWS Secret Key",
			Pattern: regexp.MustCompile(`(?i)(aws[_-]?secret[_-]?access[_-]?key|AWS_SECRET_ACCESS_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9/+=]{40})["\']?`),
			Redact:  "${1}=\"[REDACTED_AWS_SECRET]\"",
		},
		
		// Database Credentials
		{
			Name:    "Database URL with Password",
			Pattern: regexp.MustCompile(`(?i)(postgres|mysql|mongodb|redis)://([^:]+):([^@]+)@`),
			Redact:  "${1}://${2}:[REDACTED_DB_PASSWORD]@",
		},
		{
			Name:    "Database Password",
			Pattern: regexp.MustCompile(`(?i)(db[_-]?password|database[_-]?password|DB_PASSWORD)\s*[=:]\s*["\']?([^\s"']+)["\']?`),
			Redact:  "${1}=\"[REDACTED_DB_PASSWORD]\"",
		},
		
		// OAuth and Social Media
		{
			Name:    "GitHub Token",
			Pattern: regexp.MustCompile(`(?i)(github[_-]?token|gh[_-]?token|GITHUB_TOKEN)\s*[=:]\s*["\']?(gh[ps]_[a-zA-Z0-9_\-]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_GITHUB_TOKEN]\"",
		},
		{
			Name:    "Google API Key",
			Pattern: regexp.MustCompile(`(?i)(google[_-]?api[_-]?key|GOOGLE_API_KEY|GEMINI_API_KEY)\s*[=:]\s*["\']?(AIza[a-zA-Z0-9_\-]{35})["\']?`),
			Redact:  "${1}=\"[REDACTED_GOOGLE_API_KEY]\"",
		},
		{
			Name:    "OpenAI API Key",
			Pattern: regexp.MustCompile(`(?i)(openai[_-]?api[_-]?key|OPENAI_API_KEY)\s*[=:]\s*["\']?(sk-[a-zA-Z0-9\-]{10,})["\']?`),
			Redact:  "${1}=\"[REDACTED_OPENAI_KEY]\"",
		},
		{
			Name:    "Anthropic/Claude API Key",
			Pattern: regexp.MustCompile(`(?i)(claude[_-]?api[_-]?key|anthropic[_-]?api[_-]?key|CLAUDE_API_KEY)\s*[=:]\s*["\']?(sk-ant-[a-zA-Z0-9\-_]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_CLAUDE_KEY]\"",
		},
		{
			Name:    "Grok/X.AI API Key",
			Pattern: regexp.MustCompile(`(?i)(grok[_-]?api[_-]?key|xai[_-]?api[_-]?key|GROK_API_KEY)\s*[=:]\s*["\']?(xai-[a-zA-Z0-9\-_]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_GROK_KEY]\"",
		},
		{
			Name:    "Slack Token",
			Pattern: regexp.MustCompile(`(?i)(slack[_-]?token|SLACK_TOKEN)\s*[=:]\s*["\']?(xox[baprs]-[a-zA-Z0-9\-]{10,})["\']?`),
			Redact:  "${1}=\"[REDACTED_SLACK_TOKEN]\"",
		},
		
		// Private Keys
		{
			Name:    "Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----).*?(-----END (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_PRIVATE_KEY]\n${2}",
		},
		
		// JWT Tokens
		{
			Name:    "JWT Token",
			Pattern: regexp.MustCompile(`(?i)(jwt|token)\s*[=:]\s*["\']?(eyJ[a-zA-Z0-9_\-]*\.eyJ[a-zA-Z0-9_\-]*\.[a-zA-Z0-9_\-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_JWT_TOKEN]\"",
		},
		
		// Generic Passwords
		{
			Name:    "Password",
			Pattern: regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*["\']([^\s"']{8,})["\']`),
			Redact:  "${1}=\"[REDACTED_PASSWORD]\"",
		},
		
		// Generic Secrets
		{
			Name:    "Secret",
			Pattern: regexp.MustCompile(`(?i)(secret|SECRET)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_SECRET]\"",
		},
		
		// Environment Variable Assignments (catch-all for .env patterns)
		{
			Name:    "Generic Token",
			Pattern: regexp.MustCompile(`(?i)(access[_-]?token|auth[_-]?token|client[_-]?secret|private[_-]?key)\s*[=:]\s*["\']?([a-zA-Z0-9_\-\.]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_TOKEN]\"",
		},
		
		// Credit Card Numbers (basic pattern)
		{
			Name:    "Credit Card",
			Pattern: regexp.MustCompile(`\b([0-9]{4}[\s\-]?){3}[0-9]{4}\b`),
			Redact:  "[REDACTED_CREDIT_CARD]",
		},
		
		// Email in credentials context
		{
			Name:    "Email in Credentials",
			Pattern: regexp.MustCompile(`(?i)(email|user|username)\s*[=:]\s*["\']?([a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,})["\']?`),
			Redact:  "${1}=\"[REDACTED_EMAIL]\"",
		},
	}
)

// ScrubDiff removes sensitive information from git diff output
func ScrubDiff(diff string) string {
	scrubbed := diff
	
	// Apply each pattern
	for _, pattern := range sensitivePatterns {
		scrubbed = pattern.Pattern.ReplaceAllString(scrubbed, pattern.Redact)
	}
	
	return scrubbed
}

// ScrubLines removes sensitive information line by line
// This is useful for more granular control
func ScrubLines(content string) string {
	lines := strings.Split(content, "\n")
	scrubbedLines := make([]string, len(lines))
	
	for i, line := range lines {
		scrubbedLine := line
		for _, pattern := range sensitivePatterns {
			scrubbedLine = pattern.Pattern.ReplaceAllString(scrubbedLine, pattern.Redact)
		}
		scrubbedLines[i] = scrubbedLine
	}
	
	return strings.Join(scrubbedLines, "\n")
}

// HasSensitiveData checks if the content contains any sensitive patterns
func HasSensitiveData(content string) bool {
	for _, pattern := range sensitivePatterns {
		if pattern.Pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// GetDetectedPatterns returns names of all detected sensitive patterns
func GetDetectedPatterns(content string) []string {
	var detected []string
	for _, pattern := range sensitivePatterns {
		if pattern.Pattern.MatchString(content) {
			detected = append(detected, pattern.Name)
		}
	}
	return detected
}

// ScrubEnvFile specifically handles .env file patterns
func ScrubEnvFile(content string) string {
	lines := strings.Split(content, "\n")
	scrubbedLines := make([]string, len(lines))
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			scrubbedLines[i] = line
			continue
		}
		
		// Check if line contains an assignment
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				// Redact the value if it looks like sensitive data
				upperKey := strings.ToUpper(strings.TrimSpace(key))
				if strings.Contains(upperKey, "KEY") ||
					strings.Contains(upperKey, "SECRET") ||
					strings.Contains(upperKey, "TOKEN") ||
					strings.Contains(upperKey, "PASSWORD") ||
					strings.Contains(upperKey, "PASS") ||
					strings.Contains(upperKey, "API") ||
					strings.Contains(upperKey, "AUTH") {
					scrubbedLines[i] = key + "=[REDACTED]"
					continue
				}
			}
		}
		
		// Apply normal scrubbing
		scrubbedLines[i] = ScrubDiff(line)
	}
	
	return strings.Join(scrubbedLines, "\n")
}
