package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/users (Create User)
// ============================================================================

func TestCreateUserBoundary_UsernameEmpty(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty username should be rejected (401 for missing auth, 400 for bad input once authenticated)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Should return auth error (401) or validation error (400)")
}

func TestCreateUserBoundary_UsernameTooShort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "ab",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Username less than 3 chars should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Username shorter than 3 chars should return 400")
}

func TestCreateUserBoundary_UsernameAtMinBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "abc",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Username at min=3 should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"Username at min boundary (3 chars) should succeed")
}

func TestCreateUserBoundary_UsernameAtMaxBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	maxUsername := strings.Repeat("a", 255)

	createUserReq := models.CreateUserRequest{
		Username: maxUsername,
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Username at max=255 should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"Username at max boundary (255 chars) should succeed")
}

func TestCreateUserBoundary_UsernameTooLong(t *testing.T) {
	router, _, _ := newTestEnv(t)

	tooLongUsername := strings.Repeat("a", 256)

	createUserReq := models.CreateUserRequest{
		Username: tooLongUsername,
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Username over 255 should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Username exceeding max length should return 400")
}

func TestCreateUserBoundary_InvalidEmail(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "not-an-email",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid email should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Invalid email format should return 400")
}

func TestCreateUserBoundary_PasswordTooShort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "short",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Password less than 8 chars should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Password shorter than 8 chars should return 400")
}

func TestCreateUserBoundary_PasswordAtMinBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "12345678",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Password at min=8 should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"Password at min boundary (8 chars) should succeed")
}

func TestCreateUserBoundary_InvalidRole(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "superadmin", // Invalid role
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Invalid role should be rejected (401 for missing auth, 400 for bad input once authenticated)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Should return auth error (401) or validation error (400)")
}

func TestCreateUserBoundary_RoleAdmin(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newadmin",
		Email:    "newadmin@example.com",
		Password: "ValidPassword123!",
		FullName: "New Admin",
		Role:     "admin",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// admin role should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"admin role should be accepted")
}

func TestCreateUserBoundary_RoleUser(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// user role should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"user role should be accepted")
}

func TestCreateUserBoundary_RoleViewer(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newviewer",
		Email:    "newviewer@example.com",
		Password: "ValidPassword123!",
		FullName: "New Viewer",
		Role:     "viewer",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// viewer role should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"viewer role should be accepted")
}

func TestCreateUserBoundary_FullNameMaxLength(t *testing.T) {
	router, _, _ := newTestEnv(t)

	maxFullName := strings.Repeat("a", 255)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: maxFullName,
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Full name at max should be accepted (200 OK or 201 Created both acceptable)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"User creation should succeed with 200 OK or 201 Created")
}

func TestCreateUserBoundary_FullNameExceedsMax(t *testing.T) {
	router, _, _ := newTestEnv(t)

	exceedsMaxFullName := strings.Repeat("a", 256)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: exceedsMaxFullName,
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Full name exceeding max should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Full name exceeding max length should return 400")
}

func TestCreateUserBoundary_SQLInjectionInUsername(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "admin' OR '1'='1",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// SQL injection should be safely handled (either accepted as literal string or rejected)
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"SQL injection should be handled safely")
}

func TestCreateUserBoundary_VeryLongPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	veryLongPassword := strings.Repeat("a", 10000)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: veryLongPassword,
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Very long password should be handled gracefully
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Very long password should be handled gracefully")
}

func TestCreateUserBoundary_MissingRole(t *testing.T) {
	router, _, _ := newTestEnv(t)

	// JSON without role field
	jsonData := map[string]interface{}{
		"username":  "newuser",
		"email":     "newuser@example.com",
		"password":  "ValidPassword123!",
		"full_name": "New User",
	}

	body, _ := json.Marshal(jsonData)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Missing required role should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Missing required role field should return 400")
}

func TestCreateUserBoundary_EmptyRole(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!",
		FullName: "New User",
		Role:     "",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty role should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"Empty role field should return 400")
}

func TestCreateUserBoundary_PasswordWithSpecialCharacters(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "P@$$w0rd!#%&*(){}[]<>?/\\|",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Special characters in password should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"Special characters in password should be accepted")
}

func TestCreateUserBoundary_PasswordWithUnicodeCharacters(t *testing.T) {
	router, _, _ := newTestEnv(t)

	createUserReq := models.CreateUserRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "ValidPassword123!™©®",
		FullName: "New User",
		Role:     "user",
	}

	body, _ := json.Marshal(createUserReq)
	req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Unicode characters should be accepted
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated,
		"Unicode characters in password should be accepted")
}
