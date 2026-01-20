package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// FileType represents the type of file
type FileType string

const (
	FileTypeMusic    FileType = "music"
	FileTypePhoto    FileType = "photo"
	FileTypeVideo    FileType = "video"
	FileTypeDocument FileType = "document"
	FileTypeInvoice  FileType = "invoice" // Special: use for any invoice-like docs
)

// FileProcessingStatus represents the processing status of a file
type FileProcessingStatus string

const (
	FileStatusPending    FileProcessingStatus = "pending"
	FileStatusProcessing FileProcessingStatus = "processing"
	FileStatusCompleted  FileProcessingStatus = "completed"
	FileStatusFailed     FileProcessingStatus = "failed"
)

// File represents a file in the file management system
type File struct {
	ID               uint                 `gorm:"primaryKey" json:"id"`
	UserID           string               `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Title            string               `gorm:"not null;type:varchar(255)" json:"title"`
	Summary          string               `gorm:"type:text" json:"summary"`
	Content          string               `gorm:"type:text" json:"content"` // Parsed text content
	FileType         FileType             `gorm:"type:varchar(20);default:'document'" json:"file_type"`
	FolderID         *uint                `gorm:"index" json:"folder_id"`
	Folder           *Folder              `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Tags             []Tag                `gorm:"many2many:file_tags" json:"tags,omitempty"`
	S3Key            string               `gorm:"uniqueIndex;not null" json:"s3_key"`
	OriginalFilename string               `gorm:"not null;type:varchar(255)" json:"original_filename"`
	MimeType         string               `gorm:"type:varchar(255)" json:"mime_type"`
	Size             int64                `json:"size"`
	ProcessingStatus FileProcessingStatus `gorm:"type:varchar(20);default:'pending'" json:"processing_status"`
	ProcessingError  string               `gorm:"type:text" json:"processing_error,omitempty"`
	HasEmbedding     bool                 `gorm:"default:false" json:"has_embedding"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	DeletedAt        gorm.DeletedAt       `gorm:"index" json:"-"`
}

// TableName specifies the table name for File
func (File) TableName() string {
	return "files"
}

// DetectFileTypeFromMimeType returns the initial file type based on MIME type
func DetectFileTypeFromMimeType(mimeType string) FileType {
	mimeType = strings.ToLower(mimeType)

	// Music types
	if strings.HasPrefix(mimeType, "audio/") {
		return FileTypeMusic
	}

	// Photo types
	if strings.HasPrefix(mimeType, "image/") {
		return FileTypePhoto
	}

	// Video types
	if strings.HasPrefix(mimeType, "video/") {
		return FileTypeVideo
	}

	// Default to document for everything else
	return FileTypeDocument
}

// IsInvoiceContent checks if the content appears to be an invoice
// Returns true if invoice-related keywords or patterns are detected
func IsInvoiceContent(content string) bool {
	if content == "" {
		return false
	}

	content = strings.ToLower(content)

	// Invoice-related keywords
	keywords := []string{
		"invoice",
		"bill",
		"receipt",
		"payment due",
		"amount due",
		"total amount",
		"subtotal",
		"tax",
		"invoice number",
		"invoice #",
		"inv #",
		"bill to",
		"sold to",
		"remit to",
		"payment terms",
		"due date",
		"balance due",
	}

	// Count matches
	matchCount := 0
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			matchCount++
		}
	}

	// If 2 or more keywords match, consider it an invoice
	return matchCount >= 2
}
