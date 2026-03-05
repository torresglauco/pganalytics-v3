package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    int
	Token     string
	IPAddress string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateSecureRandomString generates a secure random alphanumeric string
func generateSecureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	return string(bytes)
}

// parseInt parses a string to an integer
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

// parseInt64 parses a string to an int64
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// SessionManager manages user sessions with Redis backend
type SessionManager struct {
	redisClient interface{} // Would be redis.Client in real implementation
	tokenTTL    time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient interface{}) *SessionManager {
	return &SessionManager{
		redisClient: redisClient,
		tokenTTL:    24 * time.Hour, // Default 24-hour TTL
	}
}

// CreateSession creates a new user session
func (sm *SessionManager) CreateSession(userID int, ipAddress, userAgent string) (*Session, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        generateSessionID(),
		UserID:    userID,
		Token:     token,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.tokenTTL),
	}

	// In a real implementation, would store in Redis
	// For now, just return the session
	return session, nil
}

// ValidateSession validates a session
func (sm *SessionManager) ValidateSession(sessionID, token string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	if token == "" {
		return fmt.Errorf("token is required")
	}

	// In a real implementation, would lookup in Redis
	// For now, just validate format
	if len(sessionID) < 1 || len(token) < 1 {
		return fmt.Errorf("invalid session or token")
	}

	return nil
}

// RevokeSession revokes a session
func (sm *SessionManager) RevokeSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	// In a real implementation, would delete from Redis
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (sm *SessionManager) RevokeAllUserSessions(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// In a real implementation, would delete all user sessions from Redis
	return nil
}

// IsSessionExpired checks if a session has expired
func (s *Session) IsSessionExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// GetRemainingTTL returns the remaining time-to-live for a session
func (s *Session) GetRemainingTTL() time.Duration {
	ttl := s.ExpiresAt.Sub(time.Now())
	if ttl < 0 {
		return 0
	}
	return ttl
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions() error {
	// In a real implementation, would scan Redis for expired sessions
	// and delete them
	return nil
}

// ValidateIPAddress validates an IP address
func ValidateIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ExtractIPFromRequest extracts IP address from request
// Handles X-Forwarded-For, X-Real-IP, and direct RemoteAddr
func ExtractIPFromRequest(forwardedFor, realIP, remoteAddr string) string {
	// Check X-Forwarded-For first (handles proxies)
	if forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, get the first
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ValidateIPAddress(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP
	if realIP != "" && ValidateIPAddress(realIP) {
		return realIP
	}

	// Use RemoteAddr as fallback
	if remoteAddr != "" {
		// Remove port if present
		if idx := strings.LastIndex(remoteAddr, ":"); idx > 0 {
			ip := remoteAddr[:idx]
			if ValidateIPAddress(ip) {
				return ip
			}
		}
		if ValidateIPAddress(remoteAddr) {
			return remoteAddr
		}
	}

	return ""
}

// SessionStats represents session statistics
type SessionStats struct {
	TotalSessions      int
	ActiveSessions     int
	ExpiredSessions    int
	AvgSessionDuration time.Duration
}

// GetSessionStats returns statistics about sessions
func (sm *SessionManager) GetSessionStats() (*SessionStats, error) {
	// In a real implementation, would gather stats from Redis
	return &SessionStats{
		TotalSessions:      0,
		ActiveSessions:     0,
		ExpiredSessions:    0,
		AvgSessionDuration: 0,
	}, nil
}

// RefreshSession refreshes a session's TTL
func (sm *SessionManager) RefreshSession(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	// In a real implementation, would update TTL in Redis
	return nil
}

// GetSessionByID retrieves a session by ID
func (sm *SessionManager) GetSessionByID(sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	// In a real implementation, would retrieve from Redis
	return nil, fmt.Errorf("session not found")
}

// GetUserSessions retrieves all sessions for a user
func (sm *SessionManager) GetUserSessions(userID int) ([]*Session, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// In a real implementation, would query Redis for all user sessions
	return []*Session{}, nil
}

// Helper function to calculate time duration
func calculateDuration(start, end time.Time) time.Duration {
	if start.After(end) {
		start, end = end, start
	}

	// Handle edge cases
	diff := end.Sub(start)
	if diff < 0 {
		diff = 0
	}
	if diff > math.MaxInt64/time.Nanosecond {
		diff = time.Duration(math.MaxInt64)
	}

	return diff
}
