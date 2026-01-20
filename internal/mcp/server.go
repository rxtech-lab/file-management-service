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
) *MCPServer {
	mcpServer := &MCPServer{
		dbService: dbService,
	}
	mcpServer.initializeTools(tagService, folderService, fileService, uploadService, searchService)
	return mcpServer
}

// initializeTools registers all file management tools
func (s *MCPServer) initializeTools(
	tagService services.TagService,
	folderService services.FolderService,
	fileService services.FileService,
	uploadService services.UploadService,
	searchService services.SearchService,
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

	// Tag Tools
	createTagTool := tools.NewCreateTagTool(tagService)
	srv.AddTool(createTagTool.GetTool(), createTagTool.GetHandler())

	listTagsTool := tools.NewListTagsTool(tagService)
	srv.AddTool(listTagsTool.GetTool(), listTagsTool.GetHandler())

	getTagTool := tools.NewGetTagTool(tagService)
	srv.AddTool(getTagTool.GetTool(), getTagTool.GetHandler())

	updateTagTool := tools.NewUpdateTagTool(tagService)
	srv.AddTool(updateTagTool.GetTool(), updateTagTool.GetHandler())

	deleteTagTool := tools.NewDeleteTagTool(tagService)
	srv.AddTool(deleteTagTool.GetTool(), deleteTagTool.GetHandler())

	// Folder Tools
	createFolderTool := tools.NewCreateFolderTool(folderService)
	srv.AddTool(createFolderTool.GetTool(), createFolderTool.GetHandler())

	listFoldersTool := tools.NewListFoldersTool(folderService)
	srv.AddTool(listFoldersTool.GetTool(), listFoldersTool.GetHandler())

	getFolderTool := tools.NewGetFolderTool(folderService)
	srv.AddTool(getFolderTool.GetTool(), getFolderTool.GetHandler())

	updateFolderTool := tools.NewUpdateFolderTool(folderService)
	srv.AddTool(updateFolderTool.GetTool(), updateFolderTool.GetHandler())

	deleteFolderTool := tools.NewDeleteFolderTool(folderService)
	srv.AddTool(deleteFolderTool.GetTool(), deleteFolderTool.GetHandler())

	moveFolderTool := tools.NewMoveFolderTool(folderService)
	srv.AddTool(moveFolderTool.GetTool(), moveFolderTool.GetHandler())

	getFolderTreeTool := tools.NewGetFolderTreeTool(folderService)
	srv.AddTool(getFolderTreeTool.GetTool(), getFolderTreeTool.GetHandler())

	addTagsToFolderTool := tools.NewAddTagsToFolderTool(folderService)
	srv.AddTool(addTagsToFolderTool.GetTool(), addTagsToFolderTool.GetHandler())

	removeTagsFromFolderTool := tools.NewRemoveTagsFromFolderTool(folderService)
	srv.AddTool(removeTagsFromFolderTool.GetTool(), removeTagsFromFolderTool.GetHandler())

	// File Tools
	createFileTool := tools.NewCreateFileTool(fileService)
	srv.AddTool(createFileTool.GetTool(), createFileTool.GetHandler())

	listFilesTool := tools.NewListFilesTool(fileService)
	srv.AddTool(listFilesTool.GetTool(), listFilesTool.GetHandler())

	getFileTool := tools.NewGetFileTool(fileService)
	srv.AddTool(getFileTool.GetTool(), getFileTool.GetHandler())

	updateFileTool := tools.NewUpdateFileTool(fileService)
	srv.AddTool(updateFileTool.GetTool(), updateFileTool.GetHandler())

	deleteFileTool := tools.NewDeleteFileTool(fileService)
	srv.AddTool(deleteFileTool.GetTool(), deleteFileTool.GetHandler())

	moveFilesTool := tools.NewMoveFilesTool(fileService)
	srv.AddTool(moveFilesTool.GetTool(), moveFilesTool.GetHandler())

	addTagsToFileTool := tools.NewAddTagsToFileTool(fileService)
	srv.AddTool(addTagsToFileTool.GetTool(), addTagsToFileTool.GetHandler())

	removeTagsFromFileTool := tools.NewRemoveTagsFromFileTool(fileService)
	srv.AddTool(removeTagsFromFileTool.GetTool(), removeTagsFromFileTool.GetHandler())

	getFileDownloadURLTool := tools.NewGetFileDownloadURLTool(fileService, uploadService)
	srv.AddTool(getFileDownloadURLTool.GetTool(), getFileDownloadURLTool.GetHandler())

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
		return `Tag Management Tools:

1. create_tag - Create a new tag
   Parameters: name (required), description, color

2. list_tags - List all tags with optional search
   Parameters: keyword, limit, offset

3. get_tag - Get a tag by ID
   Parameters: tag_id (required)

4. update_tag - Update an existing tag
   Parameters: tag_id (required), name, description, color

5. delete_tag - Delete a tag
   Parameters: tag_id (required)`

	case "folder":
		return `Folder Management Tools:

1. create_folder - Create a new folder
   Parameters: name (required), description, parent_id

2. list_folders - List folders with optional filtering
   Parameters: keyword, parent_id, tag_ids, limit, offset

3. get_folder - Get a folder by ID
   Parameters: folder_id (required)

4. update_folder - Update an existing folder
   Parameters: folder_id (required), name, description

5. delete_folder - Delete a folder
   Parameters: folder_id (required)

6. move_folder - Move a folder to a new parent
   Parameters: folder_id (required), parent_id

7. get_folder_tree - Get folder tree structure
   Parameters: parent_id (optional)

8. add_tags_to_folder - Add tags to a folder
   Parameters: folder_id (required), tag_ids (required)

9. remove_tags_from_folder - Remove tags from a folder
   Parameters: folder_id (required), tag_ids (required)`

	case "file":
		return `File Management Tools:

1. create_file - Create a new file record
   Parameters: title (required), s3_key (required), original_filename (required),
               file_type, folder_id, mime_type, size

2. list_files - List files with filtering
   Parameters: keyword, folder_id, file_type, tag_ids, status, sort_by, sort_order, limit, offset

3. get_file - Get a file by ID
   Parameters: file_id (required)

4. update_file - Update an existing file
   Parameters: file_id (required), title, summary, file_type, folder_id

5. delete_file - Delete a file
   Parameters: file_id (required)

6. move_files - Move files to a different folder
   Parameters: file_ids (required), folder_id

7. add_tags_to_file - Add tags to a file
   Parameters: file_id (required), tag_ids (required)

8. remove_tags_from_file - Remove tags from a file
   Parameters: file_id (required), tag_ids (required)

9. get_file_download_url - Get presigned download URL
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

TAG MANAGEMENT (5 tools):
- create_tag: Create a new tag
- list_tags: List tags with search
- get_tag: Get tag details
- update_tag: Update a tag
- delete_tag: Delete a tag

FOLDER MANAGEMENT (9 tools):
- create_folder: Create a new folder
- list_folders: List folders with filters
- get_folder: Get folder details
- update_folder: Update a folder
- delete_folder: Delete a folder
- move_folder: Move folder to new parent
- get_folder_tree: Get folder hierarchy
- add_tags_to_folder: Tag a folder
- remove_tags_from_folder: Untag a folder

FILE MANAGEMENT (9 tools):
- create_file: Create file record
- list_files: List with filters
- get_file: Get file details
- update_file: Update a file
- delete_file: Delete a file
- move_files: Batch move files
- add_tags_to_file: Tag a file
- remove_tags_from_file: Untag a file
- get_file_download_url: Get download URL

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
