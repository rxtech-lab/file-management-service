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

// ContentParserConfig holds configuration for the content parser service
type ContentParserConfig struct {
	EndpointURL string // e.g., https://your-python-service/convert
	APIKey      string // ADMIN_API_KEY for authentication
}

// ParsedContent represents the result of content parsing
type ParsedContent struct {
	TextContent string `json:"content"`
}

// ContentParserService handles file content parsing via external Python service
type ContentParserService interface {
	ParseFileContent(ctx context.Context, fileURL string) (*ParsedContent, error)
}

type contentParserService struct {
	config ContentParserConfig
	client *http.Client
}

// NewContentParserService creates a new ContentParserService
func NewContentParserService(config ContentParserConfig) ContentParserService {
	return &contentParserService{
		config: config,
		client: &http.Client{},
	}
}

// convertRequest is the request body for the convert API
type convertRequest struct {
	File string `json:"file"`
}

// ParseFileContent parses file content from a URL using the Python service
func (s *contentParserService) ParseFileContent(ctx context.Context, fileURL string) (*ParsedContent, error) {
	if fileURL == "" {
		return nil, fmt.Errorf("file URL cannot be empty")
	}

	reqBody := convertRequest{
		File: fileURL,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := strings.TrimSuffix(s.config.EndpointURL, "/") + "/convert"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("content parser error (status %d): %s", resp.StatusCode, string(body))
	}

	var parsed ParsedContent
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &parsed, nil
}

// MockContentParserService is a mock implementation for TESTING ONLY.
// Do not use in production. Production code requires proper CONTENT_PARSER_ENDPOINT
// environment variable.
type MockContentParserService struct{}

// NewMockContentParserService creates a mock content parser service for TESTING ONLY.
func NewMockContentParserService() ContentParserService {
	return &MockContentParserService{}
}

func (m *MockContentParserService) ParseFileContent(ctx context.Context, fileURL string) (*ParsedContent, error) {
	// Return mock content based on URL
	return &ParsedContent{
		TextContent: fmt.Sprintf("Mock parsed content from: %s\n\nThis is sample text content for testing purposes.", fileURL),
	}, nil
}

// GenerateSummary creates a summary from the content
// This is a simple implementation - in production, you might use an LLM
func GenerateSummary(content string, maxLength int) string {
	if content == "" {
		return ""
	}

	// Simple summary: take the first paragraph or maxLength characters
	content = strings.TrimSpace(content)

	// Find first paragraph
	paragraphEnd := strings.Index(content, "\n\n")
	if paragraphEnd > 0 && paragraphEnd < maxLength {
		return strings.TrimSpace(content[:paragraphEnd])
	}

	// Truncate at maxLength
	if len(content) > maxLength {
		// Try to break at word boundary
		truncated := content[:maxLength]
		lastSpace := strings.LastIndex(truncated, " ")
		if lastSpace > maxLength/2 {
			return truncated[:lastSpace] + "..."
		}
		return truncated + "..."
	}

	return content
}
