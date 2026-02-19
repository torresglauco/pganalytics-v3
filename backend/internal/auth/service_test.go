package auth

import (
	"testing"
	"time"

	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserStore implements UserStore for testing
type MockUserStore struct {
	users map[string]*models.User
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
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
			"inactiveuser": {
				ID:        2,
				Username:  "inactiveuser",
				Email:     "inactive@example.com",
				Role:      "user",
				IsActive:  false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}
}

func (m *MockUserStore) GetUserByUsername(username string) (*models.User, error) {
	user, exists := m.users[username]
	if !exists {
		return nil, nil
	}
	return user, nil
}

func (m *MockUserStore) GetUserByID(id int) (*models.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserStore) UpdateUserLastLogin(userID int, timestamp time.Time) error {
	for _, user := range m.users {
		if user.ID == userID {
			user.LastLogin = &timestamp
			return nil
		}
	}
	return nil
}

// MockCollectorStore implements CollectorStore for testing
type MockCollectorStore struct {
	collectors map[uuid.UUID]*models.Collector
}

func NewMockCollectorStore() *MockCollectorStore {
	return &MockCollectorStore{
		collectors: make(map[uuid.UUID]*models.Collector),
	}
}

func (m *MockCollectorStore) CreateCollector(collector *models.Collector) (uuid.UUID, error) {
	m.collectors[collector.ID] = collector
	return collector.ID, nil
}

func (m *MockCollectorStore) GetCollectorByID(id uuid.UUID) (*models.Collector, error) {
	collector, exists := m.collectors[id]
	if !exists {
		return nil, nil
	}
	return collector, nil
}

func (m *MockCollectorStore) UpdateCollectorStatus(id uuid.UUID, status string) error {
	if collector, exists := m.collectors[id]; exists {
		collector.Status = status
	}
	return nil
}

func (m *MockCollectorStore) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	if collector, exists := m.collectors[id]; exists {
		collector.CertificateThumbprint = &thumbprint
		collector.CertificateExpiresAt = &expiresAt
	}
	return nil
}

// MockTokenStore implements TokenStore for testing
type MockTokenStore struct {
	tokens map[int]*models.APIToken
}

func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{
		tokens: make(map[int]*models.APIToken),
	}
}

func (m *MockTokenStore) CreateAPIToken(token *models.APIToken) (int, error) {
	id := len(m.tokens) + 1
	token.ID = id
	m.tokens[id] = token
	return id, nil
}

func (m *MockTokenStore) GetAPITokenByHash(hash string) (*models.APIToken, error) {
	for _, token := range m.tokens {
		if token.TokenHash == hash {
			return token, nil
		}
	}
	return nil, nil
}

func (m *MockTokenStore) UpdateAPITokenLastUsed(id int, timestamp time.Time) error {
	if token, exists := m.tokens[id]; exists {
		token.LastUsed = &timestamp
	}
	return nil
}

// Tests

func TestAuthService_LoginUser_Success(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Test successful login
	resp, err := authService.LoginUser("testuser", "password123")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "testuser", resp.User.Username)
	assert.Equal(t, "test@example.com", resp.User.Email)
}

func TestAuthService_LoginUser_UserNotFound(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Test login with non-existent user
	resp, err := authService.LoginUser("nonexistent", "password123")

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAuthService_LoginUser_InactiveUser(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Test login with inactive user
	resp, err := authService.LoginUser("inactiveuser", "password123")

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestAuthService_RefreshUserToken_Success(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Generate a refresh token
	user, _ := userStore.GetUserByUsername("testuser")
	refreshToken, _, _ := jm.GenerateUserRefreshToken(user)

	// Test refresh
	resp, err := authService.RefreshUserToken(refreshToken)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "testuser", resp.User.Username)
}

func TestAuthService_RegisterCollector_Success(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Test registration
	req := &models.CollectorRegisterRequest{
		Name:     "test-collector",
		Hostname: "db-server-01",
		Address:  nil,
		Version:  nil,
	}

	resp, err := authService.RegisterCollector(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotZero(t, resp.CollectorID)
	assert.NotEmpty(t, resp.Token)
	assert.NotEmpty(t, resp.Certificate)
	assert.NotEmpty(t, resp.PrivateKey)
}

func TestAuthService_RegisterCollector_InvalidRequest(t *testing.T) {
	jm := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, 30*time.Minute)
	pm := NewPasswordManager()
	cm, _ := NewCertificateManager("", "")
	userStore := NewMockUserStore()
	collectorStore := NewMockCollectorStore()
	tokenStore := NewMockTokenStore()

	authService := NewAuthService(jm, pm, cm, userStore, collectorStore, tokenStore)

	// Test registration with empty name
	req := &models.CollectorRegisterRequest{
		Name:     "",
		Hostname: "db-server-01",
	}

	resp, err := authService.RegisterCollector(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestPasswordManager_HashAndVerify(t *testing.T) {
	pm := NewPasswordManager()
	password := "my-secure-password"

	// Test hashing
	hash, err := pm.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Test verification - correct password
	assert.True(t, pm.VerifyPassword(hash, password))

	// Test verification - wrong password
	assert.False(t, pm.VerifyPassword(hash, "wrong-password"))
}
