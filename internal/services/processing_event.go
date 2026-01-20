package services

// ProcessingEvent represents a real-time status update during file processing
// This unified event type can represent events from different sources (content parsing, invoice, agent)
type ProcessingEvent struct {
	Type    string      `json:"type"`              // "status", "tool_call", "tool_result", "thinking", "result", "error", "invoice", "complete"
	Source  string      `json:"source"`            // "system", "invoice", "agent" - identifies which service emitted the event
	Message string      `json:"message"`           // Human-readable status message
	Data    interface{} `json:"data,omitempty"`    // Optional additional data
	Tool    string      `json:"tool,omitempty"`    // Tool name if type is tool_call
	FileID  uint        `json:"file_id,omitempty"` // File ID being processed
}

// NewProcessingEvent creates a new processing event
func NewProcessingEvent(source, eventType, message string, fileID uint) ProcessingEvent {
	return ProcessingEvent{
		Type:    eventType,
		Source:  source,
		Message: message,
		FileID:  fileID,
	}
}

// FromAgentEvent converts an AgentEvent to ProcessingEvent
func FromAgentEvent(event AgentEvent) ProcessingEvent {
	return ProcessingEvent{
		Type:    event.Type,
		Source:  "agent",
		Message: event.Message,
		Data:    event.Data,
		Tool:    event.Tool,
		FileID:  event.FileID,
	}
}

// FromInvoiceEvent converts an InvoiceStreamEvent to ProcessingEvent
func FromInvoiceEvent(event InvoiceStreamEvent, fileID uint) ProcessingEvent {
	return ProcessingEvent{
		Type:    "invoice",
		Source:  "invoice",
		Message: event.Message,
		Data: map[string]interface{}{
			"status":     event.Status,
			"tool_name":  event.ToolName,
			"invoice_id": event.InvoiceID,
		},
		Tool:   event.ToolName,
		FileID: fileID,
	}
}
