package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// AuthSuite tests authentication middleware behavior
type AuthSuite struct {
	suite.Suite
	setup *TestSetup
}

func (s *AuthSuite) SetupTest() {
	s.setup = NewTestSetup(s.T())
}

func (s *AuthSuite) TearDownTest() {
	s.setup.Cleanup()
}

// TestAuthMiddlewareOrder verifies that auth middleware runs before route handlers.
// This test prevents regression of the middleware ordering bug where routes were
// registered before authentication middleware, causing auth to never run.
func (s *AuthSuite) TestAuthMiddlewareOrder() {
	// Test various endpoints that require authentication
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/tags"},
		{"POST", "/api/tags"},
		{"GET", "/api/folders"},
		{"POST", "/api/folders"},
		{"GET", "/api/files"},
		{"POST", "/api/files"},
	}

	for _, ep := range endpoints {
		s.Run(ep.method+" "+ep.path+" requires auth", func() {
			// Request WITHOUT auth header should fail
			req := httptest.NewRequest(ep.method, ep.path, nil)
			req.Header.Set("Content-Type", "application/json")
			// Deliberately NOT setting X-Test-User-ID header

			resp, err := s.setup.App.Test(req, -1)
			s.Require().NoError(err)
			s.Equal(http.StatusUnauthorized, resp.StatusCode,
				"Request without auth should return 401 Unauthorized")
		})
	}
}

// TestAuthenticatedRequestsSucceed verifies that properly authenticated requests work
func (s *AuthSuite) TestAuthenticatedRequestsSucceed() {
	// GET requests with auth should succeed (200)
	resp, err := s.setup.MakeRequest("GET", "/api/tags", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode,
		"Authenticated GET request should succeed")

	resp, err = s.setup.MakeRequest("GET", "/api/folders", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode,
		"Authenticated GET request should succeed")

	resp, err = s.setup.MakeRequest("GET", "/api/files", nil)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode,
		"Authenticated GET request should succeed")
}

// TestUnauthenticatedHealthCheck verifies health check doesn't require auth
func (s *AuthSuite) TestUnauthenticatedHealthCheck() {
	req := httptest.NewRequest("GET", "/health", nil)
	// No auth header

	resp, err := s.setup.App.Test(req, -1)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode,
		"Health check should not require authentication")
}

// TestUserIsolation verifies that users can only access their own data
func (s *AuthSuite) TestUserIsolation() {
	// Create a tag as user1
	tagID, err := s.setup.CreateTestTag("User1 Tag")
	s.Require().NoError(err)

	// User1 should be able to access their tag
	resp, err := s.setup.MakeAuthenticatedRequest("GET", "/api/tags/"+uintToStringHelper(tagID), nil, s.setup.TestUserID)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode, "User1 should access their own tag")

	// User2 should NOT be able to access User1's tag
	resp, err = s.setup.MakeAuthenticatedRequest("GET", "/api/tags/"+uintToStringHelper(tagID), nil, "different-user-456")
	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode, "User2 should not access User1's tag")
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}
