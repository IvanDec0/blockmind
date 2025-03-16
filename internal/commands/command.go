package commands

import (
	"blockmind/internal/security"
	"context"
	"strings"
)

// Command represents a command that can be executed
type Command interface {
	// Name returns the name of the command
	Name() string
	// Aliases returns alternative names for the command
	Aliases() []string
	// Description returns the description of the command
	Description() string
	// Execute executes the command with the given arguments
	Execute(ctx context.Context, args []string) (string, error)
}

// Manager handles command registration and execution
type Manager struct {
	commands       map[string]Command
	defaultHandler func(context.Context, string) (string, error)
}

// NewManager creates a new command manager
func NewManager(defaultHandler func(context.Context, string) (string, error)) *Manager {
	return &Manager{
		commands:       make(map[string]Command),
		defaultHandler: defaultHandler,
	}
}

// Register registers a command
func (m *Manager) Register(cmd Command) {
	m.commands[cmd.Name()] = cmd

	// Register aliases
	for _, alias := range cmd.Aliases() {
		m.commands[alias] = cmd
	}
}

// Execute executes a command
func (m *Manager) Execute(ctx context.Context, input string) (string, error) {
	// First sanitize the entire input
	input = security.SanitizeInput(input)

	// Split input into command and arguments
	var command string
	var args []string

	parts := splitCommandText(input)
	if len(parts) == 0 {
		return "Send a command like '/price Bitcoin' or ask a question.", nil
	}

	command = parts[0]
	if len(parts) > 1 {
		args = parts[1:]
	}

	// Check if it's a command
	if isCommand(command) {
		// Remove the "/" prefix
		cmdName := command[1:]

		// Further sanitize command and args
		cmdName, args = security.SanitizeCommand(cmdName, args)

		cmd, exists := m.commands[cmdName]
		if exists {
			return cmd.Execute(ctx, args)
		}
		return "Unknown command. Type /help for a list of commands.", nil
	}

	// If not a command, use the default handler for questions
	if m.defaultHandler != nil {
		return m.defaultHandler(ctx, input)
	}

	return "I don't understand that. Try typing /help for assistance.", nil
}

// GetCommands returns all registered commands
func (m *Manager) GetCommands() map[string]Command {
	return m.commands
}

// Helper functions
func splitCommandText(text string) []string {
	if text == "" {
		return []string{}
	}

	// Improved version that handles quoted arguments
	parts := []string{}
	var currentPart strings.Builder
	inQuotes := false

	for _, char := range text {
		switch char {
		case ' ':
			if !inQuotes {
				if currentPart.Len() > 0 {
					parts = append(parts, currentPart.String())
					currentPart.Reset()
				}
			} else {
				currentPart.WriteRune(char)
			}
		case '"':
			inQuotes = !inQuotes
		default:
			currentPart.WriteRune(char)
		}
	}

	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	return parts
}

func isCommand(text string) bool {
	return len(text) > 0 && text[0] == '/'
}
