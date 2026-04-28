package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// HTTP STATUS CODES: 200 OK (Successful Requests)
// ============================================================================

func TestHTTPStatusCodes_200OK(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"Health endpoint returns 200", "GET", "/api/v1/health", nil},
		{"Valid login returns 200", "POST", "/api/v1/auth/login", models.LoginRequest{
			Username: "testuser",
			Password: "password123",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Status code mismatch")
		})
	}
}

// ============================================================================
// HTTP STATUS CODES: 400 Bad Request (Validation Errors)
// ============================================================================

func TestHTTPStatusCodes_400BadRequest(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tests := []struct {
		name        string
		method      string
		path        string
		body        interface{}
		rawBody     string
		expectCodes []int // Allow multiple acceptable status codes
	}{
		{"Empty username returns 400", "POST", "/api/v1/auth/login", models.LoginRequest{
			Username: "",
			Password: "password123",
		}, "", []int{http.StatusBadRequest}},
		{"Empty password returns 400", "POST", "/api/v1/auth/login", models.LoginRequest{
			Username: "testuser",
			Password: "",
		}, "", []int{http.StatusBadRequest}},
		{"Invalid JSON body returns 400", "POST", "/api/v1/auth/login", nil, "{invalid json}", []int{http.StatusBadRequest}},
		{"Whitespace-only username returns 400 or 401", "POST", "/api/v1/auth/login", models.LoginRequest{
			Username: "   ",
			Password: "password123",
		}, "", []int{http.StatusBadRequest, http.StatusUnauthorized}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.rawBody != "" {
				body.WriteString(tt.rawBody)
			} else if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check if the status code is one of the expected codes
			statusOK := false
			for _, code := range tt.expectCodes {
				if w.Code == code {
					statusOK = true
					break
				}
			}
			assert.True(t, statusOK, "Expected status code %v, got %d", tt.expectCodes, w.Code)
		})
	}
}

// ============================================================================
// HTTP STATUS CODES: 401 Unauthorized (Missing/Invalid Auth)
// ============================================================================

func TestHTTPStatusCodes_401Unauthorized(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		authHeader string
	}{
		{"Protected endpoint without token returns 401", "GET", "/api/v1/users", nil, ""},
		{"Invalid credentials returns 401", "POST", "/api/v1/auth/login", models.LoginRequest{
			Username: "nonexistent",
			Password: "wrongpass",
		}, ""},
		{"Invalid auth token returns 401", "GET", "/api/v1/users", nil, "Bearer invalid-token"},
		{"Malformed auth token returns 401", "GET", "/api/v1/users", nil, "Bearer"},
		{"Empty Bearer token returns 401", "GET", "/api/v1/users", nil, "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "Status code mismatch")
		})
	}
}

// ============================================================================
// HTTP STATUS CODES: 403 Forbidden (Insufficient Permissions)
// ============================================================================

func TestHTTPStatusCodes_403Forbidden(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		// Setup endpoint disabled returns 403
		{"Disabled setup endpoint returns 403", "POST", "/api/v1/auth/setup", models.SignupRequest{
			Username: "newadmin",
			Email:    "admin@example.com",
			Password: "password123",
			FullName: "Admin",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, &body)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "Status code mismatch")
		})
	}
}

// ============================================================================
// HTTP STATUS CODES: 404 Not Found (Resource Doesn't Exist)
// ============================================================================

func TestHTTPStatusCodes_404NotFound(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		// Unknown endpoint returns 404 (no auth required for non-existent routes)
		{"Unknown endpoint returns 404", "GET", "/api/v1/nonexistent"},
		// Unknown route under known path
		{"Unknown route returns 404", "GET", "/api/v1/auth/nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Status code mismatch")
		})
	}
}

// TestHTTPStatusCodes_404NotFound_ProtectedEndpoint tests that protected endpoints return 401 (auth required)
// when no auth is provided, rather than 404 (the resource exists but requires auth)
func TestHTTPStatusCodes_404NotFound_ProtectedEndpoint(t *testing.T) {
	router, _, _ := newTestEnv(t)

	validUUID := uuid.New().String()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		// Protected endpoints return 401 when no auth is provided (auth check before route resolution)
		{"Protected collector endpoint without auth returns 401", "GET", "/api/v1/collectors/" + validUUID, http.StatusUnauthorized},
		{"Protected collector delete without auth returns 401", "DELETE", "/api/v1/collectors/" + validUUID, http.StatusUnauthorized},
		// Non-existent routes return 404
		{"Unknown endpoint returns 404", "GET", "/api/v1/nonexistent", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "Status code mismatch")
		})
	}
}
