package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// FileTool handles all file operations via a single tool with an action parameter
type FileTool struct {
	service          services.FileService
	uploadService    services.UploadService
	embeddingService services.EmbeddingService
	invoiceService   services.InvoiceService
}

func NewFileTool(service services.FileService, uploadService services.UploadService, embeddingService services.EmbeddingService, invoiceService services.InvoiceService) *FileTool {
	return &FileTool{
		service:          service,
		uploadService:    uploadService,
		embeddingService: embeddingService,
		invoiceService:   invoiceService,
	}
}

func (t *FileTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_files",
		mcp.WithDescription("Manage files. Use the 'action' parameter to specify the operation."),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: create, list, get, update, delete, move, add_tags, remove_tags, get_download_url")),
		mcp.WithNumber("file_id", mcp.Description("File ID (required for get, update, delete, add_tags, remove_tags, get_download_url)")),
		mcp.WithString("file_ids", mcp.Description("Comma-separated file IDs (required for move)")),
		mcp.WithString("title", mcp.Description("File title (required for create, optional for update)")),
		mcp.WithString("s3_key", mcp.Description("S3 key from upload (required for create)")),
		mcp.WithString("original_filename", mcp.Description("Original filename (required for create)")),
		mcp.WithString("file_type", mcp.Description("File type: music, photo, video, document, invoice (for create, update, list)")),
		mcp.WithNumber("folder_id", mcp.Description("Folder ID (for create, update, list, move)")),
		mcp.WithString("mime_type", mcp.Description("MIME type of the file (for create)")),
		mcp.WithNumber("size", mcp.Description("File size in bytes (for create)")),
		mcp.WithString("summary", mcp.Description("File summary (for update)")),
		mcp.WithString("tag_ids", mcp.Description("Comma-separated tag IDs (for add_tags, remove_tags, list)")),
		mcp.WithString("keyword", mcp.Description("Search keyword (for list)")),
		mcp.WithString("status", mcp.Description("Filter by processing status: pending, processing, completed, failed (for list)")),
		mcp.WithString("sort_by", mcp.Description("Sort by: created_at, title, size, updated_at (for list)")),
		mcp.WithString("sort_order", mcp.Description("Sort order: asc, desc (for list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of files to return, default 100 (for list)")),
		mcp.WithNumber("offset", mcp.Description("Number of files to skip for pagination (for list)")),
	)
}

func (t *FileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		action := getStringArg(args, "action")

		switch action {
		case "create":
			return t.handleCreate(userID, args)
		case "list":
			return t.handleList(userID, args)
		case "get":
			return t.handleGet(userID, args)
		case "update":
			return t.handleUpdate(userID, args)
		case "delete":
			return t.handleDelete(ctx, userID, args)
		case "move":
			return t.handleMove(userID, args)
		case "add_tags":
			return t.handleAddTags(userID, args)
		case "remove_tags":
			return t.handleRemoveTags(userID, args)
		case "get_download_url":
			return t.handleGetDownloadURL(ctx, userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Invalid action: %s. Use create, list, get, update, delete, move, add_tags, remove_tags, or get_download_url", action)), nil
		}
	}
}

func (t *FileTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	title, _ := args["title"].(string)
	s3Key, _ := args["s3_key"].(string)
	originalFilename, _ := args["original_filename"].(string)

	if title == "" || s3Key == "" || originalFilename == "" {
		return mcp.NewToolResultError("title, s3_key, and original_filename are required for create action"), nil
	}

	file := &models.File{
		Title:            title,
		S3Key:            s3Key,
		OriginalFilename: originalFilename,
		MimeType:         getStringArg(args, "mime_type"),
		ProcessingStatus: models.FileStatusPending,
	}

	if size := getIntArg(args, "size", 0); size > 0 {
		file.Size = int64(size)
	}

	if folderID := getUintArg(args, "folder_id"); folderID > 0 {
		file.FolderID = &folderID
	}

	if fileType := getStringArg(args, "file_type"); fileType != "" {
		file.FileType = models.FileType(fileType)
	} else {
		file.FileType = models.DetectFileTypeFromMimeType(file.MimeType)
	}

	if err := t.service.CreateFile(userID, file); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create file: %v", err)), nil
	}

	created, _ := t.service.GetFileByID(userID, file.ID)
	result, _ := json.Marshal(fileToMap(created))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	opts := services.FileListOptions{
		Keyword:   getStringArg(args, "keyword"),
		Limit:     getIntArg(args, "limit", 100),
		Offset:    getIntArg(args, "offset", 0),
		TagIDs:    parseTagIDs(getStringArg(args, "tag_ids")),
		SortBy:    getStringArg(args, "sort_by"),
		SortOrder: getStringArg(args, "sort_order"),
	}

	if folderID := getUintArg(args, "folder_id"); folderID > 0 {
		opts.FolderID = &folderID
	}

	if fileType := getStringArg(args, "file_type"); fileType != "" {
		opts.FileTypes = []models.FileType{models.FileType(fileType)}
	}

	if status := getStringArg(args, "status"); status != "" {
		s := models.FileProcessingStatus(status)
		opts.Status = &s
	}

	files, total, err := t.service.ListFiles(userID, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list files: %v", err)), nil
	}

	fileList := make([]map[string]any, len(files))
	for i, file := range files {
		fileList[i] = fileToMap(&file)
	}

	result, _ := json.Marshal(map[string]any{
		"data":   fileList,
		"total":  total,
		"limit":  opts.Limit,
		"offset": opts.Offset,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for get action"), nil
	}

	file, err := t.service.GetFileByID(userID, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}
	if file == nil {
		return mcp.NewToolResultError("File not found"), nil
	}

	result, _ := json.Marshal(fileToMap(file))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for update action"), nil
	}

	existing, err := t.service.GetFileByID(userID, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}
	if existing == nil {
		return mcp.NewToolResultError("File not found"), nil
	}

	if title, ok := args["title"].(string); ok && title != "" {
		existing.Title = title
	}
	if summary, ok := args["summary"].(string); ok {
		existing.Summary = summary
	}
	if fileType, ok := args["file_type"].(string); ok && fileType != "" {
		existing.FileType = models.FileType(fileType)
	}
	if folderID := getUintArg(args, "folder_id"); folderID > 0 {
		existing.FolderID = &folderID
	}

	if err := t.service.UpdateFile(userID, existing); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update file: %v", err)), nil
	}

	updated, _ := t.service.GetFileByID(userID, fileID)
	result, _ := json.Marshal(fileToMap(updated))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleDelete(ctx context.Context, userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for delete action"), nil
	}

	file, err := t.service.GetFileByID(userID, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}
	if file == nil {
		return mcp.NewToolResultError("File not found"), nil
	}

	if err := t.service.DeleteFile(userID, fileID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete file: %v", err)), nil
	}

	if file.S3Key != "" && t.uploadService != nil {
		_ = t.uploadService.DeleteFile(ctx, file.S3Key)
	}

	if file.HasEmbedding && t.embeddingService != nil {
		_ = t.embeddingService.DeleteFileEmbedding(userID, file.ID)
	}

	if file.InvoiceID != nil && t.invoiceService != nil && t.invoiceService.IsEnabled() {
		authToken, _ := utils.GetRawAuthToken(ctx)
		if authToken != "" {
			invoiceID := *file.InvoiceID
			go func() {
				deleteCtx := context.Background()
				_ = t.invoiceService.DeleteInvoice(deleteCtx, invoiceID, authToken)
			}()
		}
	}

	result, _ := json.Marshal(map[string]any{
		"message": "File deleted successfully",
		"file_id": fileID,
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleMove(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileIDsStr := getStringArg(args, "file_ids")
	fileIDs := parseUintSlice(fileIDsStr)
	if len(fileIDs) == 0 {
		return mcp.NewToolResultError("file_ids is required for move action"), nil
	}

	var targetFolderID *uint
	if folderID := getUintArg(args, "folder_id"); folderID > 0 {
		targetFolderID = &folderID
	}

	if err := t.service.MoveFiles(userID, fileIDs, targetFolderID); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move files: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]any{
		"message":     "Files moved successfully",
		"moved_count": len(fileIDs),
	})
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleAddTags(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for add_tags action"), nil
	}

	tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
	if len(tagIDs) == 0 {
		return mcp.NewToolResultError("tag_ids is required for add_tags action"), nil
	}

	if err := t.service.AddTagsToFile(userID, fileID, tagIDs); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add tags: %v", err)), nil
	}

	updated, _ := t.service.GetFileByID(userID, fileID)
	result, _ := json.Marshal(fileToMap(updated))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleRemoveTags(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for remove_tags action"), nil
	}

	tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
	if len(tagIDs) == 0 {
		return mcp.NewToolResultError("tag_ids is required for remove_tags action"), nil
	}

	if err := t.service.RemoveTagsFromFile(userID, fileID, tagIDs); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to remove tags: %v", err)), nil
	}

	updated, _ := t.service.GetFileByID(userID, fileID)
	result, _ := json.Marshal(fileToMap(updated))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *FileTool) handleGetDownloadURL(ctx context.Context, userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	fileID := getUintArg(args, "file_id")
	if fileID == 0 {
		return mcp.NewToolResultError("file_id is required for get_download_url action"), nil
	}

	file, err := t.service.GetFileByID(userID, fileID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
	}
	if file == nil {
		return mcp.NewToolResultError("File not found"), nil
	}

	downloadURL, err := t.uploadService.GetPresignedDownloadURL(ctx, file.S3Key)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get download URL: %v", err)), nil
	}

	result, _ := json.Marshal(map[string]any{
		"download_url": downloadURL,
		"filename":     file.OriginalFilename,
		"key":          file.S3Key,
	})
	return mcp.NewToolResultText(string(result)), nil
}

// Helper functions
func fileToMap(file *models.File) map[string]any {
	m := map[string]any{
		"id":                file.ID,
		"title":             file.Title,
		"summary":           file.Summary,
		"file_type":         file.FileType,
		"s3_key":            file.S3Key,
		"original_filename": file.OriginalFilename,
		"mime_type":         file.MimeType,
		"size":              file.Size,
		"processing_status": file.ProcessingStatus,
		"processing_error":  file.ProcessingError,
		"has_embedding":     file.HasEmbedding,
		"folder_id":         file.FolderID,
		"created_at":        file.CreatedAt,
		"updated_at":        file.UpdatedAt,
	}

	if file.Folder != nil {
		m["folder"] = folderToMap(file.Folder)
	}

	if len(file.Tags) > 0 {
		tags := make([]map[string]any, len(file.Tags))
		for i, tag := range file.Tags {
			tags[i] = tagToMap(&tag)
		}
		m["tags"] = tags
	}

	return m
}

func parseUintSlice(s string) []uint {
	if s == "" {
		return nil
	}

	var result []uint
	for _, part := range splitAndTrim(s, ",") {
		if id := parseUint(part); id > 0 {
			result = append(result, id)
		}
	}
	return result
}
