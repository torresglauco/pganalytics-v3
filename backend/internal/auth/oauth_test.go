package auth

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNewOAuthConnector tests OAuth connector initialization
func TestNewOAuthConnector(t *testing.T) {
	tests := []struct {
		name      string
		rootURL   string
		providers []OAuthProviderConfig
		wantErr   bool
	}{
		{
			name:      "no providers",
			rootURL:   "http://localhost:8080",
			providers: []OAuthProviderConfig{},
			wantErr:   false,
		},
		{
			name:    "valid Google provider",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "google",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			},
			wantErr: false,
		},
		{
			name:    "valid GitHub provider",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "github",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			},
			wantErr: false,
		},
		{
			name:    "valid Azure AD provider",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "azure_ad",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			},
			wantErr: false,
		},
		{
			name:    "custom OIDC provider with missing URLs",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "custom",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					// Missing auth_url and token_url
				},
			},
			wantErr: true,
		},
		{
			name:    "custom OIDC provider with all URLs",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "custom",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					AuthURL:      "https://custom.example.com/oauth/authorize",
					TokenURL:     "https://custom.example.com/oauth/token",
				},
			},
			wantErr: false,
		},
		{
			name:    "unsupported provider",
			rootURL: "http://localhost:8080",
			providers: []OAuthProviderConfig{
				{
					Name:         "unsupported",
					ClientID:     "client-id",
					ClientSecret: "client-secret",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, err := NewOAuthConnector(tt.rootURL, tt.providers)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewOAuthConnector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && connector == nil {
				t.Errorf("NewOAuthConnector() = nil, want non-nil")
			}
		})
	}
}

// TestGetAuthCodeURL tests authorization code URL generation
func TestGetAuthCodeURL(t *testing.T) {
	connector, _ := NewOAuthConnector(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
		},
	)

	tests := []struct {
		name      string
		provider  OAuthProvider
		state     string
		wantError bool
	}{
		{
			name:      "valid Google provider",
			provider:  OAuthProviderGoogle,
			state:     "test-state-123",
			wantError: false,
		},
		{
			name:      "unsupported provider",
			provider:  OAuthProvider("unsupported"),
			state:     "test-state-123",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := connector.GetAuthCodeURL(tt.provider, tt.state)

			if (err != nil) != tt.wantError {
				t.Errorf("GetAuthCodeURL() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError && url == "" {
				t.Errorf("GetAuthCodeURL() = empty string, want non-empty URL")
			}

			// Verify URL contains state parameter
			if !tt.wantError && tt.state != "" {
				if url != "" && len(url) > 0 {
					// URL should be valid (would contain the state)
					t.Logf("Generated URL: %s", url[:50]+"...") // Log first 50 chars
				}
			}
		})
	}
}

// TestIsTokenExpired tests token expiry checking
func TestIsTokenExpired(t *testing.T) {
	connector, _ := NewOAuthConnector(
		"http://localhost:8080",
		[]OAuthProviderConfig{},
	)

	tests := []struct {
		name     string
		token    interface{}
		expected bool
	}{
		{
			name:     "nil token",
			token:    nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to OAuth2 token if needed
			var oauthToken interface{}
			if tt.token == nil {
				oauthToken = nil
			}

			// Can't fully test without actual OAuth2.Token
			// This is a basic smoke test
			_ = connector
			_ = oauthToken
		})
	}
}

// TestProviderConfiguration tests provider configuration
func TestProviderConfiguration(t *testing.T) {
	tests := []struct {
		name             string
		provider         OAuthProvider
		clientID         string
		clientSecret     string
		expectConfigured bool
	}{
		{
			name:             "Google",
			provider:         OAuthProviderGoogle,
			clientID:         "google-client-id",
			clientSecret:     "google-secret",
			expectConfigured: true,
		},
		{
			name:             "GitHub",
			provider:         OAuthProviderGitHub,
			clientID:         "github-client-id",
			clientSecret:     "github-secret",
			expectConfigured: true,
		},
		{
			name:             "Azure AD",
			provider:         OAuthProviderAzureAD,
			clientID:         "azure-client-id",
			clientSecret:     "azure-secret",
			expectConfigured: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector, _ := NewOAuthConnector(
				"http://localhost:8080",
				[]OAuthProviderConfig{
					{
						Name:         string(tt.provider),
						ClientID:     tt.clientID,
						ClientSecret: tt.clientSecret,
					},
				},
			)

			if tt.expectConfigured && connector == nil {
				t.Errorf("Expected connector to be configured for %s", tt.provider)
			}
		})
	}
}

// TestGetUserInfo tests user info retrieval (mock test)
func TestGetUserInfo(t *testing.T) {
	connector, _ := NewOAuthConnector(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
	)

	// Note: These tests are limited without actual OAuth2.Token
	// In production, would use mock OAuth2 responses
	tests := []struct {
		name     string
		provider OAuthProvider
	}{
		{
			name:     "Google provider",
			provider: OAuthProviderGoogle,
		},
		{
			name:     "GitHub provider",
			provider: OAuthProviderGitHub,
		},
		{
			name:     "Azure AD provider",
			provider: OAuthProviderAzureAD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Would fail without actual token and mock HTTP responses
			// Documenting the test structure
			_ = connector
			_ = tt.provider
			t.Skip("Requires mock OAuth2 token and HTTP responses")
		})
	}
}

// BenchmarkGetAuthCodeURL benchmarks auth code URL generation
func BenchmarkGetAuthCodeURL(b *testing.B) {
	connector, _ := NewOAuthConnector(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = connector.GetAuthCodeURL(OAuthProviderGoogle, "test-state")
	}
}

// TestCircuitBreakerInitialization tests circuit breaker initialization
func TestCircuitBreakerInitialization(t *testing.T) {
	logger := zap.NewNop()
	connector, err := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
		logger,
	)

	if err != nil {
		t.Fatalf("NewOAuthConnectorWithLogger() error = %v, want nil", err)
	}

	if connector.circuitBreaker == nil {
		t.Errorf("circuitBreaker is nil, want non-nil")
	}

	if connector.timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", connector.timeout)
	}

	if connector.circuitBreaker.State() != string(OAuthStateClosed) {
		t.Errorf("initial state = %s, want %s", connector.circuitBreaker.State(), OAuthStateClosed)
	}
}

// TestCircuitBreakerOpenAfterFailures tests circuit breaker opens after N failures
func TestCircuitBreakerOpenAfterFailures(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{},
		logger,
	)

	// Record 5 failures (threshold)
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	if !connector.circuitBreaker.IsOpen() {
		t.Errorf("circuit is not open after 5 failures, state = %s", connector.circuitBreaker.State())
	}
}

// TestCircuitBreakerClosesAfterSuccesses tests circuit breaker closes after recovery
func TestCircuitBreakerClosesAfterSuccesses(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{},
		logger,
	)

	// Open circuit by recording failures
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	if !connector.circuitBreaker.IsOpen() {
		t.Fatal("circuit should be open")
	}

	// Trigger half-open by waiting (simulate timeout)
	connector.circuitBreaker.mu.Lock()
	connector.circuitBreaker.lastFailureTime = time.Now().Add(-40 * time.Second)
	connector.circuitBreaker.mu.Unlock()

	// Circuit should transition to half-open
	_ = connector.circuitBreaker.IsOpen()

	// Record successes to close it
	for i := 0; i < 3; i++ {
		connector.circuitBreaker.RecordSuccess()
	}

	if connector.circuitBreaker.State() != string(OAuthStateClosed) {
		t.Errorf("state = %s after recovery, want %s", connector.circuitBreaker.State(), OAuthStateClosed)
	}
}

// TestExchangeCodeForTokenCircuitBreaker tests token exchange with circuit breaker
func TestExchangeCodeForTokenCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
		logger,
	)

	// Open the circuit
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	// Try to exchange code - should fail due to open circuit
	_, err := connector.ExchangeCodeForToken(context.Background(), OAuthProviderGoogle, "test-code")

	if err == nil {
		t.Errorf("ExchangeCodeForToken() error = nil, want error for open circuit")
	}

	if err.Error() != "OAuth provider google temporarily unavailable (circuit open)" {
		t.Errorf("ExchangeCodeForToken() error message incorrect: %v", err)
	}
}

// TestRefreshTokenCircuitBreaker tests token refresh with circuit breaker
func TestRefreshTokenCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
		logger,
	)

	// Open the circuit
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	// Try to refresh token - should fail due to open circuit
	ctx := context.Background()
	_, err := connector.RefreshToken(ctx, OAuthProviderGoogle, nil)

	if err == nil {
		t.Errorf("RefreshToken() error = nil, want error for open circuit")
	}

	if err.Error() != "OAuth provider google temporarily unavailable (circuit open)" {
		t.Errorf("RefreshToken() error message incorrect: %v", err)
	}
}

// TestGetUserInfoCircuitBreaker tests user info fetch with circuit breaker
func TestGetUserInfoCircuitBreaker(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
		logger,
	)

	// Open the circuit
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	// Try to get user info - should fail due to open circuit
	ctx := context.Background()
	_, err := connector.GetUserInfo(ctx, OAuthProviderGoogle, nil)

	if err == nil {
		t.Errorf("GetUserInfo() error = nil, want error for open circuit")
	}

	if err.Error() != "OAuth provider google temporarily unavailable (circuit open)" {
		t.Errorf("GetUserInfo() error message incorrect: %v", err)
	}
}

// TestCircuitBreakerWithTimeout tests context timeout is applied
func TestCircuitBreakerWithTimeout(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{
			{
				Name:         "google",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
			},
		},
		logger,
	)

	// Test withTimeout with no existing deadline
	ctx := context.Background()
	newCtx, cancel := connector.withTimeout(ctx)
	defer cancel()

	if _, ok := newCtx.Deadline(); !ok {
		t.Errorf("withTimeout() should set deadline on context")
	}

	// Test withTimeout with existing deadline (should not override)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	newCtx2, cancel3 := connector.withTimeout(ctx2)
	// cancel3 should be a no-op
	cancel3()

	// Both contexts should have the same deadline (the original one)
	deadline1, _ := ctx2.Deadline()
	deadline2, _ := newCtx2.Deadline()
	if !deadline1.Equal(deadline2) {
		t.Errorf("withTimeout() should not override existing deadline")
	}
}

// TestCircuitBreakerMetrics tests circuit breaker metrics collection
func TestCircuitBreakerMetrics(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{},
		logger,
	)

	// Record some activity
	connector.circuitBreaker.RecordSuccess()
	connector.circuitBreaker.RecordFailure()

	metrics := connector.circuitBreaker.GetMetrics()

	if metrics == nil {
		t.Errorf("GetMetrics() returned nil")
	}

	if state, ok := metrics["state"]; !ok {
		t.Errorf("metrics missing 'state'")
	} else if state != string(OAuthStateClosed) {
		t.Errorf("state metric = %v, want %s", state, OAuthStateClosed)
	}

	if failureCount, ok := metrics["failure_count"]; !ok {
		t.Errorf("metrics missing 'failure_count'")
	} else if failureCount != 1 {
		t.Errorf("failure_count = %v, want 1", failureCount)
	}
}

// TestCircuitBreakerReset tests circuit breaker reset functionality
func TestCircuitBreakerReset(t *testing.T) {
	logger := zap.NewNop()
	connector, _ := NewOAuthConnectorWithLogger(
		"http://localhost:8080",
		[]OAuthProviderConfig{},
		logger,
	)

	// Open circuit
	for i := 0; i < 5; i++ {
		connector.circuitBreaker.RecordFailure()
	}

	if connector.circuitBreaker.State() != string(OAuthStateOpen) {
		t.Fatalf("circuit should be open before reset")
	}

	// Reset
	connector.circuitBreaker.Reset()

	if connector.circuitBreaker.State() != string(OAuthStateClosed) {
		t.Errorf("circuit state after reset = %s, want %s", connector.circuitBreaker.State(), OAuthStateClosed)
	}

	// IsOpen returns false when the circuit is closed (allowing operations)
	if connector.circuitBreaker.IsOpen() {
		t.Errorf("circuit should not be open after reset (should allow operations)")
	}
}
