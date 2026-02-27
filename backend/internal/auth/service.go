package auth

import (
	"fmt"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/google/uuid"
)

// AuthService handles authentication operations
type AuthService struct {
	JWTManager      *JWTManager
	PasswordManager *PasswordManager
	CertManager     *CertificateManager
	userStore       UserStore
	collectorStore  CollectorStore
	tokenStore      TokenStore
}

// UserStore defines user data access interface
type UserStore interface {
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUserLastLogin(userID int, timestamp time.Time) error
}

// CollectorStore defines collector data access interface
type CollectorStore interface {
	CreateCollector(collector *models.Collector) (uuid.UUID, error)
	GetCollectorByID(id uuid.UUID) (*models.Collector, error)
	UpdateCollectorStatus(id uuid.UUID, status string) error
	UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error
}

// TokenStore defines token data access interface
type TokenStore interface {
	CreateAPIToken(token *models.APIToken) (int, error)
	GetAPITokenByHash(hash string) (*models.APIToken, error)
	UpdateAPITokenLastUsed(id int, timestamp time.Time) error
}

// NewAuthService creates a new authentication service
func NewAuthService(
	jwtManager *JWTManager,
	passwordManager *PasswordManager,
	certManager *CertificateManager,
	userStore UserStore,
	collectorStore CollectorStore,
	tokenStore TokenStore,
) *AuthService {
	return &AuthService{
		JWTManager:      jwtManager,
		PasswordManager: passwordManager,
		CertManager:     certManager,
		userStore:       userStore,
		collectorStore:  collectorStore,
		tokenStore:      tokenStore,
	}
}

// LoginUser authenticates a user and returns tokens
func (as *AuthService) LoginUser(username, password string) (*models.LoginResponse, error) {
	// Get user from store
	user, err := as.userStore.GetUserByUsername(username)
	if err != nil {
		return nil, apperrors.InvalidCredentials()
	}

	if user == nil {
		return nil, apperrors.InvalidCredentials()
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.Unauthorized("User account is inactive", "")
	}

	// Verify password
	passwordMatch := as.PasswordManager.VerifyPassword(user.PasswordHash, password)
	if !passwordMatch {
		return nil, apperrors.InvalidCredentials()
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

	// Update last login
	_ = as.userStore.UpdateUserLastLogin(user.ID, time.Now())

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    *expiresAt,
		User:         user,
	}, nil
}

// RefreshUserToken generates a new access token from a refresh token
func (as *AuthService) RefreshUserToken(refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := as.JWTManager.ValidateUserRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get user from store
	user, err := as.userStore.GetUserByID(claims.UserID)
	if err != nil {
		return nil, apperrors.Unauthorized("User not found", "")
	}

	if user == nil {
		return nil, apperrors.Unauthorized("User not found", "")
	}

	// Generate new access token
	accessToken, expiresAt, err := as.JWTManager.GenerateUserToken(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken, // Return same refresh token
		ExpiresAt:    *expiresAt,
		User:         user,
	}, nil
}

// RegisterCollector registers a new collector and generates credentials
func (as *AuthService) RegisterCollector(req *models.CollectorRegisterRequest) (*models.CollectorRegisterResponse, error) {
	// Validate request
	if req.Name == "" || req.Hostname == "" {
		return nil, apperrors.BadRequest("Invalid collector data", "Name and hostname are required")
	}

	// Create collector record
	collector := &models.Collector{
		ID:                  uuid.New(),
		Name:                req.Name,
		Hostname:            req.Hostname,
		Description:         "",
		Address:             req.Address,
		Version:             req.Version,
		Status:              "registered",
		ConfigVersion:       1,
		MetricsCountTotal:   0,
		MetricsCount24h:     0,
		HealthCheckInterval: 60, // Default: check health every 60 seconds
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Create in database
	_, err := as.collectorStore.CreateCollector(collector)
	if err != nil {
		return nil, apperrors.DatabaseError("Failed to create collector", err.Error())
	}

	// Generate certificate
	certPair, err := as.CertManager.GenerateCollectorCertificate(
		collector.ID,
		req.Hostname,
		365, // 1 year validity
	)
	if err != nil {
		return nil, apperrors.InternalServerError("Certificate generation failed", err.Error())
	}

	// Update collector with certificate info
	err = as.collectorStore.UpdateCollectorCertificate(collector.ID, certPair.Thumbprint, certPair.ExpiresAt)
	if err != nil {
		return nil, apperrors.DatabaseError("Failed to store certificate info", err.Error())
	}

	// Generate JWT token for collector
	token, expiresAt, err := as.JWTManager.GenerateCollectorToken(collector)
	if err != nil {
		return nil, err
	}

	return &models.CollectorRegisterResponse{
		CollectorID: collector.ID,
		Token:       token,
		Certificate: certPair.Certificate,
		PrivateKey:  certPair.PrivateKey,
		ExpiresAt:   *expiresAt,
	}, nil
}

// ValidateCollectorToken validates a collector's JWT token
func (as *AuthService) ValidateCollectorToken(token string) (*models.Collector, error) {
	claims, err := as.JWTManager.ValidateCollectorToken(token)
	if err != nil {
		return nil, err
	}

	// Parse collector ID from claims
	collectorID, err := uuid.Parse(claims.CollectorID)
	if err != nil {
		return nil, apperrors.InvalidToken("Invalid collector ID in token")
	}

	// Get collector from store
	collector, err := as.collectorStore.GetCollectorByID(collectorID)
	if err != nil {
		return nil, apperrors.Unauthorized("Collector not found", "")
	}

	if collector == nil {
		return nil, apperrors.Unauthorized("Collector not found", "")
	}

	// Check if collector is active
	if collector.Status == "inactive" || collector.Status == "error" {
		return nil, apperrors.Unauthorized(
			fmt.Sprintf("Collector is %s", collector.Status),
			"",
		)
	}

	return collector, nil
}

// ValidateUserToken validates a user's JWT token
func (as *AuthService) ValidateUserToken(token string) (*models.User, error) {
	claims, err := as.JWTManager.ValidateUserToken(token)
	if err != nil {
		return nil, err
	}

	// Get user from store
	user, err := as.userStore.GetUserByID(claims.UserID)
	if err != nil {
		return nil, apperrors.Unauthorized("User not found", "")
	}

	if user == nil {
		return nil, apperrors.Unauthorized("User not found", "")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.Unauthorized("User is inactive", "")
	}

	return user, nil
}
