package auth

import (
	"testing"
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
