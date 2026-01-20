package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rxtech-lab/invoice-management/internal/models"
)

// AgentConfig holds configuration for the AI agent service
type AgentConfig struct {
	GatewayURL string // AI_GATEWAY_URL env var
	APIKey     string // AI_GATEWAY_API_KEY env var
	Model      string // AGENT_MODEL env var (default: gpt-4o-mini)
	MaxTurns   int    // AGENT_MAX_TURNS env var (default: 10)
	Enabled    bool   // AGENT_ENABLED env var (default: true)
}

// AgentEvent represents a real-time status update from the agent
type AgentEvent struct {
	Type    string      `json:"type"`              // "status", "tool_call", "tool_result", "thinking", "result", "error"
	Message string      `json:"message"`           // Human-readable status message
	Data    interface{} `json:"data,omitempty"`    // Optional additional data
	Tool    string      `json:"tool,omitempty"`    // Tool name if type is tool_call
	FileID  uint        `json:"file_id,omitempty"` // File ID being processed
}

// AgentService handles AI-powered file organization
type AgentService interface {
	// ProcessFileWithAgent runs the agent to tag and organize a file
	// eventChan receives real-time status updates for SSE streaming
	ProcessFileWithAgent(ctx context.Context, userID string, fileID uint,
		content, summary string, eventChan chan<- AgentEvent) error

	// OrganizeFile lets user trigger AI to reorganize an existing file
	OrganizeFile(ctx context.Context, userID string, fileID uint,
		eventChan chan<- AgentEvent) error

	// IsEnabled returns whether the agent is enabled
	IsEnabled() bool
}

type agentService struct {
	config        AgentConfig
	client        *http.Client
	tagService    TagService
	fileService   FileService
	folderService FolderService
}

// NewAgentService creates a new AgentService
func NewAgentService(
	config AgentConfig,
	tagService TagService,
	fileService FileService,
	folderService FolderService,
) AgentService {
	if config.MaxTurns <= 0 {
		config.MaxTurns = 10
	}
	if config.Model == "" {
		config.Model = "gpt-4o-mini"
	}
	return &agentService{
		config:        config,
		client:        &http.Client{},
		tagService:    tagService,
		fileService:   fileService,
		folderService: folderService,
	}
}

func (s *agentService) IsEnabled() bool {
	return s.config.Enabled && s.config.GatewayURL != "" && s.config.APIKey != ""
}

// Tool definitions for OpenAI-compatible function calling
type toolDefinition struct {
	Type     string         `json:"type"`
	Function functionSchema `json:"function"`
}

type functionSchema struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Parameters  parametersSchema `json:"parameters"`
}

type parametersSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Chat completion request/response types
type agentChatRequest struct {
	Model      string           `json:"model"`
	Messages   []agentMessage   `json:"messages"`
	Tools      []toolDefinition `json:"tools,omitempty"`
	ToolChoice interface{}      `json:"tool_choice,omitempty"`
}

type agentMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function functionCall `json:"function"`
}

type functionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type agentChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int          `json:"index"`
		Message      agentMessage `json:"message"`
		FinishReason string       `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// getTools returns the tool definitions for the agent
func (s *agentService) getTools() []toolDefinition {
	return []toolDefinition{
		{
			Type: "function",
			Function: functionSchema{
				Name:        "search_tags",
				Description: "Search for existing tags by keyword. Use this to find relevant tags before creating new ones.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"keyword": map[string]interface{}{
							"type":        "string",
							"description": "Search keyword to find matching tags",
						},
					},
					Required: []string{"keyword"},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "list_all_tags",
				Description: "List all existing tags for the user. Use this to see what tags are available.",
				Parameters: parametersSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "create_tag",
				Description: "Create a new tag. Only create tags if no suitable existing tag was found.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Tag name (lowercase, use hyphens for spaces, e.g., 'project-report')",
						},
						"color": map[string]interface{}{
							"type":        "string",
							"description": "Hex color code for the tag (e.g., '#FF5733')",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Brief description of what this tag represents",
						},
					},
					Required: []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "add_tags_to_file",
				Description: "Add one or more tags to the file being processed.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"tag_ids": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "integer"},
							"description": "Array of tag IDs to add to the file",
						},
					},
					Required: []string{"tag_ids"},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "get_folder_tree",
				Description: "Get the hierarchical folder structure. Use this to understand how files are organized.",
				Parameters: parametersSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "list_folders",
				Description: "List folders at a specific level. Pass parent_id to list children of a folder, or omit for root folders.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"parent_id": map[string]interface{}{
							"type":        "integer",
							"description": "Parent folder ID. Omit or use null to list root folders.",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "move_file",
				Description: "Move the file to a different folder. Use this to organize the file into the appropriate folder.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"folder_id": map[string]interface{}{
							"type":        "integer",
							"description": "Target folder ID to move the file to. Use null to move to root.",
						},
					},
					Required: []string{"folder_id"},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "get_file_info",
				Description: "Get current information about the file being processed, including its current folder and tags.",
				Parameters: parametersSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: functionSchema{
				Name:        "create_folder",
				Description: "Create a new folder. Only create if no suitable existing folder was found.",
				Parameters: parametersSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Folder name",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Brief description of what this folder is for",
						},
						"parent_id": map[string]interface{}{
							"type":        "integer",
							"description": "Parent folder ID. Omit for root level folder.",
						},
					},
					Required: []string{"name"},
				},
			},
		},
	}
}

// getSystemPrompt returns the system prompt for the agent
func (s *agentService) getSystemPrompt() string {
	return `You are a file organization assistant. Your job is to analyze file content and organize it properly by:

1. **Tagging**: Search for existing tags that match the file content. Add relevant existing tags. Only create new tags if no suitable existing tags are found.

2. **Folder Organization**: If the file is in the root folder (no folder assigned), analyze the folder structure and move it to the most appropriate folder. If no suitable folder exists, you may create one.

Guidelines:
- Always search for existing tags first before creating new ones
- Tag names should be lowercase with hyphens (e.g., "project-report", "meeting-notes")
- Be conservative with new tag creation - prefer existing tags
- Consider the file type, content topics, and existing organization patterns
- Explain your reasoning briefly before taking actions
- When moving files, choose the most specific appropriate folder

Work efficiently - you have limited turns to complete the organization.`
}

// ProcessFileWithAgent runs the agent to tag and organize a file
func (s *agentService) ProcessFileWithAgent(
	ctx context.Context,
	userID string,
	fileID uint,
	content, summary string,
	eventChan chan<- AgentEvent,
) error {
	if !s.IsEnabled() {
		return nil
	}

	// Get file info
	file, err := s.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		eventChan <- AgentEvent{Type: "error", Message: "Failed to get file info", FileID: fileID}
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Truncate content for context window
	truncatedContent := content
	if len(truncatedContent) > 5000 {
		truncatedContent = truncatedContent[:5000] + "..."
	}

	// Build user prompt
	userPrompt := fmt.Sprintf(`Please organize this file:

**File Information:**
- Title: %s
- Type: %s
- Current Folder: %s

**Summary:**
%s

**Content (truncated):**
%s

Please:
1. First, list all existing tags to see what's available
2. Search for and add relevant existing tags
3. Create new tags only if needed for important topics not covered by existing tags
4. If the file is in the root folder, examine the folder structure and move it to an appropriate folder

Start by examining the existing tags and folder structure.`,
		file.Title,
		file.FileType,
		getFolderName(file),
		summary,
		truncatedContent,
	)

	// Initialize messages
	messages := []agentMessage{
		{Role: "system", Content: s.getSystemPrompt()},
		{Role: "user", Content: userPrompt},
	}

	eventChan <- AgentEvent{
		Type:    "status",
		Message: "AI agent started analyzing file...",
		FileID:  fileID,
	}

	// Agent loop
	for turn := 0; turn < s.config.MaxTurns; turn++ {
		eventChan <- AgentEvent{
			Type:    "thinking",
			Message: fmt.Sprintf("AI is thinking... (turn %d/%d)", turn+1, s.config.MaxTurns),
			FileID:  fileID,
		}

		// Call LLM
		response, err := s.callChatCompletions(ctx, messages)
		if err != nil {
			eventChan <- AgentEvent{Type: "error", Message: fmt.Sprintf("AI error: %v", err), FileID: fileID}
			return fmt.Errorf("chat completion failed: %w", err)
		}

		if len(response.Choices) == 0 {
			eventChan <- AgentEvent{Type: "error", Message: "No response from AI", FileID: fileID}
			return fmt.Errorf("no response from AI")
		}

		choice := response.Choices[0]
		assistantMsg := choice.Message

		// Add assistant message to history
		messages = append(messages, assistantMsg)

		// Check if model wants to call tools
		if choice.FinishReason == "tool_calls" && len(assistantMsg.ToolCalls) > 0 {
			// Execute each tool call
			for _, tc := range assistantMsg.ToolCalls {
				eventChan <- AgentEvent{
					Type:    "tool_call",
					Message: fmt.Sprintf("Executing: %s", formatToolCallMessage(tc.Function.Name, tc.Function.Arguments)),
					Tool:    tc.Function.Name,
					FileID:  fileID,
					Data:    tc,
				}

				// Execute tool
				result, err := s.executeTool(ctx, userID, fileID, tc)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}

				eventChan <- AgentEvent{
					Type:    "tool_result",
					Message: fmt.Sprintf("Result from %s", tc.Function.Name),
					Tool:    tc.Function.Name,
					FileID:  fileID,
					Data:    result,
				}

				// Add tool result to messages
				messages = append(messages, agentMessage{
					Role:       "tool",
					Content:    result,
					ToolCallID: tc.ID,
				})
			}
		} else if choice.FinishReason == "stop" {
			// Agent finished
			eventChan <- AgentEvent{
				Type:    "result",
				Message: assistantMsg.Content,
				FileID:  fileID,
			}
			return nil
		} else {
			// Unexpected finish reason, but might have content
			if assistantMsg.Content != "" {
				eventChan <- AgentEvent{
					Type:    "result",
					Message: assistantMsg.Content,
					FileID:  fileID,
				}
			}
			return nil
		}
	}

	eventChan <- AgentEvent{
		Type:    "error",
		Message: "Agent reached maximum turns without completing",
		FileID:  fileID,
	}
	return fmt.Errorf("max turns exceeded")
}

// OrganizeFile lets user trigger AI to reorganize an existing file
func (s *agentService) OrganizeFile(
	ctx context.Context,
	userID string,
	fileID uint,
	eventChan chan<- AgentEvent,
) error {
	// Get file with content
	file, err := s.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	return s.ProcessFileWithAgent(ctx, userID, fileID, file.Content, file.Summary, eventChan)
}

// callChatCompletions makes the API request to the AI gateway
func (s *agentService) callChatCompletions(ctx context.Context, messages []agentMessage) (*agentChatResponse, error) {
	reqBody := agentChatRequest{
		Model:      s.config.Model,
		Messages:   messages,
		Tools:      s.getTools(),
		ToolChoice: "auto",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(s.config.GatewayURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp agentChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	return &chatResp, nil
}

// executeTool runs the specified tool and returns the result
func (s *agentService) executeTool(ctx context.Context, userID string, fileID uint, tc toolCall) (string, error) {
	var args map[string]interface{}
	if tc.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	switch tc.Function.Name {
	case "search_tags":
		return s.executeSearchTags(userID, args)
	case "list_all_tags":
		return s.executeListAllTags(userID)
	case "create_tag":
		return s.executeCreateTag(userID, args)
	case "add_tags_to_file":
		return s.executeAddTagsToFile(userID, fileID, args)
	case "get_folder_tree":
		return s.executeGetFolderTree(userID)
	case "list_folders":
		return s.executeListFolders(userID, args)
	case "move_file":
		return s.executeMoveFile(userID, fileID, args)
	case "get_file_info":
		return s.executeGetFileInfo(userID, fileID)
	case "create_folder":
		return s.executeCreateFolder(userID, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", tc.Function.Name)
	}
}

// Tool implementations

func (s *agentService) executeSearchTags(userID string, args map[string]interface{}) (string, error) {
	keyword, _ := args["keyword"].(string)
	tags, _, err := s.tagService.ListTags(userID, keyword, 20, 0)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return fmt.Sprintf("No tags found matching '%s'", keyword), nil
	}

	result := fmt.Sprintf("Found %d tags matching '%s':\n", len(tags), keyword)
	for _, tag := range tags {
		result += fmt.Sprintf("- ID: %d, Name: %s", tag.ID, tag.Name)
		if tag.Description != "" {
			result += fmt.Sprintf(" (%s)", tag.Description)
		}
		result += "\n"
	}
	return result, nil
}

func (s *agentService) executeListAllTags(userID string) (string, error) {
	tags, _, err := s.tagService.ListTags(userID, "", 100, 0)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "No tags exist yet. You can create new tags as needed.", nil
	}

	result := fmt.Sprintf("All existing tags (%d total):\n", len(tags))
	for _, tag := range tags {
		result += fmt.Sprintf("- ID: %d, Name: %s", tag.ID, tag.Name)
		if tag.Description != "" {
			result += fmt.Sprintf(" (%s)", tag.Description)
		}
		result += "\n"
	}
	return result, nil
}

func (s *agentService) executeCreateTag(userID string, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name is required")
	}

	tag := &models.Tag{
		Name:        strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Color:       getStringArg(args, "color", ""),
		Description: getStringArg(args, "description", ""),
	}

	if err := s.tagService.CreateTag(userID, tag); err != nil {
		return "", err
	}

	return fmt.Sprintf("Created tag: ID=%d, Name='%s'", tag.ID, tag.Name), nil
}

func (s *agentService) executeAddTagsToFile(userID string, fileID uint, args map[string]interface{}) (string, error) {
	tagIDsRaw, ok := args["tag_ids"].([]interface{})
	if !ok || len(tagIDsRaw) == 0 {
		return "", fmt.Errorf("tag_ids is required and must be a non-empty array")
	}

	tagIDs := make([]uint, 0, len(tagIDsRaw))
	for _, id := range tagIDsRaw {
		switch v := id.(type) {
		case float64:
			tagIDs = append(tagIDs, uint(v))
		case int:
			tagIDs = append(tagIDs, uint(v))
		}
	}

	if len(tagIDs) == 0 {
		return "", fmt.Errorf("no valid tag IDs provided")
	}

	if err := s.fileService.AddTagsToFile(userID, fileID, tagIDs); err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully added %d tag(s) to the file", len(tagIDs)), nil
}

func (s *agentService) executeGetFolderTree(userID string) (string, error) {
	folders, err := s.folderService.GetFolderTree(userID, nil)
	if err != nil {
		return "", err
	}

	if len(folders) == 0 {
		return "No folders exist yet. Files are in the root folder.", nil
	}

	result := "Folder structure:\n"
	result += formatFolderTree(folders, 0)
	return result, nil
}

func (s *agentService) executeListFolders(userID string, args map[string]interface{}) (string, error) {
	opts := FolderListOptions{
		Limit:  100,
		Offset: 0,
	}

	if parentID, ok := args["parent_id"].(float64); ok {
		pid := uint(parentID)
		opts.ParentID = &pid
	}

	folders, _, err := s.folderService.ListFolders(userID, opts)
	if err != nil {
		return "", err
	}

	if len(folders) == 0 {
		if opts.ParentID == nil {
			return "No root folders exist.", nil
		}
		return "No subfolders in this folder.", nil
	}

	result := fmt.Sprintf("Found %d folder(s):\n", len(folders))
	for _, folder := range folders {
		result += fmt.Sprintf("- ID: %d, Name: %s", folder.ID, folder.Name)
		if folder.Description != "" {
			result += fmt.Sprintf(" (%s)", folder.Description)
		}
		result += "\n"
	}
	return result, nil
}

func (s *agentService) executeMoveFile(userID string, fileID uint, args map[string]interface{}) (string, error) {
	var targetFolderID *uint
	if folderID, ok := args["folder_id"].(float64); ok {
		fid := uint(folderID)
		targetFolderID = &fid
	}

	if err := s.fileService.MoveFiles(userID, []uint{fileID}, targetFolderID); err != nil {
		return "", err
	}

	if targetFolderID == nil {
		return "Moved file to root folder", nil
	}

	// Get folder name for better feedback
	folder, _ := s.folderService.GetFolderByID(userID, *targetFolderID)
	if folder != nil {
		return fmt.Sprintf("Moved file to folder '%s'", folder.Name), nil
	}
	return fmt.Sprintf("Moved file to folder ID %d", *targetFolderID), nil
}

func (s *agentService) executeGetFileInfo(userID string, fileID uint) (string, error) {
	file, err := s.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		return "", fmt.Errorf("failed to get file")
	}

	result := fmt.Sprintf(`File Information:
- ID: %d
- Title: %s
- Type: %s
- Folder: %s
- Tags: %s
- Processing Status: %s`,
		file.ID,
		file.Title,
		file.FileType,
		getFolderName(file),
		formatFileTags(file),
		file.ProcessingStatus,
	)

	return result, nil
}

func (s *agentService) executeCreateFolder(userID string, args map[string]interface{}) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name is required")
	}

	folder := &models.Folder{
		Name:        name,
		Description: getStringArg(args, "description", ""),
	}

	if parentID, ok := args["parent_id"].(float64); ok {
		pid := uint(parentID)
		folder.ParentID = &pid
	}

	if err := s.folderService.CreateFolder(userID, folder); err != nil {
		return "", err
	}

	return fmt.Sprintf("Created folder: ID=%d, Name='%s'", folder.ID, folder.Name), nil
}

// Helper functions

func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}

func getFolderName(file *models.File) string {
	if file.Folder != nil {
		return file.Folder.Name
	}
	if file.FolderID == nil {
		return "Root (no folder)"
	}
	return fmt.Sprintf("Folder ID %d", *file.FolderID)
}

func formatFileTags(file *models.File) string {
	if len(file.Tags) == 0 {
		return "None"
	}
	names := make([]string, len(file.Tags))
	for i, tag := range file.Tags {
		names[i] = tag.Name
	}
	return strings.Join(names, ", ")
}

func formatFolderTree(folders []models.Folder, depth int) string {
	result := ""
	indent := strings.Repeat("  ", depth)
	for _, folder := range folders {
		result += fmt.Sprintf("%s- ID: %d, Name: %s", indent, folder.ID, folder.Name)
		if folder.Description != "" {
			result += fmt.Sprintf(" (%s)", folder.Description)
		}
		result += "\n"
		if len(folder.Children) > 0 {
			result += formatFolderTree(folder.Children, depth+1)
		}
	}
	return result
}

func formatToolCallMessage(toolName, args string) string {
	switch toolName {
	case "search_tags":
		var a map[string]interface{}
		json.Unmarshal([]byte(args), &a)
		if keyword, ok := a["keyword"].(string); ok {
			return fmt.Sprintf("Searching tags for '%s'", keyword)
		}
		return "Searching tags"
	case "list_all_tags":
		return "Listing all tags"
	case "create_tag":
		var a map[string]interface{}
		json.Unmarshal([]byte(args), &a)
		if name, ok := a["name"].(string); ok {
			return fmt.Sprintf("Creating tag '%s'", name)
		}
		return "Creating new tag"
	case "add_tags_to_file":
		return "Adding tags to file"
	case "get_folder_tree":
		return "Getting folder structure"
	case "list_folders":
		return "Listing folders"
	case "move_file":
		var a map[string]interface{}
		json.Unmarshal([]byte(args), &a)
		if folderID, ok := a["folder_id"].(float64); ok {
			return fmt.Sprintf("Moving file to folder %d", int(folderID))
		}
		return "Moving file"
	case "get_file_info":
		return "Getting file information"
	case "create_folder":
		var a map[string]interface{}
		json.Unmarshal([]byte(args), &a)
		if name, ok := a["name"].(string); ok {
			return fmt.Sprintf("Creating folder '%s'", name)
		}
		return "Creating new folder"
	default:
		return toolName
	}
}

// MockAgentService is a mock implementation for testing
type MockAgentService struct {
	enabled bool
}

func NewMockAgentService() AgentService {
	return &MockAgentService{enabled: true}
}

func (m *MockAgentService) ProcessFileWithAgent(ctx context.Context, userID string, fileID uint,
	content, summary string, eventChan chan<- AgentEvent) error {
	eventChan <- AgentEvent{Type: "status", Message: "Mock agent processing...", FileID: fileID}
	eventChan <- AgentEvent{Type: "result", Message: "Mock organization complete", FileID: fileID}
	return nil
}

func (m *MockAgentService) OrganizeFile(ctx context.Context, userID string, fileID uint,
	eventChan chan<- AgentEvent) error {
	return m.ProcessFileWithAgent(ctx, userID, fileID, "", "", eventChan)
}

func (m *MockAgentService) IsEnabled() bool {
	return m.enabled
}
