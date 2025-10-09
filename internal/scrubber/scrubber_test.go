package scrubber

import (
	"strings"
	"testing"
)

func TestScrubAPIKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Generic API key",
			input:    `api_key="sk-1234567890abcdefghij"`,
			expected: `api_key="[REDACTED_API_KEY]"`,
		},
		{
			name:     "API key with colon",
			input:    `apikey: "abcdefghijklmnopqrstuvwxyz123456"`,
			expected: `apikey="[REDACTED_API_KEY]"`,
		},
		{
			name:     "OpenAI API key",
			input:    `OPENAI_API_KEY=sk-proj-abcdefghijklmnopqrst`,
			expected: `OPENAI_API_KEY="[REDACTED_OPENAI_KEY]"`,
		},
		{
			name:     "Gemini API key",
			input:    `GEMINI_API_KEY="AIzaSyDabcdefghijklmnopqrstuvwxyz12345"`,
			expected: `GEMINI_API_KEY="[REDACTED_GEMINI_API_KEY]"`,
		},
		{
			name:     "Claude API key",
			input:    `CLAUDE_API_KEY=sk-ant-api03-abcdefghijklmnopqrst`,
			expected: `CLAUDE_API_KEY="[REDACTED_CLAUDE_KEY]"`,
		},
		{
			name:     "Grok API key",
			input:    `GROK_API_KEY=xai-abcdefghijklmnopqrst1234567890`,
			expected: `GROK_API_KEY="[REDACTED_GROK_KEY]"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if !strings.Contains(result, "[REDACTED") {
				t.Errorf("ScrubDiff() failed to redact sensitive data.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubAWSCredentials(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "AWS Access Key",
			input: `AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE`,
		},
		{
			name:  "AWS Secret Key",
			input: `AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if !strings.Contains(result, "[REDACTED") {
				t.Errorf("ScrubDiff() failed to redact AWS credentials.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubDatabaseCredentials(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "PostgreSQL URL",
			input: `postgres://username:mypassword123@localhost:5432/mydb`,
		},
		{
			name:  "MySQL URL",
			input: `mysql://admin:secretpass@db.example.com:3306/database`,
		},
		{
			name:  "MongoDB URL",
			input: `mongodb://user:pass123@mongodb.example.com:27017/myapp`,
		},
		{
			name:  "DB Password variable",
			input: `DB_PASSWORD=supersecretpassword123`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if strings.Contains(result, "password123") ||
				strings.Contains(result, "secretpass") ||
				strings.Contains(result, "pass123") ||
				strings.Contains(result, "supersecretpassword123") {
				t.Errorf("ScrubDiff() failed to redact database password.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubJWTTokens(t *testing.T) {
	input := `token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"`
	result := ScrubDiff(input)

	if !strings.Contains(result, "[REDACTED_JWT_TOKEN]") {
		t.Errorf("ScrubDiff() failed to redact JWT token.\nInput: %s\nOutput: %s", input, result)
	}
}

func TestScrubPrivateKeys(t *testing.T) {
	input := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1234567890abcdef
ghijklmnopqrstuvwxyz
-----END RSA PRIVATE KEY-----`

	result := ScrubDiff(input)

	if strings.Contains(result, "MIIEpAIBAAKCAQEA") || strings.Contains(result, "ghijklmnopqrstuvwxyz") {
		t.Errorf("ScrubDiff() failed to redact private key.\nOutput: %s", result)
	}
	if !strings.Contains(result, "[REDACTED_RSA_PRIVATE_KEY]") {
		t.Errorf("ScrubDiff() did not add redaction marker.\nOutput: %s", result)
	}
}

func TestScrubBearerToken(t *testing.T) {
	input := `Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9abcdefghijk`
	result := ScrubDiff(input)

	if !strings.Contains(result, "[REDACTED") {
		t.Errorf("ScrubDiff() failed to redact bearer token.\nInput: %s\nOutput: %s", input, result)
	}
}

func TestScrubGitHubToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "GitHub personal access token",
			input: `GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvw`,
		},
		{
			name:  "GitHub app token",
			input: `gh_token: "ghs_1234567890abcdefghijklmnopqrstuvw"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if !strings.Contains(result, "[REDACTED") {
				t.Errorf("ScrubDiff() failed to redact GitHub token.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubSlackToken(t *testing.T) {
	input := `SLACK_TOKEN=xoxb-1234567890-1234567890-abcdefghijk`
	result := ScrubDiff(input)

	if !strings.Contains(result, "[REDACTED_SLACK_TOKEN]") {
		t.Errorf("ScrubDiff() failed to redact Slack token.\nInput: %s\nOutput: %s", input, result)
	}
}

func TestScrubPasswords(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Password with equals",
			input: `password="mySecretPass123"`,
		},
		{
			name:  "Password with colon",
			input: `PASSWORD: "AnotherSecret456"`,
		},
		{
			name:  "passwd variant",
			input: `passwd='ComplexPassword789'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if !strings.Contains(result, "[REDACTED") {
				t.Errorf("ScrubDiff() failed to redact password.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubEnvFile(t *testing.T) {
	input := `# Environment variables
DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=sk-1234567890abcdefghij
SECRET_TOKEN=mysecrettoken123
DEBUG=true
PORT=3000
OPENAI_API_KEY=sk-proj-abcdefghijk`

	result := ScrubEnvFile(input)

	// Check that sensitive values are redacted
	if strings.Contains(result, "sk-1234567890") ||
		strings.Contains(result, "mysecrettoken123") ||
		strings.Contains(result, "sk-proj-abcdefghijk") {
		t.Errorf("ScrubEnvFile() failed to redact sensitive values.\nOutput: %s", result)
	}

	// Check that non-sensitive values are preserved
	if !strings.Contains(result, "DEBUG=true") || !strings.Contains(result, "PORT=3000") {
		t.Errorf("ScrubEnvFile() incorrectly redacted non-sensitive values.\nOutput: %s", result)
	}

	// Check that comments are preserved
	if !strings.Contains(result, "# Environment variables") {
		t.Errorf("ScrubEnvFile() removed comments.\nOutput: %s", result)
	}
}

func TestHasSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Contains API key",
			input:    `api_key="sk-1234567890abcdefghij"`,
			expected: true,
		},
		{
			name:     "Contains password",
			input:    `password="mySecretPass123"`,
			expected: true,
		},
		{
			name:     "No sensitive data",
			input:    `console.log("Hello World")`,
			expected: false,
		},
		{
			name:     "Normal code",
			input:    `const port = 3000; app.listen(port);`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasSensitiveData(tt.input)
			if result != tt.expected {
				t.Errorf("HasSensitiveData() = %v, expected %v for input: %s", result, tt.expected, tt.input)
			}
		})
	}
}

func TestGetDetectedPatterns(t *testing.T) {
	input := `
	OPENAI_API_KEY=sk-proj-abcdefghijklmnop
	password="mySecretPass123"
	GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvw
	`

	patterns := GetDetectedPatterns(input)

	if len(patterns) == 0 {
		t.Error("GetDetectedPatterns() returned no patterns for input with sensitive data")
	}

	// Should detect at least these patterns
	hasOpenAI := false
	hasPassword := false
	hasGitHub := false

	for _, p := range patterns {
		if strings.Contains(p, "OpenAI") {
			hasOpenAI = true
		}
		if strings.Contains(p, "Password") {
			hasPassword = true
		}
		if strings.Contains(p, "GitHub") {
			hasGitHub = true
		}
	}

	if !hasOpenAI {
		t.Error("GetDetectedPatterns() did not detect OpenAI API key")
	}
	if !hasPassword {
		t.Error("GetDetectedPatterns() did not detect password")
	}
	if !hasGitHub {
		t.Error("GetDetectedPatterns() did not detect GitHub token")
	}
}

func TestScrubComplexGitDiff(t *testing.T) {
	// Simulate a realistic git diff with sensitive data
	input := `diff --git a/.env b/.env
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/.env
@@ -0,0 +1,5 @@
+DATABASE_URL=postgres://admin:secretpass@localhost/mydb
+OPENAI_API_KEY=sk-proj-abcdefghijklmnopqrstuvwxyz
+JWT_SECRET=mysupersecretjwtkey123456
+PORT=3000
+NODE_ENV=development

diff --git a/config.js b/config.js
index abcdef..123456 100644
--- a/config.js
+++ b/config.js
@@ -1,3 +1,4 @@
 module.exports = {
   port: process.env.PORT || 3000,
+  apiKey: "sk-1234567890abcdefghijklmnop",
 };`

	result := ScrubDiff(input)

	// Check that sensitive values are removed
	if strings.Contains(result, "secretpass") {
		t.Error("Failed to scrub database password from diff")
	}
	if strings.Contains(result, "sk-proj-abcdefghijklmnopqrstuvwxyz") {
		t.Error("Failed to scrub OpenAI API key from diff")
	}
	if strings.Contains(result, "mysupersecretjwtkey") {
		t.Error("Failed to scrub JWT secret from diff")
	}
	if strings.Contains(result, "sk-1234567890abcdefghijklmnop") {
		t.Error("Failed to scrub API key from config file")
	}

	// Check that non-sensitive values are preserved
	if !strings.Contains(result, "PORT=3000") {
		t.Error("Incorrectly removed non-sensitive PORT value")
	}
	if !strings.Contains(result, "NODE_ENV=development") {
		t.Error("Incorrectly removed non-sensitive NODE_ENV value")
	}
	if !strings.Contains(result, "diff --git") {
		t.Error("Removed git diff headers")
	}
}

func TestScrubEmailInCredentials(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Email in username field",
			input: `username: "user@example.com"`,
		},
		{
			name:  "Email in email field",
			input: `email="admin@company.org"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if !strings.Contains(result, "[REDACTED_EMAIL]") {
				t.Errorf("ScrubDiff() failed to redact email.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}

func TestScrubCreditCard(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Credit card with spaces",
			input: `card: "4532 1234 5678 9010"`,
		},
		{
			name:  "Credit card with dashes",
			input: `4532-1234-5678-9010`,
		},
		{
			name:  "Credit card no separators",
			input: `4532123456789010`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrubDiff(tt.input)
			if strings.Contains(result, "4532") && strings.Contains(result, "9010") {
				t.Errorf("ScrubDiff() failed to redact credit card.\nInput: %s\nOutput: %s", tt.input, result)
			}
		})
	}
}
