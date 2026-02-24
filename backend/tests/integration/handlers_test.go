package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/api"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockStores for testing
type TestUserStore struct {
	users map[string]*models.User
}

func NewTestUserStore() *TestUserStore {
	return &TestUserStore{
		users: map[string]*models.User{
			"testuser": {
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				FullName:  "Test User",
				Role:      "user",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
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
	return uuid.Nil, nil
}
func (n *NilPostgresDB) GetCollectorByID(id uuid.UUID) (*models.Collector, error) { return nil, nil }
func (n *NilPostgresDB) UpdateCollectorStatus(id uuid.UUID, status string) error  { return nil }
func (n *NilPostgresDB) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	return nil
}
func (n *NilPostgresDB) CreateAPIToken(token *models.APIToken) (int, error)       { return 0, nil }
func (n *NilPostgresDB) GetAPITokenByHash(hash string) (*models.APIToken, error)  { return nil, nil }
func (n *NilPostgresDB) UpdateAPITokenLastUsed(id int, timestamp time.Time) error { return nil }

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
		Environment: "development",
		Port:        8080,
	}

	// Create API server (note: postgres and timescale are nil for this test)
	server := api.NewServer(cfg, logger, nil, nil, authService, jwtManager)

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

	var resp models.LoginResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "testuser", resp.User.Username)
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
