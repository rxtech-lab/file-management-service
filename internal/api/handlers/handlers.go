package handlers

import (
	"context"
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// Ensure StrictHandlers implements the StrictServerInterface
var _ generated.StrictServerInterface = (*StrictHandlers)(nil)

// Common errors
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// StrictHandlers implements the generated StrictServerInterface
type StrictHandlers struct {
	tagService           services.TagService
	folderService        services.FolderService
	fileService          services.FileService
	uploadService        services.UploadService
	embeddingService     services.EmbeddingService
	contentParserService services.ContentParserService
	searchService        services.SearchService
	summaryService       services.SummaryService
}

// NewStrictHandlers creates a new StrictHandlers instance
func NewStrictHandlers(
	tagService services.TagService,
	folderService services.FolderService,
	fileService services.FileService,
	uploadService services.UploadService,
	embeddingService services.EmbeddingService,
	contentParserService services.ContentParserService,
	searchService services.SearchService,
	summaryService services.SummaryService,
) *StrictHandlers {
	return &StrictHandlers{
		tagService:           tagService,
		folderService:        folderService,
		fileService:          fileService,
		uploadService:        uploadService,
		embeddingService:     embeddingService,
		contentParserService: contentParserService,
		searchService:        searchService,
		summaryService:       summaryService,
	}
}

// getUserID extracts user ID from context (set by authentication middleware)
func getUserID(ctx context.Context) (string, error) {
	user, ok := utils.GetAuthenticatedUser(ctx)
	if !ok || user == nil {
		return "", ErrUnauthorized
	}
	return user.Sub, nil
}

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// deref safely dereferences a pointer, returning zero value if nil
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// derefInt safely dereferences an int pointer with a default value
func derefInt(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}

// Error response helpers

func unauthorized() generated.UnauthorizedJSONResponse {
	return generated.UnauthorizedJSONResponse{Error: "Unauthorized"}
}

func badRequest(msg string) generated.BadRequestJSONResponse {
	return generated.BadRequestJSONResponse{Error: msg}
}

func notFound(msg string) generated.NotFoundJSONResponse {
	return generated.NotFoundJSONResponse{Error: msg}
}
