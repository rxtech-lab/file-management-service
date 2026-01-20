package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// AgentHandlers handles AI agent-related HTTP endpoints
type AgentHandlers struct {
	agentService services.AgentService
	fileService  services.FileService
}

// NewAgentHandlers creates a new AgentHandlers instance
func NewAgentHandlers(agentService services.AgentService, fileService services.FileService) *AgentHandlers {
	return &AgentHandlers{
		agentService: agentService,
		fileService:  fileService,
	}
}

// StreamAgentProgress handles SSE streaming of agent progress
// GET /api/files/:id/agent-stream
func (h *AgentHandlers) StreamAgentProgress(c *fiber.Ctx) error {
	// Get authenticated user
	user := c.Locals(middleware.AuthenticatedUserContextKey)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	authenticatedUser := user.(*utils.AuthenticatedUser)
	userID := authenticatedUser.Sub

	// Parse file ID
	fileIDStr := c.Params("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Check if agent is enabled
	if h.agentService == nil || !h.agentService.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI agent is not enabled"})
	}

	// Get file to verify ownership and get content
	file, err := h.fileService.GetFileByID(userID, uint(fileID))
	if err != nil || file == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create event channel
	eventChan := make(chan services.AgentEvent, 100)

	// Run agent in goroutine
	go func() {
		defer close(eventChan)
		err := h.agentService.OrganizeFile(ctx, userID, uint(fileID), eventChan)
		if err != nil {
			eventChan <- services.AgentEvent{
				Type:    "error",
				Message: fmt.Sprintf("Agent error: %v", err),
				FileID:  uint(fileID),
			}
		}
	}()

	// Stream events to client
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Send initial connection event
		data, _ := json.Marshal(services.AgentEvent{
			Type:    "connected",
			Message: "Connected to agent stream",
			FileID:  uint(fileID),
		})
		fmt.Fprintf(w, "data: %s\n\n", data)
		w.Flush()

		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					// Channel closed, send done event
					data, _ := json.Marshal(services.AgentEvent{
						Type:    "done",
						Message: "Agent processing complete",
						FileID:  uint(fileID),
					})
					fmt.Fprintf(w, "data: %s\n\n", data)
					w.Flush()
					return
				}
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.Flush()

				// If this is a result or error, we're done
				if event.Type == "result" || event.Type == "error" {
					// Send done event
					doneData, _ := json.Marshal(services.AgentEvent{
						Type:    "done",
						Message: "Stream complete",
						FileID:  uint(fileID),
					})
					fmt.Fprintf(w, "data: %s\n\n", doneData)
					w.Flush()
					return
				}

			case <-ctx.Done():
				// Context cancelled/timeout
				data, _ := json.Marshal(services.AgentEvent{
					Type:    "error",
					Message: "Request timeout",
					FileID:  uint(fileID),
				})
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.Flush()
				return
			}
		}
	})

	return nil
}

// TriggerAgentOrganize triggers the AI agent to organize a file
// POST /api/files/:id/organize
func (h *AgentHandlers) TriggerAgentOrganize(c *fiber.Ctx) error {
	// Get authenticated user
	user := c.Locals(middleware.AuthenticatedUserContextKey)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	authenticatedUser := user.(*utils.AuthenticatedUser)
	userID := authenticatedUser.Sub

	// Parse file ID
	fileIDStr := c.Params("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Check if agent is enabled
	if h.agentService == nil || !h.agentService.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "AI agent is not enabled"})
	}

	// Verify file ownership
	file, err := h.fileService.GetFileByID(userID, uint(fileID))
	if err != nil || file == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// Return the stream URL for the client to subscribe to
	return c.JSON(fiber.Map{
		"message":    "Agent organization started",
		"file_id":    fileID,
		"stream_url": fmt.Sprintf("/api/files/%d/agent-stream", fileID),
	})
}

// GetAgentStatus returns the current status of the agent service
// GET /api/agent/status
func (h *AgentHandlers) GetAgentStatus(c *fiber.Ctx) error {
	enabled := h.agentService != nil && h.agentService.IsEnabled()
	return c.JSON(fiber.Map{
		"enabled": enabled,
	})
}
