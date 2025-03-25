package handlers

import (
	"blockmind/internal/commands"
	"blockmind/internal/config"
	"blockmind/internal/ia"
	"blockmind/internal/middleware"
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// WhatsAppHandler handles WhatsApp message processing
type WhatsAppHandler struct {
	client         *whatsmeow.Client
	commandManager *commands.Manager
	config         *config.Config
	handlerChain   middleware.HandlerFunc
}

// NewWhatsAppHandler creates a new WhatsApp handler
func NewWhatsAppHandler(client *whatsmeow.Client, cfg *config.Config) *WhatsAppHandler {
	// Create default handler for non-command messages
	defaultHandler := func(ctx context.Context, text string) (string, error) {
		return ia.AskQuestion(text, cfg)
	}

	// Create command manager
	manager := commands.NewManager(defaultHandler)

	// Register commands
	manager.Register(commands.NewPriceCommand(cfg))
	manager.Register(commands.NewRecommendCommand(cfg))

	// Help command needs a reference to the manager
	helpCmd := commands.NewHelpCommand(manager)
	manager.Register(helpCmd)

	// Create the handler chain once
	handler := manager.Execute
	handler = middleware.StructuredLogger(handler)
	handler = middleware.RateLimiter(cfg.RateLimit, cfg.RateLimitPeriod)(handler)
	handler = middleware.Timeout(cfg.CommandTimeout)(handler)

	return &WhatsAppHandler{
		client:         client,
		commandManager: manager,
		config:         cfg,
		handlerChain:   handler,
	}
}

// HandleMessage processes incoming WhatsApp messages
func (h *WhatsAppHandler) HandleMessage(message *waE2E.Message, chatJID types.JID) {
	// Extract text from message
	text := message.GetConversation()
	if text == "" {
		return // Ignore non-text messages
	}

	// Create context with timeout and user info
	ctx := context.Background()
	ctx = context.WithValue(ctx, "user_jid", chatJID.String())
	ctx, cancel := context.WithTimeout(ctx, h.config.CommandTimeout)
	defer cancel()

	// Use the existing handler chain
	response, err := h.handlerChain(ctx, text)
	if err != nil {
		response = "Sorry, I encountered an error while processing your request."
		fmt.Printf("Error processing message: %v\n", err)
	}

	// Send response if any
	if response != "" {
		h.SendMessage(ctx, chatJID, response)
	}
}

// SendMessage sends a message to WhatsApp
func (h *WhatsAppHandler) SendMessage(ctx context.Context, recipient types.JID, text string) {
	_, err := h.client.SendMessage(ctx, recipient, &waE2E.Message{
		Conversation: &text,
	})
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
	}
}
