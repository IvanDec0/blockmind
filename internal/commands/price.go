package commands

import (
	"blockmind/internal/config"
	"blockmind/internal/crypto"
	"context"
	"strings"
)

// PriceCommand handles price inquiries for cryptocurrencies
type PriceCommand struct {
	cfg *config.Config
}

// NewPriceCommand creates a new price command
func NewPriceCommand(cfg *config.Config) *PriceCommand {
	return &PriceCommand{
		cfg: cfg,
	}
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

	var cryptoName, targetCurrency string

	// Check if we have both crypto and currency
	if len(args) >= 2 && (strings.ToLower(args[len(args)-2]) == "in" || strings.ToLower(args[len(args)-2]) == "to" || strings.ToLower(args[len(args)-2]) == "en") {
		cryptoName = strings.Join(args[:len(args)-2], " ")
		targetCurrency = args[len(args)-1]
	} else {
		cryptoName = strings.Join(args, " ")
		targetCurrency = ""
	}
	price, err := crypto.GetCryptoPrice(cryptoName, targetCurrency, c.cfg)
	if err != nil {
		return "", err
	}

	return price, nil
}
