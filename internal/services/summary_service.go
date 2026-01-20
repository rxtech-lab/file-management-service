package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SummaryConfig holds configuration for the summary service
type SummaryConfig struct {
	GatewayURL string // e.g., https://ai-gateway.vercel.sh/v1
	APIKey     string // AI Gateway API key
	Model      string // e.g., openai/gpt-4o-mini
}

// SummaryService handles AI-powered summary generation
type SummaryService interface {
	GenerateSummary(ctx context.Context, content string, maxLength int) (string, error)
}

type summaryService struct {
	config SummaryConfig
	client *http.Client
}

// NewSummaryService creates a new SummaryService
func NewSummaryService(config SummaryConfig) SummaryService {
	return &summaryService{
		config: config,
		client: &http.Client{},
	}
}

// chatRequest is the request body for the chat completions API
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the response from the chat completions API
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// GenerateSummary generates a summary using AI
func (s *summaryService) GenerateSummary(ctx context.Context, content string, maxLength int) (string, error) {
	if content == "" {
		return "", nil
	}

	// Truncate content if too long (to fit in context window)
	if len(content) > 15000 {
		content = content[:15000]
	}

	prompt := fmt.Sprintf(`Summarize the following document content in a concise manner.
The summary should be no longer than %d characters and capture the key points.
Do not include any preamble like "Here is a summary" - just provide the summary directly.

Document content:
%s`, maxLength, content)

	reqBody := chatRequest{
		Model: s.config.Model,
		Messages: []chatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(s.config.GatewayURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("summary API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("summary API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from summary API")
	}

	summary := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// Truncate if still too long
	if len(summary) > maxLength {
		summary = summary[:maxLength-3] + "..."
	}

	return summary, nil
}

// MockSummaryService is a mock implementation for TESTING ONLY.
// Do not use in production. Production code requires proper AI_GATEWAY_URL
// and AI_GATEWAY_API_KEY environment variables.
type MockSummaryService struct{}

// NewMockSummaryService creates a mock summary service for TESTING ONLY.
func NewMockSummaryService() SummaryService {
	return &MockSummaryService{}
}

func (m *MockSummaryService) GenerateSummary(ctx context.Context, content string, maxLength int) (string, error) {
	// Fall back to simple truncation for testing
	return GenerateSummary(content, maxLength), nil
}
