package ia

import (
	"blockmind/internal/config"
	"blockmind/internal/security"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AskQuestion sends a question to the Hugging Face API and returns the answer
func AskQuestion(question string, cfg *config.Config) (string, error) {
	// Sanitize the input question first
	question = security.SanitizeInput(question)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.AITimeout)
	defer cancel()

	// Validate config (should already be validated, but just in case)
	if cfg.HuggingFaceAPIKey == "" || cfg.HuggingFaceModel == "" {
		return "", fmt.Errorf("missing required API key or model name in configuration")
	}

	// Improved prompt structure with clear separation of instructions
	systemPrompt := `You are a precise question-answering system.
Rules:
1. Respond ONLY with the answer, no extra text or formatting
2. Match the question's language exactly
3. Never use markdown or special characters
4. Stop generation immediately after answer`

	userPrompt := fmt.Sprintf("Question: %s", question)

	requestBody := map[string]interface{}{
		"model": cfg.HuggingFaceModel,
		"messages": []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"temperature": cfg.AITemperature,
		"max_tokens":  cfg.AIMaxTokens,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		cfg.GetHuggingFaceAPIURL(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+cfg.HuggingFaceAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request with custom client and timeout
	client := &http.Client{
		Timeout: cfg.AITimeout - 1*time.Second, // Slightly less than context timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Read response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %s: %s", resp.Status, string(respBody))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract answer using a more robust approach with type assertions
	if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	// Try to extract the answer from different possible keys
	if answer, ok := response["answer"].(string); ok {
		return answer, nil
	}

	// Some models might return an array of answers
	if answers, ok := response["answers"].([]interface{}); ok && len(answers) > 0 {
		if firstAnswer, ok := answers[0].(map[string]interface{}); ok {
			if text, ok := firstAnswer["answer"].(string); ok {
				return text, nil
			}
		}
	}

	// If we can't find a structured answer, return a generic message
	return "I couldn't understand the response from the AI service.", nil
}
