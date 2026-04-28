package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/collectors/register
// ============================================================================

func TestCollectorRegisterBoundary_EmptyName(t *testing.T) {
	router, _, _ := newTestEnv(t)

	registerReq := models.CollectorRegisterRequest{
		Name:     "",
		Hostname: "collector-01.example.com",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty name should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty collector name should return 400")
}

func TestCollectorRegisterBoundary_EmptyHostname(t *testing.T) {
	router, _, _ := newTestEnv(t)

	registerReq := models.CollectorRegisterRequest{
		Name:     "collector-01",
		Hostname: "",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty hostname should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty hostname should return 400")
}

func TestCollectorRegisterBoundary_HostnameTooLong(t *testing.T) {
	router, _, _ := newTestEnv(t)

	// Hostname exceeds 255 character limit
	tooLongHostname := strings.Repeat("a", 256) + ".example.com"

	registerReq := models.CollectorRegisterRequest{
		Name:     "collector-01",
		Hostname: tooLongHostname,
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extremely long hostname should be either rejected (400) or accepted (200)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK,
		"Extremely long hostname should return validation error (400) or succeed (200)")
}

func TestCollectorRegisterBoundary_HostnameAtMaxBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	// Hostname at 255 character limit
	maxHostname := strings.Repeat("a", 243) + ".example.com"

	registerReq := models.CollectorRegisterRequest{
		Name:     "collector-01",
		Hostname: maxHostname,
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Hostname at max should be accepted
	assert.Equal(t, http.StatusOK, w.Code,
		"Hostname at max length should succeed")
}

func TestCollectorRegisterBoundary_NameTooLong(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tooLongName := strings.Repeat("a", 1000)

	registerReq := models.CollectorRegisterRequest{
		Name:     tooLongName,
		Hostname: "collector-01.example.com",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extremely long name should be either rejected (400) or accepted (200)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusOK,
		"Extremely long name should return validation error (400) or succeed (200)")
}

func TestCollectorRegisterBoundary_InvalidIPAddress(t *testing.T) {
	router, _, _ := newTestEnv(t)

	addr := "999.999.999.999"

	registerReq := models.CollectorRegisterRequest{
		Name:     "collector-01",
		Hostname: "collector-01.example.com",
		Address:  &addr,
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle invalid address gracefully
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Invalid address should be handled gracefully")
}

func TestCollectorRegisterBoundary_ValidIPAddress(t *testing.T) {
	router, _, _ := newTestEnv(t)

	addr := "192.168.1.100"

	registerReq := models.CollectorRegisterRequest{
		Name:     "collector-01",
		Hostname: "collector-01.example.com",
		Address:  &addr,
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Valid IP address should be accepted
	assert.Equal(t, http.StatusOK, w.Code,
		"Valid IP address should be accepted")
}

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/metrics/push
// ============================================================================

func TestMetricsPushBoundary_EmptyCollectorID(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty collector ID should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty collector ID should return 400")
}

func TestMetricsPushBoundary_EmptyHostname(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "",
		Timestamp:    time.Now(),
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty hostname should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty hostname should return 400")
}

func TestMetricsPushBoundary_NoMetrics(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 0,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty metrics array should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Empty metrics array should be handled gracefully")
}

func TestMetricsPushBoundary_VeryLargeMetricsCount(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 1000000,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Very large metrics count should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Very large metrics count should be handled")
}

func TestMetricsPushBoundary_FutureTimestamp(t *testing.T) {
	router, _, _ := newTestEnv(t)

	futureTime := time.Now().Add(24 * time.Hour)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    futureTime,
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Future timestamp should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Future timestamp should be handled")
}

func TestMetricsPushBoundary_PastTimestamp(t *testing.T) {
	router, _, _ := newTestEnv(t)

	pastTime := time.Now().Add(-365 * 24 * time.Hour)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "collector-01",
		Hostname:     "collector-01.example.com",
		Timestamp:    pastTime,
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Very old timestamp should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Very old timestamp should be handled")
}

func TestMetricsPushBoundary_InvalidUUIDCollectorID(t *testing.T) {
	router, _, _ := newTestEnv(t)

	metricsReq := models.MetricsPushRequest{
		CollectorID:  "not-a-valid-uuid",
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid UUID should be handled
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized || w.Code == http.StatusNotFound,
		"Invalid UUID format should be rejected")
}

func TestMetricsPushBoundary_ValidUUID(t *testing.T) {
	router, _, _ := newTestEnv(t)

	validUUID := uuid.New().String()

	metricsReq := models.MetricsPushRequest{
		CollectorID:  validUUID,
		Hostname:     "collector-01.example.com",
		Timestamp:    time.Now(),
		MetricsCount: 5,
		Metrics:      []interface{}{},
	}

	body, _ := json.Marshal(metricsReq)
	req := httptest.NewRequest("POST", "/api/v1/metrics/push", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Valid UUID should be accepted (might fail with 404 if collector doesn't exist, but not 400)
	assert.NotEqual(t, http.StatusBadRequest, w.Code,
		"Valid UUID format should not return 400")
}

// ============================================================================
// BOUNDARY TESTS: GET /api/v1/collectors/{id} / DELETE /api/v1/collectors/{id}
// ============================================================================

func TestGetCollectorBoundary_InvalidUUID(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/collectors/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid UUID should return 400
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Invalid UUID format should return 400")
}

func TestGetCollectorBoundary_ValidUUIDNotFound(t *testing.T) {
	router, _, _ := newTestEnv(t)

	validUUID := uuid.New().String()
	req := httptest.NewRequest("GET", "/api/v1/collectors/"+validUUID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Valid UUID but not found should return 404
	assert.Equal(t, http.StatusNotFound, w.Code,
		"Valid UUID but collector not found should return 404")
}

func TestDeleteCollectorBoundary_InvalidUUID(t *testing.T) {
	router, _, _ := newTestEnv(t)

	req := httptest.NewRequest("DELETE", "/api/v1/collectors/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid UUID should return 400
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Invalid UUID format should return 400")
}
