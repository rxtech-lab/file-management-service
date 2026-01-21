package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *FileTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *FileTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *FileTestSuite) TestCreateFile() {
	file := map[string]interface{}{
		"title":             "Test Document",
		"s3_key":            "files/test-user-123/doc.pdf",
		"original_filename": "doc.pdf",
		"mime_type":         "application/pdf",
		"size":              1024,
		"file_type":         "document",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/files", file)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Document", result["title"])
	s.Equal("files/test-user-123/doc.pdf", result["s3_key"])
	s.Equal("doc.pdf", result["original_filename"])
	s.Equal("document", result["file_type"])
	s.Equal("pending", result["processing_status"])
	s.NotNil(result["id"])
}

func (s *FileTestSuite) TestCreateFileInFolder() {
	// Create folder first
	folderID, err := s.setup.CreateTestFolder("Documents", nil)
	s.Require().NoError(err)

	file := map[string]interface{}{
		"title":             "Document in Folder",
		"s3_key":            "files/test-user-123/folder-doc.pdf",
		"original_filename": "folder-doc.pdf",
		"folder_id":         folderID,
	}

	resp, err := s.setup.MakeRequest("POST", "/api/files", file)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal(float64(folderID), result["folder_id"])
}

func (s *FileTestSuite) TestCreateFileEmptyBody() {
	// Empty body should fail
	resp, err := s.setup.MakeRequest("POST", "/api/files", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *FileTestSuite) TestListFiles() {
	// Create some files
	_, err := s.setup.CreateTestFile("File1", "files/test-user-123/file1.pdf", "file1.pdf", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("File2", "files/test-user-123/file2.pdf", "file2.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/files", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 2)
	s.Equal(float64(2), result["total"])
}

func (s *FileTestSuite) TestListFilesInFolder() {
	// Create folder and files
	folderID, err := s.setup.CreateTestFolder("Documents", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("InFolder", "files/test-user-123/in-folder.pdf", "in-folder.pdf", &folderID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("NotInFolder", "files/test-user-123/not-in-folder.pdf", "not-in-folder.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/files?folder_id=%d", folderID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 1)
}

func (s *FileTestSuite) TestListFilesRootFolderDoesNotShowSubfolderFiles() {
	// Create folder
	folderID, err := s.setup.CreateTestFolder("Documents", nil)
	s.Require().NoError(err)

	// Create file in folder
	_, err = s.setup.CreateTestFile("InFolder", "files/test-user-123/in-folder.pdf", "in-folder.pdf", &folderID)
	s.Require().NoError(err)

	// Create file in root (no folder)
	_, err = s.setup.CreateTestFile("InRoot", "files/test-user-123/in-root.pdf", "in-root.pdf", nil)
	s.Require().NoError(err)

	// List files without folder_id (root folder view)
	resp, err := s.setup.MakeRequest("GET", "/api/files", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	// Should only return the root file, not the one in the folder
	data := result["data"].([]interface{})
	s.Len(data, 1)
	s.Equal(float64(1), result["total"])

	// Verify it's the root file
	firstFile := data[0].(map[string]interface{})
	s.Equal("InRoot", firstFile["title"])
}

func (s *FileTestSuite) TestListFilesWithKeyword() {
	// Create files
	_, err := s.setup.CreateTestFile("Invoice 2024", "files/test-user-123/invoice.pdf", "invoice.pdf", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("Contract", "files/test-user-123/contract.pdf", "contract.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/files?keyword=Invoice", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 1)
}

func (s *FileTestSuite) TestGetFile() {
	fileID, err := s.setup.CreateTestFile("Test File", "files/test-user-123/test.pdf", "test.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test File", result["title"])
	s.Equal(float64(fileID), result["id"])
}

func (s *FileTestSuite) TestGetFileNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/files/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FileTestSuite) TestUpdateFile() {
	fileID, err := s.setup.CreateTestFile("Original Title", "files/test-user-123/test.pdf", "test.pdf", nil)
	s.Require().NoError(err)

	update := map[string]interface{}{
		"title":     "Updated Title",
		"summary":   "File summary",
		"file_type": "invoice",
	}

	resp, err := s.setup.MakeRequest("PUT", fmt.Sprintf("/api/files/%d", fileID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Title", result["title"])
	s.Equal("File summary", result["summary"])
	s.Equal("invoice", result["file_type"])
}

func (s *FileTestSuite) TestDeleteFile() {
	fileID, err := s.setup.CreateTestFile("To Delete", "files/test-user-123/delete.pdf", "delete.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/%d", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify it's deleted
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FileTestSuite) TestMoveFiles() {
	// Create folder and files
	folderID, err := s.setup.CreateTestFolder("Target", nil)
	s.Require().NoError(err)
	fileID1, err := s.setup.CreateTestFile("File1", "files/test-user-123/file1.pdf", "file1.pdf", nil)
	s.Require().NoError(err)
	fileID2, err := s.setup.CreateTestFile("File2", "files/test-user-123/file2.pdf", "file2.pdf", nil)
	s.Require().NoError(err)

	move := map[string]interface{}{
		"file_ids":  []int{int(fileID1), int(fileID2)},
		"folder_id": folderID,
	}

	resp, err := s.setup.MakeRequest("POST", "/api/files/move", move)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal(float64(2), result["moved_count"])

	// Verify files are in the folder
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", fileID1), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	fileResult, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)
	s.Equal(float64(folderID), fileResult["folder_id"])
}

func (s *FileTestSuite) TestAddTagsToFile() {
	// Create file and tag
	fileID, err := s.setup.CreateTestFile("Tagged File", "files/test-user-123/tagged.pdf", "tagged.pdf", nil)
	s.Require().NoError(err)
	tagID, err := s.setup.CreateTestTag("Important")
	s.Require().NoError(err)

	body := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}

	resp, err := s.setup.MakeRequest("POST", fmt.Sprintf("/api/files/%d/tags", fileID), body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	tags := result["tags"].([]interface{})
	s.Len(tags, 1)
}

func (s *FileTestSuite) TestRemoveTagsFromFile() {
	// Create file and tag, then add tag
	fileID, err := s.setup.CreateTestFile("Tagged File", "files/test-user-123/tagged.pdf", "tagged.pdf", nil)
	s.Require().NoError(err)
	tagID, err := s.setup.CreateTestTag("ToRemove")
	s.Require().NoError(err)

	// Add tag first
	addBody := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}
	_, err = s.setup.MakeRequest("POST", fmt.Sprintf("/api/files/%d/tags", fileID), addBody)
	s.Require().NoError(err)

	// Remove tag
	removeBody := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}
	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/%d/tags", fileID), removeBody)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	// Tags should be empty (nil or empty array)
	if tags, ok := result["tags"].([]interface{}); ok {
		s.Len(tags, 0)
	}
	// If tags is nil, that's also acceptable (means no tags)
}

func (s *FileTestSuite) TestGetFileDownloadURL() {
	fileID, err := s.setup.CreateTestFile("Download Test", "files/test-user-123/download.pdf", "download.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d/download", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["download_url"])
	s.Equal("download.pdf", result["filename"])
	s.Equal("files/test-user-123/download.pdf", result["key"])
}

func (s *FileTestSuite) TestProcessFile() {
	fileID, err := s.setup.CreateTestFile("Process Test", "files/test-user-123/process.pdf", "process.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("POST", fmt.Sprintf("/api/files/%d/process", fileID), nil)
	s.Require().NoError(err)
	// Should return 202 Accepted as processing is async
	s.Equal(http.StatusAccepted, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.NotNil(result["message"])
	s.Equal("processing", result["status"])
}

func (s *FileTestSuite) TestUnlinkFileInvoice() {
	// Create a file with invoice_id set
	fileID, err := s.setup.CreateTestFile("Invoice Document", "files/test-user-123/invoice.pdf", "invoice.pdf", nil)
	s.Require().NoError(err)

	// Manually set invoice_id using the database
	db := s.setup.DBService.GetDB()
	invoiceID := int64(12345)
	err = db.Exec("UPDATE files SET invoice_id = ? WHERE id = ?", invoiceID, fileID).Error
	s.Require().NoError(err)

	// Verify invoice_id is set
	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)
	s.Equal(float64(invoiceID), result["invoice_id"])

	// Unlink the invoice using invoice_id query parameter
	resp, err = s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/invoice?invoice_id=%d", invoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify invoice_id is now NULL
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", fileID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	result, err = s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)
	s.Nil(result["invoice_id"])
}

func (s *FileTestSuite) TestUnlinkFileInvoiceNotFound() {
	// Try to unlink invoice with non-existent invoice_id
	resp, err := s.setup.MakeRequest("DELETE", "/api/files/invoice?invoice_id=99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FileTestSuite) TestUnlinkFileInvoiceUnauthorized() {
	// Create a file with a different user ID and set invoice_id
	fileID, err := s.setup.CreateTestFile("Other User File", "files/other-user/doc.pdf", "doc.pdf", nil)
	s.Require().NoError(err)

	// Manually set the file to belong to a different user and set invoice_id
	db := s.setup.DBService.GetDB()
	invoiceID := int64(54321)
	err = db.Exec("UPDATE files SET user_id = ?, invoice_id = ? WHERE id = ?", "other-user-456", invoiceID, fileID).Error
	s.Require().NoError(err)

	// Try to unlink invoice as test user (should fail due to user isolation)
	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/invoice?invoice_id=%d", invoiceID), nil)
	s.Require().NoError(err)
	// Should return 404 because the file with this invoice_id doesn't belong to this user
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FileTestSuite) TestUnlinkFileInvoiceInvalidParam() {
	// Try to unlink with invalid invoice_id (0)
	resp, err := s.setup.MakeRequest("DELETE", "/api/files/invoice?invoice_id=0", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	// Try to unlink with negative invoice_id
	resp, err = s.setup.MakeRequest("DELETE", "/api/files/invoice?invoice_id=-1", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *FileTestSuite) TestUnlinkFileInvoiceDifferentUsersSameInvoiceID() {
	// This test verifies that when two different users have files with the same invoice_id,
	// each user can only unlink their own file's invoice

	db := s.setup.DBService.GetDB()
	sameInvoiceID := int64(99999)

	// User 1 (test-user-123) creates a file with invoice_id
	file1ID, err := s.setup.CreateTestFile("User1 Invoice", "files/test-user-123/invoice1.pdf", "invoice1.pdf", nil)
	s.Require().NoError(err)
	err = db.Exec("UPDATE files SET invoice_id = ? WHERE id = ?", sameInvoiceID, file1ID).Error
	s.Require().NoError(err)

	// User 2 (other-user-456) creates a file with the SAME invoice_id
	file2ID, err := s.setup.CreateTestFile("User2 Invoice", "files/other-user-456/invoice2.pdf", "invoice2.pdf", nil)
	s.Require().NoError(err)
	err = db.Exec("UPDATE files SET user_id = ?, invoice_id = ? WHERE id = ?", "other-user-456", sameInvoiceID, file2ID).Error
	s.Require().NoError(err)

	// Verify both files have the same invoice_id but different users
	var file1Count, file2Count int64
	db.Model(&struct {
		ID        uint
		UserID    string
		InvoiceID *int64
	}{}).Table("files").Where("id = ? AND user_id = ? AND invoice_id = ?", file1ID, "test-user-123", sameInvoiceID).Count(&file1Count)
	s.Equal(int64(1), file1Count)
	db.Model(&struct {
		ID        uint
		UserID    string
		InvoiceID *int64
	}{}).Table("files").Where("id = ? AND user_id = ? AND invoice_id = ?", file2ID, "other-user-456", sameInvoiceID).Count(&file2Count)
	s.Equal(int64(1), file2Count)

	// User 1 (test-user-123) tries to unlink - should succeed for their file
	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/invoice?invoice_id=%d", sameInvoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify User 1's file has invoice_id unlinked
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/files/%d", file1ID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)
	s.Nil(result["invoice_id"])

	// Verify User 2's file STILL has invoice_id (not affected by User 1's unlink)
	var file2InvoiceID *int64
	err = db.Raw("SELECT invoice_id FROM files WHERE id = ? AND user_id = ?", file2ID, "other-user-456").Scan(&file2InvoiceID).Error
	s.Require().NoError(err)
	s.NotNil(file2InvoiceID)
	s.Equal(sameInvoiceID, *file2InvoiceID)

	// User 1 tries to unlink again - should fail (404) because their file no longer has this invoice_id
	resp, err = s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/files/invoice?invoice_id=%d", sameInvoiceID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)

	// Verify User 2's file STILL has the invoice_id intact
	err = db.Raw("SELECT invoice_id FROM files WHERE id = ? AND user_id = ?", file2ID, "other-user-456").Scan(&file2InvoiceID).Error
	s.Require().NoError(err)
	s.NotNil(file2InvoiceID)
	s.Equal(sameInvoiceID, *file2InvoiceID)
}

func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}
