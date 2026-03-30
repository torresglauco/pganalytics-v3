package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"go.uber.org/zap"
)

// TestRequestIDMiddleware_GeneratesUUID tests that RequestIDMiddleware generates a UUID for each request
func TestRequestIDMiddleware_GeneratesUUID(t *testing.T) {
	// Create a logger for the server
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create a minimal server with required fields
	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	// Create a test router
	router := gin.New()
	router.Use(server.RequestIDMiddleware())

	// Add a simple handler that returns the request ID from context
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Process the request
	router.ServeHTTP(w, req)

	// Verify that the request was processed
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response contains request_id
	var response map[string]interface{}
	_ = response // We'll check the header instead

	// Verify that X-Request-ID header is set
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should not be empty")
}

// TestRequestIDMiddleware_ValidUUIDFormat tests that the generated request ID is a valid UUID format
func TestRequestIDMiddleware_ValidUUIDFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	_, err := uuid.Parse(requestID)
	assert.NoError(t, err, "X-Request-ID should be a valid UUID format")
}

// TestRequestIDMiddleware_RequestIDInContext tests that request ID is stored in context
func TestRequestIDMiddleware_RequestIDInContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = c.GetString("request_id")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	headerRequestID := w.Header().Get("X-Request-ID")

	// Both should be set
	assert.NotEmpty(t, capturedRequestID, "Request ID should be stored in context")
	assert.NotEmpty(t, headerRequestID, "Request ID should be in response header")

	// They should match
	assert.Equal(t, capturedRequestID, headerRequestID, "Context and header request IDs should match")
}

// TestRequestIDMiddleware_UniquePerRequest tests that different requests get different request IDs
func TestRequestIDMiddleware_UniquePerRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Create multiple requests
	requestIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		requestIDs[i] = w.Header().Get("X-Request-ID")
	}

	// Verify all IDs are unique
	for i := 0; i < len(requestIDs); i++ {
		for j := i + 1; j < len(requestIDs); j++ {
			assert.NotEqual(t, requestIDs[i], requestIDs[j], "Request IDs should be unique across different requests")
		}
	}
}

// TestRequestIDMiddleware_HeaderKey tests that X-Request-ID is the correct header name
func TestRequestIDMiddleware_HeaderKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the exact header name
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be present")

	// Verify it's a valid UUID
	_, err := uuid.Parse(requestID)
	assert.NoError(t, err, "X-Request-ID should be a valid UUID")
}

// TestRequestIDMiddleware_ContextKey tests that the context key is "request_id"
func TestRequestIDMiddleware_ContextKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())

	var contextRequestID interface{}
	router.GET("/test", func(c *gin.Context) {
		// Try to get with the key "request_id"
		contextRequestID, _ = c.Get("request_id")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify the context value matches the header
	assert.NotNil(t, contextRequestID, "request_id should be present in context")
	headerRequestID := w.Header().Get("X-Request-ID")
	assert.Equal(t, contextRequestID.(string), headerRequestID, "Context and header request IDs should match")
}

// TestRequestIDMiddleware_PreservesChain tests that middleware doesn't break the request chain
func TestRequestIDMiddleware_PreservesChain(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())

	// Track if handler was called
	handlerCalled := false
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify handler was called
	assert.True(t, handlerCalled, "Handler should be called after middleware")
	assert.Equal(t, http.StatusOK, w.Code, "Should return OK status")
}

// TestRequestIDMiddleware_WithMultipleMiddleware tests that RequestIDMiddleware works with other middleware
func TestRequestIDMiddleware_WithMultipleMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	server := &Server{
		logger: logger,
		config: &config.Config{
			Environment: "test",
		},
	}

	router := gin.New()
	router.Use(server.RequestIDMiddleware())

	// Add another middleware that also uses context
	router.Use(func(c *gin.Context) {
		c.Set("test_value", "test_marker")
		c.Next()
	})

	var requestID, testValue string
	router.GET("/test", func(c *gin.Context) {
		requestID = c.GetString("request_id")
		testValue = c.GetString("test_value")
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify both middleware values are present
	assert.NotEmpty(t, requestID, "Request ID should be set")
	assert.Equal(t, "test_marker", testValue, "Other middleware values should be preserved")
	assert.Equal(t, requestID, w.Header().Get("X-Request-ID"), "Header should match context")
}
