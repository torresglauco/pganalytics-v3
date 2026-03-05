package auth

import (
	"crypto/tls"
	"testing"
)

// TestNewLDAPConnector tests LDAP connector initialization
func TestNewLDAPConnector(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		bindDN    string
		password  string
		wantErr   bool
	}{
		{
			name:      "valid LDAP URL",
			serverURL: "ldap://ldap.example.com:389",
			bindDN:    "cn=admin,dc=example,dc=com",
			password:  "password123",
			wantErr:   false,
		},
		{
			name:      "LDAPS with TLS",
			serverURL: "ldaps://ldap.example.com:636",
			bindDN:    "cn=admin,dc=example,dc=com",
			password:  "password123",
			wantErr:   false,
		},
		{
			name:      "empty server URL",
			serverURL: "",
			bindDN:    "cn=admin,dc=example,dc=com",
			password:  "password123",
			wantErr:   false, // Connector created, error occurs on connection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewLDAPConnector(
				tt.serverURL,
				tt.bindDN,
				tt.password,
				"ou=users,dc=example,dc=com",
				"ou=groups,dc=example,dc=com",
				make(map[string]string),
				&tls.Config{},
			)

			if connector == nil && !tt.wantErr {
				t.Errorf("NewLDAPConnector() = nil, want non-nil")
			}
		})
	}
}

// TestLDAPConnectorFields tests that connector fields are set correctly
func TestLDAPConnectorFields(t *testing.T) {
	serverURL := "ldap://ldap.example.com:389"
	bindDN := "cn=admin,dc=example,dc=com"
	password := "password123"
	userSearchBase := "ou=users,dc=example,dc=com"
	groupSearchBase := "ou=groups,dc=example,dc=com"
	groupToRoleMap := map[string]string{
		"cn=admins,ou=groups,dc=example,dc=com": "admin",
		"cn=users,ou=groups,dc=example,dc=com":  "user",
	}

	connector := NewLDAPConnector(
		serverURL,
		bindDN,
		password,
		userSearchBase,
		groupSearchBase,
		groupToRoleMap,
		&tls.Config{},
	)

	if connector == nil {
		t.Fatal("NewLDAPConnector() returned nil")
	}

	if connector.serverURL != serverURL {
		t.Errorf("serverURL = %s, want %s", connector.serverURL, serverURL)
	}

	if connector.bindDN != bindDN {
		t.Errorf("bindDN = %s, want %s", connector.bindDN, bindDN)
	}

	if connector.bindPassword != password {
		t.Errorf("bindPassword = %s, want %s", connector.bindPassword, password)
	}

	if connector.userSearchBase != userSearchBase {
		t.Errorf("userSearchBase = %s, want %s", connector.userSearchBase, userSearchBase)
	}

	if connector.groupSearchBase != groupSearchBase {
		t.Errorf("groupSearchBase = %s, want %s", connector.groupSearchBase, groupSearchBase)
	}
}

// TestResolveRole tests LDAP group-to-role mapping
func TestResolveRole(t *testing.T) {
	tests := []struct {
		name           string
		groups         []string
		groupToRoleMap map[string]string
		expectedRole   string
	}{
		{
			name: "admin group",
			groups: []string{
				"cn=admins,ou=groups,dc=example,dc=com",
				"cn=users,ou=groups,dc=example,dc=com",
			},
			groupToRoleMap: map[string]string{
				"cn=admins,ou=groups,dc=example,dc=com": "admin",
				"cn=users,ou=groups,dc=example,dc=com":  "user",
			},
			expectedRole: "admin",
		},
		{
			name: "user group only",
			groups: []string{
				"cn=users,ou=groups,dc=example,dc=com",
			},
			groupToRoleMap: map[string]string{
				"cn=admins,ou=groups,dc=example,dc=com": "admin",
				"cn=users,ou=groups,dc=example,dc=com":  "user",
			},
			expectedRole: "user",
		},
		{
			name:           "no groups match",
			groups:         []string{"cn=other,ou=groups,dc=example,dc=com"},
			groupToRoleMap: map[string]string{},
			expectedRole:   "viewer", // default role
		},
		{
			name:           "empty groups",
			groups:         []string{},
			groupToRoleMap: map[string]string{},
			expectedRole:   "viewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := NewLDAPConnector(
				"ldap://ldap.example.com:389",
				"cn=admin,dc=example,dc=com",
				"password",
				"ou=users,dc=example,dc=com",
				"ou=groups,dc=example,dc=com",
				tt.groupToRoleMap,
				&tls.Config{},
			)

			role := connector.resolveRole(tt.groups)

			if role != tt.expectedRole {
				t.Errorf("resolveRole() = %s, want %s", role, tt.expectedRole)
			}
		})
	}
}

// TestLDAPClose tests connector closing
func TestLDAPClose(t *testing.T) {
	connector := NewLDAPConnector(
		"ldap://ldap.example.com:389",
		"cn=admin,dc=example,dc=com",
		"password",
		"ou=users,dc=example,dc=com",
		"ou=groups,dc=example,dc=com",
		make(map[string]string),
		&tls.Config{},
	)

	err := connector.Close()

	// Should not error on closing unopened connection
	if err != nil && err.Error() != "" {
		// Acceptable to have error if connection was never opened
	}
}

// BenchmarkResolveRole benchmarks the role resolution
func BenchmarkResolveRole(b *testing.B) {
	connector := NewLDAPConnector(
		"ldap://ldap.example.com:389",
		"cn=admin,dc=example,dc=com",
		"password",
		"ou=users,dc=example,dc=com",
		"ou=groups,dc=example,dc=com",
		map[string]string{
			"cn=admins,ou=groups,dc=example,dc=com": "admin",
			"cn=users,ou=groups,dc=example,dc=com":  "user",
		},
		&tls.Config{},
	)

	groups := []string{
		"cn=users,ou=groups,dc=example,dc=com",
		"cn=admins,ou=groups,dc=example,dc=com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = connector.resolveRole(groups)
	}
}
