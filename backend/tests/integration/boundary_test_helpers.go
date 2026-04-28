package integration

import (
	"testing"

	"github.com/gin-gonic/gin"
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
