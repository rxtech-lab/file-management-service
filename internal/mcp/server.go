package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/tools"
)

// MCPServer wraps the MCP server with file management tools
type MCPServer struct {
	server    *server.MCPServer
	dbService services.DBService
}

// NewMCPServer creates a new MCP server with file management tools
func NewMCPServer(
	dbService services.DBService,
	tagService services.TagService,
	folderService services.FolderService,
	fileService services.FileService,
	uploadService services.UploadService,
	searchService services.SearchService,
	embeddingService services.EmbeddingService,
	invoiceService services.InvoiceService,
) *MCPServer {
	mcpServer := &MCPServer{
		dbService: dbService,
	}
	mcpServer.initializeTools(tagService, folderService, fileService, uploadService, searchService, embeddingService, invoiceService)
	return mcpServer
}

// initializeTools registers all file management tools
func (s *MCPServer) initializeTools(
	tagService services.TagService,
	folderService services.FolderService,
	fileService services.FileService,
	uploadService services.UploadService,
	searchService services.SearchService,
	embeddingService services.EmbeddingService,
	invoiceService services.InvoiceService,
) {
	srv := server.NewMCPServer(
		"File Management MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	srv.EnableSampling()

	// Add usage prompt
	srv.AddPrompt(mcp.NewPrompt("file-management-usage",
		mcp.WithPromptDescription("Instructions and guidance for using file management tools"),
		mcp.WithArgument("tool_category",
			mcp.ArgumentDescription("Category of tools to get instructions for (tag, folder, file, search, upload, or all)"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		category := request.Params.Arguments["tool_category"]
		if category == "" {
			return nil, fmt.Errorf("tool_category is required")
		}

		instructions := getToolInstructions(category)

		return mcp.NewGetPromptResult(
			fmt.Sprintf("File Management Tools - %s", category),
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(instructions),
				),
			},
		), nil
	})

	// Tag Tool (unified CRUD)
	tagTool := tools.NewTagTool(tagService)
	srv.AddTool(tagTool.GetTool(), tagTool.GetHandler())

	// Folder Tool (unified CRUD + move, tree, tags)
	folderTool := tools.NewFolderTool(folderService, fileService, uploadService, embeddingService)
	srv.AddTool(folderTool.GetTool(), folderTool.GetHandler())

	// File Tool (unified CRUD + move, tags, download URL)
	fileTool := tools.NewFileTool(fileService, uploadService, embeddingService, invoiceService)
	srv.AddTool(fileTool.GetTool(), fileTool.GetHandler())

	// Upload Tools
	getPresignedURLTool := tools.NewGetPresignedURLTool(uploadService)
	srv.AddTool(getPresignedURLTool.GetTool(), getPresignedURLTool.GetHandler())

	// Search Tools
	searchFilesTool := tools.NewSearchFilesTool(searchService)
	srv.AddTool(searchFilesTool.GetTool(), searchFilesTool.GetHandler())

	s.server = srv
}

// SendMessageToAiClient sends a message to the AI client
func (s *MCPServer) SendMessageToAiClient(messages []mcp.SamplingMessage) error {
	samplingRequest := mcp.CreateMessageRequest{
		CreateMessageParams: mcp.CreateMessageParams{
			Messages: messages,
		},
	}

	samplingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	serverFromCtx := server.ServerFromContext(samplingCtx)
	_, err := serverFromCtx.RequestSampling(samplingCtx, samplingRequest)
	if err != nil {
		return err
	}
	return nil
}

// getToolInstructions returns instructions for the specified tool category
func getToolInstructions(category string) string {
	switch category {
	case "tag":
		return `Tag Management Tool:

manage_tags - Unified tag management tool
   action: create - Create a new tag
      Parameters: name (required), description, color

   action: list - List all tags with optional search
      Parameters: keyword, limit, offset

   action: get - Get a tag by ID
      Parameters: tag_id (required)

   action: update - Update an existing tag
      Parameters: tag_id (required), name, description, color

   action: delete - Delete a tag
      Parameters: tag_id (required)`

	case "folder":
		return `Folder Management Tool:

manage_folders - Unified folder management tool
   action: create - Create a new folder
      Parameters: name (required), description, parent_id

   action: list - List folders with optional filtering
      Parameters: keyword, parent_id, tag_ids, limit, offset

   action: get - Get a folder by ID
      Parameters: folder_id (required)

   action: update - Update an existing folder
      Parameters: folder_id (required), name, description

   action: delete - Delete a folder
      Parameters: folder_id (required)

   action: move - Move a folder to a new parent
      Parameters: folder_id (required), parent_id

   action: get_tree - Get folder tree structure
      Parameters: parent_id (optional)

   action: add_tags - Add tags to a folder
      Parameters: folder_id (required), tag_ids (required)

   action: remove_tags - Remove tags from a folder
      Parameters: folder_id (required), tag_ids (required)`

	case "file":
		return `File Management Tool:

manage_files - Unified file management tool
   action: create - Create a new file record
      Parameters: title (required), s3_key (required), original_filename (required),
                  file_type, folder_id, mime_type, size

   action: list - List files with filtering
      Parameters: keyword, folder_id, file_type, tag_ids, status, sort_by, sort_order, limit, offset

   action: get - Get a file by ID
      Parameters: file_id (required)

   action: update - Update an existing file
      Parameters: file_id (required), title, summary, file_type, folder_id

   action: delete - Delete a file
      Parameters: file_id (required)

   action: move - Move files to a different folder
      Parameters: file_ids (required), folder_id

   action: add_tags - Add tags to a file
      Parameters: file_id (required), tag_ids (required)

   action: remove_tags - Remove tags from a file
      Parameters: file_id (required), tag_ids (required)

   action: get_download_url - Get presigned download URL
      Parameters: file_id (required)`

	case "search":
		return `Search Tools:

1. search_files - Search files with different modes
   Parameters: query (required), type (fulltext|semantic|hybrid),
               folder_id, file_type, tag_ids, limit, offset

   Search types:
   - fulltext: Traditional text search on title, summary, and content
   - semantic: Vector-based semantic search using embeddings
   - hybrid: Combines fulltext and semantic search for best results`

	case "upload":
		return `File Upload Tools:

1. get_presigned_url - Get a presigned URL for uploading a file
   Parameters: filename (required), content_type

   Usage: Use this to get a URL for directly uploading files to S3.
   The returned URL can be used with PUT request to upload the file.
   After upload, use the returned key as the s3_key when creating a file record.`

	case "all":
		return `File Management MCP Tools Overview:

This MCP server provides tools for managing files, folders, tags, and search.
Each resource has a single unified tool with an 'action' parameter.

TAG MANAGEMENT (1 tool):
- manage_tags: Unified tag operations
  Actions: create, list, get, update, delete

FOLDER MANAGEMENT (1 tool):
- manage_folders: Unified folder operations
  Actions: create, list, get, update, delete, move, get_tree, add_tags, remove_tags

FILE MANAGEMENT (1 tool):
- manage_files: Unified file operations
  Actions: create, list, get, update, delete, move, add_tags, remove_tags, get_download_url

SEARCH (1 tool):
- search_files: Search with fulltext, semantic, or hybrid mode

FILE UPLOAD (1 tool):
- get_presigned_url: Get URL for file upload

All tools require authentication. Files are user-scoped.`

	default:
		return `Invalid category. Available categories: tag, folder, file, search, upload, all`
	}
}

// StartStdioServer starts the MCP server with stdio interface
func (s *MCPServer) StartStdioServer() error {
	return server.ServeStdio(s.server)
}

// StartStreamableHTTPServer starts the MCP server with streamable HTTP interface
func (s *MCPServer) StartStreamableHTTPServer() *server.StreamableHTTPServer {
	return server.NewStreamableHTTPServer(s.server)
}

// GetDBService returns the database service
func (s *MCPServer) GetDBService() services.DBService {
	return s.dbService
}

// GetServer returns the underlying MCP server
func (s *MCPServer) GetServer() *server.MCPServer {
	return s.server
}
