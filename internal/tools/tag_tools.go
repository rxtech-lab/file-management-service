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

// TagTool handles all tag CRUD operations via a single tool with an action parameter
type TagTool struct {
	service services.TagService
}

func NewTagTool(service services.TagService) *TagTool {
	return &TagTool{service: service}
}

func (t *TagTool) GetTool() mcp.Tool {
	return mcp.NewTool("manage_tags",
		mcp.WithDescription("Manage tags for organizing files and folders. Use the 'action' parameter to specify the operation."),
		mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform: create, list, get, update, delete")),
		mcp.WithNumber("tag_id", mcp.Description("Tag ID (required for get, update, delete)")),
		mcp.WithString("name", mcp.Description("Tag name (required for create, optional for update)")),
		mcp.WithString("description", mcp.Description("Tag description (optional for create, update)")),
		mcp.WithString("color", mcp.Description("Hex color code e.g. #FF5733 (optional for create, update)")),
		mcp.WithString("keyword", mcp.Description("Search keyword to filter tags (for list)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of tags to return, default 100 (for list)")),
		mcp.WithNumber("offset", mcp.Description("Number of tags to skip for pagination (for list)")),
	)
}

func (t *TagTool) GetHandler() server.ToolHandlerFunc {
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
			return t.handleDelete(userID, args)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Invalid action: %s. Use create, list, get, update, or delete", action)), nil
		}
	}
}

func (t *TagTool) handleCreate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name is required for create action"), nil
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

func (t *TagTool) handleList(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
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

func (t *TagTool) handleGet(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required for get action"), nil
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

func (t *TagTool) handleUpdate(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required for update action"), nil
	}

	existing, err := t.service.GetTagByID(userID, tagID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get tag: %v", err)), nil
	}
	if existing == nil {
		return mcp.NewToolResultError("Tag not found"), nil
	}

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

	updated, _ := t.service.GetTagByID(userID, tagID)
	result, _ := json.Marshal(tagToMap(updated))
	return mcp.NewToolResultText(string(result)), nil
}

func (t *TagTool) handleDelete(userID string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	tagID := getUintArg(args, "tag_id")
	if tagID == 0 {
		return mcp.NewToolResultError("tag_id is required for delete action"), nil
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
