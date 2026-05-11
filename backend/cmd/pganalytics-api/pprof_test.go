package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPprofEndpointsAvailable(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test server using the DefaultServeMux (where pprof registers)
	testServer := httptest.NewServer(http.DefaultServeMux)
	defer testServer.Close()

	tests := []struct {
		name        string
		endpoint    string
		expectCode  int
		description string
	}{
		{
			name:        "pprof index",
			endpoint:    "/debug/pprof/",
			expectCode:  http.StatusOK,
			description: "Should return profile index",
		},
		{
			name:        "pprof heap profile",
			endpoint:    "/debug/pprof/heap",
			expectCode:  http.StatusOK,
			description: "Should return heap profile",
		},
		{
			name:        "pprof goroutine profile",
			endpoint:    "/debug/pprof/goroutine",
			expectCode:  http.StatusOK,
			description: "Should return goroutine profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the test server URL with the endpoint
			resp, err := http.Get(testServer.URL + tt.endpoint)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", tt.endpoint, err)
			}
			defer resp.Body.Close()

			assert.Equal(t, tt.expectCode, resp.StatusCode, tt.description)
		})
	}
}

func TestPprofCPUProfile(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test server using the DefaultServeMux
	testServer := httptest.NewServer(http.DefaultServeMux)
	defer testServer.Close()

	// Test CPU profile with a short duration
	resp, err := http.Get(testServer.URL + "/debug/pprof/profile?seconds=1")
	if err != nil {
		t.Fatalf("Failed to make request to CPU profile: %v", err)
	}
	defer resp.Body.Close()

	// CPU profile should return 200
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return CPU profile")

	// Content type should be binary
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/octet-stream", "Should return binary profile data")
}

func TestPprofImportPresent(t *testing.T) {
	// This test verifies that the pprof import is present in the main package
	// The import _ "net/http/pprof" registers pprof handlers

	// If pprof endpoints work (as tested above), the import is present
	// This is a compile-time check through the tests above

	// Additional runtime check: verify that pprof handlers are registered
	testServer := httptest.NewServer(http.DefaultServeMux)
	defer testServer.Close()

	// Try to access pprof endpoint - if import is missing, this will fail
	resp, err := http.Get(testServer.URL + "/debug/pprof/")
	if err != nil {
		t.Fatalf("pprof endpoints not available: %v", err)
	}
	defer resp.Body.Close()

	// Should get 200 OK if pprof is imported
	assert.Equal(t, http.StatusOK, resp.StatusCode, "pprof import should be present")
}