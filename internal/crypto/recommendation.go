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
	"time"
)

func GetCryptoRecommendation(cryptoName string, cfg *config.Config) (string, error) {

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.AITimeout)
	defer cancel()

	// To lowercase
	cryptoName = strings.ToLower(cryptoName)

	url := fmt.Sprintf("%s/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&locale=en&precision=full", cfg.CoingeckoBaseURL, url.QueryEscape(cryptoName))

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
	var marketData []map[string]interface{}
	if err := json.Unmarshal(body, &marketData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if we got any data
	if len(marketData) == 0 {
		return "", fmt.Errorf("no data found for cryptocurrency: %s", cryptoName)
	}

	// Get the first item in the array
	data := marketData[0]

	// Format a nice response with relevant information
	recommendation := fmt.Sprintf("*%s (%s)*\n\n",
		data["name"],
		strings.ToUpper(data["symbol"].(string)))

	// Display all available data fields
	for key, value := range data {
		if value == nil {
			recommendation += fmt.Sprintf("• %s: N/A\n", formatKeyName(key))
			continue
		}

		switch v := value.(type) {
		case float64:
			// Format price-related fields with dollar signs and 2 decimal places
			if strings.Contains(key, "percentage") || strings.Contains(key, "change_percentage") {
				// Format percentage fields
				recommendation += fmt.Sprintf("• %s: %.2f%%\n", formatKeyName(key), v)
			} else if strings.Contains(key, "market_cap") || strings.Contains(key, "volume") || strings.Contains(key, "valuation") {
				// Format large numbers with commas
				recommendation += fmt.Sprintf("• %s: %s\n", formatKeyName(key), formatLargeNumber(v))
			} else if strings.Contains(key, "price") || strings.Contains(key, "ath") || strings.Contains(key, "atl") ||
				strings.Contains(key, "high") || strings.Contains(key, "low") {
				recommendation += fmt.Sprintf("• %s: $%.2f\n", formatKeyName(key), v)
			} else {
				// Format other numbers normally
				recommendation += fmt.Sprintf("• %s: %.0f\n", formatKeyName(key), v)
			}
		case string:
			if strings.Contains(key, "date") {
				// Format dates more nicely
				recommendation += fmt.Sprintf("• %s: %s\n", formatKeyName(key), formatDate(v))
			} else if strings.Contains(key, "image") {
				// Skip image URLs to keep output cleaner
				continue
			} else {
				recommendation += fmt.Sprintf("• %s: %s\n", formatKeyName(key), v)
			}
		default:
			// Just convert to string for other types
			recommendation += fmt.Sprintf("• %s: %v\n", formatKeyName(key), v)
		}
	}

	return recommendation, nil
}

func GetSentimentAndHistoricalData(data string, cryptoName string, cfg *config.Config) (string, error) {

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.AITimeout)
	defer cancel()

	// To lowercase
	cryptoName = strings.ToLower(cryptoName)

	url := fmt.Sprintf("%s/coins/%s?localization=false&tickers=true&market_data=true&community_data=true&developer_data=false", cfg.CoingeckoBaseURL, url.QueryEscape(cryptoName))

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

	// Parse JSON to extract the data - using a single map not an array
	var coinData map[string]interface{}
	if err := json.Unmarshal(body, &coinData); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Start with the existing data if provided
	result := data

	// Extract sentiment data
	if sentimentUp, ok := coinData["sentiment_votes_up_percentage"].(float64); ok {
		result += fmt.Sprintf("• %s: %.2f%%\n", formatKeyName("sentiment_votes_up_percentage"), sentimentUp)
	}

	if sentimentDown, ok := coinData["sentiment_votes_down_percentage"].(float64); ok {
		result += fmt.Sprintf("• %s: %.2f%%\n", formatKeyName("sentiment_votes_down_percentage"), sentimentDown)
	}

	// Extract market data and price change percentages
	if marketData, ok := coinData["market_data"].(map[string]interface{}); ok {
		// Extract price change percentages
		priceChangeKeys := []string{
			"price_change_percentage_7d",
			"price_change_percentage_14d",
			"price_change_percentage_30d",
			"price_change_percentage_60d",
		}

		for _, key := range priceChangeKeys {
			if val, ok := marketData[key].(float64); ok {
				result += fmt.Sprintf("• %s: %.2f%%\n", formatKeyName(key), val)
			}
		}
	}

	return result, nil
}

// Helper function to format key names more nicely
func formatKeyName(key string) string {
	// Convert snake_case to Title Case
	words := strings.Split(key, "_")
	for i := 0; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}
	return strings.Join(words, " ")
}

// Helper function to format dates more nicely
func formatDate(dateStr string) string {
	if len(dateStr) > 19 {
		dateStr = dateStr[:19]
	}
	t, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan 02, 2006 15:04:05")
}

// Helper function to format large numbers with commas
func formatLargeNumber(num float64) string {
	if num >= 1000000000 {
		return fmt.Sprintf("%.2f billion", num/1000000000)
	} else if num >= 1000000 {
		return fmt.Sprintf("%.2f million", num/1000000)
	} else if num >= 1000 {
		return fmt.Sprintf("%.2f thousand", num/1000)
	}
	return fmt.Sprintf("%.2f", num)
}
