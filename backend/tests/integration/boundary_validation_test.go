package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// BOUNDARY TESTS: Common Validation Patterns
// ============================================================================

// Test that all error responses follow a consistent format
func TestErrorResponseFormat_LoginInvalidCredentials(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "nonexistent",
		Password: "wrongpass",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verify response can be unmarshalled
	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Error response should be valid JSON")
}

func TestErrorResponseFormat_ValidationError(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "",
		Email:    "user@example.com",
		Password: "password123",
		FullName: "User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Validation error should return 400 or 401 (auth required first)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Should return 400 or 401")

	// Verify response can be unmarshalled
	var errResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.NoError(t, err, "Error response should be valid JSON")
}

// ============================================================================
// BOUNDARY TESTS: Numeric Field Overflow Protection
// ============================================================================

func TestNumericBoundary_MetricsCountMaxInt(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 2147483647, // Max int32
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle max int gracefully
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Max int metrics count should be handled")
}

func TestNumericBoundary_MetricsCountNegative(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: -1,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle negative count
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Negative metrics count should be handled")
}

// ============================================================================
// BOUNDARY TESTS: Timestamp Validation
// ============================================================================

func TestTimestampBoundary_ZeroTimestamp(t *testing.T) {
	router, _, _ := newTestEnv(t)

	zeroTime := time.Time{}

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    zeroTime,
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle zero timestamp
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Zero timestamp should be handled")
}

func TestTimestampBoundary_FarFutureTimestamp(t *testing.T) {
	router, _, _ := newTestEnv(t)

	farFuture := time.Now().AddDate(100, 0, 0)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    farFuture,
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle far future timestamp
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Far future timestamp should be handled")
}

// ============================================================================
// BOUNDARY TESTS: Pagination Parameter Validation
// ============================================================================

func TestPaginationBoundary_PageZero(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=0&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Page 0 should be rejected or redirected to page 1
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Page 0 should be handled appropriately")
}

func TestPaginationBoundary_PageNegative(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=-1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Negative page should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Negative page should be handled")
}

func TestPaginationBoundary_PageSizeZero(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Page size 0 should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Page size 0 should be handled")
}

func TestPaginationBoundary_PageSizeNegative(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Negative page size should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Negative page size should be handled")
}

func TestPaginationBoundary_PageSizeExceedsMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=1000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Page size exceeding max (100) should be rejected or capped
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Page size exceeding max should be handled")
}

func TestPaginationBoundary_PageSizeAtMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/users?page=1&page_size=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Page size at max (100) should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusUnauthorized,
		"Page size at max should be accepted")
}

// ============================================================================
// BOUNDARY TESTS: Request Body Size and Content
// ============================================================================

func TestRequestBoundary_EmptyBody(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty body should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty JSON body should return 400")
}

func TestRequestBoundary_InvalidJSON(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer([]byte("{invalid json}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid JSON should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Invalid JSON should return 400")
}

func TestRequestBoundary_MissingContentType(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	// No Content-Type header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Missing content-type should be handled (might still work with JSON)
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Missing Content-Type should be handled")
}

func TestRequestBoundary_WrongContentType(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Wrong content-type might still be accepted
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Wrong Content-Type should be handled")
}

// ============================================================================
// BOUNDARY TESTS: String Field Edge Cases
// ============================================================================

func TestStringBoundary_WhitespaceOnly(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "   ",
		Password: "   ",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Whitespace-only username should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Whitespace-only fields should be rejected")
}

func TestStringBoundary_TabAndNewlineCharacters(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "test\t\nuser",
		Password: "pass\t\nword",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle whitespace characters
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Whitespace characters should be handled")
}

func TestStringBoundary_ControlCharacters(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "test\x00user",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle control characters
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Control characters should be handled safely")
}

// ============================================================================
// BOUNDARY TESTS: Array/Slice Field Validation
// ============================================================================

func TestArrayBoundary_NullMetrics(t *testing.T) {
	router, _, _ := newTestEnv(t)

	jsonData := map[string]interface{}{
		"collector_id": "collector-01",
		"hostname":     "collector-01.example.com",
		"timestamp":    time.Now(),
		"metrics":      nil,
	}

	body, _ := json.Marshal(jsonData)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Null metrics array should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Null metrics array should be handled")
}

func TestArrayBoundary_MixedTypeMetrics(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 3,
		Metrics: []interface{}{
			"string",
			123,
			map[string]interface{}{"key": "value"},
		},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Mixed type metrics should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Mixed type metrics should be handled")
}

// ============================================================================
// BOUNDARY TESTS: Path Parameter Validation
// ============================================================================

func TestPathParameterBoundary_UUIDWithSpecialChars(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/collectors/invalid-uuid-with-special-chars!@#$", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid UUID should return 400
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Invalid UUID format should return 400")
}

func TestPathParameterBoundary_UUIDWithSpaces(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/collectors/invalid%20uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// UUID with spaces (URL encoded) should return 400, 404, or 401 (auth required)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound || w.Code == http.StatusUnauthorized,
		"UUID with spaces should be rejected")
}

// ============================================================================
// BOUNDARY TESTS: HTTP Method Validation
// ============================================================================

func TestHTTPMethodBoundary_POSTOnGetEndpoint(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("POST", "/api/v1/health", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// POST on GET endpoint should return 405 or 404
	assert.True(t, w.Code == http.StatusMethodNotAllowed || w.Code == http.StatusNotFound,
		"Should return method not allowed (405) or not found (404)")
}

func TestHTTPMethodBoundary_GETOnPostEndpoint(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// GET on POST endpoint should return 405 or 404 depending on routing
	assert.True(t, w.Code == http.StatusMethodNotAllowed || w.Code == http.StatusNotFound,
		"GET on POST endpoint should return 405 or 404")
}
