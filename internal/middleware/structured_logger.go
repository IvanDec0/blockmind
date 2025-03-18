package middleware

import (
	"blockmind/internal/logger"
	"context"
	"time"
)

// StructuredLogger is middleware that logs command execution using structured logging
func StructuredLogger(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, input string) (string, error) {
		start := time.Now()

		// Get logger from context
		log := logger.FromContext(ctx)

		// Log the incoming request
		log.Debug().
			Str("input", input).
			Msg("Processing command")

		// Execute the handler
		result, err := next(ctx, input)

		// Calculate duration
		duration := time.Since(start)

		// Log the result
		if err != nil {
			log.Error().
				Err(err).
				Str("input", input).
				Msg("Error processing command")
		} else {
			log.Info().
				Str("input", input).
				Dur("duration", duration).
				Msg("Command completed")
		}

		return result, err
	}
}
