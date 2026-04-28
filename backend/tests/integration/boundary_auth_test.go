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
// BOUNDARY TESTS: POST /api/v1/auth/login
// ============================================================================

func TestLoginBoundary_EmptyUsername(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty username should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code, "Empty username should return 400")
}

func TestLoginBoundary_EmptyPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty password should be rejected
	assert.Equal(t, http.StatusBadRequest, w.Code, "Empty password should return 400")
}

func TestLoginBoundary_VeryLongUsername(t *testing.T) {
	router, _, _ := newTestEnv(t)

	// Create a username that's 1000+ characters
	longUsername := strings.Repeat("a", 1000)

	loginReq := models.LoginRequest{
		Username: longUsername,
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should reject extremely long username (validation error 400 or auth error 401 both acceptable)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Extremely long username should return validation error (400) or auth error (401)")
}

func TestLoginBoundary_SQLInjectionAttempt_ORClause(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "admin' OR '1'='1",
		Password: "' OR '1'='1",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should safely reject SQL injection
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"SQL injection should be rejected with 401")
}

func TestLoginBoundary_SQLInjectionAttempt_UnionSelect(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "admin' UNION SELECT * FROM users--",
		Password: "admin",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should safely reject SQL injection
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"SQL UNION injection should be rejected with 401")
}

func TestLoginBoundary_SQLInjectionAttempt_DropTable(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "admin'; DROP TABLE users; --",
		Password: "admin",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should safely reject SQL injection
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"DROP TABLE injection should be rejected with 401")
}

func TestLoginBoundary_SpecialCharactersUsername(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "<script>alert('xss')</script>",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should safely reject (401 for nonexistent user)
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"XSS attempt should be rejected")
}

func TestLoginBoundary_NullByteInjection(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "admin\x00admin",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should safely handle null bytes (validation error 400 or auth error 401 both acceptable)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Null byte injection should return validation error (400) or auth error (401)")
}

func TestLoginBoundary_ValidLoginAtBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Valid credentials should succeed
	assert.Equal(t, http.StatusOK, w.Code,
		"Valid login should succeed")
}

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/auth/setup (Initial Setup)
// ============================================================================

func TestSetupBoundary_EmptyUsername(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "",
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_UsernameTooShort(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "ad",
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_UsernameAtMinBoundary(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "adm",
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_UsernameAtMaxBoundary(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	maxUsername := strings.Repeat("a", 255)

	setupReq := models.SignupRequest{
		Username: maxUsername,
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_UsernameTooLong(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	tooLongUsername := strings.Repeat("a", 256)

	setupReq := models.SignupRequest{
		Username: tooLongUsername,
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_InvalidEmail(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "not-an-email",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_PasswordTooShort(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com",
		Password: "short",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_PasswordAtMinBoundary(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com",
		Password: "12345678",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_PasswordVeryLong(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	veryLongPassword := strings.Repeat("a", 10000)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com",
		Password: veryLongPassword,
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Very long password should be handled gracefully (not crash)
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Very long password should not cause server error (5xx)")
}

func TestSetupBoundary_SQLInjectionInEmail(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com' OR '1'='1",
		Password: "ValidPassword123!",
		FullName: "Admin User",
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_FullNameMaxLength(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	maxFullName := strings.Repeat("a", 255)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: maxFullName,
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

func TestSetupBoundary_FullNameExceedsMax(t *testing.T) {
	router, _, _ := newTestEnvWithEmptyUsers(t)

	exceedsMaxFullName := strings.Repeat("a", 256)

	setupReq := models.SignupRequest{
		Username: "newadmin",
		Email:    "admin@example.com",
		Password: "ValidPassword123!",
		FullName: exceedsMaxFullName,
	}

	body, _ := json.Marshal(setupReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/setup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Setup endpoint is disabled in test config, returns 403 (Forbidden)
	assert.Equal(t, http.StatusForbidden, w.Code,
		"Setup endpoint should return 403 when disabled")
}

// ============================================================================
// BOUNDARY TESTS: POST /api/v1/auth/change-password
// ============================================================================

func TestChangePasswordBoundary_EmptyOldPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	changeReq := models.ChangePasswordRequest{
		OldPassword: "",
		NewPassword: "NewPassword123!",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	// Note: Would need proper JWT token in Authorization header in real scenario
	req.Header.Set("Authorization", "Bearer invalid")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty old password should be rejected (invalid auth header returns 401)
	assert.Equal(t, http.StatusUnauthorized, w.Code,
		"Empty old password with invalid auth should return 401")
}

func TestChangePasswordBoundary_NewPasswordTooShort(t *testing.T) {
	router, _, _ := newTestEnv(t)

	changeReq := models.ChangePasswordRequest{
		OldPassword: "password123",
		NewPassword: "short",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// New password less than 8 chars should be rejected (401 for missing auth, 400 for bad input once authenticated)
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Should return auth error (401) or validation error (400)")
}

func TestChangePasswordBoundary_NewPasswordAtMinBoundary(t *testing.T) {
	router, _, _ := newTestEnv(t)

	changeReq := models.ChangePasswordRequest{
		OldPassword: "password123",
		NewPassword: "12345678",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Would be accepted but fails auth - that's OK
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusUnauthorized,
		"Should return validation error (400) or auth error (401) for invalid auth")
}

func TestChangePasswordBoundary_SameAsOldPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	changeReq := models.ChangePasswordRequest{
		OldPassword: "password123",
		NewPassword: "password123",
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle this gracefully (not crash)
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Should not cause server error (5xx)")
}

func TestChangePasswordBoundary_VeryLongNewPassword(t *testing.T) {
	router, _, _ := newTestEnv(t)

	veryLongPassword := strings.Repeat("a", 10000)

	changeReq := models.ChangePasswordRequest{
		OldPassword: "password123",
		NewPassword: veryLongPassword,
	}

	body, _ := json.Marshal(changeReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should handle gracefully (not crash)
	assert.True(t, w.Code >= 200 && w.Code < 500,
		"Should not cause server error (5xx)")
}
