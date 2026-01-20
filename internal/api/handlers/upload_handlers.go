package handlers

import (
	"context"
	"io"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
)

// UploadFile implements generated.StrictServerInterface
func (h *StrictHandlers) UploadFile(
	ctx context.Context,
	request generated.UploadFileRequestObject,
) (generated.UploadFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UploadFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Read file from multipart request
	file, err := request.Body.NextPart()
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("No file provided")}, nil
	}
	defer file.Close()

	filename := file.FileName()
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest("Failed to read file")}, nil
	}

	// Upload to S3 - returns the key
	key, err := h.uploadService.UploadFile(ctx, userID, filename, content, contentType)
	if err != nil {
		return generated.UploadFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Return the S3 key - file record is created separately via POST /api/files
	return generated.UploadFile201JSONResponse{
		Key:         key,
		Filename:    filename,
		Size:        len(content),
		ContentType: contentType,
	}, nil
}

// GetPresignedURL implements generated.StrictServerInterface
func (h *StrictHandlers) GetPresignedURL(
	ctx context.Context,
	request generated.GetPresignedURLRequestObject,
) (generated.GetPresignedURLResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetPresignedURL401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	filename := request.Params.Filename
	if filename == "" {
		return generated.GetPresignedURL400JSONResponse{BadRequestJSONResponse: badRequest("Filename is required")}, nil
	}

	contentType := "application/octet-stream"
	if request.Params.ContentType != nil {
		contentType = *request.Params.ContentType
	}

	uploadURL, key, err := h.uploadService.GetPresignedUploadURL(ctx, userID, filename, contentType)
	if err != nil {
		return generated.GetPresignedURL400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.GetPresignedURL200JSONResponse{
		UploadUrl:   uploadURL,
		Key:         key,
		ContentType: contentType,
	}, nil
}
