package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FolderTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *FolderTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *FolderTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *FolderTestSuite) TestCreateFolder() {
	folder := map[string]interface{}{
		"name":        "Documents",
		"description": "My documents folder",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/folders", folder)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Documents", result["name"])
	s.Equal("My documents folder", result["description"])
	s.NotNil(result["id"])
}

func (s *FolderTestSuite) TestCreateFolderWithParent() {
	// Create parent folder first
	parentID, err := s.setup.CreateTestFolder("Parent", nil)
	s.Require().NoError(err)

	folder := map[string]interface{}{
		"name":      "Child",
		"parent_id": parentID,
	}

	resp, err := s.setup.MakeRequest("POST", "/api/folders", folder)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Child", result["name"])
	s.Equal(float64(parentID), result["parent_id"])
}

func (s *FolderTestSuite) TestCreateFolderEmptyBody() {
	// Empty body should fail
	resp, err := s.setup.MakeRequest("POST", "/api/folders", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *FolderTestSuite) TestCreateFolderInvalidParent() {
	folder := map[string]interface{}{
		"name":      "Child",
		"parent_id": 99999,
	}

	resp, err := s.setup.MakeRequest("POST", "/api/folders", folder)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *FolderTestSuite) TestListFolders() {
	// Create some folders
	_, err := s.setup.CreateTestFolder("Folder1", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Folder2", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/folders", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 2)
	s.Equal(float64(2), result["total"])
}

func (s *FolderTestSuite) TestListFoldersWithParent() {
	// Create parent and child folders
	parentID, err := s.setup.CreateTestFolder("Parent", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Child1", &parentID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Child2", &parentID)
	s.Require().NoError(err)

	// List children of parent
	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/folders?parent_id=%d", parentID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 2)
}

func (s *FolderTestSuite) TestGetFolder() {
	folderID, err := s.setup.CreateTestFolder("Test Folder", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/folders/%d", folderID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Folder", result["name"])
	s.Equal(float64(folderID), result["id"])
}

func (s *FolderTestSuite) TestGetFolderNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/folders/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FolderTestSuite) TestUpdateFolder() {
	folderID, err := s.setup.CreateTestFolder("Original Name", nil)
	s.Require().NoError(err)

	update := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
	}

	resp, err := s.setup.MakeRequest("PUT", fmt.Sprintf("/api/folders/%d", folderID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Name", result["name"])
	s.Equal("Updated description", result["description"])
}

func (s *FolderTestSuite) TestDeleteFolder() {
	folderID, err := s.setup.CreateTestFolder("To Delete", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/folders/%d", folderID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify it's deleted
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/folders/%d", folderID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *FolderTestSuite) TestMoveFolder() {
	// Create two parent folders and a child
	parent1ID, err := s.setup.CreateTestFolder("Parent1", nil)
	s.Require().NoError(err)
	parent2ID, err := s.setup.CreateTestFolder("Parent2", nil)
	s.Require().NoError(err)
	childID, err := s.setup.CreateTestFolder("Child", &parent1ID)
	s.Require().NoError(err)

	// Move child to parent2
	move := map[string]interface{}{
		"parent_id": parent2ID,
	}

	resp, err := s.setup.MakeRequest("POST", fmt.Sprintf("/api/folders/%d/move", childID), move)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal(float64(parent2ID), result["parent_id"])
}

func (s *FolderTestSuite) TestMoveFolderToRoot() {
	// Create parent and child
	parentID, err := s.setup.CreateTestFolder("Parent", nil)
	s.Require().NoError(err)
	childID, err := s.setup.CreateTestFolder("Child", &parentID)
	s.Require().NoError(err)

	// Move child to root (no parent_id)
	move := map[string]interface{}{}

	resp, err := s.setup.MakeRequest("POST", fmt.Sprintf("/api/folders/%d/move", childID), move)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Nil(result["parent_id"])
}

func (s *FolderTestSuite) TestGetFolderTree() {
	// Create a folder structure
	parent1ID, err := s.setup.CreateTestFolder("Parent1", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Child1", &parent1ID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Child2", &parent1ID)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFolder("Parent2", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/folders/tree", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBodyArray(resp)
	s.Require().NoError(err)

	// Should have 2 root folders
	s.Len(result, 2)

	// Find Parent1 and check it has children
	for _, f := range result {
		folder := f.(map[string]interface{})
		if folder["name"] == "Parent1" {
			children := folder["children"].([]interface{})
			s.Len(children, 2)
		}
	}
}

func (s *FolderTestSuite) TestAddTagsToFolder() {
	// Create folder and tag
	folderID, err := s.setup.CreateTestFolder("Tagged Folder", nil)
	s.Require().NoError(err)
	tagID, err := s.setup.CreateTestTag("Important")
	s.Require().NoError(err)

	body := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}

	resp, err := s.setup.MakeRequest("POST", fmt.Sprintf("/api/folders/%d/tags", folderID), body)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	// Tags may be nil or an array
	if tags, ok := result["tags"].([]interface{}); ok {
		s.Len(tags, 1)
	} else {
		s.Fail("Expected tags to be an array")
	}
}

func (s *FolderTestSuite) TestRemoveTagsFromFolder() {
	// Create folder and tag, then add tag
	folderID, err := s.setup.CreateTestFolder("Tagged Folder", nil)
	s.Require().NoError(err)
	tagID, err := s.setup.CreateTestTag("ToRemove")
	s.Require().NoError(err)

	// Add tag first
	addBody := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}
	_, err = s.setup.MakeRequest("POST", fmt.Sprintf("/api/folders/%d/tags", folderID), addBody)
	s.Require().NoError(err)

	// Remove tag
	removeBody := map[string]interface{}{
		"tag_ids": []int{int(tagID)},
	}
	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/folders/%d/tags", folderID), removeBody)
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

func TestFolderSuite(t *testing.T) {
	suite.Run(t, new(FolderTestSuite))
}
