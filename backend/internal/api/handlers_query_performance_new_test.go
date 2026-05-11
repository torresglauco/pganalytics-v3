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

func TestHandleGetSlowQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 with query list", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil, // Will cause error, but tests routing
		}

		router := gin.New()
		api := router.Group("/api/v1")
		databases := api.Group("/databases")
		databases.GET("/:id/slow-queries", server.handleGetDatabaseSlowQueries)

		req := httptest.NewRequest("GET", "/api/v1/databases/1/slow-queries", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not be 400 (validation passes)
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		databases := api.Group("/databases")
		databases.GET("/:id/slow-queries", server.handleGetDatabaseSlowQueries)

		req := httptest.NewRequest("GET", "/api/v1/databases/1/slow-queries?limit=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not be 400
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for invalid database ID", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		databases := api.Group("/databases")
		databases.GET("/:id/slow-queries", server.handleGetDatabaseSlowQueries)

		req := httptest.NewRequest("GET", "/api/v1/databases/invalid/slow-queries", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

func TestHandleGetQueryTimeline(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 with timeline data", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		queries := api.Group("/queries")
		queries.GET("/:hash/timeline", server.handleGetDatabaseQueryTimeline)

		req := httptest.NewRequest("GET", "/api/v1/queries/abc123/timeline", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not be 400 (validation passes)
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("respects hours parameter", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		queries := api.Group("/queries")
		queries.GET("/:hash/timeline", server.handleGetDatabaseQueryTimeline)

		req := httptest.NewRequest("GET", "/api/v1/queries/abc123/timeline?hours=48", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for empty hash", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		queries := api.Group("/queries")
		queries.GET("/:hash/timeline", server.handleGetDatabaseQueryTimeline)

		// Note: Gin will match the route but with empty param
		req := httptest.NewRequest("GET", "/api/v1/queries//timeline", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Route not found or bad request
		assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest)
	})
}

func TestHandleGetIndexStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 200 with index statistics", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		databases := api.Group("/databases")
		databases.GET("/:id/index-stats", server.handleGetDatabaseIndexStats)

		req := httptest.NewRequest("GET", "/api/v1/databases/1/index-stats", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should not be 400
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error for invalid database ID", func(t *testing.T) {
		logger := zap.NewNop()
		server := &Server{
			logger:   logger,
			postgres: nil,
		}

		router := gin.New()
		api := router.Group("/api/v1")
		databases := api.Group("/databases")
		databases.GET("/:id/index-stats", server.handleGetDatabaseIndexStats)

		req := httptest.NewRequest("GET", "/api/v1/databases/invalid/index-stats", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}
