package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================================
// PERMISSION TESTING HELPERS
// ============================================================================

// TestJWTSecret is the secret key used for signing test JWT tokens
// This must match the secret used in createTestServer (handlers_test.go)
const TestJWTSecret = "test-secret-key"

// authenticateAs creates a JWT token for the specified role and returns
// the Authorization header value with "Bearer " prefix.
//
// Usage:
//
//	authHeader := authenticateAs(t, "admin")
//	req.Header.Set("Authorization", authHeader)
func authenticateAs(t *testing.T, role string) string {
	t.Helper()
	return authenticateAsUser(t, 1, "testuser", role)
}

// authenticateAsUser creates a JWT token for a specific user ID and role.
// This allows testing scenarios where user ID matters (e.g., self-edit permissions).
//
// Usage:
//
//	authHeader := authenticateAsUser(t, 2, "regularuser", "user")
//	req.Header.Set("Authorization", authHeader)
func authenticateAsUser(t *testing.T, userID int, username, role string) string {
	t.Helper()

	// Create a test JWT token matching the Claims structure used in production
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"email":    username + "@test.com",
		"role":     role,
		"type":     "access",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Unix(),
		"sub":      "user:" + string(rune(userID)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(TestJWTSecret))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	return "Bearer " + tokenString
}

// makeAuthenticatedRequest creates and executes an HTTP request with authentication.
// This helper reduces boilerplate in permission boundary tests.
//
// Usage:
//
//	w := makeAuthenticatedRequest(t, router, "GET", "/api/v1/users", nil, "admin")
//	assert.Equal(t, http.StatusOK, w.Code)
func makeAuthenticatedRequest(t *testing.T, router *gin.Engine, method, path string, body []byte, role string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authenticateAs(t, role))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// makeAuthenticatedRequestAsUser creates and executes an HTTP request with a specific user context.
// Use this when testing user-specific permissions (e.g., user editing their own profile).
//
// Usage:
//
//	w := makeAuthenticatedRequestAsUser(t, router, "PUT", "/api/v1/users/2", body, 2, "regularuser", "user")
func makeAuthenticatedRequestAsUser(t *testing.T, router *gin.Engine, method, path string, body []byte, userID int, username, role string) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authenticateAsUser(t, userID, username, role))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w
}

// permissionTestCase represents a single permission test scenario for table-driven tests.
// This structure standardizes permission boundary testing across the codebase.
//
// Example:
//
//	tests := []permissionTestCase{
//	    {
//	        name:           "Admin_can_list_users",
//	        method:         "GET",
//	        path:           "/api/v1/users",
//	        role:           "admin",
//	        expectedStatus: http.StatusOK,
//	        description:    "Admin should be able to list users",
//	    },
//	    {
//	        name:           "User_cannot_create_user",
//	        method:         "POST",
//	        path:           "/api/v1/users",
//	        body:           createUserBody,
//	        role:           "user",
//	        expectedStatus: http.StatusForbidden,
//	        description:    "Regular user should not be able to create users",
//	    },
//	}
type permissionTestCase struct {
	name           string // Test case name (used in t.Run)
	method         string // HTTP method (GET, POST, PUT, DELETE)
	path           string // Request path
	body           []byte // Request body (nil for GET/DELETE)
	role           string // User role: "admin", "user", or "viewer"
	expectedStatus int    // Expected HTTP status code
	description    string // Human-readable description of the test
	allowStatus    []int  // Additional acceptable status codes (e.g., 401 if auth not wired)
}

// runPermissionTests executes a table of permission tests.
// This helper provides consistent test output and reduces code duplication.
//
// Usage:
//
//	tests := []permissionTestCase{...}
//	runPermissionTests(t, router, tests)
func runPermissionTests(t *testing.T, router *gin.Engine, tests []permissionTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := makeAuthenticatedRequest(t, router, tt.method, tt.path, tt.body, tt.role)

			// Check if the status matches expected or is in the allow list
			statusOK := w.Code == tt.expectedStatus
			for _, allowed := range tt.allowStatus {
				if w.Code == allowed {
					statusOK = true
					break
				}
			}

			assert.True(t, statusOK,
				"%s: expected status %d, got %d. Response: %s",
				tt.description, tt.expectedStatus, w.Code, w.Body.String())
		})
	}
}

// runPermissionTestsWithUser executes permission tests with specific user contexts.
// Use this when user ID matters for the permission check.
//
// Usage:
//
//	tests := []permissionTestCaseWithUser{
//	    {
//	        permissionTestCase: permissionTestCase{...},
//	        userID:   2,
//	        username: "regularuser",
//	    },
//	}
func runPermissionTestsWithUser(t *testing.T, router *gin.Engine, tests []permissionTestCaseWithUser) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := makeAuthenticatedRequestAsUser(t, router, tt.method, tt.path, tt.body,
				tt.userID, tt.username, tt.role)

			// Check if the status matches expected or is in the allow list
			statusOK := w.Code == tt.expectedStatus
			for _, allowed := range tt.allowStatus {
				if w.Code == allowed {
					statusOK = true
					break
				}
			}

			assert.True(t, statusOK,
				"%s: expected status %d, got %d. Response: %s",
				tt.description, tt.expectedStatus, w.Code, w.Body.String())
		})
	}
}

// permissionTestCaseWithUser extends permissionTestCase with specific user identity.
type permissionTestCaseWithUser struct {
	permissionTestCase
	userID   int    // Specific user ID for the request
	username string // Specific username for the request
}

// assertPermissionDenied checks that a response indicates permission denied (403 or 401).
// This helper is useful for negative permission tests.
//
// Usage:
//
//	w := makeAuthenticatedRequest(t, router, "DELETE", "/api/v1/users/2", nil, "user")
//	assertPermissionDenied(t, w, "Regular user should not be able to delete users")
func assertPermissionDenied(t *testing.T, w *httptest.ResponseRecorder, message string) {
	t.Helper()
	assert.True(t, w.Code == http.StatusForbidden || w.Code == http.StatusUnauthorized,
		"%s: expected 403 Forbidden or 401 Unauthorized, got %d. Response: %s",
		message, w.Code, w.Body.String())
}

// assertPermissionGranted checks that a response indicates success (200, 201, or 204).
// This helper is useful for positive permission tests.
//
// Usage:
//
//	w := makeAuthenticatedRequest(t, router, "GET", "/api/v1/users", nil, "admin")
//	assertPermissionGranted(t, w, "Admin should be able to list users")
func assertPermissionGranted(t *testing.T, w *httptest.ResponseRecorder, message string) {
	t.Helper()
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusNoContent,
		"%s: expected 200 OK, 201 Created, or 204 No Content, got %d. Response: %s",
		message, w.Code, w.Body.String())
}
