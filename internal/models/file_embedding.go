package models

// FileEmbedding stores vector embeddings for files
// Note: For Turso, the actual embedding column uses F32_BLOB type
// which requires raw SQL for creation and vector operations
type FileEmbedding struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	FileID uint   `gorm:"uniqueIndex;not null" json:"file_id"`
	UserID string `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	// Embedding is stored as JSON text for GORM compatibility
	// For Turso vector operations, use raw SQL with F32_BLOB
	Embedding string `gorm:"type:text" json:"embedding"`
}

// TableName specifies the table name for FileEmbedding
func (FileEmbedding) TableName() string {
	return "file_embeddings"
}
