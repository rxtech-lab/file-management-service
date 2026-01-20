package services

import (
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// FolderListOptions contains options for listing folders
type FolderListOptions struct {
	Keyword  string
	ParentID *uint // nil = root folders only, pointer to uint = specific parent
	TagIDs   []uint
	Limit    int
	Offset   int
}

// FolderService handles folder-related operations
type FolderService interface {
	CreateFolder(userID string, folder *models.Folder) error
	GetFolderByID(userID string, id uint) (*models.Folder, error)
	ListFolders(userID string, opts FolderListOptions) ([]models.Folder, int64, error)
	UpdateFolder(userID string, folder *models.Folder) error
	DeleteFolder(userID string, id uint) error
	MoveFolder(userID string, folderID uint, newParentID *uint) error
	GetFolderTree(userID string, parentID *uint) ([]models.Folder, error)
	AddTagsToFolder(userID string, folderID uint, tagIDs []uint) error
	RemoveTagsFromFolder(userID string, folderID uint, tagIDs []uint) error
	GetFolderPath(userID string, folderID uint) ([]models.Folder, error)
}

type folderService struct {
	db *gorm.DB
}

// NewFolderService creates a new FolderService
func NewFolderService(db *gorm.DB) FolderService {
	return &folderService{db: db}
}

// CreateFolder creates a new folder
func (s *folderService) CreateFolder(userID string, folder *models.Folder) error {
	folder.UserID = userID

	// Validate parent folder if specified
	if folder.ParentID != nil {
		parent, err := s.GetFolderByID(userID, *folder.ParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return errors.New("parent folder not found")
		}
	}

	return s.db.Create(folder).Error
}

// GetFolderByID retrieves a folder by ID with children and tags
func (s *folderService) GetFolderByID(userID string, id uint) (*models.Folder, error) {
	var folder models.Folder
	err := s.db.Preload("Tags").Preload("Children").
		Where("id = ? AND user_id = ?", id, userID).
		First(&folder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &folder, nil
}

// ListFolders lists folders with filtering options
func (s *folderService) ListFolders(userID string, opts FolderListOptions) ([]models.Folder, int64, error) {
	var folders []models.Folder
	var total int64

	query := s.db.Model(&models.Folder{}).Where("user_id = ?", userID)

	// Filter by parent ID
	if opts.ParentID != nil {
		query = query.Where("parent_id = ?", *opts.ParentID)
	} else {
		// Root folders only (no parent)
		query = query.Where("parent_id IS NULL")
	}

	// Keyword search
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// Filter by tags
	if len(opts.TagIDs) > 0 {
		query = query.Joins("JOIN folder_tags ON folder_tags.folder_id = folders.id").
			Where("folder_tags.tag_id IN ?", opts.TagIDs).
			Group("folders.id")
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Reset query for actual results with preloading
	query = s.db.Model(&models.Folder{}).Where("user_id = ?", userID)
	if opts.ParentID != nil {
		query = query.Where("parent_id = ?", *opts.ParentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}
	if len(opts.TagIDs) > 0 {
		query = query.Joins("JOIN folder_tags ON folder_tags.folder_id = folders.id").
			Where("folder_tags.tag_id IN ?", opts.TagIDs).
			Group("folders.id")
	}

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	if err := query.Preload("Tags").Order("name ASC").Find(&folders).Error; err != nil {
		return nil, 0, err
	}

	return folders, total, nil
}

// UpdateFolder updates a folder
func (s *folderService) UpdateFolder(userID string, folder *models.Folder) error {
	// Verify ownership
	existing, err := s.GetFolderByID(userID, folder.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("folder not found")
	}

	// Update only allowed fields
	updates := map[string]interface{}{
		"name":        folder.Name,
		"description": folder.Description,
	}

	return s.db.Model(&models.Folder{}).Where("id = ? AND user_id = ?", folder.ID, userID).Updates(updates).Error
}

// DeleteFolder deletes a folder and all its contents recursively
func (s *folderService) DeleteFolder(userID string, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Get folder to verify ownership
		var folder models.Folder
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&folder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("folder not found")
			}
			return err
		}

		// Recursively delete children folders
		var children []models.Folder
		if err := tx.Where("parent_id = ? AND user_id = ?", id, userID).Find(&children).Error; err != nil {
			return err
		}
		for _, child := range children {
			if err := s.deleteFolderRecursive(tx, userID, child.ID); err != nil {
				return err
			}
		}

		// Delete files in this folder
		if err := tx.Where("folder_id = ? AND user_id = ?", id, userID).Delete(&models.File{}).Error; err != nil {
			return err
		}

		// Clear folder_tags associations
		if err := tx.Exec("DELETE FROM folder_tags WHERE folder_id = ?", id).Error; err != nil {
			return err
		}

		// Delete the folder
		return tx.Delete(&folder).Error
	})
}

// deleteFolderRecursive is a helper for recursive folder deletion
func (s *folderService) deleteFolderRecursive(tx *gorm.DB, userID string, folderID uint) error {
	// Get children
	var children []models.Folder
	if err := tx.Where("parent_id = ? AND user_id = ?", folderID, userID).Find(&children).Error; err != nil {
		return err
	}

	// Delete children recursively
	for _, child := range children {
		if err := s.deleteFolderRecursive(tx, userID, child.ID); err != nil {
			return err
		}
	}

	// Delete files in this folder
	if err := tx.Where("folder_id = ? AND user_id = ?", folderID, userID).Delete(&models.File{}).Error; err != nil {
		return err
	}

	// Clear folder_tags associations
	if err := tx.Exec("DELETE FROM folder_tags WHERE folder_id = ?", folderID).Error; err != nil {
		return err
	}

	// Delete the folder
	return tx.Where("id = ? AND user_id = ?", folderID, userID).Delete(&models.Folder{}).Error
}

// MoveFolder moves a folder to a new parent
func (s *folderService) MoveFolder(userID string, folderID uint, newParentID *uint) error {
	// Verify the folder exists and belongs to user
	folder, err := s.GetFolderByID(userID, folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		return errors.New("folder not found")
	}

	// Verify new parent exists if specified
	if newParentID != nil {
		parent, err := s.GetFolderByID(userID, *newParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return errors.New("target parent folder not found")
		}

		// Prevent moving a folder into itself or its descendants
		if *newParentID == folderID {
			return errors.New("cannot move a folder into itself")
		}
		if s.isDescendant(userID, *newParentID, folderID) {
			return errors.New("cannot move a folder into its descendant")
		}
	}

	return s.db.Model(&models.Folder{}).
		Where("id = ? AND user_id = ?", folderID, userID).
		Update("parent_id", newParentID).Error
}

// isDescendant checks if potentialDescendant is a descendant of folderID
func (s *folderService) isDescendant(userID string, potentialDescendant, folderID uint) bool {
	var folder models.Folder
	if err := s.db.Where("id = ? AND user_id = ?", potentialDescendant, userID).First(&folder).Error; err != nil {
		return false
	}

	if folder.ParentID == nil {
		return false
	}

	if *folder.ParentID == folderID {
		return true
	}

	return s.isDescendant(userID, *folder.ParentID, folderID)
}

// GetFolderTree gets the folder tree structure starting from a parent
func (s *folderService) GetFolderTree(userID string, parentID *uint) ([]models.Folder, error) {
	var folders []models.Folder

	query := s.db.Where("user_id = ?", userID)
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if err := query.Preload("Tags").Order("name ASC").Find(&folders).Error; err != nil {
		return nil, err
	}

	// Load children recursively
	for i := range folders {
		children, err := s.GetFolderTree(userID, &folders[i].ID)
		if err != nil {
			return nil, err
		}
		folders[i].Children = children
	}

	return folders, nil
}

// AddTagsToFolder adds tags to a folder
func (s *folderService) AddTagsToFolder(userID string, folderID uint, tagIDs []uint) error {
	// Verify folder exists
	folder, err := s.GetFolderByID(userID, folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		return errors.New("folder not found")
	}

	// Get tags and verify they belong to user
	var tags []models.Tag
	if err := s.db.Where("id IN ? AND user_id = ?", tagIDs, userID).Find(&tags).Error; err != nil {
		return err
	}

	// Add tags using association
	return s.db.Model(folder).Association("Tags").Append(tags)
}

// RemoveTagsFromFolder removes tags from a folder
func (s *folderService) RemoveTagsFromFolder(userID string, folderID uint, tagIDs []uint) error {
	// Verify folder exists
	folder, err := s.GetFolderByID(userID, folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		return errors.New("folder not found")
	}

	// Get tags
	var tags []models.Tag
	if err := s.db.Where("id IN ? AND user_id = ?", tagIDs, userID).Find(&tags).Error; err != nil {
		return err
	}

	// Remove tags using association
	return s.db.Model(folder).Association("Tags").Delete(tags)
}

// GetFolderPath returns the path from root to the specified folder
func (s *folderService) GetFolderPath(userID string, folderID uint) ([]models.Folder, error) {
	var path []models.Folder
	currentID := &folderID

	for currentID != nil {
		var folder models.Folder
		if err := s.db.Where("id = ? AND user_id = ?", *currentID, userID).First(&folder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, err
		}
		path = append([]models.Folder{folder}, path...)
		currentID = folder.ParentID
	}

	return path, nil
}
