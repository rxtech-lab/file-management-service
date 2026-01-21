package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// ListFiles implements generated.StrictServerInterface
func (h *StrictHandlers) ListFiles(
	ctx context.Context,
	request generated.ListFilesRequestObject,
) (generated.ListFilesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ListFiles401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	opts := services.FileListOptions{
		Keyword: deref(request.Params.Keyword),
		Limit:   derefInt(request.Params.Limit, 100),
		Offset:  derefInt(request.Params.Offset, 0),
	}

	// Handle folder_id
	if request.Params.FolderId != nil {
		folderID := uint(*request.Params.FolderId)
		opts.FolderID = &folderID
	}

	// Handle all_folders
	if request.Params.AllFolders != nil && *request.Params.AllFolders {
		opts.AllFolders = true
	}

	// Handle file_type
	if request.Params.FileType != nil {
		ft := models.FileType(*request.Params.FileType)
		opts.FileTypes = []models.FileType{ft}
	}

	// Handle status
	if request.Params.Status != nil {
		status := models.FileProcessingStatus(*request.Params.Status)
		opts.Status = &status
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

	// Handle sorting
	if request.Params.SortBy != nil {
		opts.SortBy = string(*request.Params.SortBy)
	}
	if request.Params.SortOrder != nil {
		opts.SortOrder = string(*request.Params.SortOrder)
	}

	files, total, err := h.fileService.ListFiles(userID, opts)
	if err != nil {
		return nil, err
	}

	return generated.ListFiles200JSONResponse{
		Data:   fileListToGenerated(files),
		Total:  int(total),
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}, nil
}

// CreateFile implements generated.StrictServerInterface
func (h *StrictHandlers) CreateFile(
	ctx context.Context,
	request generated.CreateFileRequestObject,
) (generated.CreateFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.CreateFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.CreateFile400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	file := &models.File{
		Title:            request.Body.Title,
		S3Key:            request.Body.S3Key,
		OriginalFilename: request.Body.OriginalFilename,
		MimeType:         deref(request.Body.MimeType),
		Size:             deref(request.Body.Size),
		ProcessingStatus: models.FileStatusPending,
	}

	// Set folder if provided
	if request.Body.FolderId != nil {
		folderID := uint(*request.Body.FolderId)
		file.FolderID = &folderID
	}

	// Set file type - detect from mime type if not provided
	if request.Body.FileType != nil {
		file.FileType = models.FileType(*request.Body.FileType)
	} else {
		file.FileType = models.DetectFileTypeFromMimeType(file.MimeType)
	}

	if err := h.fileService.CreateFile(userID, file); err != nil {
		return generated.CreateFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch file with relations
	created, err := h.fileService.GetFileByID(userID, file.ID)
	if err != nil {
		return nil, err
	}

	return generated.CreateFile201JSONResponse(fileModelToGenerated(created)), nil
}

// GetFile implements generated.StrictServerInterface
func (h *StrictHandlers) GetFile(
	ctx context.Context,
	request generated.GetFileRequestObject,
) (generated.GetFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.GetFile404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	return generated.GetFile200JSONResponse(fileModelToGenerated(file)), nil
}

// UpdateFile implements generated.StrictServerInterface
func (h *StrictHandlers) UpdateFile(
	ctx context.Context,
	request generated.UpdateFileRequestObject,
) (generated.UpdateFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UpdateFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.UpdateFile400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Get existing file
	existing, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return generated.UpdateFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Update fields
	if request.Body.Title != nil {
		existing.Title = *request.Body.Title
	}
	if request.Body.Summary != nil {
		existing.Summary = *request.Body.Summary
	}
	if request.Body.FileType != nil {
		existing.FileType = models.FileType(*request.Body.FileType)
	}
	// Handle folder_id - even if nil (moving to root)
	if request.Body.FolderId != nil {
		folderID := uint(*request.Body.FolderId)
		existing.FolderID = &folderID
	}

	if err := h.fileService.UpdateFile(userID, existing); err != nil {
		return generated.UpdateFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated file
	updated, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.UpdateFile200JSONResponse(fileModelToGenerated(updated)), nil
}

// DeleteFile implements generated.StrictServerInterface
func (h *StrictHandlers) DeleteFile(
	ctx context.Context,
	request generated.DeleteFileRequestObject,
) (generated.DeleteFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.DeleteFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get file to delete S3 object
	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.DeleteFile404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Delete from database first
	if err := h.fileService.DeleteFile(userID, uint(request.Id)); err != nil {
		return generated.DeleteFile404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	// Delete from S3 (best effort - don't fail if S3 delete fails)
	_ = h.uploadService.DeleteFile(ctx, file.S3Key)

	// Delete embedding if exists
	if file.HasEmbedding {
		_ = h.embeddingService.DeleteFileEmbedding(userID, file.ID)
	}

	return generated.DeleteFile204Response{}, nil
}

// MoveFiles implements generated.StrictServerInterface
func (h *StrictHandlers) MoveFiles(
	ctx context.Context,
	request generated.MoveFilesRequestObject,
) (generated.MoveFilesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.MoveFiles401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.MoveFiles400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Convert file IDs
	fileIDs := make([]uint, len(request.Body.FileIds))
	for i, id := range request.Body.FileIds {
		fileIDs[i] = uint(id)
	}

	// Convert target folder ID
	var targetFolderID *uint
	if request.Body.FolderId != nil {
		fid := uint(*request.Body.FolderId)
		targetFolderID = &fid
	}

	if err := h.fileService.MoveFiles(userID, fileIDs, targetFolderID); err != nil {
		return generated.MoveFiles400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	return generated.MoveFiles200JSONResponse{
		Message:    "Files moved successfully",
		MovedCount: len(fileIDs),
	}, nil
}

// ProcessFile implements generated.StrictServerInterface
func (h *StrictHandlers) ProcessFile(
	ctx context.Context,
	request generated.ProcessFileRequestObject,
) (generated.ProcessFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.ProcessFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get file
	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.ProcessFile400JSONResponse{BadRequestJSONResponse: badRequest("File not found")}, nil
	}

	// Check if already processing
	if file.ProcessingStatus == models.FileStatusProcessing {
		return generated.ProcessFile400JSONResponse{BadRequestJSONResponse: badRequest("File is already being processed")}, nil
	}

	// Update status to processing
	if err := h.fileService.UpdateFileProcessingStatus(userID, file.ID, models.FileStatusProcessing, ""); err != nil {
		return nil, err
	}

	// Extract auth token for invoice processing
	authToken, _ := utils.GetRawAuthToken(ctx)

	// Start async processing
	go h.processFileAsync(userID, file.ID, authToken)

	return generated.ProcessFile202JSONResponse{
		Message: "File processing started",
		Status:  generated.Processing,
	}, nil
}

// processFileAsync handles file processing in a background goroutine
func (h *StrictHandlers) processFileAsync(userID string, fileID uint, authToken string) {
	ctx := context.Background()

	// Get file
	file, err := h.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to get file")
		return
	}

	// Get presigned download URL for the file
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, file.S3Key)
	if err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to get download URL: "+err.Error())
		return
	}

	// Parse content using content parser service
	parsedContent, err := h.contentParserService.ParseFileContent(ctx, downloadURL)
	if err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to parse content: "+err.Error())
		return
	}

	// Generate summary from the text content using AI
	summary, err := h.summaryService.GenerateSummary(ctx, parsedContent.TextContent, 500)
	if err != nil {
		// Fall back to simple truncation if AI summary fails
		summary = services.GenerateSummary(parsedContent.TextContent, 500)
	}

	// Detect file type from content (especially for invoice detection)
	detectedFileType := file.FileType
	if models.IsInvoiceContent(parsedContent.TextContent) {
		detectedFileType = models.FileTypeInvoice
	}

	// Update file with parsed content
	if err := h.fileService.UpdateFileContent(userID, fileID, parsedContent.TextContent, summary, detectedFileType); err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to update content: "+err.Error())
		return
	}

	// Process invoice via external API if file is detected as invoice
	if detectedFileType == models.FileTypeInvoice && h.invoiceService != nil && h.invoiceService.IsEnabled() && authToken != "" {
		log.Printf("[Invoice] Processing file %d as invoice", fileID)

		// Create event channel for logging
		invoiceEventChan := make(chan services.InvoiceStreamEvent, 100)
		go func() {
			for event := range invoiceEventChan {
				log.Printf("[Invoice] File %d: %s - %s", fileID, event.Status, event.Message)
			}
		}()

		result, err := h.invoiceService.ProcessInvoice(ctx, downloadURL, authToken, invoiceEventChan)
		if err != nil {
			log.Printf("[Invoice] File %d processing warning: %v", fileID, err)
			// Don't fail file processing - invoice processing is best-effort
		} else if result != nil {
			if err := h.fileService.UpdateFileInvoiceID(userID, fileID, result.InvoiceID); err != nil {
				log.Printf("[Invoice] File %d: failed to store invoice_id: %v", fileID, err)
			} else {
				log.Printf("[Invoice] File %d: stored invoice_id=%d", fileID, result.InvoiceID)
			}
		}
	}

	// Run AI agent to organize file (best-effort, non-blocking errors)
	if h.agentService != nil && h.agentService.IsEnabled() {
		// Create event channel for logging
		eventChan := make(chan services.AgentEvent, 100)
		go func() {
			for event := range eventChan {
				// Log agent events for debugging
				switch event.Type {
				case "error":
					log.Printf("[Agent] File %d error: %s", fileID, event.Message)
				case "tool_call":
					log.Printf("[Agent] File %d: %s", fileID, event.Message)
				case "result":
					log.Printf("[Agent] File %d completed: %s", fileID, event.Message)
				}
			}
		}()

		if err := h.agentService.ProcessFileWithAgent(ctx, userID, fileID, parsedContent.TextContent, summary, eventChan); err != nil {
			log.Printf("[Agent] File %d processing warning: %v", fileID, err)
			// Don't fail the file processing, agent is best-effort
		}
	}

	// Generate embedding
	embedding, err := h.embeddingService.GenerateEmbedding(ctx, parsedContent.TextContent)
	if err != nil {
		// Content parsed successfully but embedding failed - still mark as completed
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "Embedding generation failed: "+err.Error())
		return
	}

	// Store embedding
	if err := h.embeddingService.StoreFileEmbedding(userID, fileID, embedding); err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "Embedding storage failed: "+err.Error())
		return
	}

	// Mark as completed
	h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "")
}

// GetFileDownloadURL implements generated.StrictServerInterface
func (h *StrictHandlers) GetFileDownloadURL(
	ctx context.Context,
	request generated.GetFileDownloadURLRequestObject,
) (generated.GetFileDownloadURLResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.GetFileDownloadURL401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Get file to verify ownership and get filename
	file, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}
	if file == nil {
		return generated.GetFileDownloadURL404JSONResponse{NotFoundJSONResponse: notFound("File not found")}, nil
	}

	// Get presigned download URL
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, file.S3Key)
	if err != nil {
		return nil, err
	}

	// Calculate expiration time (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	return generated.GetFileDownloadURL200JSONResponse{
		DownloadUrl: downloadURL,
		Key:         file.S3Key,
		Filename:    file.OriginalFilename,
		ExpiresAt:   expiresAt,
	}, nil
}

// AddTagsToFile implements generated.StrictServerInterface
func (h *StrictHandlers) AddTagsToFile(
	ctx context.Context,
	request generated.AddTagsToFileRequestObject,
) (generated.AddTagsToFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.AddTagsToFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.AddTagsToFile400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Convert tag IDs
	tagIDs := make([]uint, len(request.Body.TagIds))
	for i, id := range request.Body.TagIds {
		tagIDs[i] = uint(id)
	}

	if err := h.fileService.AddTagsToFile(userID, uint(request.Id), tagIDs); err != nil {
		return generated.AddTagsToFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated file
	updated, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.AddTagsToFile200JSONResponse(fileModelToGenerated(updated)), nil
}

// RemoveTagsFromFile implements generated.StrictServerInterface
func (h *StrictHandlers) RemoveTagsFromFile(
	ctx context.Context,
	request generated.RemoveTagsFromFileRequestObject,
) (generated.RemoveTagsFromFileResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.RemoveTagsFromFile401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil {
		return generated.RemoveTagsFromFile400JSONResponse{BadRequestJSONResponse: badRequest("Request body is required")}, nil
	}

	// Convert tag IDs
	tagIDs := make([]uint, len(request.Body.TagIds))
	for i, id := range request.Body.TagIds {
		tagIDs[i] = uint(id)
	}

	if err := h.fileService.RemoveTagsFromFile(userID, uint(request.Id), tagIDs); err != nil {
		return generated.RemoveTagsFromFile400JSONResponse{BadRequestJSONResponse: badRequest(err.Error())}, nil
	}

	// Fetch updated file
	updated, err := h.fileService.GetFileByID(userID, uint(request.Id))
	if err != nil {
		return nil, err
	}

	return generated.RemoveTagsFromFile200JSONResponse(fileModelToGenerated(updated)), nil
}

// UnlinkFileInvoice implements generated.StrictServerInterface
func (h *StrictHandlers) UnlinkFileInvoice(
	ctx context.Context,
	request generated.UnlinkFileInvoiceRequestObject,
) (generated.UnlinkFileInvoiceResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.UnlinkFileInvoice401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	// Validate invoice_id parameter
	if request.Params.InvoiceId <= 0 {
		return generated.UnlinkFileInvoice400JSONResponse{BadRequestJSONResponse: badRequest("invoice_id must be a positive integer")}, nil
	}

	// Unlink the invoice from the file (verifies user ownership)
	if err := h.fileService.UnlinkFileInvoiceByInvoiceID(userID, request.Params.InvoiceId); err != nil {
		return generated.UnlinkFileInvoice404JSONResponse{NotFoundJSONResponse: notFound(err.Error())}, nil
	}

	return generated.UnlinkFileInvoice204Response{}, nil
}

// BatchDownloadFiles implements generated.StrictServerInterface
// This handler streams a ZIP file containing multiple files
func (h *StrictHandlers) BatchDownloadFiles(
	ctx context.Context,
	request generated.BatchDownloadFilesRequestObject,
) (generated.BatchDownloadFilesResponseObject, error) {
	userID, err := getUserID(ctx)
	if err != nil {
		return generated.BatchDownloadFiles401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}

	if request.Body == nil || len(request.Body.FileIds) == 0 {
		return generated.BatchDownloadFiles400JSONResponse{BadRequestJSONResponse: badRequest("file_ids is required")}, nil
	}

	// Fetch all files
	var files []*models.File
	for _, id := range request.Body.FileIds {
		file, err := h.fileService.GetFileByID(userID, uint(id))
		if err != nil {
			return nil, err
		}
		if file == nil {
			return generated.BatchDownloadFiles400JSONResponse{
				BadRequestJSONResponse: badRequest(fmt.Sprintf("File with ID %d not found", id)),
			}, nil
		}
		files = append(files, file)
	}

	// Create a pipe for streaming
	pr, pw := io.Pipe()

	// Create ZIP writer
	zipWriter := zip.NewWriter(pw)

	// Start goroutine to write files to ZIP
	go func() {
		defer pw.Close()
		defer zipWriter.Close()

		for _, file := range files {
			// Get presigned download URL
			downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, file.S3Key)
			if err != nil {
				continue // Skip files that can't be downloaded
			}

			// Download file content
			resp, err := http.Get(downloadURL)
			if err != nil {
				continue
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				continue
			}

			// Create file in ZIP
			w, err := zipWriter.Create(file.OriginalFilename)
			if err != nil {
				resp.Body.Close()
				continue
			}

			// Copy content to ZIP
			_, err = io.Copy(w, resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}
		}
	}()

	// Return streaming response
	return generated.BatchDownloadFiles200ApplicationzipResponse{
		Body:          pr,
		ContentLength: 0, // Unknown length for streaming
	}, nil
}
