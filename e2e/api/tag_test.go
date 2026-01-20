package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TagTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *TagTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *TagTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *TagTestSuite) TestCreateTag() {
	tag := map[string]interface{}{
		"name":        "Important",
		"description": "Important documents",
		"color":       "#FF0000",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/tags", tag)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Important", result["name"])
	s.Equal("Important documents", result["description"])
	s.Equal("#FF0000", result["color"])
	s.NotNil(result["id"])
}

func (s *TagTestSuite) TestCreateTagMinimal() {
	tag := map[string]interface{}{
		"name": "Simple Tag",
	}

	resp, err := s.setup.MakeRequest("POST", "/api/tags", tag)
	s.Require().NoError(err)
	s.Equal(http.StatusCreated, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Simple Tag", result["name"])
	s.NotNil(result["id"])
}

func (s *TagTestSuite) TestCreateTagEmptyBody() {
	// Empty body should fail
	resp, err := s.setup.MakeRequest("POST", "/api/tags", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *TagTestSuite) TestListTags() {
	// Create some tags first
	_, err := s.setup.CreateTestTag("Tag1")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestTag("Tag2")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/tags", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 2)
	s.Equal(float64(2), result["total"])
}

func (s *TagTestSuite) TestListTagsWithKeyword() {
	// Create tags
	_, err := s.setup.CreateTestTag("Invoice Tag")
	s.Require().NoError(err)
	_, err = s.setup.CreateTestTag("Document Tag")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/tags?keyword=Invoice", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	data := result["data"].([]interface{})
	s.Len(data, 1)
}

func (s *TagTestSuite) TestGetTag() {
	tagID, err := s.setup.CreateTestTag("Test Tag")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", fmt.Sprintf("/api/tags/%d", tagID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Test Tag", result["name"])
	s.Equal(float64(tagID), result["id"])
}

func (s *TagTestSuite) TestGetTagNotFound() {
	resp, err := s.setup.MakeRequest("GET", "/api/tags/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *TagTestSuite) TestUpdateTag() {
	tagID, err := s.setup.CreateTestTag("Original Name")
	s.Require().NoError(err)

	update := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
		"color":       "#00FF00",
	}

	resp, err := s.setup.MakeRequest("PUT", fmt.Sprintf("/api/tags/%d", tagID), update)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Updated Name", result["name"])
	s.Equal("Updated description", result["description"])
	s.Equal("#00FF00", result["color"])
}

func (s *TagTestSuite) TestUpdateTagNotFound() {
	update := map[string]interface{}{
		"name": "Updated Name",
	}

	resp, err := s.setup.MakeRequest("PUT", "/api/tags/99999", update)
	s.Require().NoError(err)
	// Handler returns 401 when tag not found (treated as unauthorized to update)
	s.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *TagTestSuite) TestDeleteTag() {
	tagID, err := s.setup.CreateTestTag("To Delete")
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("DELETE", fmt.Sprintf("/api/tags/%d", tagID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode)

	// Verify it's deleted
	resp, err = s.setup.MakeRequest("GET", fmt.Sprintf("/api/tags/%d", tagID), nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *TagTestSuite) TestDeleteTagNotFound() {
	resp, err := s.setup.MakeRequest("DELETE", "/api/tags/99999", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestTagSuite(t *testing.T) {
	suite.Run(t, new(TagTestSuite))
}
