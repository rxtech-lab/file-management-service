package services

import (
	"errors"

	"github.com/rxtech-lab/invoice-management/internal/models"
	"gorm.io/gorm"
)

// FileListOptions contains options for listing files
type FileListOptions struct {
	Keyword    string
	FolderID   *uint
	AllFolders bool // When true, search across all folders (ignores FolderID)
	TagIDs     []uint
	FileTypes  []models.FileType
	Status     *models.FileProcessingStatus
	SortBy     string // "created_at", "title", "size", "updated_at"
	SortOrder  string // "asc", "desc"
	Limit      int
	Offset     int
}

// FileService handles file-related operations
type FileService interface {
	// CRUD operations
	CreateFile(userID string, file *models.File) error
	GetFileByID(userID string, id uint) (*models.File, error)
	GetFileByS3Key(userID string, s3Key string) (*models.File, error)
	ListFiles(userID string, opts FileListOptions) ([]models.File, int64, error)
	UpdateFile(userID string, file *models.File) error
	DeleteFile(userID string, id uint) error

	// Move operations
	MoveFiles(userID string, fileIDs []uint, targetFolderID *uint) error

	// Tag operations
	AddTagsToFile(userID string, fileID uint, tagIDs []uint) error
	RemoveTagsFromFile(userID string, fileID uint, tagIDs []uint) error

	// Content operations
	UpdateFileContent(userID string, fileID uint, content, summary string, fileType models.FileType) error
	UpdateFileProcessingStatus(userID string, fileID uint, status models.FileProcessingStatus, errMsg string) error
	SetFileHasEmbedding(userID string, fileID uint, hasEmbedding bool) error

	// Folder operations
	GetFilesInFolderRecursive(userID string, folderID uint) ([]models.File, error)
}

type fileService struct {
	db *gorm.DB
}

// NewFileService creates a new FileService
func NewFileService(db *gorm.DB) FileService {
	return &fileService{db: db}
}

// CreateFile creates a new file record
func (s *fileService) CreateFile(userID string, file *models.File) error {
	file.UserID = userID

	// Set initial file type from MIME type if not already set
	if file.FileType == "" {
		file.FileType = models.DetectFileTypeFromMimeType(file.MimeType)
	}

	// Validate folder if specified
	if file.FolderID != nil {
		var folder models.Folder
		if err := s.db.Where("id = ? AND user_id = ?", *file.FolderID, userID).First(&folder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("folder not found")
			}
			return err
		}
	}

	return s.db.Create(file).Error
}

// GetFileByID retrieves a file by ID with folder and tags
func (s *fileService) GetFileByID(userID string, id uint) (*models.File, error) {
	var file models.File
	err := s.db.Preload("Tags").Preload("Folder").
		Where("id = ? AND user_id = ?", id, userID).
		First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// GetFileByS3Key retrieves a file by its S3 key
func (s *fileService) GetFileByS3Key(userID string, s3Key string) (*models.File, error) {
	var file models.File
	err := s.db.Preload("Tags").Preload("Folder").
		Where("s3_key = ? AND user_id = ?", s3Key, userID).
		First(&file).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &file, nil
}

// ListFiles lists files with filtering options
func (s *fileService) ListFiles(userID string, opts FileListOptions) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := s.db.Model(&models.File{}).Where("user_id = ?", userID)

	// Filter by folder (skip if AllFolders is true)
	if !opts.AllFolders {
		if opts.FolderID != nil {
			query = query.Where("folder_id = ?", *opts.FolderID)
		} else {
			query = query.Where("folder_id IS NULL")
		}
	}

	// Keyword search
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR content LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Filter by file types
	if len(opts.FileTypes) > 0 {
		query = query.Where("file_type IN ?", opts.FileTypes)
	}

	// Filter by processing status
	if opts.Status != nil {
		query = query.Where("processing_status = ?", *opts.Status)
	}

	// Filter by tags
	if len(opts.TagIDs) > 0 {
		query = query.Joins("JOIN file_tags ON file_tags.file_id = files.id").
			Where("file_tags.tag_id IN ?", opts.TagIDs).
			Group("files.id")
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build new query for results with preloading
	query = s.db.Model(&models.File{}).Where("user_id = ?", userID)
	if !opts.AllFolders {
		if opts.FolderID != nil {
			query = query.Where("folder_id = ?", *opts.FolderID)
		} else {
			query = query.Where("folder_id IS NULL")
		}
	}
	if opts.Keyword != "" {
		searchPattern := "%" + opts.Keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ? OR content LIKE ?", searchPattern, searchPattern, searchPattern)
	}
	if len(opts.FileTypes) > 0 {
		query = query.Where("file_type IN ?", opts.FileTypes)
	}
	if opts.Status != nil {
		query = query.Where("processing_status = ?", *opts.Status)
	}
	if len(opts.TagIDs) > 0 {
		query = query.Joins("JOIN file_tags ON file_tags.file_id = files.id").
			Where("file_tags.tag_id IN ?", opts.TagIDs).
			Group("files.id")
	}

	// Sorting
	sortBy := "created_at"
	if opts.SortBy != "" {
		switch opts.SortBy {
		case "title", "size", "created_at", "updated_at":
			sortBy = opts.SortBy
		}
	}
	sortOrder := "DESC"
	if opts.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	if err := query.Preload("Tags").Preload("Folder").Order(sortBy + " " + sortOrder).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// UpdateFile updates a file's metadata
func (s *fileService) UpdateFile(userID string, file *models.File) error {
	// Verify ownership
	existing, err := s.GetFileByID(userID, file.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("file not found")
	}

	// Update only allowed fields
	updates := map[string]any{
		"title":     file.Title,
		"summary":   file.Summary,
		"file_type": file.FileType,
	}

	// Only update folder_id if provided
	if file.FolderID != nil {
		// Validate the folder exists
		var folder models.Folder
		if err := s.db.Where("id = ? AND user_id = ?", *file.FolderID, userID).First(&folder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("folder not found")
			}
			return err
		}
		updates["folder_id"] = file.FolderID
	}

	return s.db.Model(&models.File{}).Where("id = ? AND user_id = ?", file.ID, userID).Updates(updates).Error
}

// DeleteFile deletes a file
func (s *fileService) DeleteFile(userID string, id uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Verify ownership
		var file models.File
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&file).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("file not found")
			}
			return err
		}

		// Clear file_tags associations
		if err := tx.Exec("DELETE FROM file_tags WHERE file_id = ?", id).Error; err != nil {
			return err
		}

		// Delete file embedding if exists
		if err := tx.Where("file_id = ?", id).Delete(&models.FileEmbedding{}).Error; err != nil {
			return err
		}

		// Delete the file
		return tx.Delete(&file).Error
	})
}

// MoveFiles moves multiple files to a target folder
func (s *fileService) MoveFiles(userID string, fileIDs []uint, targetFolderID *uint) error {
	// Validate target folder if specified
	if targetFolderID != nil {
		var folder models.Folder
		if err := s.db.Where("id = ? AND user_id = ?", *targetFolderID, userID).First(&folder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("target folder not found")
			}
			return err
		}
	}

	// Update files
	return s.db.Model(&models.File{}).
		Where("id IN ? AND user_id = ?", fileIDs, userID).
		Update("folder_id", targetFolderID).Error
}

// AddTagsToFile adds tags to a file
func (s *fileService) AddTagsToFile(userID string, fileID uint, tagIDs []uint) error {
	// Verify file exists
	file, err := s.GetFileByID(userID, fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	// Get tags and verify they belong to user
	var tags []models.Tag
	if err := s.db.Where("id IN ? AND user_id = ?", tagIDs, userID).Find(&tags).Error; err != nil {
		return err
	}

	// Add tags using association
	return s.db.Model(file).Association("Tags").Append(tags)
}

// RemoveTagsFromFile removes tags from a file
func (s *fileService) RemoveTagsFromFile(userID string, fileID uint, tagIDs []uint) error {
	// Verify file exists
	file, err := s.GetFileByID(userID, fileID)
	if err != nil {
		return err
	}
	if file == nil {
		return errors.New("file not found")
	}

	// Get tags
	var tags []models.Tag
	if err := s.db.Where("id IN ? AND user_id = ?", tagIDs, userID).Find(&tags).Error; err != nil {
		return err
	}

	// Remove tags using association
	return s.db.Model(file).Association("Tags").Delete(tags)
}

// UpdateFileContent updates a file's parsed content, summary, and file type
func (s *fileService) UpdateFileContent(userID string, fileID uint, content, summary string, fileType models.FileType) error {
	// Check if content looks like an invoice
	if fileType == models.FileTypeDocument && models.IsInvoiceContent(content) {
		fileType = models.FileTypeInvoice
	}

	updates := map[string]any{
		"content":   content,
		"summary":   summary,
		"file_type": fileType,
	}

	result := s.db.Model(&models.File{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found")
	}
	return nil
}

// UpdateFileProcessingStatus updates a file's processing status
func (s *fileService) UpdateFileProcessingStatus(userID string, fileID uint, status models.FileProcessingStatus, errMsg string) error {
	updates := map[string]any{
		"processing_status": status,
		"processing_error":  errMsg,
	}

	result := s.db.Model(&models.File{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found")
	}
	return nil
}

// SetFileHasEmbedding sets whether a file has an embedding
func (s *fileService) SetFileHasEmbedding(userID string, fileID uint, hasEmbedding bool) error {
	result := s.db.Model(&models.File{}).
		Where("id = ? AND user_id = ?", fileID, userID).
		Update("has_embedding", hasEmbedding)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("file not found")
	}
	return nil
}

// GetFilesInFolderRecursive returns all files in a folder and its subfolders
func (s *fileService) GetFilesInFolderRecursive(userID string, folderID uint) ([]models.File, error) {
	var allFiles []models.File

	// Get files directly in this folder
	var files []models.File
	if err := s.db.Where("folder_id = ? AND user_id = ?", folderID, userID).Find(&files).Error; err != nil {
		return nil, err
	}
	allFiles = append(allFiles, files...)

	// Get all subfolders
	var subfolders []models.Folder
	if err := s.db.Where("parent_id = ? AND user_id = ?", folderID, userID).Find(&subfolders).Error; err != nil {
		return nil, err
	}

	// Recursively get files from subfolders
	for _, subfolder := range subfolders {
		subFiles, err := s.GetFilesInFolderRecursive(userID, subfolder.ID)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, subFiles...)
	}

	return allFiles, nil
}
