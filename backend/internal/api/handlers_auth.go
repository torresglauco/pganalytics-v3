package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// ============================================================================
// LDAP AUTHENTICATION ENDPOINTS
// ============================================================================

// LDAPLoginRequest represents an LDAP login request
type LDAPLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Summary LDAP Login
// @Description Authenticate user via LDAP/Active Directory
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LDAPLoginRequest true "LDAP credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/auth/ldap/login [post]
func (s *Server) handleLDAPLogin(c *gin.Context) {
	var req LDAPLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	// Validate LDAP is enabled
	if !s.config.LDAPEnabled {
		errResp := apperrors.ServiceUnavailable("LDAP authentication is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create LDAP connector
	groupToRoleMap := make(map[string]string)
	if err := json.Unmarshal([]byte(s.config.LDAPGroupToRoleJSON), &groupToRoleMap); err != nil {
		s.logger.Error("Failed to parse LDAP group mapping", zap.Error(err))
		groupToRoleMap = make(map[string]string) // Default to empty
	}

	ldapConn := auth.NewLDAPConnector(
		s.config.LDAPServerURL,
		s.config.LDAPBindDN,
		s.config.LDAPBindPassword,
		s.config.LDAPUserSearchBase,
		s.config.LDAPGroupSearchBase,
		groupToRoleMap,
		nil, // TLS config handled separately
	)

	// Authenticate user
	ldapUser, err := ldapConn.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		s.logger.Warn("LDAP authentication failed", zap.String("username", req.Username), zap.Error(err))
		errResp := apperrors.Unauthorized("LDAP authentication failed", "Invalid credentials")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get user role from LDAP groups
	role, err := ldapConn.GetUserRole(req.Username)
	if err != nil {
		s.logger.Warn("Failed to get LDAP user role", zap.String("username", req.Username), zap.Error(err))
		role = "viewer" // Default role
	}

	// Find or create user in database
	user, err := s.createOrUpdateLDAPUser(ctx, ldapUser, role)
	if err != nil {
		s.logger.Error("Failed to create/update LDAP user", zap.String("username", req.Username), zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process authentication", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.authService.GenerateUserTokens(user)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to generate tokens", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create session
	sess, err := s.sessionManager.CreateSession(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		s.logger.Error("Failed to create session - authentication failed",
			zap.Int("user_id", user.ID),
			zap.String("ip", c.ClientIP()),
			zap.String("method", "ldap_login"),
			zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to create session", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Session creation succeeded - safe to proceed
	sessionToken := sess.Token

	// Log authentication event
	s.logAuthEvent(ctx, user.ID, "ldap_login", true, "")

	response := &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionToken: sessionToken,
		User:         user,
		ExpiresIn:    int(s.config.JWTExpiration.Seconds()),
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// OAUTH AUTHENTICATION ENDPOINTS
// ============================================================================

// OAuthLoginRequest initiates OAuth login
type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required"`
	State    string `json:"state" binding:"required"`
}

// @Summary OAuth Login URL
// @Description Get OAuth provider login URL
// @Tags Authentication
// @Accept json
// @Produce json
// @Param provider path string true "OAuth provider (google, github, azure_ad, custom)"
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/auth/oauth/{provider}/login [get]
func (s *Server) handleOAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	state := c.Query("state")

	if provider == "" || state == "" {
		errResp := apperrors.BadRequest("Missing provider or state", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	if !s.config.OAuthEnabled {
		errResp := apperrors.ServiceUnavailable("OAuth authentication is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get OAuth config from database or cache
	// For now, use config from environment
	var providerConfigs []auth.OAuthProviderConfig
	if err := json.Unmarshal([]byte(s.config.OAuthProvidersJSON), &providerConfigs); err != nil {
		s.logger.Error("Failed to parse OAuth config", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("OAuth is not properly configured", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/callback", s.config.APIBaseURL)
	oauthConn, err := auth.NewOAuthConnector(redirectURL, providerConfigs)
	if err != nil {
		s.logger.Error("Failed to create OAuth connector", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("OAuth initialization failed", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	authURL, err := oauthConn.GetAuthCodeURL(auth.OAuthProvider(provider), state)
	if err != nil {
		s.logger.Error("Failed to get auth code URL", zap.String("provider", provider), zap.Error(err))
		errResp := apperrors.BadRequest("Invalid OAuth provider", provider)
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// OAuthCallbackRequest represents OAuth callback
type OAuthCallbackRequest struct {
	Code     string `json:"code" binding:"required"`
	Provider string `json:"provider" binding:"required"`
	State    string `json:"state"`
}

// @Summary OAuth Callback
// @Description Handle OAuth provider callback
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body OAuthCallbackRequest true "OAuth callback data"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/auth/oauth/callback [post]
func (s *Server) handleOAuthCallback(c *gin.Context) {
	var req OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	if !s.config.OAuthEnabled {
		errResp := apperrors.ServiceUnavailable("OAuth is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Initialize OAuth connector
	var providerConfigs []auth.OAuthProviderConfig
	if err := json.Unmarshal([]byte(s.config.OAuthProvidersJSON), &providerConfigs); err != nil {
		errResp := apperrors.ServiceUnavailable("OAuth is not properly configured", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/callback", s.config.APIBaseURL)
	oauthConn, err := auth.NewOAuthConnector(redirectURL, providerConfigs)
	if err != nil {
		s.logger.Error("Failed to create OAuth connector", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("OAuth initialization failed", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Exchange code for token
	token, err := oauthConn.ExchangeCodeForToken(ctx, auth.OAuthProvider(req.Provider), req.Code)
	if err != nil {
		s.logger.Error("Failed to exchange OAuth code", zap.String("provider", req.Provider), zap.Error(err))
		errResp := apperrors.Unauthorized("OAuth authentication failed", "Failed to exchange code")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get user info from provider
	userInfo, err := oauthConn.GetUserInfo(ctx, auth.OAuthProvider(req.Provider), token)
	if err != nil {
		s.logger.Error("Failed to get OAuth user info", zap.String("provider", req.Provider), zap.Error(err))
		errResp := apperrors.Unauthorized("OAuth authentication failed", "Failed to get user info")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Find or create user
	user, err := s.createOrUpdateOAuthUser(ctx, userInfo)
	if err != nil {
		s.logger.Error("Failed to create/update OAuth user", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process authentication", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.authService.GenerateUserTokens(user)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to generate tokens", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create session
	sess, err := s.sessionManager.CreateSession(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		s.logger.Error("Failed to create session - authentication failed",
			zap.Int("user_id", user.ID),
			zap.String("ip", c.ClientIP()),
			zap.String("provider", req.Provider),
			zap.String("method", "oauth_callback"),
			zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to create session", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Session creation succeeded - safe to proceed
	sessionToken := sess.Token

	// Log authentication event
	s.logAuthEvent(ctx, user.ID, fmt.Sprintf("oauth_%s_login", req.Provider), true, "")

	response := &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionToken: sessionToken,
		User:         user,
		ExpiresIn:    int(s.config.JWTExpiration.Seconds()),
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// SAML AUTHENTICATION ENDPOINTS
// ============================================================================

// SAMLMetadataRequest represents SAML metadata request
type SAMLMetadataRequest struct{}

// @Summary SAML Metadata
// @Description Get SAML Service Provider metadata
// @Tags Authentication
// @Produce xml
// @Success 200 {string} string "SAML metadata XML"
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/auth/saml/metadata [get]
func (s *Server) handleSAMLMetadata(c *gin.Context) {
	if !s.config.SAMLEnabled {
		errResp := apperrors.ServiceUnavailable("SAML is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	samlConn, err := auth.NewSAMLConnector(&auth.SAMLConfig{
		CertPath: s.config.SAMLCertPath,
		KeyPath:  s.config.SAMLKeyPath,
		IDPURL:   s.config.SAMLIDPMetadataURL,
		EntityID: s.config.SAMLEntityID,
		RootURL:  s.config.APIBaseURL,
	})
	if err != nil {
		s.logger.Error("Failed to create SAML connector", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("SAML initialization failed", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	metadata, err := samlConn.GetMetadata()
	if err != nil {
		s.logger.Error("Failed to get SAML metadata", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to generate metadata", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, metadata)
}

// SAMLACSRequest represents SAML Assertion Consumer Service request
type SAMLACSRequest struct {
	SAMLResponse string `form:"SAMLResponse" binding:"required"`
	RelayState   string `form:"RelayState"`
}

// @Summary SAML Assertion Consumer Service
// @Description Handle SAML assertion response
// @Tags Authentication
// @Accept x-www-form-urlencoded
// @Produce json
// @Param SAMLResponse formData string true "SAML Response"
// @Param RelayState formData string false "Relay State"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Failure 503 {object} apperrors.AppError
// @Router /api/v1/auth/saml/acs [post]
func (s *Server) handleSAMLACS(c *gin.Context) {
	samlResponse := c.PostForm("SAMLResponse")
	if samlResponse == "" {
		errResp := apperrors.BadRequest("Missing SAML response", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	ctx := c.Request.Context()

	if !s.config.SAMLEnabled {
		errResp := apperrors.ServiceUnavailable("SAML is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	samlConn, err := auth.NewSAMLConnector(&auth.SAMLConfig{
		CertPath: s.config.SAMLCertPath,
		KeyPath:  s.config.SAMLKeyPath,
		IDPURL:   s.config.SAMLIDPMetadataURL,
		EntityID: s.config.SAMLEntityID,
		RootURL:  s.config.APIBaseURL,
	})
	if err != nil {
		s.logger.Error("Failed to create SAML connector", zap.Error(err))
		errResp := apperrors.ServiceUnavailable("SAML initialization failed", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Process SAML assertion
	assertion, err := samlConn.ProcessAssertionResponse(samlResponse)
	if err != nil {
		s.logger.Warn("Failed to process SAML assertion", zap.Error(err))
		errResp := apperrors.Unauthorized("SAML assertion validation failed", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Validate assertion
	if err := samlConn.ValidateAssertion(assertion); err != nil {
		s.logger.Warn("SAML assertion validation failed", zap.Error(err))
		errResp := apperrors.Unauthorized("SAML assertion is invalid", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Find or create user
	user, err := s.createOrUpdateSAMLUser(ctx, assertion)
	if err != nil {
		s.logger.Error("Failed to create/update SAML user", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to process authentication", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.authService.GenerateUserTokens(user)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to generate tokens", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Create session
	sess, err := s.sessionManager.CreateSession(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		s.logger.Error("Failed to create session - authentication failed",
			zap.Int("user_id", user.ID),
			zap.String("ip", c.ClientIP()),
			zap.String("method", "saml_acs"),
			zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to create session", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Session creation succeeded - safe to proceed
	sessionToken := sess.Token

	// Log authentication event
	s.logAuthEvent(ctx, user.ID, "saml_login", true, "")

	response := &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionToken: sessionToken,
		User:         user,
		ExpiresIn:    int(s.config.JWTExpiration.Seconds()),
	}

	c.JSON(http.StatusOK, response)
}

// ============================================================================
// MFA ENDPOINTS
// ============================================================================

// MFASetupRequest requests MFA setup
type MFASetupRequest struct {
	Type string `json:"type" binding:"required"` // totp, sms, email
}

// MFASetupResponse responds with MFA setup information
type MFASetupResponse struct {
	Type              string `json:"type"`
	Secret            string `json:"secret,omitempty"`  // For TOTP
	QRCode            string `json:"qr_code,omitempty"` // Base64 encoded QR code
	PhoneNumberNeeded bool   `json:"phone_number_needed,omitempty"`
}

// @Summary Setup MFA
// @Description Initiate MFA setup for current user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body MFASetupRequest true "MFA type"
// @Success 200 {object} MFASetupResponse
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/users/mfa/setup [post]
func (s *Server) handleMFASetup(c *gin.Context) {
	var req MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	if !s.config.MFAEnabled {
		errResp := apperrors.ServiceUnavailable("MFA is not enabled", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Generate TOTP secret
	key, err := s.mfaManager.GenerateTOTPSecret(user.Username)
	if err != nil {
		s.logger.Error("Failed to generate TOTP secret", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to setup MFA", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Store setup (not yet verified)
	_, err = s.mfaManager.SetupTOTP(user.ID, key.Secret())
	if err != nil {
		s.logger.Error("Failed to setup TOTP in database", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to setup MFA", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	response := MFASetupResponse{
		Type:   req.Type,
		Secret: key.Secret(),
		QRCode: key.URL(), // In production, convert to base64 QR code image
	}

	c.JSON(http.StatusOK, response)
}

// MFAVerifyRequest verifies MFA setup
type MFAVerifyRequest struct {
	Code string `json:"code" binding:"required"`
}

// @Summary Verify MFA
// @Description Verify and enable MFA for current user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body MFAVerifyRequest true "MFA code"
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/users/mfa/verify [post]
func (s *Server) handleMFAVerify(c *gin.Context) {
	var req MFAVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get current user
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	// Verify and enable TOTP
	err := s.mfaManager.VerifyAndEnableTOTP(user.ID, req.Code)
	if err != nil {
		s.logger.Warn("Failed to verify TOTP", zap.String("user_id", fmt.Sprint(user.ID)), zap.Error(err))
		errResp := apperrors.BadRequest("Invalid verification code", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Log authentication event
	ctx := c.Request.Context()
	s.logAuthEvent(ctx, user.ID, "mfa_enabled", true, "")

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA successfully enabled",
	})
}

// MFAChallengeRequest requests MFA challenge
type MFAChallengeRequest struct {
	Code string `json:"code" binding:"required"`
}

// @Summary MFA Challenge
// @Description Complete MFA challenge during login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body MFAChallengeRequest true "MFA code"
// @Success 200 {object} gin.H
// @Failure 400 {object} apperrors.AppError
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/auth/mfa/challenge [post]
func (s *Server) handleMFAChallenge(c *gin.Context) {
	var req MFAChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResp := apperrors.BadRequest("Invalid request", err.Error())
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// Get user from MFA session context
	mfaSessionID := c.GetString("mfa_session_id")
	if mfaSessionID == "" {
		errResp := apperrors.Unauthorized("No active MFA session", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	// In production, retrieve user from MFA session cache/database
	// For now, this is a placeholder implementation

	c.JSON(http.StatusOK, gin.H{
		"message": "MFA challenge completed",
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func (s *Server) createOrUpdateLDAPUser(ctx context.Context, ldapUser *auth.LDAPUser, role string) (*models.User, error) {
	// This would interact with the database to find or create a user
	// based on LDAP information
	// Placeholder implementation - to be completed
	return &models.User{
		Username: ldapUser.Username,
		Email:    ldapUser.Email,
		FullName: ldapUser.FullName,
		Role:     role,
		IsActive: true,
	}, nil
}

func (s *Server) createOrUpdateOAuthUser(ctx context.Context, userInfo *auth.OAuthUserInfo) (*models.User, error) {
	// This would interact with the database to find or create a user
	// based on OAuth provider information
	// Placeholder implementation - to be completed
	return &models.User{
		Email:    userInfo.Email,
		FullName: userInfo.FullName,
		Role:     "user",
		IsActive: true,
	}, nil
}

func (s *Server) createOrUpdateSAMLUser(ctx context.Context, assertion *auth.SAMLAssertion) (*models.User, error) {
	// This would interact with the database to find or create a user
	// based on SAML assertion information
	// Placeholder implementation - to be completed
	return &models.User{
		Email:    assertion.Email,
		FullName: assertion.FullName,
		Role:     "user",
		IsActive: true,
	}, nil
}

func (s *Server) logAuthEvent(ctx context.Context, userID int, action string, success bool, details string) {
	// Log authentication event to audit log
	// Placeholder implementation - to be completed with audit logging
}

// ============================================================================
// PASSWORD CHANGE FLOW HANDLERS
// ============================================================================

// PasswordChangeRequiredResponse indicates if password change is required
type PasswordChangeRequiredResponse struct {
	PasswordChangeRequired bool   `json:"password_change_required"`
	Message                string `json:"message,omitempty"`
}

// @Summary Check Password Change Requirement
// @Description Check if the current user is required to change their password
// @Tags Authentication
// @Security Bearer
// @Produce json
// @Success 200 {object} PasswordChangeRequiredResponse
// @Failure 401 {object} apperrors.AppError
// @Router /api/v1/auth/password-change-required [get]
func (s *Server) handleCheckPasswordChangeRequired(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		errResp := apperrors.Unauthorized("Authentication required", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	user := currentUser.(*models.User)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	userRecord, err := s.postgres.GetUserByID(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		errResp := apperrors.InternalServerError("Failed to check password status", "")
		c.JSON(errResp.StatusCode, errResp)
		return
	}

	response := PasswordChangeRequiredResponse{
		PasswordChangeRequired: !userRecord.PasswordChanged,
	}

	if !userRecord.PasswordChanged {
		response.Message = "Password change is required on first login"
	}

	c.JSON(http.StatusOK, response)
}

// RegisterAuthHandlers registers all authentication handlers
func (s *Server) RegisterAuthHandlers(engine *gin.Engine) {
	authGroup := engine.Group("/api/v1/auth")

	// LDAP authentication
	authGroup.POST("/ldap/login", s.handleLDAPLogin)

	// OAuth authentication
	authGroup.GET("/oauth/:provider/login", s.handleOAuthLogin)
	authGroup.POST("/oauth/callback", s.handleOAuthCallback)

	// SAML authentication
	authGroup.GET("/saml/metadata", s.handleSAMLMetadata)
	authGroup.POST("/saml/acs", s.handleSAMLACS)

	// MFA
	mfaGroup := engine.Group("/api/v1/users")
	mfaGroup.POST("/mfa/setup", s.AuthMiddleware(), s.handleMFASetup)
	mfaGroup.POST("/mfa/verify", s.AuthMiddleware(), s.handleMFAVerify)

	// MFA challenge
	authGroup.POST("/mfa/challenge", s.handleMFAChallenge)

	// Password change flow
	authGroup.GET("/password-change-required", s.AuthMiddleware(), s.handleCheckPasswordChangeRequired)
}
