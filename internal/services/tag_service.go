package services

import (
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// TagService handles tag-related operations
type TagService interface {
	CreateTag(userID string, tag *models.Tag) error
	GetTagByID(userID string, id uint) (*models.Tag, error)
	ListTags(userID string, keyword string, limit, offset int) ([]models.Tag, int64, error)
	UpdateTag(userID string, tag *models.Tag) error
	DeleteTag(userID string, id uint) error
	GetTagsByIDs(userID string, ids []uint) ([]models.Tag, error)
}

type tagService struct {
	db *gorm.DB
}

// NewTagService creates a new TagService
func NewTagService(db *gorm.DB) TagService {
	return &tagService{db: db}
}

// CreateTag creates a new tag
func (s *tagService) CreateTag(userID string, tag *models.Tag) error {
	tag.UserID = userID
	return s.db.Create(tag).Error
}

// GetTagByID retrieves a tag by ID for a specific user
func (s *tagService) GetTagByID(userID string, id uint) (*models.Tag, error) {
	var tag models.Tag
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// ListTags lists all tags for a user with optional keyword search
func (s *tagService) ListTags(userID string, keyword string, limit, offset int) ([]models.Tag, int64, error) {
	var tags []models.Tag
	var total int64

	query := s.db.Model(&models.Tag{}).Where("user_id = ?", userID)

	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("created_at DESC").Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// UpdateTag updates a tag
func (s *tagService) UpdateTag(userID string, tag *models.Tag) error {
	// Verify ownership
	existing, err := s.GetTagByID(userID, tag.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("tag not found")
	}

	// Update only allowed fields
	updates := map[string]interface{}{
		"name":        tag.Name,
		"color":       tag.Color,
		"description": tag.Description,
	}

	return s.db.Model(&models.Tag{}).Where("id = ? AND user_id = ?", tag.ID, userID).Updates(updates).Error
}

// DeleteTag deletes a tag
func (s *tagService) DeleteTag(userID string, id uint) error {
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("tag not found")
	}
	return nil
}

// GetTagsByIDs retrieves multiple tags by their IDs
func (s *tagService) GetTagsByIDs(userID string, ids []uint) ([]models.Tag, error) {
	var tags []models.Tag
	if len(ids) == 0 {
		return tags, nil
	}
	err := s.db.Where("id IN ? AND user_id = ?", ids, userID).Find(&tags).Error
	return tags, err
}
