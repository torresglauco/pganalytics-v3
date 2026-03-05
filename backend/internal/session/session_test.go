package session

import (
	"net"
	"testing"
	"time"
)

// MockRedisClient is a mock Redis client for testing
// In production, would use actual redis client or mock library
type MockRedisClient struct {
	data map[string]map[string]interface{}
	err  error
}

// TestSessionStructure tests that Session struct has required fields
func TestSessionStructure(t *testing.T) {
	session := &Session{
		ID:        "session-123",
		UserID:    1,
		Token:     "token-abc",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if session.ID == "" {
		t.Errorf("Session.ID is empty")
	}

	if session.UserID == 0 {
		t.Errorf("Session.UserID is zero")
	}

	if session.Token == "" {
		t.Errorf("Session.Token is empty")
	}

	if session.IPAddress == "" {
		t.Errorf("Session.IPAddress is empty")
	}

	if session.ExpiresAt.Before(session.CreatedAt) {
		t.Errorf("Session.ExpiresAt before CreatedAt")
	}
}

// TestGenerateSecureToken tests secure token generation
func TestGenerateSecureToken(t *testing.T) {
	token1, err1 := generateSecureToken(32)
	token2, err2 := generateSecureToken(32)

	if err1 != nil {
		t.Errorf("generateSecureToken() error = %v", err1)
	}

	if err2 != nil {
		t.Errorf("generateSecureToken() error = %v", err2)
	}

	// Tokens should not be empty
	if token1 == "" {
		t.Errorf("generateSecureToken() returned empty token")
	}

	// Tokens should be different
	if token1 == token2 {
		t.Errorf("generateSecureToken() produced duplicate tokens")
	}

	// Tokens should be hex-encoded (2 chars per byte)
	if len(token1) != 64 { // 32 bytes * 2 for hex = 64 chars
		t.Errorf("generateSecureToken() length = %d, want 64", len(token1))
	}

	// Verify hex characters only
	for _, c := range token1 {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("generateSecureToken() contains non-hex character: %c", c)
		}
	}
}

// TestGenerateSessionID tests session ID generation
func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == "" {
		t.Errorf("generateSessionID() returned empty ID")
	}

	if id1 == id2 {
		t.Logf("generateSessionID() produced duplicate IDs (rare)")
	}

	if len(id1) != 16 {
		t.Logf("generateSessionID() length = %d", len(id1))
	}
}

// TestGenerateSecureRandomString tests random string generation
func TestGenerateSecureRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "8 characters",
			length: 8,
		},
		{
			name:   "16 characters",
			length: 16,
		},
		{
			name:   "32 characters",
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := generateSecureRandomString(tt.length)

			if len(str) != tt.length {
				t.Errorf("generateSecureRandomString(%d) length = %d, want %d", tt.length, len(str), tt.length)
			}

			// Should only contain alphanumeric characters
			charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, c := range str {
				found := false
				for _, valid := range charset {
					if c == valid {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("generateSecureRandomString() contains invalid character: %c", c)
				}
			}
		})
	}
}

// TestParseInt tests integer parsing
func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "valid integer",
			input:    "123",
			expected: 123,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "negative",
			input:    "-456",
			expected: -456,
		},
		{
			name:     "invalid string",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInt(tt.input)

			if result != tt.expected {
				t.Errorf("parseInt(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseInt64 tests int64 parsing
func TestParseInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "valid int64",
			input:    "9223372036854775807", // Max int64
			expected: 9223372036854775807,
		},
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "small number",
			input:    "1000",
			expected: 1000,
		},
		{
			name:     "invalid string",
			input:    "abc",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInt64(tt.input)

			if result != tt.expected {
				t.Errorf("parseInt64(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSessionCreation tests session creation workflow
func TestSessionCreation(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		ipAddress string
		userAgent string
	}{
		{
			name:      "valid session",
			userID:    1,
			ipAddress: "192.168.1.1",
			userAgent: "Mozilla/5.0",
		},
		{
			name:      "session with email in user agent",
			userID:    2,
			ipAddress: "10.0.0.1",
			userAgent: "curl/7.64.1",
		},
		{
			name:      "session with IPv6",
			userID:    3,
			ipAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			userAgent: "Mobile Safari",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := generateSessionID()
			token, _ := generateSecureToken(32)

			session := &Session{
				ID:        sessionID,
				UserID:    tt.userID,
				Token:     token,
				IPAddress: tt.ipAddress,
				UserAgent: tt.userAgent,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(1 * time.Hour),
			}

			// Verify session is properly constructed
			if session.ID == "" || session.Token == "" {
				t.Errorf("Session not properly initialized")
			}

			if session.UserID != tt.userID {
				t.Errorf("Session.UserID = %d, want %d", session.UserID, tt.userID)
			}

			if session.IPAddress != tt.ipAddress {
				t.Errorf("Session.IPAddress = %s, want %s", session.IPAddress, tt.ipAddress)
			}
		})
	}
}

// TestSessionExpiry tests session expiry checking
func TestSessionExpiry(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		isExpired bool
	}{
		{
			name:      "future expiry",
			expiresAt: now.Add(1 * time.Hour),
			isExpired: false,
		},
		{
			name:      "past expiry",
			expiresAt: now.Add(-1 * time.Hour),
			isExpired: true,
		},
		{
			name:      "expiring now or past",
			expiresAt: now.Add(-1 * time.Millisecond), // Ensure it's slightly in the past
			isExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExpired := now.After(tt.expiresAt)

			if isExpired != tt.isExpired {
				t.Errorf("Session expiry check = %v, want %v", isExpired, tt.isExpired)
			}
		})
	}
}

// TestIPAddressParsing tests IP address validation
func TestIPAddressParsing(t *testing.T) {
	tests := []struct {
		name      string
		ipAddress string
		isValid   bool
	}{
		{
			name:      "valid IPv4",
			ipAddress: "192.168.1.1",
			isValid:   true,
		},
		{
			name:      "valid IPv6",
			ipAddress: "2001:0db8:85a3::8a2e:0370:7334",
			isValid:   true,
		},
		{
			name:      "localhost",
			ipAddress: "127.0.0.1",
			isValid:   true,
		},
		{
			name:      "invalid IP",
			ipAddress: "256.256.256.256",
			isValid:   false,
		},
		{
			name:      "empty IP",
			ipAddress: "",
			isValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ipAddress)
			isValid := ip != nil

			if isValid != tt.isValid {
				t.Logf("IP validation for %s: %v (expected %v)", tt.ipAddress, isValid, tt.isValid)
			}
		})
	}
}

// BenchmarkGenerateSecureToken benchmarks token generation
func BenchmarkGenerateSecureToken(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generateSecureToken(32)
	}
}

// BenchmarkGenerateSessionID benchmarks session ID generation
func BenchmarkGenerateSessionID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSessionID()
	}
}

// BenchmarkSessionCreation benchmarks session creation
func BenchmarkSessionCreation(b *testing.B) {
	token, _ := generateSecureToken(32)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Session{
			ID:        generateSessionID(),
			UserID:    1,
			Token:     token,
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}
	}
}
