package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/api"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// MockStores for testing
type TestUserStore struct {
	users map[string]*models.User
}

func NewTestUserStore() *TestUserStore {
	// Create password manager to hash the test password
	pm := auth.NewPasswordManager()
	// Hash "password123" for the test user
	hash, _ := pm.HashPassword("password123")

	return &TestUserStore{
		users: map[string]*models.User{
			"testuser": {
				ID:           1,
				Username:     "testuser",
				Email:        "test@example.com",
				PasswordHash: hash,
				FullName:     "Test User",
				Role:         "user",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			},
		},
	}
}

func (m *TestUserStore) GetUserByUsername(username string) (*models.User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *TestUserStore) GetUserByID(id int) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *TestUserStore) UpdateUserLastLogin(userID int, timestamp time.Time) error {
	for _, user := range m.users {
		if user.ID == userID {
			user.LastLogin = &timestamp
			return nil
		}
	}
	return nil
}

type TestCollectorStore struct {
	collectors map[uuid.UUID]*models.Collector
}

func NewTestCollectorStore() *TestCollectorStore {
	return &TestCollectorStore{
		collectors: make(map[uuid.UUID]*models.Collector),
	}
}

func (m *TestCollectorStore) CreateCollector(collector *models.Collector) (uuid.UUID, error) {
	m.collectors[collector.ID] = collector
	return collector.ID, nil
}

func (m *TestCollectorStore) GetCollectorByID(id uuid.UUID) (*models.Collector, error) {
	collector, exists := m.collectors[id]
	if !exists {
		return nil, nil
	}
	return collector, nil
}

func (m *TestCollectorStore) UpdateCollectorStatus(id uuid.UUID, status string) error {
	if collector, exists := m.collectors[id]; exists {
		collector.Status = status
	}
	return nil
}

func (m *TestCollectorStore) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	if collector, exists := m.collectors[id]; exists {
		collector.CertificateThumbprint = &thumbprint
		collector.CertificateExpiresAt = &expiresAt
	}
	return nil
}

type TestTokenStore struct {
	tokens map[int]*models.APIToken
}

func NewTestTokenStore() *TestTokenStore {
	return &TestTokenStore{
		tokens: make(map[int]*models.APIToken),
	}
}

func (m *TestTokenStore) CreateAPIToken(token *models.APIToken) (int, error) {
	id := len(m.tokens) + 1
	token.ID = id
	m.tokens[id] = token
	return id, nil
}

func (m *TestTokenStore) GetAPITokenByHash(hash string) (*models.APIToken, error) {
	for _, token := range m.tokens {
		if token.TokenHash == hash {
			return token, nil
		}
	}
	return nil, nil
}

func (m *TestTokenStore) UpdateAPITokenLastUsed(id int, timestamp time.Time) error {
	if token, exists := m.tokens[id]; exists {
		token.LastUsed = &timestamp
	}
	return nil
}

// Nil stores for database access (handlers don't use them in these tests)
type NilPostgresDB struct{}

func (n *NilPostgresDB) GetUserByUsername(username string) (*models.User, error)   { return nil, nil }
func (n *NilPostgresDB) GetUserByID(id int) (*models.User, error)                  { return nil, nil }
func (n *NilPostgresDB) UpdateUserLastLogin(userID int, timestamp time.Time) error { return nil }
func (n *NilPostgresDB) CreateCollector(collector *models.Collector) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (n *NilPostgresDB) GetCollectorByID(id uuid.UUID) (*models.Collector, error) { return nil, nil }
func (n *NilPostgresDB) UpdateCollectorStatus(id uuid.UUID, status string) error  { return nil }
func (n *NilPostgresDB) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	return nil
}
func (n *NilPostgresDB) CreateAPIToken(token *models.APIToken) (int, error)       { return 0, nil }
func (n *NilPostgresDB) GetAPITokenByHash(hash string) (*models.APIToken, error)  { return nil, nil }
func (n *NilPostgresDB) UpdateAPITokenLastUsed(id int, timestamp time.Time) error { return nil }
func (n *NilPostgresDB) ValidateRegistrationSecret(ctx context.Context, secret string) (*models.RegistrationSecret, error) {
	return nil, nil
}
func (n *NilPostgresDB) RecordRegistrationSecretUsage(ctx context.Context, secretID int, collectorID string, hostname string, status string, expiresAt *time.Time, ipAddress string) error {
	return nil
}

// Helper to create test server
func createTestServer(userStore auth.UserStore, collectorStore auth.CollectorStore, tokenStore auth.TokenStore) (*api.Server, *gin.Engine) {
	// Create logger
	logger, _ := zap.NewDevelopment()

	// Create JWT manager
	jwtManager := auth.NewJWTManager(
		"test-secret-key",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	// Create auth services
	passwordManager := auth.NewPasswordManager()
	certManager, _ := auth.NewCertificateManager("", "")

	authService := auth.NewAuthService(
		jwtManager,
		passwordManager,
		certManager,
		userStore,
		collectorStore,
		tokenStore,
	)

	// Create config
	cfg := &config.Config{
		Environment:        "development",
		Port:               8080,
		RegistrationSecret: "test-secret",
	}

	// Create API server (note: postgres, timescale, and secretManager are nil for this test)
	server := api.NewServer(cfg, logger, nil, nil, authService, jwtManager, nil)

	// Create router
	router := gin.New()

	// Register routes
	server.RegisterRoutes(router)

	return server, router
}

// Tests

func TestLoginHandler_Success(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	loginReq := models.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response format - tokens are now in cookies, not in JSON body
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	// Response should have message and user
	assert.Equal(t, "Login successful", resp["message"])
	assert.NotNil(t, resp["user"])
	assert.NotEmpty(t, resp["csrf_token"])

	// Check that auth_token cookie is set
	cookies := w.Result().Cookies()
	authTokenFound := false
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authTokenFound = true
			assert.NotEmpty(t, cookie.Value)
			assert.True(t, cookie.HttpOnly)
			break
		}
	}
	assert.True(t, authTokenFound, "auth_token cookie should be set")
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	loginReq := models.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCollectorRegisterHandler_Success(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	registerReq := models.CollectorRegisterRequest{
		Name:     "test-collector",
		Hostname: "db-server-01",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Registration-Secret", "test-secret")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CollectorRegisterResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)

	require.NoError(t, err)
	assert.NotZero(t, resp.CollectorID)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.Certificate)
	assert.NotEmpty(t, resp.PrivateKey)
}

func TestCollectorRegisterHandler_InvalidRequest(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	registerReq := models.CollectorRegisterRequest{
		Name:     "", // Empty name
		Hostname: "db-server-01",
	}

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/v1/collectors/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHealthHandler(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Health check requires database connections, so this will likely fail
	// but we're testing that the endpoint exists and responds
	assert.True(t, w.Code == http.StatusServiceUnavailable || w.Code == http.StatusOK)
}

func TestVersionHandler(t *testing.T) {
	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()

	_, router := createTestServer(userStore, collectorStore, tokenStore)

	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.NotEmpty(t, resp["version"])
	assert.NotEmpty(t, resp["api"])
}
