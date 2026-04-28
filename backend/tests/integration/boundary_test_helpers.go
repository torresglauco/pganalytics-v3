package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// newTestEnv creates a new test environment with router and stores.
// This helper reduces duplication across boundary tests by encapsulating the common setup:
// - Creates fresh test stores (UserStore, CollectorStore, TokenStore)
// - Initializes the test server with all necessary services
// - Returns the router for making test requests
//
// Typical usage:
//
//	router, _, _ := newTestEnv(t)  // Ignore unused stores if not needed
//
// Or if you need to interact with stores:
//
//	router, userStore, collectorStore := newTestEnv(t)
func newTestEnv(t *testing.T) (*gin.Engine, *TestUserStore, *TestCollectorStore) {
	t.Helper()

	userStore := NewTestUserStore()
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()
	_, router := createTestServer(userStore, collectorStore, tokenStore)

	return router, userStore, collectorStore
}

// newTestEnvWithEmptyUsers creates a test environment with empty user store.
// Useful for testing registration/setup scenarios where no pre-existing users should exist.
func newTestEnvWithEmptyUsers(t *testing.T) (*gin.Engine, *TestUserStore, *TestCollectorStore) {
	t.Helper()

	userStore := &TestUserStore{users: make(map[string]*models.User)}
	collectorStore := NewTestCollectorStore()
	tokenStore := NewTestTokenStore()
	_, router := createTestServer(userStore, collectorStore, tokenStore)

	return router, userStore, collectorStore
}

// ============================================================================
// MOCK POSTGRES FOR PERMISSION TESTS
// ============================================================================

// MockPostgresDB provides a mock database for permission boundary tests.
// It returns minimal valid data to satisfy auth middleware requirements.
type MockPostgresDB struct {
	users map[int]*models.User
}

// NewMockPostgresDB creates a mock database with predefined test users.
func NewMockPostgresDB() *MockPostgresDB {
	now := time.Now()
	return &MockPostgresDB{
		users: map[int]*models.User{
			1: {
				ID:        1,
				Username:  "testuser",
				Email:     "testuser@test.com",
				Role:      "admin",
				FullName:  "Test Admin",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			2: {
				ID:        2,
				Username:  "regularuser",
				Email:     "regularuser@test.com",
				Role:      "user",
				FullName:  "Regular User",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			3: {
				ID:        3,
				Username:  "vieweruser",
				Email:     "vieweruser@test.com",
				Role:      "viewer",
				FullName:  "Viewer User",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

// GetUserByID returns a mock user for the given ID.
func (m *MockPostgresDB) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, nil
}

// The following methods are stubs to satisfy the storage interface requirements.
// They return nil/empty values as they are not needed for permission boundary tests.

func (m *MockPostgresDB) GetUserByUsername(username string) (*models.User, error) {
	return nil, nil
}

func (m *MockPostgresDB) UpdateUserLastLogin(userID int, timestamp time.Time) error {
	return nil
}

func (m *MockPostgresDB) CreateCollector(collector *models.Collector) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *MockPostgresDB) GetCollectorByID(id uuid.UUID) (*models.Collector, error) {
	return nil, nil
}

func (m *MockPostgresDB) UpdateCollectorStatus(id uuid.UUID, status string) error {
	return nil
}

func (m *MockPostgresDB) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	return nil
}

func (m *MockPostgresDB) CreateAPIToken(token *models.APIToken) (int, error) {
	return 0, nil
}

func (m *MockPostgresDB) GetAPITokenByHash(hash string) (*models.APIToken, error) {
	return nil, nil
}

func (m *MockPostgresDB) UpdateAPITokenLastUsed(id int, timestamp time.Time) error {
	return nil
}

func (m *MockPostgresDB) ValidateRegistrationSecret(ctx context.Context, secret string) (*models.RegistrationSecret, error) {
	return nil, nil
}

func (m *MockPostgresDB) RecordRegistrationSecretUsage(ctx context.Context, secretID int, collectorID string, hostname string, status string, expiresAt *time.Time, ipAddress string) error {
	return nil
}
