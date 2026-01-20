package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// SearchResult represents a search result with relevance score
type SearchResult struct {
	File    models.File `json:"file"`
	Score   float64     `json:"score"`
	Snippet string      `json:"snippet,omitempty"`
}

// SearchOptions contains options for search operations
type SearchOptions struct {
	FolderID  *uint
	TagIDs    []uint
	FileTypes []models.FileType
	Limit     int
	Offset    int
}

// SearchService handles search operations including full-text and vector search
type SearchService interface {
	// FullTextSearch searches files using SQL LIKE on title and content
	FullTextSearch(userID string, query string, opts SearchOptions) ([]SearchResult, int64, error)

	// VectorSearch performs semantic search using embeddings
	VectorSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error)

	// HybridSearch combines full-text and vector search
	HybridSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error)
}

type searchService struct {
	db               *gorm.DB
	embeddingService EmbeddingService
}

// NewSearchService creates a new SearchService
func NewSearchService(db *gorm.DB, embeddingService EmbeddingService) SearchService {
	return &searchService{
		db:               db,
		embeddingService: embeddingService,
	}
}

// FullTextSearch performs a full-text search on files
func (s *searchService) FullTextSearch(userID string, query string, opts SearchOptions) ([]SearchResult, int64, error) {
	var files []models.File
	var total int64

	dbQuery := s.db.Model(&models.File{}).
		Where("user_id = ?", userID).
		Where("processing_status = ?", models.FileStatusCompleted)

	// Search in title, summary, and content
	searchPattern := "%" + query + "%"
	dbQuery = dbQuery.Where(
		"title LIKE ? OR summary LIKE ? OR content LIKE ?",
		searchPattern, searchPattern, searchPattern,
	)

	// Apply filters
	if opts.FolderID != nil {
		dbQuery = dbQuery.Where("folder_id = ?", *opts.FolderID)
	}

	if len(opts.FileTypes) > 0 {
		dbQuery = dbQuery.Where("file_type IN ?", opts.FileTypes)
	}

	if len(opts.TagIDs) > 0 {
		dbQuery = dbQuery.Joins("JOIN file_tags ON file_tags.file_id = files.id").
			Where("file_tags.tag_id IN ?", opts.TagIDs).
			Group("files.id")
	}

	// Count total
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and fetch results
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := dbQuery.Preload("Tags").Preload("Folder").
		Limit(limit).Offset(opts.Offset).
		Order("updated_at DESC").
		Find(&files).Error; err != nil {
		return nil, 0, err
	}

	// Convert to SearchResult with simple relevance scoring
	results := make([]SearchResult, len(files))
	for i, file := range files {
		score := s.calculateFullTextScore(file, query)
		snippet := s.generateSnippet(file.Content, query, 200)
		results[i] = SearchResult{
			File:    file,
			Score:   score,
			Snippet: snippet,
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, total, nil
}

// VectorSearch performs semantic search using embeddings
func (s *searchService) VectorSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get files with embeddings
	var fileEmbeddings []models.FileEmbedding
	embQuery := s.db.Where("user_id = ?", userID)
	if err := embQuery.Find(&fileEmbeddings).Error; err != nil {
		return nil, err
	}

	if len(fileEmbeddings) == 0 {
		return []SearchResult{}, nil
	}

	// Calculate similarity scores
	type scoredFile struct {
		FileID uint
		Score  float64
	}
	var scoredFiles []scoredFile

	for _, fe := range fileEmbeddings {
		embedding, err := parseEmbedding(fe.Embedding)
		if err != nil {
			continue
		}

		score := cosineSimilarity(queryEmbedding, embedding)
		scoredFiles = append(scoredFiles, scoredFile{
			FileID: fe.FileID,
			Score:  score,
		})
	}

	// Sort by score descending
	sort.Slice(scoredFiles, func(i, j int) bool {
		return scoredFiles[i].Score > scoredFiles[j].Score
	})

	// Apply limit
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	if len(scoredFiles) > limit {
		scoredFiles = scoredFiles[:limit]
	}

	// Fetch the actual files
	fileIDs := make([]uint, len(scoredFiles))
	scoreMap := make(map[uint]float64)
	for i, sf := range scoredFiles {
		fileIDs[i] = sf.FileID
		scoreMap[sf.FileID] = sf.Score
	}

	var files []models.File
	dbQuery := s.db.Where("id IN ? AND user_id = ?", fileIDs, userID).
		Where("processing_status = ?", models.FileStatusCompleted)

	// Apply filters
	if opts.FolderID != nil {
		dbQuery = dbQuery.Where("folder_id = ?", *opts.FolderID)
	}

	if len(opts.FileTypes) > 0 {
		dbQuery = dbQuery.Where("file_type IN ?", opts.FileTypes)
	}

	if len(opts.TagIDs) > 0 {
		dbQuery = dbQuery.Joins("JOIN file_tags ON file_tags.file_id = files.id").
			Where("file_tags.tag_id IN ?", opts.TagIDs).
			Group("files.id")
	}

	if err := dbQuery.Preload("Tags").Preload("Folder").Find(&files).Error; err != nil {
		return nil, err
	}

	// Build results maintaining score order
	results := make([]SearchResult, 0, len(files))
	for _, file := range files {
		results = append(results, SearchResult{
			File:    file,
			Score:   scoreMap[file.ID],
			Snippet: s.generateSnippet(file.Summary, "", 200),
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// HybridSearch combines full-text and vector search
func (s *searchService) HybridSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error) {
	// Perform both searches
	fullTextResults, _, err := s.FullTextSearch(userID, query, SearchOptions{
		FolderID:  opts.FolderID,
		TagIDs:    opts.TagIDs,
		FileTypes: opts.FileTypes,
		Limit:     50, // Get more for merging
	})
	if err != nil {
		return nil, fmt.Errorf("full-text search failed: %w", err)
	}

	vectorResults, err := s.VectorSearch(ctx, userID, query, SearchOptions{
		FolderID:  opts.FolderID,
		TagIDs:    opts.TagIDs,
		FileTypes: opts.FileTypes,
		Limit:     50,
	})
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Combine and normalize scores
	scoreMap := make(map[uint]float64)
	fileMap := make(map[uint]models.File)
	snippetMap := make(map[uint]string)

	// Weight: 40% full-text, 60% semantic
	const fullTextWeight = 0.4
	const vectorWeight = 0.6

	// Normalize and add full-text scores
	maxFTScore := 0.0
	for _, r := range fullTextResults {
		if r.Score > maxFTScore {
			maxFTScore = r.Score
		}
	}
	for _, r := range fullTextResults {
		normalizedScore := 0.0
		if maxFTScore > 0 {
			normalizedScore = r.Score / maxFTScore
		}
		scoreMap[r.File.ID] = normalizedScore * fullTextWeight
		fileMap[r.File.ID] = r.File
		snippetMap[r.File.ID] = r.Snippet
	}

	// Add vector scores
	maxVectorScore := 0.0
	for _, r := range vectorResults {
		if r.Score > maxVectorScore {
			maxVectorScore = r.Score
		}
	}
	for _, r := range vectorResults {
		normalizedScore := 0.0
		if maxVectorScore > 0 {
			normalizedScore = r.Score / maxVectorScore
		}
		scoreMap[r.File.ID] += normalizedScore * vectorWeight
		if _, exists := fileMap[r.File.ID]; !exists {
			fileMap[r.File.ID] = r.File
			snippetMap[r.File.ID] = r.Snippet
		}
	}

	// Build combined results
	var results []SearchResult
	for fileID, score := range scoreMap {
		results = append(results, SearchResult{
			File:    fileMap[fileID],
			Score:   score,
			Snippet: snippetMap[fileID],
		})
	}

	// Sort by combined score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// calculateFullTextScore calculates a simple relevance score for full-text search
func (s *searchService) calculateFullTextScore(file models.File, query string) float64 {
	score := 0.0

	// Title match is most important
	if containsIgnoreCase(file.Title, query) {
		score += 10.0
	}

	// Summary match is next
	if containsIgnoreCase(file.Summary, query) {
		score += 5.0
	}

	// Content match
	if containsIgnoreCase(file.Content, query) {
		score += 2.0
	}

	return score
}

// generateSnippet creates a snippet around the query match
func (s *searchService) generateSnippet(content, query string, maxLength int) string {
	if content == "" {
		return ""
	}

	if query == "" {
		if len(content) <= maxLength {
			return content
		}
		return content[:maxLength] + "..."
	}

	// Find query in content (case insensitive)
	lowerContent := toLower(content)
	lowerQuery := toLower(query)
	idx := indexOf(lowerContent, lowerQuery)

	if idx < 0 {
		// Query not found, return beginning
		if len(content) <= maxLength {
			return content
		}
		return content[:maxLength] + "..."
	}

	// Calculate snippet window
	start := idx - maxLength/4
	if start < 0 {
		start = 0
	}

	end := start + maxLength
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

// parseEmbedding parses an embedding from JSON string
func parseEmbedding(embJSON string) ([]float32, error) {
	var embedding []float32
	if err := json.Unmarshal([]byte(embJSON), &embedding); err != nil {
		return nil, err
	}
	return embedding, nil
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt is a simple square root implementation
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// containsIgnoreCase checks if s contains substr (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return indexOf(toLower(s), toLower(substr)) >= 0
}

// toLower converts string to lowercase
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// indexOf finds the index of substr in s, returns -1 if not found
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// MockSearchService is a mock implementation for testing
type MockSearchService struct {
	files []models.File
}

// NewMockSearchService creates a mock search service for testing
func NewMockSearchService() SearchService {
	return &MockSearchService{
		files: []models.File{},
	}
}

func (m *MockSearchService) FullTextSearch(userID string, query string, opts SearchOptions) ([]SearchResult, int64, error) {
	var results []SearchResult
	for _, f := range m.files {
		if f.UserID == userID && containsIgnoreCase(f.Title+" "+f.Content, query) {
			results = append(results, SearchResult{File: f, Score: 1.0})
		}
	}
	return results, int64(len(results)), nil
}

func (m *MockSearchService) VectorSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error) {
	results, _, err := m.FullTextSearch(userID, query, opts)
	return results, err
}

func (m *MockSearchService) HybridSearch(ctx context.Context, userID string, query string, opts SearchOptions) ([]SearchResult, error) {
	results, _, err := m.FullTextSearch(userID, query, opts)
	return results, err
}
