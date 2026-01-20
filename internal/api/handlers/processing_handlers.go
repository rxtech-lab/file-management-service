package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// ProcessingHandlers handles unified file processing with SSE streaming
type ProcessingHandlers struct {
	fileService          services.FileService
	uploadService        services.UploadService
	contentParserService services.ContentParserService
	embeddingService     services.EmbeddingService
	summaryService       services.SummaryService
	agentService         services.AgentService
	invoiceService       services.InvoiceService
}

// NewProcessingHandlers creates a new ProcessingHandlers instance
func NewProcessingHandlers(
	fileService services.FileService,
	uploadService services.UploadService,
	contentParserService services.ContentParserService,
	embeddingService services.EmbeddingService,
	summaryService services.SummaryService,
	agentService services.AgentService,
	invoiceService services.InvoiceService,
) *ProcessingHandlers {
	return &ProcessingHandlers{
		fileService:          fileService,
		uploadService:        uploadService,
		contentParserService: contentParserService,
		embeddingService:     embeddingService,
		summaryService:       summaryService,
		agentService:         agentService,
		invoiceService:       invoiceService,
	}
}

// StreamFileProcessing handles SSE streaming of file processing
// GET /api/files/:id/process-stream
func (h *ProcessingHandlers) StreamFileProcessing(c *fiber.Ctx) error {
	// Get authenticated user
	user := c.Locals(middleware.AuthenticatedUserContextKey)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	authenticatedUser := user.(*utils.AuthenticatedUser)
	userID := authenticatedUser.Sub

	// Get auth token for invoice processing
	authToken := ""
	if token, ok := c.Locals(middleware.RawAuthTokenContextKey).(string); ok {
		authToken = token
	}

	// Parse file ID
	fileIDStr := c.Params("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid file ID"})
	}

	// Get file to verify ownership
	file, err := h.fileService.GetFileByID(userID, uint(fileID))
	if err != nil || file == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "File not found"})
	}

	// Check if already processing
	if file.ProcessingStatus == models.FileStatusProcessing {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "File is already being processed"})
	}

	// Update status to processing
	if err := h.fileService.UpdateFileProcessingStatus(userID, file.ID, models.FileStatusProcessing, ""); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

	// Create event channel
	eventChan := make(chan services.ProcessingEvent, 100)

	// Run processing in goroutine
	go func() {
		defer close(eventChan)
		h.processFileWithEvents(ctx, userID, uint(fileID), authToken, eventChan)
	}()

	// Stream events to client
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()

		// Send initial connection event
		sendEvent(w, services.NewProcessingEvent("system", "connected", "Connected to processing stream", uint(fileID)))

		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					// Channel closed, processing complete
					sendEvent(w, services.NewProcessingEvent("system", "done", "Processing complete", uint(fileID)))
					return
				}
				sendEvent(w, event)

				// If error, we're done
				if event.Type == "error" {
					sendEvent(w, services.NewProcessingEvent("system", "done", "Processing finished with error", uint(fileID)))
					return
				}

			case <-ctx.Done():
				sendEvent(w, services.NewProcessingEvent("system", "error", "Request timeout", uint(fileID)))
				return
			}
		}
	})

	return nil
}

// sendEvent writes a single SSE event
func sendEvent(w *bufio.Writer, event services.ProcessingEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
}

// processFileWithEvents processes a file and emits events to the channel
func (h *ProcessingHandlers) processFileWithEvents(ctx context.Context, userID string, fileID uint, authToken string, eventChan chan<- services.ProcessingEvent) {
	var wg sync.WaitGroup

	emit := func(source, eventType, message string) {
		select {
		case eventChan <- services.NewProcessingEvent(source, eventType, message, fileID):
		default:
			// Channel full, skip
		}
	}

	// Get file
	emit("system", "status", "Loading file information...")
	file, err := h.fileService.GetFileByID(userID, fileID)
	if err != nil || file == nil {
		emit("system", "error", "Failed to get file")
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to get file")
		return
	}

	// Get presigned download URL
	emit("system", "status", "Getting download URL...")
	downloadURL, err := h.uploadService.GetPresignedDownloadURL(ctx, file.S3Key)
	if err != nil {
		emit("system", "error", "Failed to get download URL: "+err.Error())
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to get download URL: "+err.Error())
		return
	}

	// Parse content
	emit("system", "status", "Parsing file content...")
	parsedContent, err := h.contentParserService.ParseFileContent(ctx, downloadURL)
	if err != nil {
		emit("system", "error", "Failed to parse content: "+err.Error())
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to parse content: "+err.Error())
		return
	}
	emit("system", "status", "Content parsed successfully")

	// Generate summary
	emit("system", "status", "Generating summary...")
	summary, err := h.summaryService.GenerateSummary(ctx, parsedContent.TextContent, 500)
	if err != nil {
		summary = services.GenerateSummary(parsedContent.TextContent, 500)
	}
	emit("system", "status", "Summary generated")

	// Detect file type
	detectedFileType := file.FileType
	if models.IsInvoiceContent(parsedContent.TextContent) {
		detectedFileType = models.FileTypeInvoice
		emit("system", "status", "Detected file as invoice")
	}

	// Update file with parsed content
	emit("system", "status", "Saving file content...")
	if err := h.fileService.UpdateFileContent(userID, fileID, parsedContent.TextContent, summary, detectedFileType); err != nil {
		emit("system", "error", "Failed to update content: "+err.Error())
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusFailed, "Failed to update content: "+err.Error())
		return
	}

	// Process invoice if detected
	log.Printf("[Invoice] File %d: type=%s, invoiceService=%v, enabled=%v, hasToken=%v",
		fileID, detectedFileType, h.invoiceService != nil,
		h.invoiceService != nil && h.invoiceService.IsEnabled(), authToken != "")

	if detectedFileType == models.FileTypeInvoice && h.invoiceService != nil && h.invoiceService.IsEnabled() && authToken != "" {
		emit("invoice", "status", "Starting invoice processing...")
		log.Printf("[Invoice] Processing file %d as invoice", fileID)

		// Create channel for invoice events
		invoiceEventChan := make(chan services.InvoiceStreamEvent, 100)

		// Forward invoice events to main channel
		wg.Add(1)
		go func() {
			defer wg.Done()
			for invoiceEvent := range invoiceEventChan {
				select {
				case eventChan <- services.FromInvoiceEvent(invoiceEvent, fileID):
				default:
				}
			}
		}()

		result, err := h.invoiceService.ProcessInvoice(ctx, downloadURL, authToken, invoiceEventChan)
		// Note: invoiceService.ProcessInvoice closes invoiceEventChan via defer
		if err != nil {
			log.Printf("[Invoice] File %d processing warning: %v", fileID, err)
			emit("invoice", "error", "Invoice processing failed: "+err.Error())
		} else if result != nil {
			if err := h.fileService.UpdateFileInvoiceID(userID, fileID, result.InvoiceID); err != nil {
				log.Printf("[Invoice] File %d: failed to store invoice_id: %v", fileID, err)
			} else {
				log.Printf("[Invoice] File %d: stored invoice_id=%d", fileID, result.InvoiceID)
				emit("invoice", "complete", fmt.Sprintf("Invoice created with ID: %d", result.InvoiceID))
			}
		}
	}

	// Run AI agent
	if h.agentService != nil && h.agentService.IsEnabled() {
		emit("agent", "status", "Starting AI organization...")

		// Create channel for agent events
		agentEventChan := make(chan services.AgentEvent, 100)

		// Forward agent events to main channel
		wg.Add(1)
		go func() {
			defer wg.Done()
			for agentEvent := range agentEventChan {
				select {
				case eventChan <- services.FromAgentEvent(agentEvent):
				default:
				}
			}
		}()

		err := h.agentService.ProcessFileWithAgent(ctx, userID, fileID, parsedContent.TextContent, summary, agentEventChan)
		close(agentEventChan) // Ensure channel is closed so forwarding goroutine can exit
		if err != nil {
			log.Printf("[Agent] File %d processing warning: %v", fileID, err)
			emit("agent", "error", "Agent processing warning: "+err.Error())
		}
	}

	// Generate embedding
	emit("system", "status", "Generating embedding...")
	embedding, err := h.embeddingService.GenerateEmbedding(ctx, parsedContent.TextContent)
	if err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "Embedding generation failed: "+err.Error())
		emit("system", "status", "Processing complete (embedding failed)")
		return
	}

	// Store embedding
	emit("system", "status", "Storing embedding...")
	if err := h.embeddingService.StoreFileEmbedding(userID, fileID, embedding); err != nil {
		h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "Embedding storage failed: "+err.Error())
		emit("system", "status", "Processing complete (embedding storage failed)")
		return
	}

	// Mark as completed
	h.fileService.UpdateFileProcessingStatus(userID, fileID, models.FileStatusCompleted, "")
	emit("system", "complete", "File processing completed successfully")

	// Wait for all forwarding goroutines to complete before returning
	// This ensures eventChan can be safely closed by the caller
	wg.Wait()
}
