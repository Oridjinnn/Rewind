package redact

import (
	"os"
	"regexp"
	"strings"
)

// RedactMode controls how secrets are handled.
type RedactMode int

const (
	// ModeOff: no redaction.
	ModeOff RedactMode = iota
	// ModeRedact: replace secrets with placeholder.
	ModeRedact
	// ModeSkip: skip commands containing secrets entirely (not recorded).
	ModeSkip
)

var (
	// Common patterns that indicate secrets.
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:--password|--token|--secret|--key|--api-key|--apikey)\s+(\S+)`),
		regexp.MustCompile(`(?i)(?:password|passwd|pwd)\s*[=:]\s*\S+`),
		regexp.MustCompile(`(?i)(?:token|api_key|apikey|secret|jwt)\s*[=:]\s*\S+`),
		regexp.MustCompile(`(?i)Authorization:\s*Bearer\s+\S+`),
		regexp.MustCompile(`(?i)export\s+\w*(?:KEY|TOKEN|SECRET|PASS|PASSWORD)\s*=\s*\S+`),
		regexp.MustCompile(`(?i)set\s+\w*(?:KEY|TOKEN|SECRET|PASS|PASSWORD)\s*=\s*\S+`),
		regexp.MustCompile(`(?i)ghp_[a-zA-Z0-9]{36}`),                    // GitHub PAT
		regexp.MustCompile(`(?i)gho_[a-zA-Z0-9]{36}`),                    // GitHub OAuth
		regexp.MustCompile(`(?i)xox[baprs]-[a-zA-Z0-9\-]{24,}`),           // Slack tokens
		regexp.MustCompile(`(?i)sk-[a-zA-Z0-9]{20,}`),                    // OpenAI key format
		regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),                        // AWS Access Key
		regexp.MustCompile(`(?i)-----BEGIN\s+(?:RSA\s+)?PRIVATE\s+KEY-----`), // Private key
	}
)

// RedactCommand returns the command with secret values replaced, or empty string if the
// entire command should be skipped (ModeSkip). Mode is determined by the REWIND_REDACT environment
// variable: "true"/"redact" → ModeRedact, "skip" → ModeSkip, anything else → ModeOff.
func RedactCommand(cmd string) string {
	mode := detectMode()
	if mode == ModeOff {
		return cmd
	}
	if mode == ModeSkip && containsSecret(cmd) {
		return ""
	}
	return applyPatterns(cmd)
}

// ContainsSecret checks if the command appears to contain sensitive data.
func ContainsSecret(cmd string) bool {
	return containsSecret(cmd)
}

// detectMode reads REWIND_REDACT env var.
func detectMode() RedactMode {
	switch strings.ToLower(os.Getenv("REWIND_REDACT")) {
	case "true", "1", "redact", "on":
		return ModeRedact
	case "skip", "block":
		return ModeSkip
	default:
		return ModeOff
	}
}

func containsSecret(s string) bool {
	for _, re := range secretPatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func applyPatterns(cmd string) string {
	for _, re := range secretPatterns {
		cmd = re.ReplaceAllString(cmd, "${1} [REDACTED]")
	}
	return cmd
}

// RedactExportKey redacts the *value* part of an "export KEY=value" pattern.
// This avoids obliterating the whole line.
func applyExportKey(cmd string) string {
	re := regexp.MustCompile(`(?i)(export\s+\w*(?:KEY|TOKEN|SECRET|PASS|PASSWORD)\s*=\s*).*`)
	return re.ReplaceAllString(cmd, "${1}[REDACTED]")
}