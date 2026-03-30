package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/internal/config"
	"go.uber.org/zap/zaptest"
)

// TestValidateAuthConfiguration_LDAPEnabled tests LDAP validation when enabled
func TestValidateAuthConfiguration_LDAPEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{"admin": "admin", "users": "user"}`,
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error for valid LDAP config, got: %v", err)
	}
}

// TestValidateAuthConfiguration_LDAPMissingServerURL tests that missing LDAP server URL is caught
func TestValidateAuthConfiguration_LDAPMissingServerURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "", // Missing
		LDAPGroupToRoleJSON: `{"admin": "admin"}`,
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing LDAP server URL, got nil")
	}
	if err.Error() != "LDAP enabled but server URL not configured (set LDAP_SERVER_URL)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_LDAPInvalidJSON tests that invalid LDAP JSON is caught
func TestValidateAuthConfiguration_LDAPInvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{invalid json}`, // Invalid JSON
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for invalid LDAP JSON, got nil")
	}
	if err.Error() == "" {
		t.Error("Expected error message describing JSON parse failure")
	}
}

// TestValidateAuthConfiguration_LDAPEmptyMappings tests that empty LDAP mappings are caught
func TestValidateAuthConfiguration_LDAPEmptyMappings(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{}`, // Empty mappings
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for empty LDAP mappings, got nil")
	}
	if err.Error() != "LDAP group mappings are empty (set LDAP_GROUP_TO_ROLE_MAPPING with valid JSON mappings)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_OAuthEnabled tests OAuth validation when enabled
func TestValidateAuthConfiguration_OAuthEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "google",
			ClientID:     "google-client-id",
			ClientSecret: "google-secret",
			Scopes:       []string{"email", "profile"},
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error for valid OAuth config, got: %v", err)
	}
}

// TestValidateAuthConfiguration_OAuthMultipleProviders tests OAuth with multiple providers
func TestValidateAuthConfiguration_OAuthMultipleProviders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "google",
			ClientID:     "google-client-id",
			ClientSecret: "google-secret",
		},
		{
			Name:         "github",
			ClientID:     "github-client-id",
			ClientSecret: "github-secret",
		},
		{
			Name:         "azure_ad",
			ClientID:     "azure-client-id",
			ClientSecret: "azure-secret",
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error for valid multi-provider OAuth config, got: %v", err)
	}
}

// TestValidateAuthConfiguration_OAuthInvalidJSON tests that invalid OAuth JSON is caught
func TestValidateAuthConfiguration_OAuthInvalidJSON(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: `{invalid json}`, // Invalid JSON
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for invalid OAuth JSON, got nil")
	}
}

// TestValidateAuthConfiguration_OAuthEmptyProviders tests that empty OAuth providers are caught
func TestValidateAuthConfiguration_OAuthEmptyProviders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: `[]`, // Empty array
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for empty OAuth providers, got nil")
	}
	if err.Error() != "OAuth enabled but no providers configured (set OAUTH_PROVIDERS with valid JSON array)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_OAuthMissingClientID tests that missing OAuth client ID is caught
func TestValidateAuthConfiguration_OAuthMissingClientID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "google",
			ClientID:     "", // Missing
			ClientSecret: "secret",
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing OAuth ClientID, got nil")
	}
}

// TestValidateAuthConfiguration_OAuthMissingClientSecret tests that missing OAuth client secret is caught
func TestValidateAuthConfiguration_OAuthMissingClientSecret(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "google",
			ClientID:     "client-id",
			ClientSecret: "", // Missing
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing OAuth ClientSecret, got nil")
	}
}

// TestValidateAuthConfiguration_OAuthCustomProviderMissingAuthURL tests that custom OIDC provider requires auth_url
func TestValidateAuthConfiguration_OAuthCustomProviderMissingAuthURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "custom",
			ClientID:     "client-id",
			ClientSecret: "secret",
			TokenURL:     "https://example.com/token",
			AuthURL:      "", // Missing
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing custom provider auth_url, got nil")
	}
}

// TestValidateAuthConfiguration_OAuthCustomProviderMissingTokenURL tests that custom OIDC provider requires token_url
func TestValidateAuthConfiguration_OAuthCustomProviderMissingTokenURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "custom",
			ClientID:     "client-id",
			ClientSecret: "secret",
			AuthURL:      "https://example.com/auth",
			TokenURL:     "", // Missing
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		OAuthEnabled:       true,
		OAuthProvidersJSON: string(providersJSON),
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing custom provider token_url, got nil")
	}
}

// TestValidateAuthConfiguration_SAMLEnabled tests SAML validation when enabled
func TestValidateAuthConfiguration_SAMLEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create temporary certificate and key files
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Write dummy cert and key files
	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       certPath,
		SAMLKeyPath:        keyPath,
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error for valid SAML config, got: %v", err)
	}
}

// TestValidateAuthConfiguration_SAMLMissingCertPath tests that missing SAML cert path is caught
func TestValidateAuthConfiguration_SAMLMissingCertPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       "", // Missing
		SAMLKeyPath:        "/tmp/key.pem",
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing SAML cert path, got nil")
	}
	if err.Error() != "SAML enabled but certificate path not configured (set SAML_CERT_PATH)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_SAMLMissingKeyPath tests that missing SAML key path is caught
func TestValidateAuthConfiguration_SAMLMissingKeyPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")

	// Write dummy cert file
	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       certPath,
		SAMLKeyPath:        "", // Missing
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing SAML key path, got nil")
	}
	if err.Error() != "SAML enabled but key path not configured (set SAML_KEY_PATH)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_SAMLCertFileNotFound tests that missing cert file is caught
func TestValidateAuthConfiguration_SAMLCertFileNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Write dummy key file
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       "/nonexistent/cert.pem",
		SAMLKeyPath:        keyPath,
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing cert file, got nil")
	}
}

// TestValidateAuthConfiguration_SAMLKeyFileNotFound tests that missing key file is caught
func TestValidateAuthConfiguration_SAMLKeyFileNotFound(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")

	// Write dummy cert file
	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       certPath,
		SAMLKeyPath:        "/nonexistent/key.pem",
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing key file, got nil")
	}
}

// TestValidateAuthConfiguration_SAMLMissingIDPMetadataURL tests that missing IdP metadata URL is caught
func TestValidateAuthConfiguration_SAMLMissingIDPMetadataURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Write dummy cert and key files
	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       certPath,
		SAMLKeyPath:        keyPath,
		SAMLIDPMetadataURL: "", // Missing
		SAMLEntityID:       "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing SAML IdP metadata URL, got nil")
	}
	if err.Error() != "SAML enabled but IdP metadata URL not configured (set SAML_IDP_METADATA_URL)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_SAMLMissingEntityID tests that missing SAML entity ID is caught
func TestValidateAuthConfiguration_SAMLMissingEntityID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	// Write dummy cert and key files
	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	cfg := &config.Config{
		SAMLEnabled:        true,
		SAMLCertPath:       certPath,
		SAMLKeyPath:        keyPath,
		SAMLIDPMetadataURL: "https://idp.example.com/metadata",
		SAMLEntityID:       "", // Missing
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for missing SAML entity ID, got nil")
	}
	if err.Error() != "SAML enabled but entity ID not configured (set SAML_ENTITY_ID)" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// TestValidateAuthConfiguration_AllDisabled tests that validation passes when auth methods are disabled
func TestValidateAuthConfiguration_AllDisabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{
		LDAPEnabled:  false,
		OAuthEnabled: false,
		SAMLEnabled:  false,
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error when all auth methods disabled, got: %v", err)
	}
}

// TestValidateAuthConfiguration_MultipleMethodsEnabled tests validation with multiple methods enabled
func TestValidateAuthConfiguration_MultipleMethodsEnabled(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Prepare temporary SAML files
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	// Prepare OAuth config
	providers := []auth.OAuthProviderConfig{
		{
			Name:         "google",
			ClientID:     "google-client-id",
			ClientSecret: "google-secret",
		},
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{"admin": "admin"}`,
		OAuthEnabled:        true,
		OAuthProvidersJSON:  string(providersJSON),
		SAMLEnabled:         true,
		SAMLCertPath:        certPath,
		SAMLKeyPath:         keyPath,
		SAMLIDPMetadataURL:  "https://idp.example.com/metadata",
		SAMLEntityID:        "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err != nil {
		t.Errorf("Expected no error with valid multi-method config, got: %v", err)
	}
}

// TestValidateAuthConfiguration_MultipleMethodsEnabledWithInvalidOAuth tests that validation fails on first invalid method
func TestValidateAuthConfiguration_MultipleMethodsEnabledWithInvalidOAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Prepare temporary SAML files
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}

	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{"admin": "admin"}`,
		OAuthEnabled:        true,
		OAuthProvidersJSON:  `[]`, // Empty - invalid
		SAMLEnabled:         true,
		SAMLCertPath:        certPath,
		SAMLKeyPath:         keyPath,
		SAMLIDPMetadataURL:  "https://idp.example.com/metadata",
		SAMLEntityID:        "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	err := server.ValidateAuthConfiguration()
	if err == nil {
		t.Error("Expected error for invalid OAuth config, got nil")
	}
}

// Benchmark tests to ensure validation is performant
// BenchmarkValidateAuthConfiguration_AllMethodsEnabled benchmarks validation performance
func BenchmarkValidateAuthConfiguration_AllMethodsEnabled(b *testing.B) {
	logger := zaptest.NewLogger(b)

	// Prepare temporary SAML files
	tmpDir := b.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	if err := os.WriteFile(certPath, []byte("dummy cert"), 0600); err != nil {
		b.Fatalf("Failed to create temp cert file: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy key"), 0600); err != nil {
		b.Fatalf("Failed to create temp key file: %v", err)
	}

	// Prepare OAuth config with multiple providers
	providers := make([]auth.OAuthProviderConfig, 0)
	for i := 0; i < 10; i++ {
		providers = append(providers, auth.OAuthProviderConfig{
			Name:         "provider" + string(rune(i)),
			ClientID:     "client-id",
			ClientSecret: "secret",
		})
	}
	providersJSON, _ := json.Marshal(providers)

	cfg := &config.Config{
		LDAPEnabled:         true,
		LDAPServerURL:       "ldap://ldap.example.com:389",
		LDAPGroupToRoleJSON: `{"admin": "admin", "users": "user", "guests": "guest"}`,
		OAuthEnabled:        true,
		OAuthProvidersJSON:  string(providersJSON),
		SAMLEnabled:         true,
		SAMLCertPath:        certPath,
		SAMLKeyPath:         keyPath,
		SAMLIDPMetadataURL:  "https://idp.example.com/metadata",
		SAMLEntityID:        "https://app.example.com",
	}

	server := &Server{
		config: cfg,
		logger: logger,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.ValidateAuthConfiguration()
	}
}
