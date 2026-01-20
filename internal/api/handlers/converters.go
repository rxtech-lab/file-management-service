package handlers

import (
	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// Tag converters

func tagModelToGenerated(tag *models.Tag) generated.Tag {
	result := generated.Tag{
		Id:        int(tag.ID),
		UserId:    tag.UserID,
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
	if tag.Color != "" {
		result.Color = &tag.Color
	}
	if tag.Description != "" {
		result.Description = &tag.Description
	}
	return result
}

func tagListToGenerated(tags []models.Tag) []generated.Tag {
	result := make([]generated.Tag, len(tags))
	for i, tag := range tags {
		result[i] = tagModelToGenerated(&tag)
	}
	return result
}

// Folder converters

func folderModelToGenerated(folder *models.Folder) generated.Folder {
	result := generated.Folder{
		Id:        int(folder.ID),
		UserId:    folder.UserID,
		Name:      folder.Name,
		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}

	if folder.ParentID != nil {
		parentID := int(*folder.ParentID)
		result.ParentId = &parentID
	}

	if folder.Description != "" {
		result.Description = &folder.Description
	}

	if len(folder.Children) > 0 {
		childList := folderListToGenerated(folder.Children)
		result.Children = &childList
	}

	if len(folder.Tags) > 0 {
		tagList := tagListToGenerated(folder.Tags)
		result.Tags = &tagList
	}

	return result
}

func folderListToGenerated(folders []models.Folder) []generated.Folder {
	result := make([]generated.Folder, len(folders))
	for i, folder := range folders {
		result[i] = folderModelToGenerated(&folder)
	}
	return result
}

// FolderTree converters - converts models.Folder to generated.FolderTree (simplified view)

func folderToTreeGenerated(folder *models.Folder) generated.FolderTree {
	result := generated.FolderTree{
		Id:   int(folder.ID),
		Name: folder.Name,
	}

	if folder.ParentID != nil {
		parentID := int(*folder.ParentID)
		result.ParentId = &parentID
	}

	if len(folder.Children) > 0 {
		childList := folderListToTreeGenerated(folder.Children)
		result.Children = &childList
	}

	return result
}

func folderListToTreeGenerated(folders []models.Folder) []generated.FolderTree {
	result := make([]generated.FolderTree, len(folders))
	for i, folder := range folders {
		result[i] = folderToTreeGenerated(&folder)
	}
	return result
}

// File converters

func fileModelToGenerated(file *models.File) generated.File {
	result := generated.File{
		Id:               int(file.ID),
		UserId:           file.UserID,
		Title:            file.Title,
		S3Key:            file.S3Key,
		OriginalFilename: file.OriginalFilename,
		FileType:         generated.FileType(file.FileType),
		ProcessingStatus: generated.ProcessingStatus(file.ProcessingStatus),
		HasEmbedding:     file.HasEmbedding,
		CreatedAt:        file.CreatedAt,
		UpdatedAt:        file.UpdatedAt,
	}

	if file.FolderID != nil {
		folderID := int(*file.FolderID)
		result.FolderId = &folderID
	}

	if file.Folder != nil {
		f := folderModelToGenerated(file.Folder)
		result.Folder = &f
	}

	if len(file.Tags) > 0 {
		tagList := tagListToGenerated(file.Tags)
		result.Tags = &tagList
	}

	if file.Summary != "" {
		result.Summary = &file.Summary
	}

	if file.Content != "" {
		result.Content = &file.Content
	}

	if file.MimeType != "" {
		result.MimeType = &file.MimeType
	}

	if file.Size > 0 {
		result.Size = &file.Size
	}

	if file.ProcessingError != "" {
		result.ProcessingError = &file.ProcessingError
	}

	return result
}

func fileListToGenerated(files []models.File) []generated.File {
	result := make([]generated.File, len(files))
	for i, file := range files {
		result[i] = fileModelToGenerated(&file)
	}
	return result
}

// Search result converters

func searchResultToGenerated(result *services.SearchResult) generated.SearchResult {
	genResult := generated.SearchResult{
		File:  fileModelToGenerated(&result.File),
		Score: result.Score,
	}
	if result.Snippet != "" {
		genResult.Snippet = &result.Snippet
	}
	return genResult
}

func searchResultListToGenerated(results []services.SearchResult) []generated.SearchResult {
	genResults := make([]generated.SearchResult, len(results))
	for i, r := range results {
		genResults[i] = searchResultToGenerated(&r)
	}
	return genResults
}
