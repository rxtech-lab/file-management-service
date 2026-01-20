package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// EmbeddingConfig holds configuration for the embedding service
type EmbeddingConfig struct {
	GatewayURL string // e.g., https://ai-gateway.vercel.sh/v1
	APIKey     string // AI Gateway API key
	Model      string // e.g., openai/text-embedding-3-small
	Dimensions int    // e.g., 1536
}

// EmbeddingService handles embedding generation and storage
type EmbeddingService interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	StoreFileEmbedding(userID string, fileID uint, embedding []float32) error
	GetFileEmbedding(userID string, fileID uint) ([]float32, error)
	DeleteFileEmbedding(userID string, fileID uint) error
}

type embeddingService struct {
	db     *gorm.DB
	config EmbeddingConfig
	client *http.Client
}

// NewEmbeddingService creates a new EmbeddingService
func NewEmbeddingService(db *gorm.DB, config EmbeddingConfig) EmbeddingService {
	return &embeddingService{
		db:     db,
		config: config,
		client: &http.Client{},
	}
}

// embeddingRequest is the request body for the embeddings API
type embeddingRequest struct {
	Input      string `json:"input"`
	Model      string `json:"model"`
	Dimensions int    `json:"dimensions,omitempty"`
}

// embeddingResponse is the response from the embeddings API
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// GenerateEmbedding generates an embedding for the given text
func (s *embeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	// Truncate text if too long (most models have token limits)
	if len(text) > 8000 {
		text = text[:8000]
	}

	reqBody := embeddingRequest{
		Input: text,
		Model: s.config.Model,
	}
	if s.config.Dimensions > 0 {
		reqBody.Dimensions = s.config.Dimensions
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", strings.TrimSuffix(s.config.GatewayURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

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
		return nil, fmt.Errorf("embedding API error (status %d): %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embResp.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", embResp.Error.Message)
	}

	if len(embResp.Data) == 0 {
		return nil, errors.New("no embedding data in response")
	}

	return embResp.Data[0].Embedding, nil
}

// StoreFileEmbedding stores an embedding for a file
func (s *embeddingService) StoreFileEmbedding(userID string, fileID uint, embedding []float32) error {
	// Convert embedding to JSON string
	embJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	fileEmbedding := models.FileEmbedding{
		FileID:    fileID,
		UserID:    userID,
		Embedding: string(embJSON),
	}

	// Upsert: update if exists, create if not
	result := s.db.Where("file_id = ?", fileID).FirstOrCreate(&fileEmbedding)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Record already exists, update it
		return s.db.Model(&models.FileEmbedding{}).
			Where("file_id = ?", fileID).
			Update("embedding", string(embJSON)).Error
	}

	return nil
}

// GetFileEmbedding retrieves the embedding for a file
func (s *embeddingService) GetFileEmbedding(userID string, fileID uint) ([]float32, error) {
	var fileEmbedding models.FileEmbedding
	err := s.db.Where("file_id = ? AND user_id = ?", fileID, userID).First(&fileEmbedding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var embedding []float32
	if err := json.Unmarshal([]byte(fileEmbedding.Embedding), &embedding); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
	}

	return embedding, nil
}

// DeleteFileEmbedding deletes the embedding for a file
func (s *embeddingService) DeleteFileEmbedding(userID string, fileID uint) error {
	result := s.db.Where("file_id = ? AND user_id = ?", fileID, userID).Delete(&models.FileEmbedding{})
	return result.Error
}

// MockEmbeddingService is a mock implementation for TESTING ONLY.
// Do not use in production. Production code requires proper AI_GATEWAY_URL
// and AI_GATEWAY_API_KEY environment variables.
type MockEmbeddingService struct {
	embeddings map[uint][]float32
}

// NewMockEmbeddingService creates a mock embedding service for TESTING ONLY.
func NewMockEmbeddingService() EmbeddingService {
	return &MockEmbeddingService{
		embeddings: make(map[uint][]float32),
	}
}

func (m *MockEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Generate a simple deterministic embedding based on text length
	embedding := make([]float32, 1536)
	for i := 0; i < 1536; i++ {
		embedding[i] = float32(len(text)%100) / 100.0
	}
	return embedding, nil
}

func (m *MockEmbeddingService) StoreFileEmbedding(userID string, fileID uint, embedding []float32) error {
	m.embeddings[fileID] = embedding
	return nil
}

func (m *MockEmbeddingService) GetFileEmbedding(userID string, fileID uint) ([]float32, error) {
	if emb, ok := m.embeddings[fileID]; ok {
		return emb, nil
	}
	return nil, nil
}

func (m *MockEmbeddingService) DeleteFileEmbedding(userID string, fileID uint) error {
	delete(m.embeddings, fileID)
	return nil
}

// EmbeddingToString converts an embedding to a string representation for Turso vector operations
func EmbeddingToString(embedding []float32) string {
	parts := make([]string, len(embedding))
	for i, v := range embedding {
		parts[i] = strconv.FormatFloat(float64(v), 'f', 6, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
