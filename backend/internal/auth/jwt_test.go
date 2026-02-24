package auth

import (
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateUserToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, expiresAt, err := jm.GenerateUserToken(user)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, expiresAt)
	assert.True(t, time.Now().Before(*expiresAt))
}

func TestJWTManager_GenerateUserToken_NilUser(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	token, expiresAt, err := jm.GenerateUserToken(nil)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, expiresAt)
}

func TestJWTManager_ValidateUserToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, _, err := jm.GenerateUserToken(user)
	require.NoError(t, err)

	claims, err := jm.ValidateUserToken(token)

	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, TokenTypeAccess, claims.Type)
	assert.False(t, claims.IsExpired())
}

func TestJWTManager_ValidateUserToken_InvalidToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	claims, err := jm.ValidateUserToken("invalid.token.format")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateUserToken_WrongSecret(t *testing.T) {
	jm1 := NewJWTManager(
		"secret-key-1",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	jm2 := NewJWTManager(
		"secret-key-2",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, _, err := jm1.GenerateUserToken(user)
	require.NoError(t, err)

	claims, err := jm2.ValidateUserToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateUserToken_RefreshTokenAsAccess(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	refreshToken, _, err := jm.GenerateUserRefreshToken(user)
	require.NoError(t, err)

	claims, err := jm.ValidateUserToken(refreshToken)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_GenerateAndValidateRefreshToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	token, _, err := jm.GenerateUserRefreshToken(user)
	require.NoError(t, err)

	claims, err := jm.ValidateUserRefreshToken(token)

	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, TokenTypeRefresh, claims.Type)
}

func TestJWTManager_RefreshUserToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	refreshToken, _, err := jm.GenerateUserRefreshToken(user)
	require.NoError(t, err)

	newAccessToken, _, err := jm.RefreshUserToken(refreshToken, user)

	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	claims, err := jm.ValidateUserToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, TokenTypeAccess, claims.Type)
}

func TestJWTManager_RefreshUserToken_MismatchedUser(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	user1 := &models.User{
		ID:       1,
		Username: "user1",
		Email:    "user1@example.com",
		Role:     "user",
	}

	user2 := &models.User{
		ID:       2,
		Username: "user2",
		Email:    "user2@example.com",
		Role:     "user",
	}

	refreshToken, _, err := jm.GenerateUserRefreshToken(user1)
	require.NoError(t, err)

	_, _, err = jm.RefreshUserToken(refreshToken, user2)

	assert.Error(t, err)
}

func TestJWTManager_GenerateCollectorToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	collector := &models.Collector{
		ID:       uuid.New(),
		Name:     "test-collector",
		Hostname: "db-server-01",
		Status:   "registered",
	}

	token, expiresAt, err := jm.GenerateCollectorToken(collector)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, expiresAt)
	assert.True(t, time.Now().Before(*expiresAt))
}

func TestJWTManager_ValidateCollectorToken(t *testing.T) {
	jm := NewJWTManager(
		"test-secret-key-should-be-long-enough",
		15*time.Minute,
		24*time.Hour,
		30*time.Minute,
	)

	collectorID := uuid.New()
	collector := &models.Collector{
		ID:       collectorID,
		Name:     "test-collector",
		Hostname: "db-server-01",
		Status:   "registered",
	}

	token, _, err := jm.GenerateCollectorToken(collector)
	require.NoError(t, err)

	claims, err := jm.ValidateCollectorToken(token)

	require.NoError(t, err)
	assert.Equal(t, collector.ID.String(), claims.CollectorID)
	assert.Equal(t, collector.Hostname, claims.Hostname)
	assert.Equal(t, TokenTypeAccess, claims.Type)
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid header",
			header:  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr: false,
		},
		{
			name:    "Empty header",
			header:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid format - no Bearer",
			header:  "Token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid format - Bearer only",
			header:  "Bearer ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid format - lowercase bearer",
			header:  "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTokenFromHeader(tt.header)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClaims_GetTokenExpiresIn(t *testing.T) {
	now := time.Now()
	futureTime := now.Add(1 * time.Hour)

	claims := &Claims{
		RegisteredClaims: struct {
			ExpiresAt *jwt.NumericDate
			IssuedAt  *jwt.NumericDate
			NotBefore *jwt.NumericDate
			Subject   string
		}{
			ExpiresAt: jwt.NewNumericDate(futureTime),
		},
	}

	expiresIn := claims.GetTokenExpiresIn()

	assert.Greater(t, expiresIn, int64(3500)) // Should be close to 3600 (1 hour)
	assert.Less(t, expiresIn, int64(3700))
	assert.False(t, claims.IsExpired())
}

func TestClaims_ExpiredToken(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour)

	claims := &Claims{
		RegisteredClaims: struct {
			ExpiresAt *jwt.NumericDate
			IssuedAt  *jwt.NumericDate
			NotBefore *jwt.NumericDate
			Subject   string
		}{
			ExpiresAt: jwt.NewNumericDate(pastTime),
		},
	}

	assert.True(t, claims.IsExpired())
	assert.Equal(t, int64(0), claims.GetTokenExpiresIn())
}
