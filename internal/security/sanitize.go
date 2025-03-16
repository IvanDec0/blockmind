package security

import (
	"regexp"
	"strings"
)

// Regular expressions for detecting potentially malicious patterns
var (
	// Matches common script injection patterns
	scriptPattern = regexp.MustCompile(`(?i)<\s*script\b[^>]*>(.*?)<\s*/\s*script\s*>`)

	// Matches SQL injection attempts
	sqlPattern = regexp.MustCompile(`(?i)(\b(select|insert|update|delete|drop|alter|create|truncate)\b.*\b(from|into|table|database|schema)\b)`)

	// Matches URL patterns that might be used for phishing
	urlPattern = regexp.MustCompile(`(?i)(https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|www\.[a-zA-Z0-9][a-zA-Z0-9-]+[a-zA-Z0-9]\.[^\s]{2,}|https?:\/\/(?:www\.|(?!www))[a-zA-Z0-9]+\.[^\s]{2,}|www\.[a-zA-Z0-9]+\.[^\s]{2,})`)

	// Controls unicode homoglyph attacks where characters look similar but are different code points
	homoglyphPattern = regexp.MustCompile(`\p{So}|\p{Cn}`)
)

// SanitizeInput cleans user input to prevent security issues
func SanitizeInput(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Remove control characters
	sanitized = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(sanitized, "")

	// Check for suspicious patterns and flag them
	if scriptPattern.MatchString(sanitized) {
		return "⚠️ Suspicious script detected in input"
	}

	if sqlPattern.MatchString(sanitized) && len(sanitized) > 15 {
		return "⚠️ Suspicious SQL syntax detected in input"
	}

	// Limit input length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000] + "... (truncated)"
	}

	return sanitized
}

// SanitizeCommand specifically sanitizes command inputs which might need different rules
func SanitizeCommand(command string, args []string) (string, []string) {
	// Sanitize command name
	sanitizedCommand := strings.TrimSpace(command)
	sanitizedCommand = strings.ToLower(sanitizedCommand)

	// Sanitize arguments
	sanitizedArgs := make([]string, len(args))
	for i, arg := range args {
		// Trim whitespace
		arg = strings.TrimSpace(arg)

		// Remove control characters
		arg = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(arg, "")

		// Replace any suspicious patterns
		arg = homoglyphPattern.ReplaceAllString(arg, "")

		sanitizedArgs[i] = arg
	}

	return sanitizedCommand, sanitizedArgs
}

// IsSafeURL checks if a URL is potentially malicious
func IsSafeURL(url string) bool {
	// Basic check: you could extend this with URL reputation checking or allowlisting
	return !strings.Contains(strings.ToLower(url), "suspicious-domain.com")
}
