package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/torresglauco/pganalytics-v3/backend/internal/session"
)

// ============================================================================
// Mock Session Manager for Testing
// ============================================================================

// MockSessionManager for testing session creation failures
type MockSessionManager struct {
	shouldFail bool
	failErr    error
}

func NewMockSessionManager(shouldFail bool, failErr error) *MockSessionManager {
	return &MockSessionManager{
		shouldFail: shouldFail,
		failErr:    failErr,
	}
}

func (m *MockSessionManager) CreateSession(userID int, ipAddress, userAgent string) (*session.Session, error) {
	if m.shouldFail {
		return nil, m.failErr
	}
	// Success case
	return &session.Session{
		ID:        "test-session-id",
		UserID:    userID,
		Token:     "test-session-token",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (m *MockSessionManager) ValidateSession(sessionID, token string) error {
	if m.shouldFail {
		return m.failErr
	}
	return nil
}

func (m *MockSessionManager) RevokeSession(sessionID string) error {
	return nil
}

func (m *MockSessionManager) RevokeAllUserSessions(userID int) error {
	return nil
}

func (m *MockSessionManager) CleanupExpiredSessions() error {
	return nil
}

func (m *MockSessionManager) RefreshSession(sessionID string) error {
	return nil
}

func (m *MockSessionManager) GetSessionByID(sessionID string) (*session.Session, error) {
	return nil, nil
}

func (m *MockSessionManager) GetUserSessions(userID int) ([]*session.Session, error) {
	return []*session.Session{}, nil
}

func (m *MockSessionManager) GetSessionStats() (*session.SessionStats, error) {
	return &session.SessionStats{}, nil
}

// ============================================================================
// Unit Tests for Session Manager Interface
// ============================================================================

// TestMockSessionManager_Success tests that the mock correctly returns a session
func TestMockSessionManager_Success(t *testing.T) {
	mock := NewMockSessionManager(false, nil)

	sess, err := mock.CreateSession(1, "127.0.0.1", "Mozilla/5.0")
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, 1, sess.UserID)
	assert.Equal(t, "test-session-token", sess.Token)
	assert.Equal(t, "127.0.0.1", sess.IPAddress)
}

// TestMockSessionManager_Failure tests that the mock correctly returns an error
func TestMockSessionManager_Failure(t *testing.T) {
	expectedErr := fmt.Errorf("redis connection failed")
	mock := NewMockSessionManager(true, expectedErr)

	sess, err := mock.CreateSession(1, "127.0.0.1", "Mozilla/5.0")
	assert.Error(t, err)
	assert.Nil(t, sess)
	assert.Equal(t, expectedErr, err)
}

// TestSessionManager_Interface verifies the interface implementation
func TestSessionManager_ImplementsInterface(t *testing.T) {
	mock := NewMockSessionManager(false, nil)

	// Verify the mock implements the interface
	var _ session.ISessionManager = mock

	// All method calls should work
	_, _ = mock.CreateSession(1, "127.0.0.1", "User-Agent")
	_ = mock.ValidateSession("session-id", "token")
	_ = mock.RevokeSession("session-id")
	_ = mock.RevokeAllUserSessions(1)
	_ = mock.RefreshSession("session-id")
	_, _ = mock.GetSessionByID("session-id")
	_, _ = mock.GetUserSessions(1)
	_, _ = mock.GetSessionStats()
	_ = mock.CleanupExpiredSessions()
}

// TestRealSessionManager_Success tests that the real SessionManager works correctly
func TestRealSessionManager_Success(t *testing.T) {
	manager := session.NewSessionManager(nil)

	sess, err := manager.CreateSession(1, "127.0.0.1", "Mozilla/5.0")
	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, 1, sess.UserID)
	assert.NotEmpty(t, sess.Token)
	assert.Equal(t, "127.0.0.1", sess.IPAddress)
	assert.NotZero(t, sess.CreatedAt)
	assert.NotZero(t, sess.ExpiresAt)
}

// TestRealSessionManager_InvalidUserID tests that invalid user ID is rejected
func TestRealSessionManager_InvalidUserID(t *testing.T) {
	manager := session.NewSessionManager(nil)

	sess, err := manager.CreateSession(0, "127.0.0.1", "Mozilla/5.0")
	assert.Error(t, err)
	assert.Nil(t, sess)
	assert.Contains(t, err.Error(), "invalid user ID")
}

// TestSessionCreationFailureHandling_Pattern demonstrates the expected pattern
func TestSessionCreationFailureHandling_Pattern(t *testing.T) {
	// Simulate what the auth handler should do
	userID := 1
	clientIP := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	// Test with successful session creation
	successManager := NewMockSessionManager(false, nil)
	sess, err := successManager.CreateSession(userID, clientIP, userAgent)

	// CORRECT pattern: check for error BEFORE using session
	if err != nil {
		t.Fatal("expected session creation to succeed")
	}
	// Only now is it safe to use sess
	assert.NotNil(t, sess)
	sessionToken := sess.Token
	assert.NotEmpty(t, sessionToken)

	// Test with failed session creation
	failManager := NewMockSessionManager(true, fmt.Errorf("redis unavailable"))
	sess, err = failManager.CreateSession(userID, clientIP, userAgent)

	// CORRECT pattern: check for error and return early
	if err != nil {
		// Should NOT try to use sess when err != nil
		assert.Nil(t, sess)
		// Should return error response here, not continue
		return
	}
	// Should not reach here in failure case
	t.Fatal("should have returned early on error")
}
