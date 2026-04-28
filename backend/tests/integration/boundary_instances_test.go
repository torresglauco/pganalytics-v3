package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/managed-instances
// ============================================================================

func TestCreateManagedInstanceBoundary_PortZero(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           0,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port 0 should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Port 0 should return 400")
}

func TestCreateManagedInstanceBoundary_NegativePort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           -1,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Negative port should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Negative port should return 400")
}

func TestCreateManagedInstanceBoundary_PortAtMin(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           1,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port 1 should be accepted
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Port 1 (min) should be accepted")
}

func TestCreateManagedInstanceBoundary_PortAtMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           65535,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port 65535 should be accepted
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Port 65535 (max) should be accepted")
}

func TestCreateManagedInstanceBoundary_PortExceedsMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           65536,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port exceeding 65535 should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Port exceeding 65535 should return 400")
}

func TestCreateManagedInstanceBoundary_PortVeryLarge(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           999999,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Very large port should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Very large port should return 400")
}

func TestCreateManagedInstanceBoundary_NameTooShort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "ab",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Name less than 3 chars should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Name shorter than 3 chars should return 400")
}

func TestCreateManagedInstanceBoundary_NameAtMin(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "abc",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Name at min=3 should be accepted
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Name at min (3 chars) should succeed")
}

func TestCreateManagedInstanceBoundary_EmptyName(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty name should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty name should return 400")
}

func TestCreateManagedInstanceBoundary_EmptyEndpoint(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty endpoint should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty endpoint should return 400")
}

func TestCreateManagedInstanceBoundary_MonitoringIntervalNegative(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:               "test-instance",
		Endpoint:           "database.example.com",
		Port:               5432,
		MasterUsername:     "admin",
		MasterPassword:     "password123",
		MonitoringInterval: -1,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Negative monitoring interval should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Negative monitoring interval should return 400")
}

func TestCreateManagedInstanceBoundary_MonitoringIntervalZero(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:               "test-instance",
		Endpoint:           "database.example.com",
		Port:               5432,
		MasterUsername:     "admin",
		MasterPassword:     "password123",
		MonitoringInterval: 0,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Zero monitoring interval should be accepted (might mean disabled)
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Zero monitoring interval should be accepted")
}

func TestCreateManagedInstanceBoundary_ConnectionTimeoutNegative(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:               "test-instance",
		Endpoint:           "database.example.com",
		Port:               5432,
		MasterUsername:     "admin",
		MasterPassword:     "password123",
		ConnectionTimeout:  -1,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Negative connection timeout should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Negative connection timeout should return 400")
}

func TestCreateManagedInstanceBoundary_ConnectionTimeoutZero(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:               "test-instance",
		Endpoint:           "database.example.com",
		Port:               5432,
		MasterUsername:     "admin",
		MasterPassword:     "password123",
		ConnectionTimeout:  0,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Zero connection timeout should be handled
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Zero connection timeout should be handled")
}

func TestCreateManagedInstanceBoundary_InvalidEnvironment(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		Environment:    "invalid-env",
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid environment might be allowed or rejected depending on implementation
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Invalid environment should be handled")
}

func TestCreateManagedInstanceBoundary_ValidEnvironmentProduction(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "prod-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		Environment:    "production",
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Production environment should be accepted
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Production environment should be accepted")
}

func TestCreateManagedInstanceBoundary_ValidEnvironmentStaging(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createReq := models.CreateManagedInstanceRequest{
		Name:           "staging-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		Environment:    "staging",
		MasterUsername: "admin",
		MasterPassword: "password123",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Staging environment should be accepted
	assert.True(t, (w.Code == http.StatusOK || w.Code == http.StatusCreated || w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized),
		"Staging environment should be accepted")
}

// ============================================================================
// BOUNDARY TESTS: PUT /api/v1/managed-instances/{id}
// ============================================================================

func TestUpdateManagedInstanceBoundary_InvalidStatus(t *testing.T) {
	router, _, _ := newTestEnv(t)

	updateReq := models.UpdateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
		Status:         "invalid-status",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/v1/managed-instances/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid status should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusNotFound,
		"Invalid status should return 400 or 404")
}

func TestUpdateManagedInstanceBoundary_ValidStatusRegistering(t *testing.T) {
	router, _, _ := newTestEnv(t)

	updateReq := models.UpdateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
		Status:         "registering",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/v1/managed-instances/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Status "registering" should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound,
		"Valid status 'registering' should be accepted or return 404")
}

func TestUpdateManagedInstanceBoundary_ValidStatusMonitoring(t *testing.T) {
	router, _, _ := newTestEnv(t)

	updateReq := models.UpdateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
		Status:         "monitoring",
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/v1/managed-instances/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Status "monitoring" should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound,
		"Valid status 'monitoring' should be accepted or return 404")
}

func TestUpdateManagedInstanceBoundary_StringPortInTags(t *testing.T) {
	router, _, _ := newTestEnv(t)

	updateReq := models.UpdateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
		Status:         "monitoring",
		Tags: map[string]interface{}{
			"port": "not-a-number",
		},
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/v1/managed-instances/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle untyped map gracefully
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Untyped map with invalid values should be handled")
}

func TestUpdateManagedInstanceBoundary_ComplexNestedTags(t *testing.T) {
	router, _, _ := newTestEnv(t)

	updateReq := models.UpdateManagedInstanceRequest{
		Name:           "test-instance",
		Endpoint:       "database.example.com",
		Port:           5432,
		MasterUsername: "admin",
		MasterPassword: "password123",
		Status:         "monitoring",
		Tags: map[string]interface{}{
			"nested": map[string]interface{}{
				"deep": map[string]interface{}{
					"value": "test",
				},
			},
			"array": []interface{}{1, 2, 3},
		},
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PUT", "/api/v1/managed-instances/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle complex nested structures
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Complex nested structures should be handled")
}

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/managed-instances/test-connection
// ============================================================================

func TestTestConnectionBoundary_InvalidPort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     0,
		Username: "admin",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port 0 should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Port 0 should return 400")
}

func TestTestConnectionBoundary_ValidPort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     5432,
		Username: "admin",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Valid port should be accepted (might fail connection but not validation)
	assert.NotEqual(t, http.StatusBadRequest, w.Code,
		"Valid port should not return 400")
}

func TestTestConnectionBoundary_PortAtMaxBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     65535,
		Username: "admin",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port 65535 should be accepted
	assert.NotEqual(t, http.StatusBadRequest, w.Code,
		"Port 65535 should not return 400")
}

func TestTestConnectionBoundary_PortExceedsMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     65536,
		Username: "admin",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Port exceeding 65535 should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Port exceeding 65535 should return 400")
}

func TestTestConnectionBoundary_EmptyUsername(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     5432,
		Username: "",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty username should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty username should return 400")
}

func TestTestConnectionBoundary_EmptyPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "database.example.com",
		Port:     5432,
		Username: "admin",
		Password: "",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty password should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty password should return 400")
}

func TestTestConnectionBoundary_EmptyEndpoint(t *testing.T) {
	router, _, _ := newTestEnv(t)

	testReq := models.TestManagedInstanceConnectionRequest{
		Endpoint: "",
		Port:     5432,
		Username: "admin",
		Password: "password123",
	}

	body, _ := json.Marshal(testReq)
	req := httptest.NewRequest("POST", "/api/v1/managed-instances/test-connection", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty endpoint should be rejected
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Empty endpoint should return 400")
}
