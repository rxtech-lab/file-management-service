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

// CreateFolderTool handles creating a new folder
type CreateFolderTool struct {
	service services.FolderService
}

func NewCreateFolderTool(service services.FolderService) *CreateFolderTool {
	return &CreateFolderTool{service: service}
}

func (t *CreateFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_folder",
		mcp.WithDescription("Create a new folder for organizing files"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Folder name")),
		mcp.WithString("description", mcp.Description("Folder description")),
		mcp.WithNumber("parent_id", mcp.Description("Parent folder ID (omit for root folder)")),
	)
}

func (t *CreateFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		name, _ := args["name"].(string)
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}

		folder := &models.Folder{
			Name:        name,
			Description: getStringArg(args, "description"),
		}

		if parentID := getUintArg(args, "parent_id"); parentID > 0 {
			folder.ParentID = &parentID
		}

		if err := t.service.CreateFolder(userID, folder); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create folder: %v", err)), nil
		}

		// Fetch created folder with tags
		created, _ := t.service.GetFolderByID(userID, folder.ID)
		result, _ := json.Marshal(folderToMap(created))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// ListFoldersTool handles listing folders
type ListFoldersTool struct {
	service services.FolderService
}

func NewListFoldersTool(service services.FolderService) *ListFoldersTool {
	return &ListFoldersTool{service: service}
}

func (t *ListFoldersTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_folders",
		mcp.WithDescription("List folders with optional filtering"),
		mcp.WithString("keyword", mcp.Description("Search keyword to filter folders")),
		mcp.WithNumber("parent_id", mcp.Description("Parent folder ID to list children (omit for root folders)")),
		mcp.WithString("tag_ids", mcp.Description("Comma-separated tag IDs to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of folders to return (default: 100)")),
		mcp.WithNumber("offset", mcp.Description("Number of folders to skip for pagination")),
	)
}

func (t *ListFoldersTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)

		opts := services.FolderListOptions{
			Keyword: getStringArg(args, "keyword"),
			Limit:   getIntArg(args, "limit", 100),
			Offset:  getIntArg(args, "offset", 0),
			TagIDs:  parseTagIDs(getStringArg(args, "tag_ids")),
		}

		if parentID := getUintArg(args, "parent_id"); parentID > 0 {
			opts.ParentID = &parentID
		}

		folders, total, err := t.service.ListFolders(userID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list folders: %v", err)), nil
		}

		folderList := make([]map[string]interface{}, len(folders))
		for i, folder := range folders {
			folderList[i] = folderToMap(&folder)
		}

		result, _ := json.Marshal(map[string]interface{}{
			"data":   folderList,
			"total":  total,
			"limit":  opts.Limit,
			"offset": opts.Offset,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// GetFolderTool handles getting a folder by ID
type GetFolderTool struct {
	service services.FolderService
}

func NewGetFolderTool(service services.FolderService) *GetFolderTool {
	return &GetFolderTool{service: service}
}

func (t *GetFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_folder",
		mcp.WithDescription("Get a folder by its ID"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID")),
	)
}

func (t *GetFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		folder, err := t.service.GetFolderByID(userID, folderID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get folder: %v", err)), nil
		}
		if folder == nil {
			return mcp.NewToolResultError("Folder not found"), nil
		}

		result, _ := json.Marshal(folderToMap(folder))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// UpdateFolderTool handles updating a folder
type UpdateFolderTool struct {
	service services.FolderService
}

func NewUpdateFolderTool(service services.FolderService) *UpdateFolderTool {
	return &UpdateFolderTool{service: service}
}

func (t *UpdateFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_folder",
		mcp.WithDescription("Update an existing folder"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID")),
		mcp.WithString("name", mcp.Description("New folder name")),
		mcp.WithString("description", mcp.Description("New folder description")),
	)
}

func (t *UpdateFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		// Get existing folder
		existing, err := t.service.GetFolderByID(userID, folderID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get folder: %v", err)), nil
		}
		if existing == nil {
			return mcp.NewToolResultError("Folder not found"), nil
		}

		// Update fields
		if name, ok := args["name"].(string); ok && name != "" {
			existing.Name = name
		}
		if description, ok := args["description"].(string); ok {
			existing.Description = description
		}

		if err := t.service.UpdateFolder(userID, existing); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update folder: %v", err)), nil
		}

		// Fetch updated folder
		updated, _ := t.service.GetFolderByID(userID, folderID)
		result, _ := json.Marshal(folderToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// DeleteFolderTool handles deleting a folder
type DeleteFolderTool struct {
	service services.FolderService
}

func NewDeleteFolderTool(service services.FolderService) *DeleteFolderTool {
	return &DeleteFolderTool{service: service}
}

func (t *DeleteFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_folder",
		mcp.WithDescription("Delete a folder and all its contents"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID")),
	)
}

func (t *DeleteFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		if err := t.service.DeleteFolder(userID, folderID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete folder: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"message":   "Folder deleted successfully",
			"folder_id": folderID,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// MoveFolderTool handles moving a folder to a new parent
type MoveFolderTool struct {
	service services.FolderService
}

func NewMoveFolderTool(service services.FolderService) *MoveFolderTool {
	return &MoveFolderTool{service: service}
}

func (t *MoveFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("move_folder",
		mcp.WithDescription("Move a folder to a new parent folder"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID to move")),
		mcp.WithNumber("parent_id", mcp.Description("New parent folder ID (omit to move to root)")),
	)
}

func (t *MoveFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		var newParentID *uint
		if parentID := getUintArg(args, "parent_id"); parentID > 0 {
			newParentID = &parentID
		}

		if err := t.service.MoveFolder(userID, folderID, newParentID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to move folder: %v", err)), nil
		}

		// Fetch updated folder
		updated, _ := t.service.GetFolderByID(userID, folderID)
		result, _ := json.Marshal(folderToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// GetFolderTreeTool handles getting the folder tree structure
type GetFolderTreeTool struct {
	service services.FolderService
}

func NewGetFolderTreeTool(service services.FolderService) *GetFolderTreeTool {
	return &GetFolderTreeTool{service: service}
}

func (t *GetFolderTreeTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_folder_tree",
		mcp.WithDescription("Get the folder tree structure"),
		mcp.WithNumber("parent_id", mcp.Description("Parent folder ID to start from (omit for full tree from root)")),
	)
}

func (t *GetFolderTreeTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		var parentID *uint
		if pid := getUintArg(args, "parent_id"); pid > 0 {
			parentID = &pid
		}

		folders, err := t.service.GetFolderTree(userID, parentID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get folder tree: %v", err)), nil
		}

		result, _ := json.Marshal(foldersToTreeMap(folders))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// AddTagsToFolderTool handles adding tags to a folder
type AddTagsToFolderTool struct {
	service services.FolderService
}

func NewAddTagsToFolderTool(service services.FolderService) *AddTagsToFolderTool {
	return &AddTagsToFolderTool{service: service}
}

func (t *AddTagsToFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("add_tags_to_folder",
		mcp.WithDescription("Add tags to a folder"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID")),
		mcp.WithString("tag_ids", mcp.Required(), mcp.Description("Comma-separated tag IDs to add")),
	)
}

func (t *AddTagsToFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
		if len(tagIDs) == 0 {
			return mcp.NewToolResultError("tag_ids is required"), nil
		}

		if err := t.service.AddTagsToFolder(userID, folderID, tagIDs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to add tags: %v", err)), nil
		}

		// Fetch updated folder
		updated, _ := t.service.GetFolderByID(userID, folderID)
		result, _ := json.Marshal(folderToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// RemoveTagsFromFolderTool handles removing tags from a folder
type RemoveTagsFromFolderTool struct {
	service services.FolderService
}

func NewRemoveTagsFromFolderTool(service services.FolderService) *RemoveTagsFromFolderTool {
	return &RemoveTagsFromFolderTool{service: service}
}

func (t *RemoveTagsFromFolderTool) GetTool() mcp.Tool {
	return mcp.NewTool("remove_tags_from_folder",
		mcp.WithDescription("Remove tags from a folder"),
		mcp.WithNumber("folder_id", mcp.Required(), mcp.Description("Folder ID")),
		mcp.WithString("tag_ids", mcp.Required(), mcp.Description("Comma-separated tag IDs to remove")),
	)
}

func (t *RemoveTagsFromFolderTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		folderID := getUintArg(args, "folder_id")
		if folderID == 0 {
			return mcp.NewToolResultError("folder_id is required"), nil
		}

		tagIDs := parseTagIDs(getStringArg(args, "tag_ids"))
		if len(tagIDs) == 0 {
			return mcp.NewToolResultError("tag_ids is required"), nil
		}

		if err := t.service.RemoveTagsFromFolder(userID, folderID, tagIDs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to remove tags: %v", err)), nil
		}

		// Fetch updated folder
		updated, _ := t.service.GetFolderByID(userID, folderID)
		result, _ := json.Marshal(folderToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// Helper functions
func folderToMap(folder *models.Folder) map[string]interface{} {
	m := map[string]interface{}{
		"id":          folder.ID,
		"name":        folder.Name,
		"description": folder.Description,
		"parent_id":   folder.ParentID,
		"created_at":  folder.CreatedAt,
		"updated_at":  folder.UpdatedAt,
	}

	if len(folder.Tags) > 0 {
		tags := make([]map[string]interface{}, len(folder.Tags))
		for i, tag := range folder.Tags {
			tags[i] = tagToMap(&tag)
		}
		m["tags"] = tags
	}

	return m
}

func foldersToTreeMap(folders []models.Folder) []map[string]interface{} {
	result := make([]map[string]interface{}, len(folders))
	for i, folder := range folders {
		m := folderToMap(&folder)
		if len(folder.Children) > 0 {
			m["children"] = foldersToTreeMap(folder.Children)
		}
		result[i] = m
	}
	return result
}

func parseTagIDs(tagIDsStr string) []uint {
	if tagIDsStr == "" {
		return nil
	}

	var tagIDs []uint
	for _, s := range splitAndTrim(tagIDsStr, ",") {
		if id := parseUint(s); id > 0 {
			tagIDs = append(tagIDs, id)
		}
	}
	return tagIDs
}

func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}

	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			part := trimSpace(s[start:i])
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	part := trimSpace(s[start:])
	if part != "" {
		parts = append(parts, part)
	}
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func parseUint(s string) uint {
	var n uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint(c-'0')
		}
	}
	return n
}
