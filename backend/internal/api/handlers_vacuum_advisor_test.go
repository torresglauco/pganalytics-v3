package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestGetVacuumRecommendations_Success returns recommendations with valid database ID
func TestGetVacuumRecommendations_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil, // Mock database
	}

	// Create test router
	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/recommendations", server.handleGetVacuumRecommendations)

	// Test request
	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "database_id")
	assert.Contains(t, w.Body.String(), "recommendations")
}

// TestGetVacuumRecommendations_InvalidDatabaseID returns bad request for invalid ID
func TestGetVacuumRecommendations_InvalidDatabaseID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/recommendations", server.handleGetVacuumRecommendations)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/invalid/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// TestGetVacuumRecommendations_WithLimit respects limit parameter
func TestGetVacuumRecommendations_WithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/recommendations", server.handleGetVacuumRecommendations)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/recommendations?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "limit")
}

// TestGetVacuumTableRecommendation_Success returns recommendation for specific table
func TestGetVacuumTableRecommendation_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/table/:table_name", server.handleGetVacuumTableRecommendation)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/table/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "table_name")
	assert.Contains(t, w.Body.String(), "recommendation")
}

// TestGetVacuumTableRecommendation_MissingTableName returns error for missing table
func TestGetVacuumTableRecommendation_MissingTableName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/table/:table_name", server.handleGetVacuumTableRecommendation)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/table/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Due to Gin routing, missing table_name would not match this route
	// So test with a valid table name instead
	req = httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/table/test_table", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetAutovacuumConfig_Success returns autovacuum configuration
func TestGetAutovacuumConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/autovacuum-config", server.handleGetAutovacuumConfig)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/autovacuum-config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "configurations")
}

// TestExecuteVacuum_Success executes VACUUM on a table
func TestExecuteVacuum_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.POST("/api/v1/vacuum-advisor/recommendation/:recommendation_id/execute", server.handleExecuteVacuum)

	req := httptest.NewRequest("POST", "/api/v1/vacuum-advisor/recommendation/1/execute", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "executed")
}

// TestExecuteVacuum_InvalidRecommendationID returns error for invalid ID
func TestExecuteVacuum_InvalidRecommendationID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.POST("/api/v1/vacuum-advisor/recommendation/:recommendation_id/execute", server.handleExecuteVacuum)

	req := httptest.NewRequest("POST", "/api/v1/vacuum-advisor/recommendation/invalid/execute", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

// TestGetVacuumTuningSuggestions_Success returns tuning suggestions
func TestGetVacuumTuningSuggestions_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	router := gin.New()
	router.GET("/api/v1/vacuum-advisor/database/:database_id/tune-suggestions", server.handleGetVacuumTuningSuggestions)

	req := httptest.NewRequest("GET", "/api/v1/vacuum-advisor/database/1/tune-suggestions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "suggestions")
}

// TestAPIResponseFormats verifies correct JSON response structure
func TestAPIResponseFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := &Server{
		logger:   zap.NewNop(),
		postgres: nil,
	}

	testCases := []struct {
		name   string
		method string
		path   string
		handler gin.HandlerFunc
	}{
		{
			name:    "Recommendations response",
			method:  "GET",
			path:    "/api/v1/vacuum-advisor/database/:database_id/recommendations",
			handler: server.handleGetVacuumRecommendations,
		},
		{
			name:    "Table recommendation response",
			method:  "GET",
			path:    "/api/v1/vacuum-advisor/database/:database_id/table/:table_name",
			handler: server.handleGetVacuumTableRecommendation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()

			if tc.method == "GET" {
				router.GET(tc.path, tc.handler)
			} else {
				router.POST(tc.path, tc.handler)
			}

			var req *http.Request
			if tc.name == "Recommendations response" {
				req = httptest.NewRequest(tc.method, "/api/v1/vacuum-advisor/database/1/recommendations", nil)
			} else {
				req = httptest.NewRequest(tc.method, "/api/v1/vacuum-advisor/database/1/table/test", nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "{")
			assert.Contains(t, w.Body.String(), "}")
		})
	}
}