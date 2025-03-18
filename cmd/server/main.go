package main

import (
	"blockmind/internal/config"
	"blockmind/internal/handlers"
	"blockmind/internal/logger"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting BlockMind WhatsApp bot...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure structured logging
	logger.SetLevel(cfg.WhatsAppLogLevel)
	logger.Info("Starting BlockMind WhatsApp bot",
		logger.Field{Key: "debug_mode", Value: cfg.Debug},
		logger.Field{Key: "model", Value: cfg.HuggingFaceModel})

	// Setup database for WhatsApp
	dbLog := waLog.Stdout("Database", cfg.WhatsAppLogLevel, cfg.Debug)
	storeContainer, err := sqlstore.New("sqlite3", cfg.WhatsAppDBPath, dbLog)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Get device or create new one
	device, err := storeContainer.GetFirstDevice()
	if err != nil {
		log.Println("Creating new device")
		device = storeContainer.NewDevice()
	}

	// Create WhatsApp client
	client := whatsmeow.NewClient(device, waLog.Stdout("WhatsApp", cfg.WhatsAppLogLevel, cfg.Debug))

	// Create WhatsApp handler
	whatsappHandler := handlers.NewWhatsAppHandler(client, cfg)

	// Register event handlers
	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.QR:
			log.Println("Scan the QR code to authenticate.")
			qrConfig := qrterminal.Config{
				Level:      qrterminal.L,
				Writer:     os.Stdout,
				BlackChar:  qrterminal.BLACK,
				WhiteChar:  qrterminal.WHITE,
				QuietZone:  0,
				HalfBlocks: false,
				WithSixel:  false,
			}
			qrterminal.GenerateWithConfig(v.Codes[0], qrConfig)

		case *events.Message:
			if v.Info.IsFromMe {
				return
			}
			whatsappHandler.HandleMessage(v.Message, v.Info.Chat)

		case *events.Connected:
			log.Println("Connected to WhatsApp")

		case *events.LoggedOut:
			log.Println("Logged out from WhatsApp")
		}
	})

	// Connect to WhatsApp
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	log.Println("WhatsApp bot is running")

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down...")
	client.Disconnect()
	time.Sleep(500 * time.Millisecond) // Give time for cleanup
}
