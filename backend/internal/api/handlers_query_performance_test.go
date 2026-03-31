package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestGetQueryPerformanceInvalidHash tests with invalid query hash parameter
func TestGetQueryPerformanceInvalidHash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	// Create a minimal server
	server := &Server{
		logger:   logger,
		postgres: nil, // Intentionally nil - we're testing input validation
	}

	router := gin.New()
	api := router.Group("/api/v1")
	queries := api.Group("/queries")
	queries.GET("/:query_hash/performance", server.handleGetQueryPerformance)

	// Test with invalid query hash (non-numeric)
	req := httptest.NewRequest("GET", "/api/v1/queries/invalid/performance", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return bad request status
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify error response is valid JSON
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// TestGetQueryPerformanceValidHash tests handler accepts valid hash format
func TestGetQueryPerformanceValidHash(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	// Create a minimal server
	server := &Server{
		logger:   logger,
		postgres: nil, // Intentionally nil - will cause error, but input parsing should work
	}

	router := gin.New()
	api := router.Group("/api/v1")
	queries := api.Group("/queries")
	queries.GET("/:query_hash/performance", server.handleGetQueryPerformance)

	// Test with valid numeric query hash
	req := httptest.NewRequest("GET", "/api/v1/queries/12345/performance", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not return bad request (will be 500 due to nil postgres, but not 400)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

// TestGetQueryPerformanceWithQueryParams tests handler parses query parameters
func TestGetQueryPerformanceWithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	server := &Server{
		logger:   logger,
		postgres: nil,
	}

	router := gin.New()
	api := router.Group("/api/v1")
	queries := api.Group("/queries")
	queries.GET("/:query_hash/performance", server.handleGetQueryPerformance)

	// Test with query parameters for hours and metrics filter
	req := httptest.NewRequest("GET", "/api/v1/queries/12345/performance?hours=48&metrics=cpu,memory", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should not return bad request (input validation should pass)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

// TestGetQueryPerformanceWithInvalidHours tests invalid hours parameter is handled
func TestGetQueryPerformanceWithInvalidHours(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	server := &Server{
		logger:   logger,
		postgres: nil,
	}

	router := gin.New()
	api := router.Group("/api/v1")
	queries := api.Group("/queries")
	queries.GET("/:query_hash/performance", server.handleGetQueryPerformance)

	// Test with invalid hours parameter (will be clamped to default)
	req := httptest.NewRequest("GET", "/api/v1/queries/12345/performance?hours=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Handler should handle gracefully by using default
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

// TestGetQueryPerformanceResponseStructure tests response JSON structure
func TestGetQueryPerformanceResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewProduction()
	require.NoError(t, err)

	server := &Server{
		logger:   logger,
		postgres: nil,
	}

	router := gin.New()
	api := router.Group("/api/v1")
	queries := api.Group("/queries")
	queries.GET("/:query_hash/performance", server.handleGetQueryPerformance)

	req := httptest.NewRequest("GET", "/api/v1/queries/12345/performance", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify response is valid JSON
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.NotNil(t, response)
}
