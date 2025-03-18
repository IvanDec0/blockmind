package middleware

import (
	"context"
	"fmt"
	"strings"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

// Context keys
const (
	UserIDKey ContextKey = "user_jid"
)

// GetUserID extracts the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	// Try to get the user_jid from context
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		return sanitizeUserID(userID), true
	}

	// Fall back to string key for backward compatibility
	if userID, ok := ctx.Value(string(UserIDKey)).(string); ok && userID != "" {
		return sanitizeUserID(userID), true
	}

	return "", false
}

// WithUserID returns a new context with the user ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// Helper function to sanitize and normalize user IDs
func sanitizeUserID(userID string) string {
	// Remove any potential harmful characters
	userID = strings.Map(func(r rune) rune {
		// Keep alphanumeric, '@', '.', and basic punctuation
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '@' || r == '.' || r == '-' || r == '_' {
			return r
		}
		return -1
	}, userID)

	// If user ID contains @ (like phone@s.whatsapp.net), extract just the phone part
	if idx := strings.Index(userID, "@"); idx > 0 {
		return userID[:idx]
	}

	return userID
}

// Helper function to get user ID from context
func getUserIDFromContext(ctx context.Context) string {
	userID, ok := GetUserID(ctx)
	if !ok {
		// If no user ID is found, generate one based on context values
		// This ensures we still have proper rate limiting even without a user ID
		values := []string{}

		// Try to get any identifying information from the context
		for _, key := range []string{"remote_addr", "user_agent", "session_id"} {
			if val, ok := ctx.Value(key).(string); ok && val != "" {
				values = append(values, val)
			}
		}

		if len(values) > 0 {
			return fmt.Sprintf("anon-%x", strings.Join(values, "-"))
		}

		// Last resort fallback
		return "default-user"
	}

	return userID
}
