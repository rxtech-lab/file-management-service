package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// SearchFiles implements generated.StrictServerInterface
func (h *StrictHandlers) SearchFiles(
	ctx context.Context,
	request generated.SearchFilesRequestObject,
) (generated.SearchFilesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.SearchFiles401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	query := request.Params.Q
	if query == "" {
		return generated.SearchFiles400JSONResponse{BadRequestJSONResponse: badRequest("Search query is required")}, nil
	}

	opts := services.SearchOptions{
		Limit:  derefInt(request.Params.Limit, 20),
		Offset: derefInt(request.Params.Offset, 0),
	}

	// Handle folder_id
	if request.Params.FolderId != nil {
		folderID := uint(*request.Params.FolderId)
		opts.FolderID = &folderID
	}

	// Handle file_type
	if request.Params.FileType != nil {
		ft := models.FileType(*request.Params.FileType)
		opts.FileTypes = []models.FileType{ft}
	}

	// Parse tag IDs
	if request.Params.TagIds != nil && *request.Params.TagIds != "" {
		tagIDStrs := strings.Split(*request.Params.TagIds, ",")
		for _, idStr := range tagIDStrs {
			id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 32)
			if err == nil {
				opts.TagIDs = append(opts.TagIDs, uint(id))
			}
		}
	}

	// Determine search type
	searchType := "fulltext"
	if request.Params.Type != nil {
		searchType = string(*request.Params.Type)
	}

	var results []services.SearchResult
	var total int64

	switch searchType {
	case "semantic":
		results, err = h.searchService.VectorSearch(ctx, userID, query, opts)
		total = int64(len(results))
	case "hybrid":
		results, err = h.searchService.HybridSearch(ctx, userID, query, opts)
		total = int64(len(results))
	default: // fulltext
		results, total, err = h.searchService.FullTextSearch(userID, query, opts)
	}

	if err != nil {
		return generated.SearchFiles400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.SearchFiles200JSONResponse{
		Data:       searchResultListToGenerated(results),
		Total:      int(total),
		Query:      query,
		SearchType: searchType,
	}, nil
}
