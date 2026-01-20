package models

import (
	"time"

	"gorm.io/gorm"
)

// Tag represents a tag that can be attached to files and folders
type Tag struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      string         `gorm:"index;not null;type:varchar(255)" json:"user_id"`
	Name        string         `gorm:"not null;type:varchar(255)" json:"name"`
	Color       string         `gorm:"type:varchar(7)" json:"color"` // Hex color code (e.g., #FF5733)
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}
