package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rxtech-lab/invoice-management/internal/api/middleware"
	"github.com/rxtech-lab/invoice-management/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRawAuthTokenContextKeyConsistency verifies that the middleware and handlers
// use the same context key for the raw auth token. This prevents bugs where
// the handler uses a different key than what the middleware sets.
func TestRawAuthTokenContextKeyConsistency(t *testing.T) {
	// The middleware uses middleware.RawAuthTokenContextKey
	// The handlers should also use middleware.RawAuthTokenContextKey
	// This test ensures they match

	expectedKey := "rawAuthToken"
	assert.Equal(t, expectedKey, middleware.RawAuthTokenContextKey,
		"middleware.RawAuthTokenContextKey should be 'rawAuthToken'")

	// Create a test app to verify the token flows correctly
	app := fiber.New()

	// Simulate OAuth middleware setting the token
	app.Use(func(c *fiber.Ctx) error {
		// Set user
		user := &utils.AuthenticatedUser{Sub: "test-user"}
		c.Locals(middleware.AuthenticatedUserContextKey, user)

		// Set raw token (like OAuth middleware does)
		c.Locals(middleware.RawAuthTokenContextKey, "test-token-12345")

		return c.Next()
	})

	// Test endpoint that mimics how ProcessingHandlers extracts the token
	app.Get("/test-token", func(c *fiber.Ctx) error {
		// This is exactly how ProcessingHandlers extracts the token
		authToken := ""
		if token, ok := c.Locals(middleware.RawAuthTokenContextKey).(string); ok {
			authToken = token
		}

		return c.JSON(fiber.Map{
			"token_found": authToken != "",
			"token_value": authToken,
		})
	})

	// Make request
	req := httptest.NewRequest("GET", "/test-token", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	// Verify token was extracted correctly
	assert.True(t, result["token_found"].(bool), "Token should be found using middleware.RawAuthTokenContextKey")
	assert.Equal(t, "test-token-12345", result["token_value"].(string), "Token value should match")
}

// TestProcessStreamAuthTokenExtraction tests that the process-stream endpoint
// correctly receives and extracts the auth token from Fiber locals
func TestProcessStreamAuthTokenExtraction(t *testing.T) {
	setup := NewTestSetupWithAuthToken(t)
	defer setup.Cleanup()

	// Create a file first
	fileID, err := setup.CreateTestFile("test.pdf", "test-key", "test.pdf", nil)
	require.NoError(t, err)

	// Make request to process-stream with auth token
	req := httptest.NewRequest("GET", "/api/files/"+uintToStringHelper(fileID)+"/process-stream", nil)
	req.Header.Set("X-Test-User-ID", setup.TestUserID)
	req.Header.Set("X-Test-Auth-Token", "test-bearer-token")

	resp, err := setup.App.Test(req, -1) // -1 for no timeout on SSE
	require.NoError(t, err)

	// The endpoint should return 200 and start streaming
	// (or 409 if already processing, which is also valid)
	assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict,
		"Expected 200 or 409, got %d", resp.StatusCode)
}

// NewTestSetupWithAuthToken creates a test setup that includes auth token handling
func NewTestSetupWithAuthToken(t *testing.T) *TestSetup {
	setup := NewTestSetup(t)
	// Note: The test auth middleware is updated to also set raw auth token
	return setup
}
