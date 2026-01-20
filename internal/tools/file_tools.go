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

// CreateFileTool handles creating a new file record
type CreateFileTool struct {
	service services.FileService
}

func NewCreateFileTool(service services.FileService) *CreateFileTool {
	return &CreateFileTool{service: service}
}

func (t *CreateFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_file",
		mcp.WithDescription("Create a new file record (after uploading to S3)"),
		mcp.WithString("title", mcp.Required(), mcp.Description("File title")),
		mcp.WithString("s3_key", mcp.Required(), mcp.Description("S3 key from upload")),
		mcp.WithString("original_filename", mcp.Required(), mcp.Description("Original filename")),
		mcp.WithString("file_type", mcp.Description("File type: music, photo, video, document, invoice")),
		mcp.WithNumber("folder_id", mcp.Description("Folder ID to place the file in")),
		mcp.WithString("mime_type", mcp.Description("MIME type of the file")),
		mcp.WithNumber("size", mcp.Description("File size in bytes")),
	)
}

func (t *CreateFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		title, _ := args["title"].(string)
		s3Key, _ := args["s3_key"].(string)
		originalFilename, _ := args["original_filename"].(string)

		if title == "" || s3Key == "" || originalFilename == "" {
			return mcp.NewToolResultError("title, s3_key, and original_filename are required"), nil
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

		// Fetch created file with relations
		created, _ := t.service.GetFileByID(userID, file.ID)
		result, _ := json.Marshal(fileToMap(created))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// ListFilesTool handles listing files
type ListFilesTool struct {
	service services.FileService
}

func NewListFilesTool(service services.FileService) *ListFilesTool {
	return &ListFilesTool{service: service}
}

func (t *ListFilesTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_files",
		mcp.WithDescription("List files with optional filtering"),
		mcp.WithString("keyword", mcp.Description("Search keyword")),
		mcp.WithNumber("folder_id", mcp.Description("Filter by folder ID")),
		mcp.WithString("file_type", mcp.Description("Filter by file type: music, photo, video, document, invoice")),
		mcp.WithString("tag_ids", mcp.Description("Comma-separated tag IDs to filter by")),
		mcp.WithString("status", mcp.Description("Filter by processing status: pending, processing, completed, failed")),
		mcp.WithString("sort_by", mcp.Description("Sort by: created_at, title, size, updated_at")),
		mcp.WithString("sort_order", mcp.Description("Sort order: asc, desc")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of files to return (default: 100)")),
		mcp.WithNumber("offset", mcp.Description("Number of files to skip for pagination")),
	)
}

func (t *ListFilesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)

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
}

// GetFileTool handles getting a file by ID
type GetFileTool struct {
	service services.FileService
}

func NewGetFileTool(service services.FileService) *GetFileTool {
	return &GetFileTool{service: service}
}

func (t *GetFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_file",
		mcp.WithDescription("Get a file by its ID"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
	)
}

func (t *GetFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
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
}

// UpdateFileTool handles updating a file
type UpdateFileTool struct {
	service services.FileService
}

func NewUpdateFileTool(service services.FileService) *UpdateFileTool {
	return &UpdateFileTool{service: service}
}

func (t *UpdateFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_file",
		mcp.WithDescription("Update an existing file"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
		mcp.WithString("title", mcp.Description("New file title")),
		mcp.WithString("summary", mcp.Description("New file summary")),
		mcp.WithString("file_type", mcp.Description("New file type")),
		mcp.WithNumber("folder_id", mcp.Description("New folder ID")),
	)
}

func (t *UpdateFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
		}

		// Get existing file
		existing, err := t.service.GetFileByID(userID, fileID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get file: %v", err)), nil
		}
		if existing == nil {
			return mcp.NewToolResultError("File not found"), nil
		}

		// Update fields
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

		// Fetch updated file
		updated, _ := t.service.GetFileByID(userID, fileID)
		result, _ := json.Marshal(fileToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// DeleteFileTool handles deleting a file
type DeleteFileTool struct {
	service services.FileService
}

func NewDeleteFileTool(service services.FileService) *DeleteFileTool {
	return &DeleteFileTool{service: service}
}

func (t *DeleteFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_file",
		mcp.WithDescription("Delete a file"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
	)
}

func (t *DeleteFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
		}

		if err := t.service.DeleteFile(userID, fileID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete file: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]any{
			"message": "File deleted successfully",
			"file_id": fileID,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// MoveFilesTool handles moving files to a different folder
type MoveFilesTool struct {
	service services.FileService
}

func NewMoveFilesTool(service services.FileService) *MoveFilesTool {
	return &MoveFilesTool{service: service}
}

func (t *MoveFilesTool) GetTool() mcp.Tool {
	return mcp.NewTool("move_files",
		mcp.WithDescription("Move multiple files to a different folder"),
		mcp.WithString("file_ids", mcp.Required(), mcp.Description("Comma-separated file IDs to move")),
		mcp.WithNumber("folder_id", mcp.Description("Target folder ID (omit to move to root)")),
	)
}

func (t *MoveFilesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileIDsStr := getStringArg(args, "file_ids")
		fileIDs := parseUintSlice(fileIDsStr)
		if len(fileIDs) == 0 {
			return mcp.NewToolResultError("file_ids is required"), nil
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
}

// AddTagsToFileTool handles adding tags to a file
type AddTagsToFileTool struct {
	service services.FileService
}

func NewAddTagsToFileTool(service services.FileService) *AddTagsToFileTool {
	return &AddTagsToFileTool{service: service}
}

func (t *AddTagsToFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("add_tags_to_file",
		mcp.WithDescription("Add tags to a file"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
		mcp.WithString("tag_ids", mcp.Required(), mcp.Description("Comma-separated tag IDs to add")),
	)
}

func (t *AddTagsToFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
		}

		tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
		if len(tagIDs) == 0 {
			return mcp.NewToolResultError("tag_ids is required"), nil
		}

		if err := t.service.AddTagsToFile(userID, fileID, tagIDs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add tags: %v", err)), nil
		}

		// Fetch updated file
		updated, _ := t.service.GetFileByID(userID, fileID)
		result, _ := json.Marshal(fileToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// RemoveTagsFromFileTool handles removing tags from a file
type RemoveTagsFromFileTool struct {
	service services.FileService
}

func NewRemoveTagsFromFileTool(service services.FileService) *RemoveTagsFromFileTool {
	return &RemoveTagsFromFileTool{service: service}
}

func (t *RemoveTagsFromFileTool) GetTool() mcp.Tool {
	return mcp.NewTool("remove_tags_from_file",
		mcp.WithDescription("Remove tags from a file"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
		mcp.WithString("tag_ids", mcp.Required(), mcp.Description("Comma-separated tag IDs to remove")),
	)
}

func (t *RemoveTagsFromFileTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
		}

		tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
		if len(tagIDs) == 0 {
			return mcp.NewToolResultError("tag_ids is required"), nil
		}

		if err := t.service.RemoveTagsFromFile(userID, fileID, tagIDs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove tags: %v", err)), nil
		}

		// Fetch updated file
		updated, _ := t.service.GetFileByID(userID, fileID)
		result, _ := json.Marshal(fileToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// GetFileDownloadURLTool handles getting a presigned download URL for a file
type GetFileDownloadURLTool struct {
	fileService   services.FileService
	uploadService services.UploadService
}

func NewGetFileDownloadURLTool(fileService services.FileService, uploadService services.UploadService) *GetFileDownloadURLTool {
	return &GetFileDownloadURLTool{
		fileService:   fileService,
		uploadService: uploadService,
	}
}

func (t *GetFileDownloadURLTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_file_download_url",
		mcp.WithDescription("Get a presigned URL for downloading a file"),
		mcp.WithNumber("file_id", mcp.Required(), mcp.Description("File ID")),
	)
}

func (t *GetFileDownloadURLTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		fileID := getUintArg(args, "file_id")
		if fileID == 0 {
			return mcp.NewToolResultError("file_id is required"), nil
		}

		file, err := t.fileService.GetFileByID(userID, fileID)
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
