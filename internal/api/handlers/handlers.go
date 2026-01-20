package handlers

import (
	"context"
	"errors"
	"fmt"

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
	agentService         services.AgentService
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
	agentService services.AgentService,
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
		agentService:         agentService,
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

// GetAgentStatus returns the status of the AI agent service
func (h *StrictHandlers) GetAgentStatus(
	ctx context.Context,
	request generated.GetAgentStatusRequestObject,
) (generated.GetAgentStatusResponseObject, error) {
	enabled := h.agentService != nil && h.agentService.IsEnabled()
	return generated.GetAgentStatus200JSONResponse{
		Enabled: enabled,
	}, nil
}

// OrganizeFile triggers the AI agent to organize a file
func (h *StrictHandlers) OrganizeFile(
	ctx context.Context,
	request generated.OrganizeFileRequestObject,
) (generated.OrganizeFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.OrganizeFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	fileID := uint(request.Id)

	// Check if agent is enabled
	if h.agentService == nil || !h.agentService.IsEnabled() {
		return generated.OrganizeFile503JSONResponse{Error: "AI agent is not enabled"}, nil
	}

	// Verify file ownership
	file, err := h.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		return generated.OrganizeFile404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Return the stream URL for the client to subscribe to
	return generated.OrganizeFile200JSONResponse{
		Message:   "Agent organization started",
		FileId:    int(fileID),
		StreamUrl: fmt.Sprintf("/api/files/%d/agent-stream", fileID),
	}, nil
}

// StreamAgentProgress streams AI agent progress events (SSE endpoint - needs special handling)
func (h *StrictHandlers) StreamAgentProgress(
	ctx context.Context,
	request generated.StreamAgentProgressRequestObject,
) (generated.StreamAgentProgressResponseObject, error) {
	// This is an SSE endpoint that requires special Fiber handling
	// The actual implementation is in AgentHandlers.StreamAgentProgress
	// Return 503 as this strict handler cannot handle SSE
	return nil, errors.New("SSE endpoints should use direct Fiber handlers")
}
