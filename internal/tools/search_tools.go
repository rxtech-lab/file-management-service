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

// SearchFilesTool handles searching files
type SearchFilesTool struct {
	service services.SearchService
}

func NewSearchFilesTool(service services.SearchService) *SearchFilesTool {
	return &SearchFilesTool{service: service}
}

func (t *SearchFilesTool) GetTool() mcp.Tool {
	return mcp.NewTool("search_files",
		mcp.WithDescription("Search files using different search modes"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
		mcp.WithString("type", mcp.Description("Search type: fulltext, semantic, or hybrid (default: hybrid)")),
		mcp.WithNumber("folder_id", mcp.Description("Filter results to a specific folder")),
		mcp.WithString("file_type", mcp.Description("Filter by file type: music, photo, video, document, invoice")),
		mcp.WithString("tag_ids", mcp.Description("Comma-separated tag IDs to filter by")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of results (default: 20)")),
		mcp.WithNumber("offset", mcp.Description("Number of results to skip for pagination")),
	)
}

func (t *SearchFilesTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := utils.GetUserID(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		query, _ := args["query"].(string)
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		searchType := getStringArg(args, "type")
		if searchType == "" {
			searchType = "hybrid"
		}

		opts := services.SearchOptions{
			Limit:  getIntArg(args, "limit", 20),
			Offset: getIntArg(args, "offset", 0),
			TagIDs: parseTagIDs(getStringArg(args, "tag_ids")),
		}

		if folderID := getUintArg(args, "folder_id"); folderID > 0 {
			opts.FolderID = &folderID
		}

		if fileType := getStringArg(args, "file_type"); fileType != "" {
			opts.FileTypes = []models.FileType{models.FileType(fileType)}
		}

		var results []services.SearchResult
		var total int64
		var err error

		switch searchType {
		case "fulltext":
			results, total, err = t.service.FullTextSearch(userID, query, opts)
		case "semantic":
			results, err = t.service.VectorSearch(ctx, userID, query, opts)
			total = int64(len(results))
		case "hybrid":
			results, err = t.service.HybridSearch(ctx, userID, query, opts)
			total = int64(len(results))
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Invalid search type: %s. Use fulltext, semantic, or hybrid", searchType)), nil
		}

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
		}

		resultList := make([]map[string]any, len(results))
		for i, r := range results {
			resultList[i] = searchResultToMap(r)
		}

		result, _ := json.Marshal(map[string]any{
			"data":        resultList,
			"total":       total,
			"query":       query,
			"search_type": searchType,
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}

// Helper function
func searchResultToMap(r services.SearchResult) map[string]any {
	return map[string]any{
		"file":    fileToMap(&r.File),
		"score":   r.Score,
		"snippet": r.Snippet,
	}
}
