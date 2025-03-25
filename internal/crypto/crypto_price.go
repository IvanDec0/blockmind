package crypto

import (
	"blockmind/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetCryptoPrice(crypto string, target string, cfg *config.Config) (string, error) {
	// Get the price of a cryptocurrency

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.AITimeout)
	defer cancel()

	// To lowercase
	crypto = strings.ToLower(crypto)
	target = strings.ToLower(target)

	// Set Default target currency to USD
	if target == "" {
		target = "usd"
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=%s&include_market_cap=false&include_24hr_vol=false&include_24hr_change=false&include_last_updated_at=false&precision=full", cfg.CoingeckoBaseURL, url.QueryEscape(crypto), url.QueryEscape(target))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Setup headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-cg-demo-api-key", cfg.CoingeckoAPIKey)

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON to extract the price
	var priceData map[string]map[string]float64
	if err := json.Unmarshal(body, &priceData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if priceInfo, ok := priceData[crypto]; ok {
		if price, ok := priceInfo[target]; ok {
			return fmt.Sprintf("%s -> %.4f %s", crypto, price, strings.ToUpper(target)), nil
		}
	}

	return "", fmt.Errorf("price data not found for %s in %s", crypto, target)
}
