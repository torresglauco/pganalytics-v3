package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockPostgresDB for metrics testing
type MockMetricsPostgresDB struct {
	schemaMetrics       *models.SchemaMetricsResponse
	lockMetrics         *models.LockMetricsResponse
	bloatMetrics        *models.BloatMetricsResponse
	cacheMetrics        *models.CacheMetricsResponse
	connectionMetrics   *models.ConnectionMetricsResponse
	extensionMetrics    *models.ExtensionMetricsResponse
	shouldReturnError   bool
	errorMessage        string
}

func (m *MockMetricsPostgresDB) GetSchemaMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.SchemaMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.schemaMetrics, nil
}

func (m *MockMetricsPostgresDB) GetLockMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.LockMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.lockMetrics, nil
}

func (m *MockMetricsPostgresDB) GetBloatMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.BloatMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.bloatMetrics, nil
}

func (m *MockMetricsPostgresDB) GetCacheMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.CacheMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.cacheMetrics, nil
}

func (m *MockMetricsPostgresDB) GetConnectionMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.ConnectionMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.connectionMetrics, nil
}

func (m *MockMetricsPostgresDB) GetExtensionMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.ExtensionMetricsResponse, error) {
	if m.shouldReturnError {
		return nil, fmt.Errorf("database error: %s", m.errorMessage)
	}
	return m.extensionMetrics, nil
}

// Helper function to create test server with mock storage
func createTestMetricsServer(mockDB *MockMetricsPostgresDB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()

	// Add mock metrics handlers
	router.GET("/api/v1/collectors/:id/schema", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetSchemaMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_schema",
			Count:      len(metrics.Tables),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/v1/collectors/:id/locks", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetLockMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_locks",
			Count:      len(metrics.ActiveLocks),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/v1/collectors/:id/bloat", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetBloatMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_bloat",
			Count:      len(metrics.TableBloat),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/v1/collectors/:id/cache-hits", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetCacheMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_cache",
			Count:      len(metrics.TableCacheHit),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/v1/collectors/:id/connections", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetConnectionMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_connections",
			Count:      len(metrics.ConnectionSummary),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.GET("/api/v1/collectors/:id/extensions", func(c *gin.Context) {
		collectorID := c.Param("id")
		metrics, err := mockDB.GetExtensionMetrics(c.Request.Context(), uuid.MustParse(collectorID), nil, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := &models.MetricsResponse{
			MetricType: "pg_extensions",
			Count:      len(metrics.Extensions),
			Timestamp:  time.Now(),
			Data:       metrics,
		}
		c.JSON(http.StatusOK, resp)
	})

	return router
}

// Tests

func TestGetSchemaMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		schemaMetrics: &models.SchemaMetricsResponse{
			Tables: []*models.SchemaTable{
				{
					CollectorID:  collectorID,
					DatabaseName: "testdb",
					SchemaName:   "public",
					TableName:    "users",
					TableType:    "BASE TABLE",
				},
			},
			Columns: []*models.SchemaColumn{
				{
					CollectorID:     collectorID,
					DatabaseName:    "testdb",
					SchemaName:      "public",
					TableName:       "users",
					ColumnName:      "id",
					DataType:        "integer",
					IsNullable:      false,
					OrdinalPosition: 1,
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/schema", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_schema")
	assert.Contains(t, w.Body.String(), "users")
}

func TestGetLockMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	username := "testuser"
	sessionState := "active"
	lockAge := 5.0

	mockDB := &MockMetricsPostgresDB{
		lockMetrics: &models.LockMetricsResponse{
			ActiveLocks: []*models.Lock{
				{
					CollectorID:   collectorID,
					DatabaseName:  "testdb",
					PID:           12345,
					LockType:      "relation",
					Mode:          "AccessExclusiveLock",
					Granted:       true,
					Username:      &username,
					SessionState:  &sessionState,
					LockAgeSeconds: &lockAge,
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/locks", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_locks")
	assert.Contains(t, w.Body.String(), "AccessExclusiveLock")
}

func TestGetBloatMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		bloatMetrics: &models.BloatMetricsResponse{
			TableBloat: []*models.TableBloat{
				{
					CollectorID:      collectorID,
					DatabaseName:     "testdb",
					SchemaName:       "public",
					TableName:        "users",
					DeadTuples:       1000,
					LiveTuples:       50000,
					DeadRatioPercent: 2.0,
					TableSize:        "8192000",
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/bloat", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_bloat")
	assert.Contains(t, w.Body.String(), "users")
}

func TestGetCacheMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		cacheMetrics: &models.CacheMetricsResponse{
			TableCacheHit: []*models.TableCacheHit{
				{
					CollectorID:       collectorID,
					DatabaseName:      "testdb",
					SchemaName:        "public",
					TableName:         "users",
					HeapBlksHit:       95000,
					HeapBlksRead:      5000,
					HeapCacheHitRatio: 95.0,
					IdxBlksHit:        5000,
					IdxBlksRead:       100,
					IdxCacheHitRatio:  98.0,
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/cache-hits", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_cache")
	assert.Contains(t, w.Body.String(), "users")
}

func TestGetConnectionMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	maxAge := 3600.0
	minAge := 10.0

	mockDB := &MockMetricsPostgresDB{
		connectionMetrics: &models.ConnectionMetricsResponse{
			ConnectionSummary: []*models.ConnectionSummary{
				{
					CollectorID:     collectorID,
					DatabaseName:    "testdb",
					ConnectionState: "active",
					ConnectionCount: 15,
					MaxAgeSeconds:   &maxAge,
					MinAgeSeconds:   &minAge,
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/connections", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_connections")
	assert.Contains(t, w.Body.String(), "active")
}

func TestGetExtensionMetrics_Success(t *testing.T) {
	collectorID := uuid.New()

	owner := "postgres"

	mockDB := &MockMetricsPostgresDB{
		extensionMetrics: &models.ExtensionMetricsResponse{
			Extensions: []*models.Extension{
				{
					CollectorID:     collectorID,
					DatabaseName:    "testdb",
					ExtensionName:   "plpgsql",
					ExtensionVersion: "1.0",
					ExtensionOwner:  &owner,
					ExtensionSchema: "pg_catalog",
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/extensions", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pg_extensions")
	assert.Contains(t, w.Body.String(), "plpgsql")
}

func TestMetricsEndpoints_InvalidCollectorID(t *testing.T) {
	// Test with valid UUID format but invalid in database
	validUUID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		schemaMetrics:      &models.SchemaMetricsResponse{},
		lockMetrics:        &models.LockMetricsResponse{},
		bloatMetrics:       &models.BloatMetricsResponse{},
		cacheMetrics:       &models.CacheMetricsResponse{},
		connectionMetrics:  &models.ConnectionMetricsResponse{},
		extensionMetrics:   &models.ExtensionMetricsResponse{},
	}

	router := createTestMetricsServer(mockDB)

	endpoints := []string{
		"/api/v1/collectors/%s/schema",
		"/api/v1/collectors/%s/locks",
		"/api/v1/collectors/%s/bloat",
		"/api/v1/collectors/%s/cache-hits",
		"/api/v1/collectors/%s/connections",
		"/api/v1/collectors/%s/extensions",
	}

	for _, endpoint := range endpoints {
		w := httptest.NewRecorder()
		url := fmt.Sprintf(endpoint, validUUID.String())
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, fmt.Sprintf("endpoint %s should return 200 with valid UUID", endpoint))
	}
}

func TestMetricsEndpoints_DatabaseError(t *testing.T) {
	collectorID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		shouldReturnError: true,
		errorMessage:      "connection timeout",
	}

	router := createTestMetricsServer(mockDB)

	endpoints := []string{
		"/api/v1/collectors/%s/schema",
		"/api/v1/collectors/%s/locks",
		"/api/v1/collectors/%s/bloat",
		"/api/v1/collectors/%s/cache-hits",
		"/api/v1/collectors/%s/connections",
		"/api/v1/collectors/%s/extensions",
	}

	for _, endpoint := range endpoints {
		w := httptest.NewRecorder()
		url := fmt.Sprintf(endpoint, collectorID.String())
		req, _ := http.NewRequest("GET", url, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code, fmt.Sprintf("endpoint %s should return 500 on database error", endpoint))
	}
}

func TestMetricsEndpoints_EmptyResults(t *testing.T) {
	collectorID := uuid.New()

	tests := []struct {
		name      string
		mockDB    *MockMetricsPostgresDB
		endpoint  string
		fieldName string
	}{
		{
			name: "schema with empty tables",
			mockDB: &MockMetricsPostgresDB{
				schemaMetrics: &models.SchemaMetricsResponse{
					Tables:   []*models.SchemaTable{},
					Columns:  []*models.SchemaColumn{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/schema",
			fieldName: "pg_schema",
		},
		{
			name: "locks with empty results",
			mockDB: &MockMetricsPostgresDB{
				lockMetrics: &models.LockMetricsResponse{
					ActiveLocks: []*models.Lock{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/locks",
			fieldName: "pg_locks",
		},
		{
			name: "bloat with empty results",
			mockDB: &MockMetricsPostgresDB{
				bloatMetrics: &models.BloatMetricsResponse{
					TableBloat:  []*models.TableBloat{},
					IndexBloat:  []*models.IndexBloat{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/bloat",
			fieldName: "pg_bloat",
		},
		{
			name: "cache with empty results",
			mockDB: &MockMetricsPostgresDB{
				cacheMetrics: &models.CacheMetricsResponse{
					TableCacheHit: []*models.TableCacheHit{},
					IndexCacheHit: []*models.IndexCacheHit{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/cache-hits",
			fieldName: "pg_cache",
		},
		{
			name: "connections with empty results",
			mockDB: &MockMetricsPostgresDB{
				connectionMetrics: &models.ConnectionMetricsResponse{
					ConnectionSummary: []*models.ConnectionSummary{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/connections",
			fieldName: "pg_connections",
		},
		{
			name: "extensions with empty results",
			mockDB: &MockMetricsPostgresDB{
				extensionMetrics: &models.ExtensionMetricsResponse{
					Extensions: []*models.Extension{},
				},
			},
			endpoint:  "/api/v1/collectors/%s/extensions",
			fieldName: "pg_extensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := createTestMetricsServer(tt.mockDB)

			w := httptest.NewRecorder()
			url := fmt.Sprintf(tt.endpoint, collectorID.String())
			req, _ := http.NewRequest("GET", url, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), tt.fieldName)
			assert.Contains(t, w.Body.String(), "\"count\":0")
		})
	}
}

func TestMetricsEndpoints_ResponseFormat(t *testing.T) {
	collectorID := uuid.New()

	mockDB := &MockMetricsPostgresDB{
		schemaMetrics: &models.SchemaMetricsResponse{
			Tables: []*models.SchemaTable{
				{
					CollectorID:  collectorID,
					DatabaseName: "testdb",
					SchemaName:   "public",
					TableName:    "test_table",
					TableType:    "BASE TABLE",
				},
			},
		},
	}

	router := createTestMetricsServer(mockDB)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/collectors/%s/schema", collectorID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Check standard response format
	assert.Contains(t, w.Body.String(), "metric_type")
	assert.Contains(t, w.Body.String(), "count")
	assert.Contains(t, w.Body.String(), "timestamp")
	assert.Contains(t, w.Body.String(), "data")
}
