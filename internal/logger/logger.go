package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	// defaultLogger is the default logger instance
	defaultLogger zerolog.Logger
)

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

func init() {
	// Configure the default logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	defaultLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// SetOutput sets the output destination for the default logger
func SetOutput(w io.Writer) {
	defaultLogger = zerolog.New(w).With().Timestamp().Logger()
}

// SetLevel sets the global log level
func SetLevel(level string) {
	switch level {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// With returns a logger with the given fields
func With(fields ...Field) zerolog.Logger {
	ctx := defaultLogger.With()
	for _, field := range fields {
		ctx = ctx.Interface(field.Key, field.Value)
	}
	return ctx.Logger()
}

// Debug logs a debug message
func Debug(msg string, fields ...Field) {
	event := defaultLogger.Debug()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Info logs an info message
func Info(msg string, fields ...Field) {
	event := defaultLogger.Info()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...Field) {
	event := defaultLogger.Warn()
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// Error logs an error message
func Error(msg string, err error, fields ...Field) {
	event := defaultLogger.Error()
	if err != nil {
		event = event.Err(err)
	}
	for _, field := range fields {
		event = event.Interface(field.Key, field.Value)
	}
	event.Msg(msg)
}

// FromContext creates a logger from a context
func FromContext(ctx context.Context) zerolog.Logger {
	if userID, ok := ctx.Value("user_jid").(string); ok {
		return With(Field{Key: "user_jid", Value: userID})
	}
	return defaultLogger
}
