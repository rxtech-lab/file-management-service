package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// InvoiceConfig holds configuration for the invoice processing service
type InvoiceConfig struct {
	ServerURL    string
	McpServerURL string
}

// InvoiceStreamEvent represents a streaming response event from the invoice API
type InvoiceStreamEvent struct {
	Status    string `json:"status"`               // "calling", "complete", "error"
	ToolName  string `json:"toolName,omitempty"`   // Tool being called (for "calling" status)
	Message   string `json:"message"`              // Human-readable message
	InvoiceID *int64 `json:"invoice_id,omitempty"` // Invoice ID (only on "complete" status)
}

// InvoiceResult represents the final result of invoice processing
type InvoiceResult struct {
	InvoiceID int64
	Message   string
}

// InvoiceService handles invoice processing via external API
type InvoiceService interface {
	// ProcessInvoice calls the external invoice API with streaming
	// Returns the invoice_id on success
	// authToken: OAuth Bearer token from user context
	// eventChan: receives real-time status updates (optional, can be nil)
	ProcessInvoice(ctx context.Context, fileURL, authToken string, eventChan chan<- InvoiceStreamEvent) (*InvoiceResult, error)

	// DeleteInvoice deletes an invoice by ID via the external API
	// Returns nil on success, if no auth token, or if service is not enabled
	// Retries 3 times on failure (total 4 attempts)
	DeleteInvoice(ctx context.Context, invoiceID int64, authToken string) error

	// IsEnabled returns whether invoice processing is configured
	IsEnabled() bool
}

type invoiceService struct {
	config InvoiceConfig
	client *http.Client
}

// NewInvoiceService creates a new invoice service instance
func NewInvoiceService(config InvoiceConfig) InvoiceService {
	return &invoiceService{
		config: config,
		client: &http.Client{},
	}
}

func (s *invoiceService) IsEnabled() bool {
	return s.config.ServerURL != ""
}

func (s *invoiceService) ProcessInvoice(ctx context.Context, fileURL, authToken string, eventChan chan<- InvoiceStreamEvent) (*InvoiceResult, error) {
	if eventChan != nil {
		defer close(eventChan)
	}

	// Prepare request body
	body := map[string]string{"file_url": fileURL}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	endpoint := s.config.ServerURL + "/api/invoices/agent"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	log.Printf("[Invoice] Calling invoice API: %s", endpoint)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Parse streaming response (SSE format: "event: type\ndata: {...}\n\n")
	scanner := bufio.NewScanner(resp.Body)
	var result *InvoiceResult

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and event type lines
		if line == "" || strings.HasPrefix(line, "event:") {
			continue
		}

		// Extract JSON from "data: {...}" lines
		jsonData := line
		if strings.HasPrefix(line, "data:") {
			jsonData = strings.TrimPrefix(line, "data:")
			jsonData = strings.TrimSpace(jsonData)
		}

		if jsonData == "" {
			continue
		}

		var event InvoiceStreamEvent
		if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
			log.Printf("[Invoice] Failed to parse event: %v (line: %s)", err, line)
			continue // Skip malformed lines
		}

		// Send event to channel if provided
		if eventChan != nil {
			select {
			case eventChan <- event:
			default:
				// Channel full, skip event
			}
		}

		// Check for completion
		if event.Status == "complete" && event.InvoiceID != nil {
			result = &InvoiceResult{
				InvoiceID: *event.InvoiceID,
				Message:   event.Message,
			}
			log.Printf("[Invoice] Received invoice_id: %d", result.InvoiceID)
		}

		// Check for error
		if event.Status == "error" {
			return nil, fmt.Errorf("invoice API error: %s", event.Message)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if result == nil {
		return nil, fmt.Errorf("no invoice_id received from server")
	}

	return result, nil
}

// DeleteInvoice deletes an invoice by ID with retry logic
// Returns nil on success, if no auth token is provided, or if service is not enabled
func (s *invoiceService) DeleteInvoice(ctx context.Context, invoiceID int64, authToken string) error {
	if authToken == "" {
		log.Printf("[Invoice] No auth token provided, skipping invoice deletion for invoice_id=%d", invoiceID)
		return nil
	}

	if !s.IsEnabled() {
		log.Printf("[Invoice] Service not enabled, skipping invoice deletion for invoice_id=%d", invoiceID)
		return nil
	}

	endpoint := fmt.Sprintf("%s/api/invoices/%d", s.config.McpServerURL, invoiceID)
	retryDelays := []time.Duration{100 * time.Millisecond, 500 * time.Millisecond, 1000 * time.Millisecond}

	var lastErr error
	for attempt := 0; attempt <= len(retryDelays); attempt++ {
		if attempt > 0 {
			log.Printf("[Invoice] Retrying invoice deletion (attempt %d/%d) for invoice_id=%d", attempt, len(retryDelays), invoiceID)
			time.Sleep(retryDelays[attempt-1])
		}

		req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}
		req.Header.Set("Authorization", "Bearer "+authToken)

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		resp.Body.Close()

		// Success cases: 200 OK, 204 No Content, 404 Not Found (already deleted)
		if resp.StatusCode == http.StatusOK ||
			resp.StatusCode == http.StatusNoContent ||
			resp.StatusCode == http.StatusNotFound {
			log.Printf("[Invoice] Successfully deleted invoice_id=%d (status: %d)", invoiceID, resp.StatusCode)
			return nil
		}

		lastErr = fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	log.Printf("[Invoice] Failed to delete invoice_id=%d after %d attempts: %v", invoiceID, len(retryDelays)+1, lastErr)
	return lastErr
}
