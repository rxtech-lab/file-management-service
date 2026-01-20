package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// ListFolders implements generated.StrictServerInterface
func (h *StrictHandlers) ListFolders(
	ctx context.Context,
	request generated.ListFoldersRequestObject,
) (generated.ListFoldersResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListFolders401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	opts := services.FolderListOptions{
		Keyword: deref(request.Params.Keyword),
		Limit:   derefInt(request.Params.Limit, 100),
		Offset:  derefInt(request.Params.Offset, 0),
	}

	// Handle parent_id
	if request.Params.ParentId != nil {
		parentID := uint(*request.Params.ParentId)
		opts.ParentID = &parentID
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

	folders, total, err := h.folderService.ListFolders(userID, opts)
	if err != nil {
		return nil, err
	}

	return generated.ListFolders200JSONResponse{
		Data:   folderListToGenerated(folders),
		Total:  int(total),
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}, nil
}

// CreateFolder implements generated.StrictServerInterface
func (h *StrictHandlers) CreateFolder(
	ctx context.Context,
	request generated.CreateFolderRequestObject,
) (generated.CreateFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.CreateFolder400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	folder := &models.Folder{
		Name:        request.Body.Name,
		Description: deref(request.Body.Description),
	}

	if request.Body.ParentId != nil {
		parentID := uint(*request.Body.ParentId)
		folder.ParentID = &parentID
	}

	if err := h.folderService.CreateFolder(userID, folder); err != nil {
		return generated.CreateFolder400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch folder with tags
	created, err := h.folderService.GetFolderByID(userID, folder.ID)
	if err != nil {
		return nil, err
	}

	return generated.CreateFolder201JSONResponse(folderModelToGenerated(created)), nil
}

// GetFolder implements generated.StrictServerInterface
func (h *StrictHandlers) GetFolder(
	ctx context.Context,
	request generated.GetFolderRequestObject,
) (generated.GetFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	folder, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if folder == nil {
		return generated.GetFolder404JSONResponse{NotFoundJSONResponse: notFound("Folder not found")}, nil
	}

	return generated.GetFolder200JSONResponse(folderModelToGenerated(folder)), nil
}

// UpdateFolder implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateFolder(
	ctx context.Context,
	request generated.UpdateFolderRequestObject,
) (generated.UpdateFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.UpdateFolder400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Get existing folder
	existing, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return generated.UpdateFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Update fields
	if request.Body.Name != nil {
		existing.Name = *request.Body.Name
	}
	if request.Body.Description != nil {
		existing.Description = *request.Body.Description
	}

	if err := h.folderService.UpdateFolder(userID, existing); err != nil {
		return generated.UpdateFolder400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated folder
	updated, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.UpdateFolder200JSONResponse(folderModelToGenerated(updated)), nil
}

// DeleteFolder implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteFolder(
	ctx context.Context,
	request generated.DeleteFolderRequestObject,
) (generated.DeleteFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if err := h.folderService.DeleteFolder(userID, uint(request.Id)); err != nil {
		return generated.DeleteFolder404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.DeleteFolder204Response{}, nil
}

// MoveFolder implements generated.StrictServerInterface
func (h *StrictHandlers) MoveFolder(
	ctx context.Context,
	request generated.MoveFolderRequestObject,
) (generated.MoveFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.MoveFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.MoveFolder400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	var newParentID *uint
	if request.Body.ParentId != nil {
		pid := uint(*request.Body.ParentId)
		newParentID = &pid
	}

	if err := h.folderService.MoveFolder(userID, uint(request.Id), newParentID); err != nil {
		return generated.MoveFolder400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated folder
	updated, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.MoveFolder200JSONResponse(folderModelToGenerated(updated)), nil
}

// GetFolderTree implements generated.StrictServerInterface
func (h *StrictHandlers) GetFolderTree(
	ctx context.Context,
	request generated.GetFolderTreeRequestObject,
) (generated.GetFolderTreeResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFolderTree401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	var parentID *uint
	if request.Params.ParentId != nil {
		pid := uint(*request.Params.ParentId)
		parentID = &pid
	}

	folders, err := h.folderService.GetFolderTree(userID, parentID)
	if err != nil {
		return nil, err
	}

	return generated.GetFolderTree200JSONResponse(folderListToTreeGenerated(folders)), nil
}

// AddTagsToFolder implements generated.StrictServerInterface
func (h *StrictHandlers) AddTagsToFolder(
	ctx context.Context,
	request generated.AddTagsToFolderRequestObject,
) (generated.AddTagsToFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.AddTagsToFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.AddTagsToFolder400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Convert tag IDs
	tagIDs := make([]uint, len(request.Body.TagIds))
	for i, id := range request.Body.TagIds {
		tagIDs[i] = uint(id)
	}

	if err := h.folderService.AddTagsToFolder(userID, uint(request.Id), tagIDs); err != nil {
		return generated.AddTagsToFolder400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated folder
	updated, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.AddTagsToFolder200JSONResponse(folderModelToGenerated(updated)), nil
}

// RemoveTagsFromFolder implements generated.StrictServerInterface
func (h *StrictHandlers) RemoveTagsFromFolder(
	ctx context.Context,
	request generated.RemoveTagsFromFolderRequestObject,
) (generated.RemoveTagsFromFolderResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.RemoveTagsFromFolder401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.RemoveTagsFromFolder400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Convert tag IDs
	tagIDs := make([]uint, len(request.Body.TagIds))
	for i, id := range request.Body.TagIds {
		tagIDs[i] = uint(id)
	}

	if err := h.folderService.RemoveTagsFromFolder(userID, uint(request.Id), tagIDs); err != nil {
		return generated.RemoveTagsFromFolder400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated folder
	updated, err := h.folderService.GetFolderByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.RemoveTagsFromFolder200JSONResponse(folderModelToGenerated(updated)), nil
}
