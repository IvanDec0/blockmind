package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// HelpCommand provides help information about available commands
type HelpCommand struct {
	manager *Manager
}

// NewHelpCommand creates a new help command
func NewHelpCommand(manager *Manager) *HelpCommand {
	return &HelpCommand{manager: manager}
}

// Name returns the name of the command
func (c *HelpCommand) Name() string {
	return "help"
}

// Aliases returns alternative names for the command
func (c *HelpCommand) Aliases() []string {
	return []string{"h", "ayuda"}
}

// Description returns the description of the command
func (c *HelpCommand) Description() string {
	return "Shows available commands and usage information"
}

// Execute executes the command with the given arguments
func (c *HelpCommand) Execute(ctx context.Context, args []string) (string, error) {
	var helpText strings.Builder

	helpText.WriteString("*Available Commands:*\n\n")

	// Get all unique commands (ignoring aliases)
	uniqueCommands := make(map[string]Command)
	for name, cmd := range c.manager.GetCommands() {
		if name == cmd.Name() {
			uniqueCommands[name] = cmd
		}
	}

	// Sort commands alphabetically
	var commandNames []string
	for name := range uniqueCommands {
		commandNames = append(commandNames, name)
	}
	sort.Strings(commandNames)

	// Build help text
	for _, name := range commandNames {
		cmd := uniqueCommands[name]

		// Show command name and description
		helpText.WriteString(fmt.Sprintf("/%s - %s\n", cmd.Name(), cmd.Description()))

		// Show aliases if any
		if aliases := cmd.Aliases(); len(aliases) > 0 {
			helpText.WriteString(fmt.Sprintf("  Aliases: /%s\n", strings.Join(aliases, ", /")))
		}

		helpText.WriteString("\n")
	}

	helpText.WriteString("You can also ask me questions directly!")

	return helpText.String(), nil
}
