package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthProvider represents an OAuth provider
type OAuthProvider string

const (
	OAuthProviderGoogle  OAuthProvider = "google"
	OAuthProviderGitHub  OAuthProvider = "github"
	OAuthProviderAzureAD OAuthProvider = "azure_ad"
	OAuthProviderCustom  OAuthProvider = "custom"
)

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
	providers map[OAuthProvider]*oauth2.Config
	rootURL   string
}

// NewOAuthConnector creates a new OAuth connector
func NewOAuthConnector(rootURL string, providerConfigs []OAuthProviderConfig) (*OAuthConnector, error) {
	oc := &OAuthConnector{
		providers: make(map[OAuthProvider]*oauth2.Config),
		rootURL:   rootURL,
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
	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// RefreshToken refreshes an OAuth token
func (oc *OAuthConnector) RefreshToken(ctx context.Context, provider OAuthProvider, token *oauth2.Token) (*oauth2.Token, error) {
	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

	// Create a token source for refreshing
	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

// GetUserInfo retrieves user information from OAuth provider
func (oc *OAuthConnector) GetUserInfo(ctx context.Context, provider OAuthProvider, token *oauth2.Token) (*OAuthUserInfo, error) {
	config, ok := oc.providers[provider]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", provider)
	}

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
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %d - %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
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
		return nil, fmt.Errorf("user email not available from provider")
	}

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
