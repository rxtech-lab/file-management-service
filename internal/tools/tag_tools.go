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

// CreateTagTool handles creating a new tag
type CreateTagTool struct {
	service services.TagService
}

func NewCreateTagTool(service services.TagService) *CreateTagTool {
	return &CreateTagTool{service: service}
}

func (t *CreateTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("create_tag",
		mcp.WithDescription("Create a new tag for organizing files and folders"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Tag name")),
		mcp.WithString("description", mcp.Description("Tag description")),
		mcp.WithString("color", mcp.Description("Hex color code (e.g., #FF5733)")),
	)
}

func (t *CreateTagTool) GetHandler() server.ToolHandlerFunc {
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

		tag := &models.Tag{
			Name:        name,
			Description: getStringArg(args, "description"),
			Color:       getStringArg(args, "color"),
		}

		if err := t.service.CreateTag(userID, tag); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create tag: %v", err)), nil
		}

		result, _ := json.Marshal(tagToMap(tag))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// ListTagsTool handles listing tags
type ListTagsTool struct {
	service services.TagService
}

func NewListTagsTool(service services.TagService) *ListTagsTool {
	return &ListTagsTool{service: service}
}

func (t *ListTagsTool) GetTool() mcp.Tool {
	return mcp.NewTool("list_tags",
		mcp.WithDescription("List all tags with optional keyword search"),
		mcp.WithString("keyword", mcp.Description("Search keyword to filter tags")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of tags to return (default: 100)")),
		mcp.WithNumber("offset", mcp.Description("Number of tags to skip for pagination")),
	)
}

func (t *ListTagsTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		keyword := getStringArg(args, "keyword")
		limit := getIntArg(args, "limit", 100)
		offset := getIntArg(args, "offset", 0)

		tags, total, err := t.service.ListTags(userID, keyword, limit, offset)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tags: %v", err)), nil
		}

		tagList := make([]map[string]interface{}, len(tags))
		for i, tag := range tags {
			tagList[i] = tagToMap(&tag)
		}

		result, _ := json.Marshal(map[string]interface{}{
			"data":   tagList,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// GetTagTool handles getting a tag by ID
type GetTagTool struct {
	service services.TagService
}

func NewGetTagTool(service services.TagService) *GetTagTool {
	return &GetTagTool{service: service}
}

func (t *GetTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_tag",
		mcp.WithDescription("Get a tag by its ID"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *GetTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		tagID := getUintArg(args, "tag_id")
		if tagID == 0 {
			return mcp.NewToolResultError("tag_id is required"), nil
		}

		tag, err := t.service.GetTagByID(userID, tagID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get tag: %v", err)), nil
		}
		if tag == nil {
			return mcp.NewToolResultError("Tag not found"), nil
		}

		result, _ := json.Marshal(tagToMap(tag))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// UpdateTagTool handles updating a tag
type UpdateTagTool struct {
	service services.TagService
}

func NewUpdateTagTool(service services.TagService) *UpdateTagTool {
	return &UpdateTagTool{service: service}
}

func (t *UpdateTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("update_tag",
		mcp.WithDescription("Update an existing tag"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
		mcp.WithString("name", mcp.Description("New tag name")),
		mcp.WithString("description", mcp.Description("New tag description")),
		mcp.WithString("color", mcp.Description("New hex color code")),
	)
}

func (t *UpdateTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		tagID := getUintArg(args, "tag_id")
		if tagID == 0 {
			return mcp.NewToolResultError("tag_id is required"), nil
		}

		// Get existing tag
		existing, err := t.service.GetTagByID(userID, tagID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get tag: %v", err)), nil
		}
		if existing == nil {
			return mcp.NewToolResultError("Tag not found"), nil
		}

		// Update fields
		if name, ok := args["name"].(string); ok && name != "" {
			existing.Name = name
		}
		if description, ok := args["description"].(string); ok {
			existing.Description = description
		}
		if color, ok := args["color"].(string); ok {
			existing.Color = color
		}

		if err := t.service.UpdateTag(userID, existing); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update tag: %v", err)), nil
		}

		// Fetch updated tag
		updated, _ := t.service.GetTagByID(userID, tagID)
		result, _ := json.Marshal(tagToMap(updated))
		return mcp.NewToolResultText(string(result)), nil
	}
}

// DeleteTagTool handles deleting a tag
type DeleteTagTool struct {
	service services.TagService
}

func NewDeleteTagTool(service services.TagService) *DeleteTagTool {
	return &DeleteTagTool{service: service}
}

func (t *DeleteTagTool) GetTool() mcp.Tool {
	return mcp.NewTool("delete_tag",
		mcp.WithDescription("Delete a tag"),
		mcp.WithNumber("tag_id", mcp.Required(), mcp.Description("Tag ID")),
	)
}

func (t *DeleteTagTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		tagID := getUintArg(args, "tag_id")
		if tagID == 0 {
			return mcp.NewToolResultError("tag_id is required"), nil
		}

		if err := t.service.DeleteTag(userID, tagID); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete tag: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"message": "Tag deleted successfully",
			"tag_id":  tagID,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// Helper functions
func tagToMap(tag *models.Tag) map[string]interface{} {
	return map[string]interface{}{
		"id":          tag.ID,
		"name":        tag.Name,
		"description": tag.Description,
		"color":       tag.Color,
		"created_at":  tag.CreatedAt,
		"updated_at":  tag.UpdatedAt,
	}
}

func getStringArg(args map[string]interface{}, key string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return ""
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

func getUintArg(args map[string]interface{}, key string) uint {
	if val, ok := args[key].(float64); ok {
		return uint(val)
	}
	return 0
}

func getUintSliceArg(args map[string]interface{}, key string) []uint {
	var result []uint
	if val, ok := args[key].([]interface{}); ok {
		for _, v := range val {
			if num, ok := v.(float64); ok {
				result = append(result, uint(num))
			}
		}
	}
	return result
}
