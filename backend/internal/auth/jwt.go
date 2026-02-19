package auth

import (
	"fmt"
	"time"

	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents JWT claims
type Claims struct {
	UserID   int       `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// CollectorClaims represents JWT claims for collectors
type CollectorClaims struct {
	CollectorID string `json:"collector_id"`
	Hostname    string `json:"hostname"`
	Type        TokenType `json:"type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secret                  string
	accessTokenExpiration   time.Duration
	refreshTokenExpiration  time.Duration
	collectorTokenExpiration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(
	secret string,
	accessExpiration, refreshExpiration, collectorExpiration time.Duration,
) *JWTManager {
	return &JWTManager{
		secret:                   secret,
		accessTokenExpiration:    accessExpiration,
		refreshTokenExpiration:   refreshExpiration,
		collectorTokenExpiration: collectorExpiration,
	}
}

// ============================================================================
// USER TOKEN OPERATIONS
// ============================================================================

// GenerateUserToken generates a JWT token for a user
func (jm *JWTManager) GenerateUserToken(user *models.User) (string, *time.Time, error) {
	if user == nil {
		return "", nil, apperrors.BadRequest("Invalid user", "User cannot be nil")
	}

	expirationTime := time.Now().Add(jm.accessTokenExpiration)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Type:     TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("user:%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.secret))
	if err != nil {
		return "", nil, apperrors.InternalServerError(
			"Token generation failed",
			fmt.Sprintf("Failed to sign token: %v", err),
		)
	}

	return tokenString, &expirationTime, nil
}

// GenerateUserRefreshToken generates a refresh token for a user
func (jm *JWTManager) GenerateUserRefreshToken(user *models.User) (string, *time.Time, error) {
	if user == nil {
		return "", nil, apperrors.BadRequest("Invalid user", "User cannot be nil")
	}

	expirationTime := time.Now().Add(jm.refreshTokenExpiration)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Type:     TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("user:%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.secret))
	if err != nil {
		return "", nil, apperrors.InternalServerError(
			"Token generation failed",
			fmt.Sprintf("Failed to sign refresh token: %v", err),
		)
	}

	return tokenString, &expirationTime, nil
}

// ValidateUserToken validates and parses a user JWT token
func (jm *JWTManager) ValidateUserToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.InvalidToken(
					fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]),
				)
			}
			return []byte(jm.secret), nil
		},
	)

	if err != nil {
		return nil, apperrors.InvalidToken(fmt.Sprintf("Token parsing failed: %v", err))
	}

	if !token.Valid {
		return nil, apperrors.InvalidToken("Token is invalid")
	}

	// Verify token type
	if claims.Type != TokenTypeAccess {
		return nil, apperrors.InvalidToken("This is not an access token")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, apperrors.TokenExpired()
	}

	return claims, nil
}

// ValidateUserRefreshToken validates a refresh token
func (jm *JWTManager) ValidateUserRefreshToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.InvalidToken(
					fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]),
				)
			}
			return []byte(jm.secret), nil
		},
	)

	if err != nil {
		return nil, apperrors.InvalidToken(fmt.Sprintf("Token parsing failed: %v", err))
	}

	if !token.Valid {
		return nil, apperrors.InvalidToken("Token is invalid")
	}

	// Verify token type
	if claims.Type != TokenTypeRefresh {
		return nil, apperrors.InvalidToken("This is not a refresh token")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, apperrors.TokenExpired()
	}

	return claims, nil
}

// RefreshUserToken generates a new access token from a refresh token
func (jm *JWTManager) RefreshUserToken(refreshTokenString string, user *models.User) (string, *time.Time, error) {
	// Validate refresh token
	claims, err := jm.ValidateUserRefreshToken(refreshTokenString)
	if err != nil {
		return "", nil, err
	}

	// Verify the user ID matches
	if claims.UserID != user.ID {
		return "", nil, apperrors.Unauthorized(
			"Token user mismatch",
			"Refresh token belongs to a different user",
		)
	}

	// Generate new access token
	return jm.GenerateUserToken(user)
}

// ============================================================================
// COLLECTOR TOKEN OPERATIONS
// ============================================================================

// GenerateCollectorToken generates a JWT token for a collector
func (jm *JWTManager) GenerateCollectorToken(collector *models.Collector) (string, *time.Time, error) {
	if collector == nil {
		return "", nil, apperrors.BadRequest("Invalid collector", "Collector cannot be nil")
	}

	expirationTime := time.Now().Add(jm.collectorTokenExpiration)

	claims := &CollectorClaims{
		CollectorID: collector.ID.String(),
		Hostname:    collector.Hostname,
		Type:        TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("collector:%s", collector.ID.String()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jm.secret))
	if err != nil {
		return "", nil, apperrors.InternalServerError(
			"Token generation failed",
			fmt.Sprintf("Failed to sign collector token: %v", err),
		)
	}

	return tokenString, &expirationTime, nil
}

// ValidateCollectorToken validates and parses a collector JWT token
func (jm *JWTManager) ValidateCollectorToken(tokenString string) (*CollectorClaims, error) {
	claims := &CollectorClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.InvalidToken(
					fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]),
				)
			}
			return []byte(jm.secret), nil
		},
	)

	if err != nil {
		return nil, apperrors.InvalidToken(fmt.Sprintf("Token parsing failed: %v", err))
	}

	if !token.Valid {
		return nil, apperrors.InvalidToken("Token is invalid")
	}

	// Verify token type
	if claims.Type != TokenTypeAccess {
		return nil, apperrors.InvalidToken("This is not an access token")
	}

	// Check expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, apperrors.TokenExpired()
	}

	return claims, nil
}

// ============================================================================
// TOKEN UTILITY FUNCTIONS
// ============================================================================

// ExtractTokenFromHeader extracts the token from an Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", apperrors.MissingAuthHeader()
	}

	// Expected format: "Bearer <token>"
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", apperrors.InvalidToken("Invalid authorization header format")
	}

	token := authHeader[7:]
	if token == "" {
		return "", apperrors.InvalidToken("Token is empty")
	}

	return token, nil
}

// GetTokenExpiration returns when the token expires
func (c *Claims) GetTokenExpiration() time.Time {
	if c.ExpiresAt != nil {
		return c.ExpiresAt.Time
	}
	return time.Time{}
}

// GetTokenExpiresIn returns how long until token expires (in seconds)
func (c *Claims) GetTokenExpiresIn() int64 {
	if c.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(c.ExpiresAt.Time).Seconds()
	if remaining < 0 {
		return 0
	}
	return int64(remaining)
}

// IsExpired checks if token is expired
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}

// GetCollectorTokenExpiration returns when the token expires
func (c *CollectorClaims) GetTokenExpiration() time.Time {
	if c.ExpiresAt != nil {
		return c.ExpiresAt.Time
	}
	return time.Time{}
}

// GetTokenExpiresIn returns how long until token expires (in seconds)
func (c *CollectorClaims) GetTokenExpiresIn() int64 {
	if c.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(c.ExpiresAt.Time).Seconds()
	if remaining < 0 {
		return 0
	}
	return int64(remaining)
}

// IsExpired checks if token is expired
func (c *CollectorClaims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}
