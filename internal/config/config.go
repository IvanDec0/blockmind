package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// AI Service
	HuggingFaceAPIKey string
	HuggingFaceModel  string
	HuggingFaceAPIURL string
	AITimeout         time.Duration
	AIMaxTokens       int
	AITemperature     float64

	// Coingecko
	CoingeckoAPIKey  string
	CoingeckoBaseURL string

	// WhatsApp
	WhatsAppDBPath   string
	WhatsAppLogLevel string

	// Rate Limiting
	RateLimit       int
	RateLimitPeriod time.Duration

	// General
	CommandTimeout time.Duration
	Debug          bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but continue if it doesn't exist
	_ = godotenv.Load()

	config := &Config{
		// Default values
		AITimeout:         20 * time.Second,
		AIMaxTokens:       250,
		AITemperature:     0.0,
		HuggingFaceAPIURL: "https://router.huggingface.co/hf-inference/models/",
		CoingeckoBaseURL:  "https://api.coingecko.com/api/v3",
		WhatsAppDBPath:    "file:whatsapp.db?_foreign_keys=on",
		WhatsAppLogLevel:  "INFO",
		RateLimit:         5,
		RateLimitPeriod:   time.Minute,
		CommandTimeout:    25 * time.Second,
		Debug:             false,
	}

	// Required values
	config.HuggingFaceAPIKey = os.Getenv("HUGGINGFACE_API_KEY")
	config.HuggingFaceModel = os.Getenv("HUGGINGFACE_MODEL")
	config.CoingeckoAPIKey = os.Getenv("COINGECKO_API_KEY")

	// Optional values with overrides
	if val := os.Getenv("AI_TIMEOUT"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			config.AITimeout = time.Duration(seconds) * time.Second
		}
	}

	if val := os.Getenv("AI_MAX_TOKENS"); val != "" {
		if tokens, err := strconv.Atoi(val); err == nil {
			config.AIMaxTokens = tokens
		}
	}

	if val := os.Getenv("AI_TEMPERATURE"); val != "" {
		if temp, err := strconv.ParseFloat(val, 64); err == nil {
			config.AITemperature = temp
		}
	}

	if val := os.Getenv("WHATSAPP_DB_PATH"); val != "" {
		config.WhatsAppDBPath = val
	}

	if val := os.Getenv("WHATSAPP_LOG_LEVEL"); val != "" {
		config.WhatsAppLogLevel = val
	}

	if val := os.Getenv("RATE_LIMIT"); val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			config.RateLimit = limit
		}
	}

	if val := os.Getenv("RATE_LIMIT_PERIOD"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			config.RateLimitPeriod = time.Duration(seconds) * time.Second
		}
	}

	if val := os.Getenv("COMMAND_TIMEOUT"); val != "" {
		if seconds, err := strconv.Atoi(val); err == nil {
			config.CommandTimeout = time.Duration(seconds) * time.Second
		}
	}

	if val := os.Getenv("COINGECKO_API_URL"); val != "" {
		config.CoingeckoBaseURL = val
	}

	if val := os.Getenv("HUGGINGFACE_BASE_URL"); val != "" {
		config.HuggingFaceAPIURL = val
	}

	if val := os.Getenv("DEBUG"); val == "true" {
		config.Debug = true
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// validate ensures all required configuration values are set
func (c *Config) validate() error {
	if c.HuggingFaceAPIKey == "" {
		return fmt.Errorf("missing required environment variable: HUGGINGFACE_API_KEY")
	}
	if c.HuggingFaceModel == "" {
		return fmt.Errorf("missing required environment variable: HUGGINGFACE_MODEL")
	}

	if c.CoingeckoAPIKey == "" {
		return fmt.Errorf("missing required environment variable: COINGECKO_API_KEY")
	}
	return nil
}

// GetHuggingFaceAPIURL returns the constructed API URL for the HuggingFace model
func (c *Config) GetHuggingFaceAPIURL() string {
	return c.HuggingFaceAPIURL + c.HuggingFaceModel + "/v1/chat/completions"
}
