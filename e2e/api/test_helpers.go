package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/services"
	"github.com/rxtech-lab/invoice-management/internal/utils"
	"github.com/stretchr/testify/require"
)

// TestSetup contains all test dependencies
type TestSetup struct {
	t                    *testing.T
	DBService            services.DBService
	TagService           services.TagService
	FolderService        services.FolderService
	FileService          services.FileService
	UploadService        services.UploadService
	EmbeddingService     services.EmbeddingService
	ContentParserService services.ContentParserService
	SearchService        services.SearchService
	InvoiceService       *services.MockInvoiceService
	APIServer            *api.APIServer
	App                  *fiber.App
	TestUserID           string
}

// NewTestSetup creates a new test setup with in-memory database
func NewTestSetup(t *testing.T) *TestSetup {
	// Create in-memory database
	dbService, err := services.NewSqliteDBService(":memory:")
	require.NoError(t, err, "Failed to create in-memory database")

	db := dbService.GetDB()

	// Create services
	tagService := services.NewTagService(db)
	folderService := services.NewFolderService(db)
	fileService := services.NewFileService(db)
	uploadService := services.NewMockUploadService()
	embeddingService := services.NewMockEmbeddingService()
	contentParserService := services.NewMockContentParserService()
	summaryService := services.NewMockSummaryService()
	searchService := services.NewSearchService(db, embeddingService)
	agentService := services.NewMockAgentService()
	invoiceService := services.NewMockInvoiceService(true)

	// Create API server
	apiServer := api.NewAPIServer(
		dbService,
		tagService,
		folderService,
		fileService,
		uploadService,
		embeddingService,
		contentParserService,
		searchService,
		summaryService,
		agentService,
		invoiceService,
		nil, // No MCP server for tests
	)

	// Add test authentication middleware before routes
	SetupTestAuthMiddleware(apiServer.GetFiberApp())

	// Setup routes
	apiServer.SetupRoutes()

	setup := &TestSetup{
		t:                    t,
		DBService:            dbService,
		TagService:           tagService,
		FolderService:        folderService,
		FileService:          fileService,
		UploadService:        uploadService,
		EmbeddingService:     embeddingService,
		ContentParserService: contentParserService,
		SearchService:        searchService,
		InvoiceService:       invoiceService,
		APIServer:            apiServer,
		App:                  apiServer.GetFiberApp(),
		TestUserID:           "test-user-123",
	}

	return setup
}

// Cleanup cleans up test resources
func (s *TestSetup) Cleanup() {
	if s.DBService != nil {
		s.DBService.Close()
	}
}

// MakeRequest makes an HTTP request to the test server
func (s *TestSetup) MakeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Add mock authentication header
	s.addAuthHeader(req)

	resp, err := s.App.Test(req, -1)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// MakeAuthenticatedRequest makes an authenticated HTTP request
func (s *TestSetup) MakeAuthenticatedRequest(method, path string, body interface{}, userID string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Add mock user context
	s.addAuthHeaderWithUserID(req, userID)

	resp, err := s.App.Test(req, -1)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// addAuthHeader adds authentication header for testing
func (s *TestSetup) addAuthHeader(req *http.Request) {
	s.addAuthHeaderWithUserID(req, s.TestUserID)
}

// addAuthHeaderWithUserID adds authentication header with a specific user ID
func (s *TestSetup) addAuthHeaderWithUserID(req *http.Request, userID string) {
	// For testing, we'll use a mock JWT or a test header
	// The actual authentication is handled by middleware
	req.Header.Set("X-Test-User-ID", userID)
}

// ReadResponseBody reads the response body as a map
func (s *TestSetup) ReadResponseBody(resp *http.Response) (map[string]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ReadResponseBodyArray reads the response body as an array
func (s *TestSetup) ReadResponseBodyArray(resp *http.Response) ([]interface{}, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateTestTag creates a test tag
func (s *TestSetup) CreateTestTag(name string) (uint, error) {
	tag := &struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}{
		Name:        name,
		Description: "Test tag",
		Color:       "#FF5733",
	}

	resp, err := s.MakeRequest("POST", "/api/tags", tag)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestFolder creates a test folder
func (s *TestSetup) CreateTestFolder(name string, parentID *uint) (uint, error) {
	folder := map[string]interface{}{
		"name":        name,
		"description": "Test folder",
	}

	if parentID != nil {
		folder["parent_id"] = *parentID
	}

	resp, err := s.MakeRequest("POST", "/api/folders", folder)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// CreateTestFile creates a test file
func (s *TestSetup) CreateTestFile(title, s3Key, filename string, folderID *uint) (uint, error) {
	file := map[string]interface{}{
		"title":             title,
		"s3_key":            s3Key,
		"original_filename": filename,
		"file_type":         "document",
	}

	if folderID != nil {
		file["folder_id"] = *folderID
	}

	resp, err := s.MakeRequest("POST", "/api/files", file)
	if err != nil {
		return 0, err
	}

	result, err := s.ReadResponseBody(resp)
	if err != nil {
		return 0, err
	}

	return uint(result["id"].(float64)), nil
}

// uintToStringHelper converts uint to string
func uintToStringHelper(n uint) string {
	return fmt.Sprintf("%d", n)
}

// SetupTestAuthMiddleware sets up a test authentication middleware
func SetupTestAuthMiddleware(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		userID := c.Get("X-Test-User-ID")
		if userID != "" {
			user := &utils.AuthenticatedUser{
				Sub: userID,
			}
			c.Locals(middleware.AuthenticatedUserContextKey, user)

			// Also set raw auth token if provided (for invoice processing tests)
			// This uses the same key as the OAuth middleware: middleware.RawAuthTokenContextKey
			authToken := c.Get("X-Test-Auth-Token")
			if authToken != "" {
				c.Locals(middleware.RawAuthTokenContextKey, authToken)
			}
		}
		return c.Next()
	})
}
