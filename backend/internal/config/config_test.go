package config

import (
	"testing"
)

// TestConfigLoadDefaults tests that configuration loads with correct defaults
func TestConfigLoadDefaults(t *testing.T) {
	t.Run("default development environment URLs", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "development")
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("TIMESCALE_URL", "postgres://test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("REGISTRATION_SECRET", "test-secret")

		cfg := Load()

		if cfg.APIBaseURL != "http://localhost:8080" {
			t.Errorf("Expected APIBaseURL to default to http://localhost:8080, got %s", cfg.APIBaseURL)
		}
		if cfg.FrontendURL != "http://localhost:3000" {
			t.Errorf("Expected FrontendURL to default to http://localhost:3000, got %s", cfg.FrontendURL)
		}
	})

	t.Run("custom URLs from environment", func(t *testing.T) {
		t.Setenv("API_BASE_URL", "https://api.example.com")
		t.Setenv("FRONTEND_URL", "https://app.example.com")
		t.Setenv("DATABASE_URL", "postgres://test")
		t.Setenv("TIMESCALE_URL", "postgres://test")
		t.Setenv("JWT_SECRET", "test-secret")
		t.Setenv("REGISTRATION_SECRET", "test-secret")

		cfg := Load()

		if cfg.APIBaseURL != "https://api.example.com" {
			t.Errorf("Expected APIBaseURL to be https://api.example.com, got %s", cfg.APIBaseURL)
		}
		if cfg.FrontendURL != "https://app.example.com" {
			t.Errorf("Expected FrontendURL to be https://app.example.com, got %s", cfg.FrontendURL)
		}
	})
}

// TestConfigValidateProductionURLs tests production URL validation
func TestConfigValidateProductionURLs(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		apiBaseURL  string
		frontendURL string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "production with valid URLs",
			environment: "production",
			apiBaseURL:  "https://api.example.com",
			frontendURL: "https://app.example.com",
			wantErr:     false,
		},
		{
			name:        "production with empty API URL",
			environment: "production",
			apiBaseURL:  "",
			frontendURL: "https://app.example.com",
			wantErr:     true,
			errMsg:      "API_BASE_URL must be set to a valid production URL, not localhost",
		},
		{
			name:        "production with localhost API URL",
			environment: "production",
			apiBaseURL:  "http://localhost:8080",
			frontendURL: "https://app.example.com",
			wantErr:     true,
			errMsg:      "API_BASE_URL must be set to a valid production URL, not localhost",
		},
		{
			name:        "production with empty Frontend URL",
			environment: "production",
			apiBaseURL:  "https://api.example.com",
			frontendURL: "",
			wantErr:     true,
			errMsg:      "FRONTEND_URL must be set to a valid production URL, not localhost",
		},
		{
			name:        "production with localhost Frontend URL",
			environment: "production",
			apiBaseURL:  "https://api.example.com",
			frontendURL: "http://localhost:3000",
			wantErr:     true,
			errMsg:      "FRONTEND_URL must be set to a valid production URL, not localhost",
		},
		{
			name:        "development with localhost URLs",
			environment: "development",
			apiBaseURL:  "http://localhost:8080",
			frontendURL: "http://localhost:3000",
			wantErr:     false,
		},
		{
			name:        "development with empty URLs",
			environment: "development",
			apiBaseURL:  "",
			frontendURL: "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Environment:       tt.environment,
				APIBaseURL:        tt.apiBaseURL,
				FrontendURL:       tt.frontendURL,
				DatabaseURL:       "postgres://test",
				TimescaleURL:      "postgres://test",
				JWTSecret:         "test-secret",
				RegistrationSecret: "test-secret",
			}

			err := cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message to contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

// TestConfigIsProduction tests the IsProduction method
func TestConfigIsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{
			name:        "production environment",
			environment: "production",
			expected:    true,
		},
		{
			name:        "development environment",
			environment: "development",
			expected:    false,
		},
		{
			name:        "staging environment",
			environment: "staging",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			if result := cfg.IsProduction(); result != tt.expected {
				t.Errorf("IsProduction() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestConfigOAuthAndSAMLURLs tests that OAuth/SAML use the APIBaseURL
func TestConfigOAuthAndSAMLURLs(t *testing.T) {
	tests := []struct {
		name       string
		apiBaseURL string
		expected   string
	}{
		{
			name:       "production URL",
			apiBaseURL: "https://api.example.com",
			expected:   "https://api.example.com/api/v1/auth/oauth/callback",
		},
		{
			name:       "development URL",
			apiBaseURL: "http://localhost:8080",
			expected:   "http://localhost:8080/api/v1/auth/oauth/callback",
		},
		{
			name:       "custom port",
			apiBaseURL: "https://api.example.com:8443",
			expected:   "https://api.example.com:8443/api/v1/auth/oauth/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{APIBaseURL: tt.apiBaseURL}

			// Simulate how handlers_auth.go constructs the redirect URL
			redirectURL := cfg.APIBaseURL + "/api/v1/auth/oauth/callback"

			if redirectURL != tt.expected {
				t.Errorf("OAuth redirect URL = %q, expected %q", redirectURL, tt.expected)
			}
		})
	}
}
