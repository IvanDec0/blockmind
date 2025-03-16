package commands

import (
	"context"
	"fmt"
	"strings"
)

// PriceCommand handles price inquiries for cryptocurrencies
type PriceCommand struct{}

// NewPriceCommand creates a new price command
func NewPriceCommand() *PriceCommand {
	return &PriceCommand{}
}

// Name returns the name of the command
func (c *PriceCommand) Name() string {
	return "price"
}

// Aliases returns alternative names for the command
func (c *PriceCommand) Aliases() []string {
	return []string{"p", "precio"}
}

// Description returns the description of the command
func (c *PriceCommand) Description() string {
	return "Get the price of a cryptocurrency"
}

// Execute executes the command with the given arguments
func (c *PriceCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Please specify a cryptocurrency (e.g., /price Bitcoin)", nil
	}

	// In a real implementation, you would call a crypto price API here
	// For now, we'll just return a placeholder response
	cryptoName := strings.Join(args, " ")
	return fmt.Sprintf("The price of %s is $1,000", cryptoName), nil
}
