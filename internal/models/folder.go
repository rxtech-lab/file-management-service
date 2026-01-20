package models

import (
	"time"

	"gorm.io/gorm"
)

// Folder represents a folder in the file management system
// Supports tree structure via self-referential ParentID
type Folder struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name        string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *uint          `gorm:"index" json:"parent_id"` // nil = root folder
	Parent      *Folder        `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Folder       `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Tags        []Tag          `gorm:"many2many:folder_tags" json:"tags,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Folder
func (Folder) TableName() string {
	return "folders"
}
