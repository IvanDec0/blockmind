package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type userLimiter struct {
	count     int
	resetTime time.Time
	mutex     sync.Mutex
}

// HandlerFunc represents a function that processes a command
type HandlerFunc func(ctx context.Context, input string) (string, error)

// RateLimiter limits the rate of command execution per user
func RateLimiter(limit int, period time.Duration) func(HandlerFunc) HandlerFunc {
	fmt.Println("RateLimiter")

	limiters := make(map[string]*userLimiter)
	var limitersMutex sync.Mutex

	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, input string) (string, error) {
			// Get user ID from context (or use a default)
			userID := getUserIDFromContext(ctx)

			limitersMutex.Lock()
			ul, exists := limiters[userID]
			if !exists {
				ul = &userLimiter{
					count:     0,
					resetTime: time.Now().Add(period),
				}
				limiters[userID] = ul
			}
			limitersMutex.Unlock()

			ul.mutex.Lock()
			defer ul.mutex.Unlock()

			// Reset counter if period has elapsed
			now := time.Now()
			if now.After(ul.resetTime) {
				ul.count = 0
				ul.resetTime = now.Add(period)
			}

			// Check if limit is reached
			if ul.count >= limit {
				return "You're sending messages too quickly. Please wait a moment.", nil
			}

			// Increment counter
			ul.count++

			// Continue processing
			return next(ctx, input)
		}
	}
}

// Timeout adds a timeout to command execution
func Timeout(duration time.Duration) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, input string) (string, error) {
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()

			// Use a channel to handle the timeout
			resultCh := make(chan struct {
				response string
				err      error
			})

			go func() {
				response, err := next(ctx, input)
				resultCh <- struct {
					response string
					err      error
				}{response, err}
			}()

			// Wait for result or timeout
			select {
			case result := <-resultCh:
				return result.response, result.err
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					return "Request timed out. Please try again later.", nil
				}
				return "", ctx.Err()
			}
		}
	}
}
