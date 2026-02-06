package handlers

import (
	"context"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
)

// GetFileContent implements generated.StrictServerInterface
func (h *StrictHandlers) GetFileContent(
	ctx context.Context,
	request generated.GetFileContentRequestObject,
) (generated.GetFileContentResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFileContent401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get file to verify ownership and get S3 key
	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.GetFileContent404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Download content from S3
	content, contentType, err := h.uploadService.DownloadFile(ctx, file.S3Key)
	if err != nil {
		return generated.GetFileContent404JSONResponse{NotFoundJSONResponse: notFound("Failed to download file content")}, nil
	}

	return generated.GetFileContent200JSONResponse{
		Content:  string(content),
		MimeType: contentType,
		Size:     int64(len(content)),
	}, nil
}

// UpdateFileContent implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateFileContent(
	ctx context.Context,
	request generated.UpdateFileContentRequestObject,
) (generated.UpdateFileContentResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateFileContent401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.UpdateFileContent400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Get file to verify ownership and get S3 key
	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.UpdateFileContent404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Determine content type
	contentType := file.MimeType
	if request.Body.MimeType != nil && *request.Body.MimeType != "" {
		contentType = *request.Body.MimeType
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload new content to S3
	contentBytes := []byte(request.Body.Content)
	if err := h.uploadService.UploadToExistingKey(ctx, file.S3Key, contentBytes, contentType); err != nil {
		return generated.UpdateFileContent400JSONResponse{BadRequestJSONResponse: badRequest("Failed to update file content: " + err.Error())}, nil
	}

	// Update file size in database
	file.Size = int64(len(contentBytes))
	if err := h.fileService.UpdateFile(userID, file); err != nil {
		return generated.UpdateFileContent400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated file with relations
	updated, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.UpdateFileContent200JSONResponse(fileModelToGenerated(updated)), nil
}
