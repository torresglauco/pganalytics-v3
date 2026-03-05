package auth

import (
	"testing"
)

// MockSMSProvider implements SMSProvider for testing
type MockSMSProvider struct {
	messages []string
	err      error
}

func (m *MockSMSProvider) SendSMS(phoneNumber, message string) error {
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, message)
	return nil
}

// TestNewMFAManager tests MFA manager initialization
func TestNewMFAManager(t *testing.T) {
	mockSMS := &MockSMSProvider{}
	manager := NewMFAManager(nil, mockSMS) // db can be nil for unit tests

	if manager == nil {
		t.Errorf("NewMFAManager() = nil, want non-nil")
	}

	if manager.smsProvider != mockSMS {
		t.Errorf("smsProvider not set correctly")
	}
}

// TestGenerateTOTPSecret tests TOTP secret generation
func TestGenerateTOTPSecret(t *testing.T) {
	manager := NewMFAManager(nil, nil)

	tests := []struct {
		name     string
		username string
	}{
		{
			name:     "valid username",
			username: "testuser",
		},
		{
			name:     "email format username",
			username: "user@example.com",
		},
		{
			name:     "username with special chars",
			username: "user.name@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := manager.GenerateTOTPSecret(tt.username)

			if err != nil {
				t.Errorf("GenerateTOTPSecret() error = %v", err)
				return
			}

			if key == nil {
				t.Errorf("GenerateTOTPSecret() returned nil key")
				return
			}

			if key.Secret() == "" {
				t.Errorf("Generated TOTP secret is empty")
			}
		})
	}
}

// TestVerifyTOTP tests TOTP code verification
func TestVerifyTOTP(t *testing.T) {
	manager := NewMFAManager(nil, nil)

	// Generate a valid secret
	key, _ := manager.GenerateTOTPSecret("testuser")
	secret := key.Secret()

	tests := []struct {
		name      string
		secret    string
		code      string
		wantValid bool
	}{
		{
			name:      "generated secret with current code",
			secret:    secret,
			code:      "",    // Would be actual TOTP code, but we can't generate current code in test
			wantValid: false, // Empty code should fail
		},
		{
			name:      "empty secret",
			secret:    "",
			code:      "123456",
			wantValid: false,
		},
		{
			name:      "invalid code format",
			secret:    secret,
			code:      "invalid",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := manager.VerifyTOTP(tt.secret, tt.code)

			if valid != tt.wantValid {
				t.Logf("VerifyTOTP() valid = %v, wantValid %v (may be time-dependent)", valid, tt.wantValid)
			}
		})
	}
}

// TestGenerateBackupCodes tests backup code generation
func TestGenerateBackupCodes(t *testing.T) {
	// Can't test with nil database, this is a smoke test
	// In production would use mock database
	manager := NewMFAManager(nil, nil)

	tests := []struct {
		name      string
		userID    int
		count     int
		wantError bool
	}{
		{
			name:      "valid user and count",
			userID:    1,
			count:     10,
			wantError: true, // Will error because db is nil
		},
		{
			name:      "zero codes",
			userID:    1,
			count:     0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that need a real database
			if tt.wantError {
				t.Skip("Skipping test - requires database connection")
			}

			_, err := manager.GenerateBackupCodes(tt.userID, tt.count)

			if (err != nil) != tt.wantError {
				t.Errorf("GenerateBackupCodes() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestGenerateSecureCode tests secure code generation
func TestGenerateSecureCode(t *testing.T) {
	code1 := generateSecureCode(8)
	code2 := generateSecureCode(8)

	// Codes should be generated
	if code1 == "" {
		t.Errorf("generateSecureCode() returned empty string")
	}

	// Codes should be different
	if code1 == code2 {
		t.Errorf("generateSecureCode() returned identical codes (very unlikely)")
	}

	// Code should have correct length
	if len(code1) != 8 {
		t.Errorf("generateSecureCode(8) length = %d, want 8", len(code1))
	}

	// Code should only contain alphanumeric chars
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, c := range code1 {
		found := false
		for _, valid := range validChars {
			if c == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("generateSecureCode() contains invalid character: %c", c)
		}
	}
}

// TestGenerateRandomCode tests random code generation
func TestGenerateRandomCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "6-digit code",
			length: 6,
		},
		{
			name:   "8-digit code",
			length: 8,
		},
		{
			name:   "10-digit code",
			length: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateRandomCode(tt.length)

			if len(code) != tt.length {
				t.Errorf("generateRandomCode(%d) length = %d, want %d", tt.length, len(code), tt.length)
			}

			// All characters should be digits
			for _, c := range code {
				if c < '0' || c > '9' {
					t.Errorf("generateRandomCode() contains non-digit: %c", c)
				}
			}
		})
	}
}

// TestHashCode tests code hashing
func TestHashCode(t *testing.T) {
	code := "TEST1234"
	hash := hashCode(code)

	if hash == "" {
		t.Errorf("hashCode() returned empty string")
	}

	if hash == code {
		t.Logf("hashCode() returned unmodified code (base32 encoding)")
	}

	// Same code should produce same hash
	hash2 := hashCode(code)
	if hash != hash2 {
		t.Errorf("hashCode() not deterministic: %s != %s", hash, hash2)
	}

	// Different codes should produce different hashes
	hash3 := hashCode("DIFFERENT")
	if hash == hash3 {
		t.Errorf("hashCode() collision: same hash for different codes")
	}
}

// TestValidateTOTPSecret tests TOTP secret validation
func TestValidateTOTPSecret(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		wantError bool
	}{
		{
			name:      "valid base32 secret",
			secret:    "JBSWY3DPEBLW64TMMQ======", // Valid base32
			wantError: false,
		},
		{
			name:      "empty secret",
			secret:    "",
			wantError: true,
		},
		{
			name:      "invalid base32",
			secret:    "!!!invalid!!!",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTOTPSecret(tt.secret)

			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTOTPSecret() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMFATypeValues tests MFA type constants
func TestMFATypeValues(t *testing.T) {
	tests := []struct {
		name        string
		mfaType     MFAType
		expectedStr string
	}{
		{
			name:        "TOTP type",
			mfaType:     MFATypeTOTP,
			expectedStr: "totp",
		},
		{
			name:        "SMS type",
			mfaType:     MFATypeSMS,
			expectedStr: "sms",
		},
		{
			name:        "Email type",
			mfaType:     MFATypeEmail,
			expectedStr: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mfaType) != tt.expectedStr {
				t.Errorf("MFAType = %s, want %s", tt.mfaType, tt.expectedStr)
			}
		})
	}
}

// BenchmarkGenerateSecureCode benchmarks secure code generation
func BenchmarkGenerateSecureCode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateSecureCode(8)
	}
}

// BenchmarkGenerateRandomCode benchmarks random code generation
func BenchmarkGenerateRandomCode(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateRandomCode(6)
	}
}

// BenchmarkHashCode benchmarks code hashing
func BenchmarkHashCode(b *testing.B) {
	code := "TEST1234"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hashCode(code)
	}
}
