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

		// Azure Credentials
		{
			Name:    "Azure Client Secret",
			Pattern: regexp.MustCompile(`(?i)(azure[_-]?client[_-]?secret|AZURE_CLIENT_SECRET)\s*[=:]\s*["\']?([a-zA-Z0-9_\-\.~]{20,})["\']?`),
			Redact:  "${1}=\"[REDACTED_AZURE_CLIENT_SECRET]\"",
		},
		{
			Name:    "Azure Subscription Key",
			Pattern: regexp.MustCompile(`(?i)(azure[_-]?subscription[_-]?key|AZURE_SUBSCRIPTION_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_AZURE_SUBSCRIPTION_KEY]\"",
		},
		{
			Name:    "Azure Storage Key",
			Pattern: regexp.MustCompile(`(?i)(azure[_-]?storage[_-]?key|AZURE_STORAGE_KEY|AccountKey)\s*[=:]\s*["\']?([a-zA-Z0-9/+=]{88})["\']?`),
			Redact:  "${1}=\"[REDACTED_AZURE_STORAGE_KEY]\"",
		},
		{
			Name:    "Azure Service Principal",
			Pattern: regexp.MustCompile(`(?i)(azure[_-]?service[_-]?principal|AZURE_SERVICE_PRINCIPAL)\s*[=:]\s*["\']?([a-f0-9-]{36})["\']?`),
			Redact:  "${1}=\"[REDACTED_AZURE_SERVICE_PRINCIPAL]\"",
		},

		// Google Cloud Credentials
		{
			Name:    "Google Cloud Service Account Key",
			Pattern: regexp.MustCompile(`(?i)(gcp[_-]?service[_-]?account[_-]?key|GOOGLE_APPLICATION_CREDENTIALS)\s*[=:]\s*["\']?([a-zA-Z0-9_\-\.@]+\.json)["\']?`),
			Redact:  "${1}=\"[REDACTED_GCP_SA_KEY]\"",
		},
		{
			Name:    "Google Cloud API Key",
			Pattern: regexp.MustCompile(`(?i)(gcp[_-]?api[_-]?key|GOOGLE_CLOUD_API_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{39})["\']?`),
			Redact:  "${1}=\"[REDACTED_GCP_API_KEY]\"",
		},
		{
			Name:    "Google Cloud OAuth Client",
			Pattern: regexp.MustCompile(`(?i)(gcp[_-]?oauth[_-]?client[_-]?secret|GOOGLE_OAUTH_CLIENT_SECRET)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{24})["\']?`),
			Redact:  "${1}=\"[REDACTED_GCP_OAUTH_SECRET]\"",
		},
		{
			Name:    "Google Cloud JSON Credentials",
			Pattern: regexp.MustCompile(`(?s)"type":\s*"service_account".*?"private_key":\s*"-----BEGIN PRIVATE KEY-----.*?-----END PRIVATE KEY-----`),
			Redact:  "\"type\": \"service_account\",\n\"private_key\": \"[REDACTED_GCP_PRIVATE_KEY]\"",
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
			Name:    "Gemini API Key",
			Pattern: regexp.MustCompile(`(?i)(google[_-]?api[_-]?key|GOOGLE_API_KEY|GEMINI_API_KEY)\s*[=:]\s*["\']?(AIza[a-zA-Z0-9_\-]{35})["\']?`),
			Redact:  "${1}=\"[REDACTED_GEMINI_API_KEY]\"",
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

		// Payment & Communication APIs
		{
			Name:    "Stripe API Key",
			Pattern: regexp.MustCompile(`(?i)(stripe[_-]?api[_-]?key|STRIPE_API_KEY)\s*[=:]\s*["\']?(sk_live_[a-zA-Z0-9]{24})["\']?`),
			Redact:  "${1}=\"[REDACTED_STRIPE_KEY]\"",
		},
		{
			Name:    "Stripe Publishable Key",
			Pattern: regexp.MustCompile(`(?i)(stripe[_-]?publishable[_-]?key|STRIPE_PUBLISHABLE_KEY)\s*[=:]\s*["\']?(pk_live_[a-zA-Z0-9]{24})["\']?`),
			Redact:  "${1}=\"[REDACTED_STRIPE_PUBLISHABLE_KEY]\"",
		},
		{
			Name:    "Twilio API Key",
			Pattern: regexp.MustCompile(`(?i)(twilio[_-]?api[_-]?key|TWILIO_API_KEY)\s*[=:]\s*["\']?(SK[a-f0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_TWILIO_API_KEY]\"",
		},
		{
			Name:    "Twilio Auth Token",
			Pattern: regexp.MustCompile(`(?i)(twilio[_-]?auth[_-]?token|TWILIO_AUTH_TOKEN)\s*[=:]\s*["\']?([a-f0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_TWILIO_AUTH_TOKEN]\"",
		},
		{
			Name:    "Twilio Account SID",
			Pattern: regexp.MustCompile(`(?i)(twilio[_-]?account[_-]?sid|TWILIO_ACCOUNT_SID)\s*[=:]\s*["\']?(AC[a-f0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_TWILIO_ACCOUNT_SID]\"",
		},

		// Cloud & Infrastructure APIs
		{
			Name:    "DigitalOcean API Key",
			Pattern: regexp.MustCompile(`(?i)(digitalocean[_-]?api[_-]?key|DIGITALOCEAN_API_KEY)\s*[=:]\s*["\']?([a-f0-9]{64})["\']?`),
			Redact:  "${1}=\"[REDACTED_DIGITALOCEAN_KEY]\"",
		},
		{
			Name:    "Heroku API Key",
			Pattern: regexp.MustCompile(`(?i)(heroku[_-]?api[_-]?key|HEROKU_API_KEY)\s*[=:]\s*["\']?([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})["\']?`),
			Redact:  "${1}=\"[REDACTED_HEROKU_KEY]\"",
		},
		{
			Name:    "Vercel API Key",
			Pattern: regexp.MustCompile(`(?i)(vercel[_-]?api[_-]?key|VERCEL_API_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{24})["\']?`),
			Redact:  "${1}=\"[REDACTED_VERCEL_KEY]\"",
		},
		{
			Name:    "Netlify API Key",
			Pattern: regexp.MustCompile(`(?i)(netlify[_-]?api[_-]?key|NETLIFY_API_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9_\-]{40})["\']?`),
			Redact:  "${1}=\"[REDACTED_NETLIFY_KEY]\"",
		},

		// Database & Analytics APIs
		{
			Name:    "MongoDB Atlas API Key",
			Pattern: regexp.MustCompile(`(?i)(mongodb[_-]?api[_-]?key|MONGODB_API_KEY)\s*[=:]\s*["\']?([a-f0-9]{24}-[a-f0-9]{24})["\']?`),
			Redact:  "${1}=\"[REDACTED_MONGODB_API_KEY]\"",
		},
		{
			Name:    "SendGrid API Key",
			Pattern: regexp.MustCompile(`(?i)(sendgrid[_-]?api[_-]?key|SENDGRID_API_KEY)\s*[=:]\s*["\']?(SG\.[a-zA-Z0-9_\-\.]{22,}\.[a-zA-Z0-9_\-\.]{43})["\']?`),
			Redact:  "${1}=\"[REDACTED_SENDGRID_KEY]\"",
		},
		{
			Name:    "Mailgun API Key",
			Pattern: regexp.MustCompile(`(?i)(mailgun[_-]?api[_-]?key|MAILGUN_API_KEY)\s*[=:]\s*["\']?(key-[a-f0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_MAILGUN_KEY]\"",
		},

		// Social & Auth APIs
		{
			Name:    "Facebook App Secret",
			Pattern: regexp.MustCompile(`(?i)(facebook[_-]?app[_-]?secret|FACEBOOK_APP_SECRET)\s*[=:]\s*["\']?([a-f0-9]{32})["\']?`),
			Redact:  "${1}=\"[REDACTED_FACEBOOK_SECRET]\"",
		},
		{
			Name:    "Twitter API Key",
			Pattern: regexp.MustCompile(`(?i)(twitter[_-]?api[_-]?key|TWITTER_API_KEY)\s*[=:]\s*["\']?([a-zA-Z0-9]{25})["\']?`),
			Redact:  "${1}=\"[REDACTED_TWITTER_API_KEY]\"",
		},
		{
			Name:    "Twitter API Secret",
			Pattern: regexp.MustCompile(`(?i)(twitter[_-]?api[_-]?secret|TWITTER_API_SECRET)\s*[=:]\s*["\']?([a-zA-Z0-9]{50})["\']?`),
			Redact:  "${1}=\"[REDACTED_TWITTER_API_SECRET]\"",
		},
		{
			Name:    "LinkedIn Client Secret",
			Pattern: regexp.MustCompile(`(?i)(linkedin[_-]?client[_-]?secret|LINKEDIN_CLIENT_SECRET)\s*[=:]\s*["\']?([a-zA-Z0-9]{16})["\']?`),
			Redact:  "${1}=\"[REDACTED_LINKEDIN_SECRET]\"",
		},

		// SSH Keys and Private Keys in various formats
		{
			Name:    "RSA Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN RSA PRIVATE KEY-----).*?(-----END RSA PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_RSA_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "EC Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN EC PRIVATE KEY-----).*?(-----END EC PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_EC_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "DSA Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN DSA PRIVATE KEY-----).*?(-----END DSA PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_DSA_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "OpenSSH Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN OPENSSH PRIVATE KEY-----).*?(-----END OPENSSH PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_OPENSSH_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "PKCS#8 Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN PRIVATE KEY-----).*?(-----END PRIVATE KEY-----)`),
			Redact:  "${1}\n[REDACTED_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "PGP Private Key",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN PGP PRIVATE KEY BLOCK-----).*?(-----END PGP PRIVATE KEY BLOCK-----)`),
			Redact:  "${1}\n[REDACTED_PGP_PRIVATE_KEY]\n${2}",
		},
		{
			Name:    "SSH Public Key (RSA)",
			Pattern: regexp.MustCompile(`(?i)(ssh[_-]?rsa[_-]?public[_-]?key|SSH_RSA_PUBLIC_KEY)\s*[=:]\s*["\']?(ssh-rsa [a-zA-Z0-9/+=]+ [a-zA-Z0-9._@-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_SSH_RSA_PUBLIC_KEY]\"",
		},
		{
			Name:    "SSH Public Key (ED25519)",
			Pattern: regexp.MustCompile(`(?i)(ssh[_-]?ed25519[_-]?public[_-]?key|SSH_ED25519_PUBLIC_KEY)\s*[=:]\s*["\']?(ssh-ed25519 [a-zA-Z0-9]+ [a-zA-Z0-9._@-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_SSH_ED25519_PUBLIC_KEY]\"",
		},
		{
			Name:    "SSH Public Key (ECDSA)",
			Pattern: regexp.MustCompile(`(?i)(ssh[_-]?ecdsa[_-]?public[_-]?key|SSH_ECDSA_PUBLIC_KEY)\s*[=:]\s*["\']?(ecdsa-sha2-[a-zA-Z0-9-]+ [a-zA-Z0-9/+=]+ [a-zA-Z0-9._@-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_SSH_ECDSA_PUBLIC_KEY]\"",
		},
		{
			Name:    "SSH Public Key (Generic)",
			Pattern: regexp.MustCompile(`(?i)(ssh[_-]?public[_-]?key|SSH_PUBLIC_KEY)\s*[=:]\s*["\']?(ssh-[a-zA-Z0-9-]+ [a-zA-Z0-9/+=]+ [a-zA-Z0-9._@-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_SSH_PUBLIC_KEY]\"",
		},
		{
			Name:    "SSH Authorized Keys",
			Pattern: regexp.MustCompile(`(?i)(ssh[_-]?authorized[_-]?keys|SSH_AUTHORIZED_KEYS)\s*[=:]\s*["\']?(ssh-[a-zA-Z0-9-]+ [a-zA-Z0-9/+=]+ [a-zA-Z0-9._@-]+)["\']?`),
			Redact:  "${1}=\"[REDACTED_SSH_AUTHORIZED_KEYS]\"",
		},
		{
			Name:    "SSH Private Key File Content",
			Pattern: regexp.MustCompile(`(?s)(-----BEGIN [A-Z ]+PRIVATE KEY-----\n)([A-Za-z0-9+/=\n]+)(\n-----END [A-Z ]+PRIVATE KEY-----)`),
			Redact:  "${1}[REDACTED_SSH_PRIVATE_KEY_CONTENT]${3}",
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
