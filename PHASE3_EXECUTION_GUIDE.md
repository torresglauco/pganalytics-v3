# Phase 3 (v3.3.0) Execution Guide

## Quick Start: Enterprise Features Implementation

This guide covers the detailed execution steps for Phase 3, which adds enterprise-grade authentication, encryption, HA/failover, and audit logging.

---

## Week 1-2: Enterprise Authentication (LDAP/SAML/OAuth/MFA)

### 1. LDAP/Active Directory Support

**File**: `/backend/internal/auth/ldap.go` (new)

```go
package auth

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
)

// LDAPConnector handles LDAP/Active Directory authentication
type LDAPConnector struct {
	serverURL       string
	bindDN          string
	bindPassword    string
	userSearchBase  string
	groupSearchBase string
	groupToRoleMap  map[string]string
	tlsConfig       *tls.Config
	mu              sync.RWMutex
	conn            *ldap.Conn
}

// NewLDAPConnector creates a new LDAP connector
func NewLDAPConnector(serverURL, bindDN, bindPassword, userSearchBase, groupSearchBase string, groupToRoleMap map[string]string, tlsConfig *tls.Config) *LDAPConnector {
	return &LDAPConnector{
		serverURL:       serverURL,
		bindDN:          bindDN,
		bindPassword:    bindPassword,
		userSearchBase:  userSearchBase,
		groupSearchBase: groupSearchBase,
		groupToRoleMap:  groupToRoleMap,
		tlsConfig:       tlsConfig,
	}
}

// AuthenticateUser authenticates a user against LDAP
func (lc *LDAPConnector) AuthenticateUser(username, password string) (map[string]interface{}, error) {
	conn, err := lc.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	// Bind with service account
	err = conn.Bind(lc.bindDN, lc.bindPassword)
	if err != nil {
		return nil, apperrors.Unauthorized("LDAP service account bind failed", "")
	}

	// Search for user
	searchRequest := ldap.NewSearchRequest(
		lc.userSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=person)(uid=%s))", ldap.EscapeFilter(username)),
		[]string{"dn", "uid", "mail", "displayName", "memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, apperrors.Unauthorized("User lookup failed", "")
	}

	if len(sr.Entries) != 1 {
		return nil, apperrors.InvalidCredentials()
	}

	userDN := sr.Entries[0].DN

	// Try to bind as user to verify password
	conn2, err := lc.dial()
	if err != nil {
		return nil, fmt.Errorf("failed to create user connection: %w", err)
	}
	defer conn2.Close()

	err = conn2.Bind(userDN, password)
	if err != nil {
		return nil, apperrors.InvalidCredentials()
	}

	// Extract user attributes
	email := sr.Entries[0].GetAttributeValue("mail")
	displayName := sr.Entries[0].GetAttributeValue("displayName")
	if displayName == "" {
		displayName = username
	}

	// Get user groups
	groups := sr.Entries[0].GetAttributeValues("memberOf")
	role := lc.resolveRole(groups)

	return map[string]interface{}{
		"username":   username,
		"email":      email,
		"fullName":   displayName,
		"role":       role,
		"groups":     groups,
		"ldapDN":     userDN,
		"source":     "ldap",
	}, nil
}

// SyncUserGroups syncs LDAP group membership for a user
func (lc *LDAPConnector) SyncUserGroups(username string) ([]string, error) {
	conn, err := lc.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	// Bind with service account
	err = conn.Bind(lc.bindDN, lc.bindPassword)
	if err != nil {
		return nil, fmt.Errorf("service account bind failed: %w", err)
	}

	// Search for user
	searchRequest := ldap.NewSearchRequest(
		lc.userSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=person)(uid=%s))", ldap.EscapeFilter(username)),
		[]string{"memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found")
	}

	return sr.Entries[0].GetAttributeValues("memberOf"), nil
}

// GetUserAttributes retrieves user attributes from LDAP
func (lc *LDAPConnector) GetUserAttributes(username string, attributes []string) (map[string][]string, error) {
	conn, err := lc.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	// Bind with service account
	err = conn.Bind(lc.bindDN, lc.bindPassword)
	if err != nil {
		return nil, fmt.Errorf("service account bind failed: %w", err)
	}

	// Search for user
	searchRequest := ldap.NewSearchRequest(
		lc.userSearchBase,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=person)(uid=%s))", ldap.EscapeFilter(username)),
		attributes,
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("user search failed: %w", err)
	}

	if len(sr.Entries) != 1 {
		return nil, fmt.Errorf("user not found")
	}

	result := make(map[string][]string)
	for _, attr := range attributes {
		result[attr] = sr.Entries[0].GetAttributeValues(attr)
	}

	return result, nil
}

// Internal helper methods

func (lc *LDAPConnector) dial() (*ldap.Conn, error) {
	url := strings.TrimPrefix(lc.serverURL, "ldap://")
	url = strings.TrimPrefix(url, "ldaps://")

	if strings.HasPrefix(lc.serverURL, "ldaps://") {
		return ldap.DialTLS("tcp", url, lc.tlsConfig)
	}
	return ldap.Dial("tcp", url)
}

func (lc *LDAPConnector) getConnection() (*ldap.Conn, error) {
	return lc.dial()
}

func (lc *LDAPConnector) resolveRole(groups []string) string {
	// Check group-to-role mapping
	for _, group := range groups {
		if role, ok := lc.groupToRoleMap[group]; ok {
			return role
		}
	}
	return "viewer" // Default role
}

// Close closes the LDAP connection
func (lc *LDAPConnector) Close() error {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if lc.conn != nil {
		lc.conn.Close()
		lc.conn = nil
	}
	return nil
}
```

**Implementation Notes**:
1. Replace `github.com/go-ldap/ldap/v3` with actual import
2. Add to `go.mod`: `go get github.com/go-ldap/ldap/v3`
3. Handle TLS configuration in config loader
4. Add comprehensive error logging

### 2. Update Config System

**File**: `/backend/internal/config/config.go` (modify)

```go
// Add to Config struct:

// LDAP Configuration
LDAPEnabled        bool
LDAPServerURL      string
LDAPBindDN         string
LDAPBindPassword   string
LDAPUserSearchBase string
LDAPGroupSearchBase string
LDAPGroupToRoleMap map[string]string // JSON-encoded

// SAML Configuration
SAMLEnabled        bool
SAMLCertPath       string
SAMLKeyPath        string
SAMLIDPUrl         string
SAMLEntityID       string

// OAuth Configuration
OAuthEnabled       bool
OAuthProviders     string // JSON-encoded

// MFA Configuration
MFAEnabled         bool
MFATOTPEnabled     bool
MFASMSEnabled      bool
MFASMSProvider     string // twilio|sns
MFASMSAPIKey       string // Encrypted

// Session Configuration
SessionBackend     string // redis|memory
SessionTTL         time.Duration
SessionInactivityTimeout time.Duration
```

**Implementation Steps**:
1. Add environment variable parsing
2. Add validation for auth settings
3. Create example `.env.production`
4. Document all new config options

### 3. Database Migrations

**File**: `/backend/migrations/011_enterprise_auth.sql` (new)

```sql
-- User MFA Methods table
CREATE TABLE user_mfa_methods (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- totp, sms, backup_codes
    secret_encrypted BYTEA, -- Base32-encoded TOTP secret
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, type)
);

CREATE INDEX idx_user_mfa_methods_user_id ON user_mfa_methods(user_id);

-- User Backup Codes table
CREATE TABLE user_backup_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, code_hash)
);

CREATE INDEX idx_user_backup_codes_user_id ON user_backup_codes(user_id);

-- User Sessions table (distributed)
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_token_hash)
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- OAuth Provider Configuration
CREATE TABLE oauth_providers (
    id BIGSERIAL PRIMARY KEY,
    provider_name VARCHAR(100) NOT NULL UNIQUE, -- google, github, azure_ad
    client_id_encrypted BYTEA NOT NULL,
    client_secret_encrypted BYTEA NOT NULL,
    config_json JSONB,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Audit table for auth events (early warning of attacks)
CREATE TABLE auth_events (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL, -- login_success, login_failed, mfa_setup, logout
    ip_address INET,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_events_user_id ON auth_events(user_id);
CREATE INDEX idx_auth_events_created_at ON auth_events(created_at);
```

### 4. Session Management

**File**: `/backend/internal/session/session.go` (new)

```go
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Session represents a user session
type Session struct {
	ID        string
	UserID    int
	Token     string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionManager manages user sessions
type SessionManager struct {
	redis *redis.Client
	ttl   time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(redis *redis.Client, ttl time.Duration) *SessionManager {
	return &SessionManager{
		redis: redis,
		ttl:   ttl,
	}
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userID int, ipAddress, userAgent string) (*Session, error) {
	// Generate secure random token
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	sessionID := generateSessionID()
	now := time.Now()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     token,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.ttl),
	}

	// Store in Redis
	key := fmt.Sprintf("session:%s", sessionID)
	data := map[string]interface{}{
		"user_id":    userID,
		"token":      token,
		"ip_address": ipAddress,
		"user_agent": userAgent,
		"created_at": now.Unix(),
		"expires_at": session.ExpiresAt.Unix(),
	}

	err = sm.redis.HSet(ctx, key, data).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	// Set expiration
	sm.redis.Expire(ctx, key, sm.ttl)

	return session, nil
}

// ValidateSession validates a session
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionID, token string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	data, err := sm.redis.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("session not found")
	}

	// Verify token matches
	if data["token"] != token {
		return nil, fmt.Errorf("token mismatch")
	}

	// Check expiration
	expiresAtUnix := parseInt64(data["expires_at"])
	if expiresAtUnix < time.Now().Unix() {
		sm.RevokeSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return &Session{
		ID:        sessionID,
		UserID:    parseInt(data["user_id"]),
		Token:     token,
		IPAddress: data["ip_address"],
		UserAgent: data["user_agent"],
	}, nil
}

// RevokeSession revokes a session
func (sm *SessionManager) RevokeSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sm.redis.Del(ctx, key).Err()
}

// RevokeAllUserSessions revokes all sessions for a user
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID int) error {
	// In production, use Redis pattern matching
	pattern := "session:*"
	iter := sm.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, _ := sm.redis.HGetAll(ctx, key).Result()
		if parseInt(data["user_id"]) == userID {
			sm.redis.Del(ctx, key)
		}
	}
	return iter.Err()
}

// Internal helpers

func generateSessionID() string {
	return generateSecureRandomString(16)
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSecureRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func parseInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}
```

### 5. Integration with Auth Service

**File**: `/backend/internal/auth/service.go` (modify)

Add LDAP support to `LoginUser` method:

```go
func (as *AuthService) LoginUserLDAP(username, password string) (*models.LoginResponse, error) {
	if !as.config.LDAPEnabled {
		return nil, apperrors.Unauthorized("LDAP not enabled", "")
	}

	// Authenticate against LDAP
	ldapUser, err := as.ldapConnector.AuthenticateUser(username, password)
	if err != nil {
		return nil, err
	}

	// Get or create user in database
	user, err := as.userStore.GetUserByUsername(username)
	if err != nil || user == nil {
		// Create new user from LDAP
		user = &models.User{
			Username: username,
			Email:    ldapUser["email"].(string),
			FullName: ldapUser["fullName"].(string),
			Role:     ldapUser["role"].(string),
			IsActive: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Note: store actual password hash for local fallback
	}

	// Generate tokens
	accessToken, expiresAt, err := as.JWTManager.GenerateUserToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := as.JWTManager.GenerateUserRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user,
	}, nil
}
```

---

## Week 2-3: Encryption at Rest & Key Management

### Implementation will continue in next message

---

## Deployment Checklist

### Pre-Deployment
- [ ] Feature flags disabled for new auth methods
- [ ] Config validated against test LDAP server
- [ ] Database migrations tested on staging
- [ ] Fallback login (JWT) tested and working
- [ ] Backup and rollback procedures documented

### Deployment
- [ ] Deploy code to staging
- [ ] Run migrations on staging
- [ ] Enable LDAP for test users only
- [ ] Monitor auth logs for errors
- [ ] Verify JWT login still works as fallback

### Post-Deployment
- [ ] Collect feedback from test users
- [ ] Monitor API latency (LDAP calls add ~500ms)
- [ ] Check audit logs for any issues
- [ ] Prepare rollback procedure
- [ ] Schedule gradual rollout to all users

---

## Testing Checklist

- [ ] LDAP auth with valid credentials
- [ ] LDAP auth with invalid credentials
- [ ] LDAP group-to-role mapping
- [ ] Session creation and validation
- [ ] Session expiration
- [ ] Concurrent session management
- [ ] Load test with 100+ simultaneous sessions
- [ ] Fallback to JWT if LDAP fails

