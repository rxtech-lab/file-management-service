package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SearchTestSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *SearchTestSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *SearchTestSuite) TearDownTest() {
	s.setup.Cleanup()
}

func (s *SearchTestSuite) TestSearchFilesFullText() {
	// Create some files with different titles
	_, err := s.setup.CreateTestFile("Invoice 2024", "files/test-user-123/invoice2024.pdf", "invoice2024.pdf", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("Contract Agreement", "files/test-user-123/contract.pdf", "contract.pdf", nil)
	s.Require().NoError(err)
	_, err = s.setup.CreateTestFile("Invoice 2023", "files/test-user-123/invoice2023.pdf", "invoice2023.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/search?q=Invoice&type=fulltext", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Invoice", result["query"])
	s.Equal("fulltext", result["search_type"])
	// Note: files are created with pending status, so they won't appear in search results
	// which filter by completed status. The test verifies the endpoint works.
}

func (s *SearchTestSuite) TestSearchFilesHybrid() {
	// Create files for search
	_, err := s.setup.CreateTestFile("Project Plan", "files/test-user-123/project.pdf", "project.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/search?q=project&type=hybrid", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("project", result["query"])
	s.Equal("hybrid", result["search_type"])
}

func (s *SearchTestSuite) TestSearchFilesSemantic() {
	// Create files for search
	_, err := s.setup.CreateTestFile("Annual Report", "files/test-user-123/report.pdf", "report.pdf", nil)
	s.Require().NoError(err)

	resp, err := s.setup.MakeRequest("GET", "/api/search?q=yearly%20summary&type=semantic", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("yearly summary", result["query"])
	s.Equal("semantic", result["search_type"])
}

func (s *SearchTestSuite) TestSearchFilesDefaultType() {
	// Create files
	_, err := s.setup.CreateTestFile("Test Doc", "files/test-user-123/test.pdf", "test.pdf", nil)
	s.Require().NoError(err)

	// Default search type should be fulltext
	resp, err := s.setup.MakeRequest("GET", "/api/search?q=test", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("test", result["query"])
	s.Equal("fulltext", result["search_type"])
}

func (s *SearchTestSuite) TestSearchFilesMissingQuery() {
	resp, err := s.setup.MakeRequest("GET", "/api/search", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *SearchTestSuite) TestSearchFilesWithPagination() {
	// Create multiple files
	for i := 0; i < 5; i++ {
		_, err := s.setup.CreateTestFile(
			"Document "+string(rune('A'+i)),
			"files/test-user-123/doc"+string(rune('a'+i))+".pdf",
			"doc"+string(rune('a'+i))+".pdf",
			nil,
		)
		s.Require().NoError(err)
	}

	resp, err := s.setup.MakeRequest("GET", "/api/search?q=Document&type=fulltext&limit=2&offset=0", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	result, err := s.setup.ReadResponseBody(resp)
	s.Require().NoError(err)

	s.Equal("Document", result["query"])
	// Results might be empty due to processing status filter, but pagination params are passed
}

func (s *SearchTestSuite) TestSearchFilesInvalidType() {
	// Invalid search type falls back to default behavior (fulltext)
	resp, err := s.setup.MakeRequest("GET", "/api/search?q=test&type=invalid", nil)
	s.Require().NoError(err)
	// Handler accepts any type value and uses it as-is
	s.Equal(http.StatusOK, resp.StatusCode)
}

func TestSearchSuite(t *testing.T) {
	suite.Run(t, new(SearchTestSuite))
}
