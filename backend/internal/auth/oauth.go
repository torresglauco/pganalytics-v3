package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"go.uber.org/zap"
)

// OAuthProvider represents an OAuth provider
type OAuthProvider string

const (
	OAuthProviderGoogle  OAuthProvider = "google"
	OAuthProviderGitHub  OAuthProvider = "github"
	OAuthProviderAzureAD OAuthProvider = "azure_ad"
	OAuthProviderCustom  OAuthProvider = "custom"
)

// OAuthCircuitBreakerState represents the state of an OAuth circuit breaker
type OAuthCircuitBreakerState string

const (
	// OAuthStateClosed means the circuit is closed (normal operation)
	OAuthStateClosed OAuthCircuitBreakerState = "closed"
	// OAuthStateOpen means the circuit is open (service unavailable)
	OAuthStateOpen OAuthCircuitBreakerState = "open"
	// OAuthStateHalfOpen means the circuit is half-open (testing recovery)
	OAuthStateHalfOpen OAuthCircuitBreakerState = "half-open"
)

// OAuthCircuitBreaker implements the circuit breaker pattern for OAuth service resilience
type OAuthCircuitBreaker struct {
	mu               sync.RWMutex
	state            OAuthCircuitBreakerState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	logger           *zap.Logger
}

// NewOAuthCircuitBreaker creates a new circuit breaker for OAuth
func NewOAuthCircuitBreaker(logger *zap.Logger) *OAuthCircuitBreaker {
	return &OAuthCircuitBreaker{
		state:            OAuthStateClosed,
		failureCount:     0,
		successCount:     0,
		failureThreshold: 5,                // Open after 5 failures
		successThreshold: 3,                // Close after 3 successes
		timeout:          30 * time.Second, // Try recovery after 30 seconds
		logger:           logger,
	}
}

// RecordSuccess records a successful OAuth call
func (cb *OAuthCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case OAuthStateClosed:
		// Success in closed state, reset counter
		cb.failureCount = 0
		cb.successCount = 0

	case OAuthStateHalfOpen:
		// Success in half-open state, increment success counter
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = OAuthStateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.logger.Info("OAuth circuit breaker closed - service recovered")
		}

	case OAuthStateOpen:
		// Ignore successes when open (waiting for timeout)
	}
}

// RecordFailure records a failed OAuth call
func (cb *OAuthCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case OAuthStateClosed:
		// Failure in closed state, increment counter
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = OAuthStateOpen
			cb.logger.Warn("OAuth circuit breaker opened - too many failures",
				zap.Int("failure_count", cb.failureCount))
		}

	case OAuthStateHalfOpen:
		// Failure in half-open state, re-open the circuit
		cb.state = OAuthStateOpen
		cb.failureCount = 0
		cb.successCount = 0
		cb.logger.Warn("OAuth circuit breaker reopened - failure during recovery")

	case OAuthStateOpen:
		// Already open, just update timestamp
		cb.lastFailureTime = time.Now()
	}
}

// IsOpen checks if the circuit is open (service unavailable)
func (cb *OAuthCircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == OAuthStateClosed {
		return false
	}

	if cb.state == OAuthStateOpen {
		// Check if timeout has elapsed to try recovery
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Upgrade to half-open
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = OAuthStateHalfOpen
			cb.failureCount = 0
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			cb.logger.Info("OAuth circuit breaker half-open - attempting recovery")
			return false
		}
		return true
	}

	// Half-open state
	return false
}

// State returns the current state as a string
func (cb *OAuthCircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return string(cb.state)
}

// Reset resets the circuit breaker to closed state
func (cb *OAuthCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = OAuthStateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.logger.Info("OAuth circuit breaker reset to closed state")
}

// GetMetrics returns the current metrics
func (cb *OAuthCircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":              string(cb.state),
		"failure_count":      cb.failureCount,
		"success_count":      cb.successCount,
		"failure_threshold":  cb.failureThreshold,
		"success_threshold":  cb.successThreshold,
		"last_failure_time":  cb.lastFailureTime,
		"time_since_failure": time.Since(cb.lastFailureTime).Seconds(),
	}
}

// OAuthConfig represents OAuth provider configuration
type OAuthProviderConfig struct {
	Name         string   `json:"name"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`
}

// OAuthConnector handles OAuth 2.0 / OIDC authentication
type OAuthConnector struct {
	providers      map[OAuthProvider]*oauth2.Config
	rootURL        string
	circuitBreaker *OAuthCircuitBreaker
	timeout        time.Duration
	logger         *zap.Logger
}

// NewOAuthConnector creates a new OAuth connector
func NewOAuthConnector(rootURL string, providerConfigs []OAuthProviderConfig) (*OAuthConnector, error) {
	return NewOAuthConnectorWithLogger(rootURL, providerConfigs, zap.NewNop())
}

// NewOAuthConnectorWithLogger creates a new OAuth connector with a custom logger
func NewOAuthConnectorWithLogger(rootURL string, providerConfigs []OAuthProviderConfig, logger *zap.Logger) (*OAuthConnector, error) {
	oc := &OAuthConnector{
		providers:      make(map[OAuthProvider]*oauth2.Config),
		rootURL:        rootURL,
		circuitBreaker: NewOAuthCircuitBreaker(logger),
		timeout:        10 * time.Second, // Default 10 second timeout
		logger:         logger,
	}

	for _, cfg := range providerConfigs {
		err := oc.addProvider(cfg)
		if err != nil {
			return nil, err
		}
	}

	return oc, nil
}

// addProvider adds an OAuth provider configuration
func (oc *OAuthConnector) addProvider(cfg OAuthProviderConfig) error {
	var config *oauth2.Config
	providerName := OAuthProvider(cfg.Name)

	switch providerName {
	case OAuthProviderGoogle:
		config = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/api/v1/auth/oauth/callback", oc.rootURL),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}

	case OAuthProviderGitHub:
		config = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/api/v1/auth/oauth/callback", oc.rootURL),
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}

	case OAuthProviderAzureAD:
		// Azure AD endpoints
		config = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/api/v1/auth/oauth/callback", oc.rootURL),
			Scopes: []string{
				"https://graph.microsoft.com/user.read",
			},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
		}

	case OAuthProviderCustom:
		// Custom OIDC provider
		if cfg.AuthURL == "" || cfg.TokenURL == "" {
			return fmt.Errorf("custom provider requires auth_url and token_url")
		}

		config = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  fmt.Sprintf("%s/api/v1/auth/oauth/callback", oc.rootURL),
			Scopes:       cfg.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
		}

	default:
		return fmt.Errorf("unsupported OAuth provider: %s", providerName)
	}

	oc.providers[providerName] = config
	return nil
}

// GetAuthCodeURL returns the authorization code URL for a provider
func (oc *OAuthConnector) GetAuthCodeURL(provider OAuthProvider, state string) (string, error) {
	config, ok := oc.providers[provider]
	if !ok {
		return "", fmt.Errorf("provider not configured: %s", provider)
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID       string
	Email    string
	FullName string
	Avatar   string
	Provider OAuthProvider
}

// ExchangeCodeForToken exchanges authorization code for token
func (oc *OAuthConnector) ExchangeCodeForToken(ctx context.Context, provider OAuthProvider, code string) (*oauth2.Token, error) {
	// Check circuit breaker
	if oc.circuitBreaker.IsOpen() {
		oc.logger.Warn("OAuth circuit breaker is open",
			zap.String("provider", string(provider)))
		return nil, fmt.Errorf("OAuth provider %s temporarily unavailable (circuit open)", provider)
	}

	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	// Add timeout to context
	ctx, cancel := oc.withTimeout(ctx)
	defer cancel()

	token, err := config.Exchange(ctx, code)
	if err != nil {
		oc.circuitBreaker.RecordFailure()
		oc.logger.Error("OAuth token exchange failed",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	oc.circuitBreaker.RecordSuccess()
	oc.logger.Debug("OAuth token exchange succeeded",
		zap.String("provider", string(provider)))
	return token, nil
}

// RefreshToken refreshes an OAuth token
func (oc *OAuthConnector) RefreshToken(ctx context.Context, provider OAuthProvider, token *oauth2.Token) (*oauth2.Token, error) {
	// Check circuit breaker
	if oc.circuitBreaker.IsOpen() {
		oc.logger.Warn("OAuth circuit breaker is open",
			zap.String("provider", string(provider)))
		return nil, fmt.Errorf("OAuth provider %s temporarily unavailable (circuit open)", provider)
	}

	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	// Add timeout to context
	ctx, cancel := oc.withTimeout(ctx)
	defer cancel()

	// Create a token source for refreshing
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		oc.circuitBreaker.RecordFailure()
		oc.logger.Error("OAuth token refresh failed",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	oc.circuitBreaker.RecordSuccess()
	oc.logger.Debug("OAuth token refresh succeeded",
		zap.String("provider", string(provider)))
	return newToken, nil
}

// GetUserInfo retrieves user information from OAuth provider
func (oc *OAuthConnector) GetUserInfo(ctx context.Context, provider OAuthProvider, token *oauth2.Token) (*OAuthUserInfo, error) {
	// Check circuit breaker
	if oc.circuitBreaker.IsOpen() {
		oc.logger.Warn("OAuth circuit breaker is open",
			zap.String("provider", string(provider)))
		return nil, fmt.Errorf("OAuth provider %s temporarily unavailable (circuit open)", provider)
	}

	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	// Add timeout to context
	ctx, cancel := oc.withTimeout(ctx)
	defer cancel()

	client := config.Client(ctx, token)

	var userInfoURL string
	switch provider {
	case OAuthProviderGoogle:
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case OAuthProviderGitHub:
		userInfoURL = "https://api.github.com/user"
	case OAuthProviderAzureAD:
		userInfoURL = "https://graph.microsoft.com/v1.0/me"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		oc.circuitBreaker.RecordFailure()
		oc.logger.Error("OAuth user info fetch failed",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		oc.circuitBreaker.RecordFailure()
		body, _ := ioutil.ReadAll(resp.Body)
		oc.logger.Error("OAuth user info returned error",
			zap.String("provider", string(provider)),
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("failed to get user info: %d - %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		oc.circuitBreaker.RecordFailure()
		oc.logger.Error("OAuth user info read failed",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to read user info: %w", err)
	}

	userInfo := &OAuthUserInfo{Provider: provider}

	// Parse response based on provider
	switch provider {
	case OAuthProviderGoogle:
		var googleUser struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			Name          string `json:"name"`
			Picture       string `json:"picture"`
			VerifiedEmail bool   `json:"verified_email"`
		}

		if err := json.Unmarshal(body, &googleUser); err != nil {
			oc.circuitBreaker.RecordFailure()
			oc.logger.Error("OAuth user info parse failed",
				zap.String("provider", string(provider)),
				zap.Error(err))
			return nil, fmt.Errorf("failed to parse Google user info: %w", err)
		}

		userInfo.ID = googleUser.ID
		userInfo.Email = googleUser.Email
		userInfo.FullName = googleUser.Name
		userInfo.Avatar = googleUser.Picture

	case OAuthProviderGitHub:
		var githubUser struct {
			ID        int    `json:"id"`
			Login     string `json:"login"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		}

		if err := json.Unmarshal(body, &githubUser); err != nil {
			oc.circuitBreaker.RecordFailure()
			oc.logger.Error("OAuth user info parse failed",
				zap.String("provider", string(provider)),
				zap.Error(err))
			return nil, fmt.Errorf("failed to parse GitHub user info: %w", err)
		}

		userInfo.ID = fmt.Sprintf("%d", githubUser.ID)
		userInfo.Email = githubUser.Email
		if userInfo.Email == "" {
			userInfo.Email = fmt.Sprintf("%s@github.com", githubUser.Login)
		}
		userInfo.FullName = githubUser.Name
		userInfo.Avatar = githubUser.AvatarURL

	case OAuthProviderAzureAD:
		var azureUser struct {
			ID                string `json:"id"`
			UserPrincipalName string `json:"userPrincipalName"`
			DisplayName       string `json:"displayName"`
			Mail              string `json:"mail"`
		}

		if err := json.Unmarshal(body, &azureUser); err != nil {
			oc.circuitBreaker.RecordFailure()
			oc.logger.Error("OAuth user info parse failed",
				zap.String("provider", string(provider)),
				zap.Error(err))
			return nil, fmt.Errorf("failed to parse Azure AD user info: %w", err)
		}

		userInfo.ID = azureUser.ID
		userInfo.Email = azureUser.Mail
		if userInfo.Email == "" {
			userInfo.Email = azureUser.UserPrincipalName
		}
		userInfo.FullName = azureUser.DisplayName
	}

	if userInfo.Email == "" {
		oc.circuitBreaker.RecordFailure()
		oc.logger.Error("OAuth user email not available",
			zap.String("provider", string(provider)))
		return nil, fmt.Errorf("user email not available from provider")
	}

	oc.circuitBreaker.RecordSuccess()
	oc.logger.Debug("OAuth user info fetch succeeded",
		zap.String("provider", string(provider)),
		zap.String("user_id", userInfo.ID))
	return userInfo, nil
}

// IsTokenExpired checks if a token is expired
func (oc *OAuthConnector) IsTokenExpired(token *oauth2.Token) bool {
	if token == nil {
		return true
	}
	return !token.Valid()
}

// GetTokenExpiry returns the token expiry time
func (oc *OAuthConnector) GetTokenExpiry(token *oauth2.Token) time.Time {
	if token == nil {
		return time.Time{}
	}
	return token.Expiry
}

// withTimeout adds timeout to context if not already set
func (oc *OAuthConnector) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	// Check if deadline already set
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {} // Return original context, no-op cancel
	}

	// Add timeout
	return context.WithTimeout(ctx, oc.timeout)
}
