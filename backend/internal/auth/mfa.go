package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFAType represents the type of MFA
type MFAType string

const (
	MFATypeTOTP  MFAType = "totp"
	MFATypeSMS   MFAType = "sms"
	MFATypeEmail MFAType = "email"
)

// MFAMethod represents a user's MFA method
type MFAMethod struct {
	ID        int
	UserID    int
	Type      MFAType
	Secret    string // Base32-encoded secret
	Verified  bool
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BackupCode represents a backup code
type BackupCode struct {
	ID      int
	UserID  int
	Code    string // Hashed
	Used    bool
	UsedAt  *time.Time
	Created time.Time
}

// MFAManager handles multi-factor authentication
type MFAManager struct {
	db          *sql.DB
	smsProvider SMSProvider
}

// SMSProvider defines SMS delivery interface
type SMSProvider interface {
	SendSMS(phoneNumber, message string) error
}

// NewMFAManager creates a new MFA manager
func NewMFAManager(db *sql.DB, smsProvider SMSProvider) *MFAManager {
	return &MFAManager{
		db:          db,
		smsProvider: smsProvider,
	}
}

// GenerateTOTPSecret generates a TOTP secret
func (m *MFAManager) GenerateTOTPSecret(username string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "pgAnalytics",
		AccountName: username,
		Period:      30,
		SecretSize:  32,
		Digits:      otp.DigitsSix,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key, nil
}

// VerifyTOTP verifies a TOTP code
func (m *MFAManager) VerifyTOTP(secret string, code string) bool {
	return totp.Validate(code, secret)
}

// SetupTOTP sets up TOTP for a user
func (m *MFAManager) SetupTOTP(userID int, secret string) (*MFAMethod, error) {
	query := `
		INSERT INTO user_mfa_methods (user_id, type, secret_encrypted, verified, enabled)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, type) DO UPDATE
		SET secret_encrypted = $3, verified = $4, enabled = $5, updated_at = NOW()
		RETURNING id, user_id, type, verified, enabled, created_at, updated_at
	`

	var method MFAMethod
	err := m.db.QueryRow(query, userID, MFATypeTOTP, secret, false, false).Scan(
		&method.ID, &method.UserID, &method.Type, &method.Verified, &method.Enabled,
		&method.CreatedAt, &method.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to setup TOTP: %w", err)
	}

	method.Secret = secret
	return &method, nil
}

// VerifyAndEnableTOTP verifies TOTP code and enables it
func (m *MFAManager) VerifyAndEnableTOTP(userID int, code string) error {
	// Get the TOTP secret
	query := `SELECT secret_encrypted FROM user_mfa_methods WHERE user_id = $1 AND type = $2`

	var secret string
	err := m.db.QueryRow(query, userID, MFATypeTOTP).Scan(&secret)
	if err != nil {
		return fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	// Verify the code
	if !m.VerifyTOTP(secret, code) {
		return fmt.Errorf("invalid TOTP code")
	}

	// Mark as verified and enabled
	updateQuery := `
		UPDATE user_mfa_methods
		SET verified = true, enabled = true, updated_at = NOW()
		WHERE user_id = $1 AND type = $2
	`

	_, err = m.db.Exec(updateQuery, userID, MFATypeTOTP)
	if err != nil {
		return fmt.Errorf("failed to enable TOTP: %w", err)
	}

	return nil
}

// SendSMSCode sends an SMS code to the user
func (m *MFAManager) SendSMSCode(userID int, phoneNumber string) (string, error) {
	if m.smsProvider == nil {
		return "", fmt.Errorf("SMS provider not configured")
	}

	// Generate random 6-digit code
	code := generateRandomCode(6)

	// Send SMS
	message := fmt.Sprintf("Your pgAnalytics verification code is: %s. Valid for 10 minutes.", code)
	err := m.smsProvider.SendSMS(phoneNumber, message)
	if err != nil {
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}

	// Store code in cache (Redis) with TTL of 10 minutes
	// For now, we return the code (in production, use Redis)

	return code, nil
}

// VerifySMSCode verifies an SMS code
// Note: In production, this should check against cached code
func (m *MFAManager) VerifySMSCode(userID int, code string, expectedCode string) bool {
	return code == expectedCode
}

// GenerateBackupCodes generates backup codes for a user
func (m *MFAManager) GenerateBackupCodes(userID int, count int) ([]string, error) {
	var codes []string

	for i := 0; i < count; i++ {
		code := generateSecureCode(8)
		codes = append(codes, code)

		// Hash and store in database
		codeHash := hashCode(code)

		query := `
			INSERT INTO user_backup_codes (user_id, code_hash, used)
			VALUES ($1, $2, $3)
		`

		_, err := m.db.Exec(query, userID, codeHash, false)
		if err != nil {
			return nil, fmt.Errorf("failed to store backup code: %w", err)
		}
	}

	return codes, nil
}

// ValidateBackupCode validates and marks a backup code as used
func (m *MFAManager) ValidateBackupCode(userID int, code string) error {
	codeHash := hashCode(code)

	query := `
		UPDATE user_backup_codes
		SET used = true, used_at = NOW()
		WHERE user_id = $1 AND code_hash = $2 AND used = false
		RETURNING id
	`

	var id int
	err := m.db.QueryRow(query, userID, codeHash).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid or already-used backup code")
		}
		return fmt.Errorf("failed to validate backup code: %w", err)
	}

	return nil
}

// GetUserMFAMethods gets all MFA methods for a user
func (m *MFAManager) GetUserMFAMethods(userID int) ([]MFAMethod, error) {
	query := `
		SELECT id, user_id, type, verified, enabled, created_at, updated_at
		FROM user_mfa_methods
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query MFA methods: %w", err)
	}
	defer rows.Close()

	var methods []MFAMethod
	for rows.Next() {
		var method MFAMethod
		err := rows.Scan(
			&method.ID, &method.UserID, &method.Type, &method.Verified,
			&method.Enabled, &method.CreatedAt, &method.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MFA method: %w", err)
		}
		methods = append(methods, method)
	}

	return methods, rows.Err()
}

// HasEnabledMFA checks if user has enabled MFA
func (m *MFAManager) HasEnabledMFA(userID int) (bool, error) {
	query := `SELECT COUNT(*) FROM user_mfa_methods WHERE user_id = $1 AND enabled = true`

	var count int
	err := m.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check MFA status: %w", err)
	}

	return count > 0, nil
}

// DisableMFA disables a specific MFA method
func (m *MFAManager) DisableMFA(userID int, mfaType MFAType) error {
	query := `
		UPDATE user_mfa_methods
		SET enabled = false, updated_at = NOW()
		WHERE user_id = $1 AND type = $2
	`

	result, err := m.db.Exec(query, userID, mfaType)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("MFA method not found")
	}

	return nil
}

// GetBackupCodeCount gets the number of unused backup codes
func (m *MFAManager) GetBackupCodeCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM user_backup_codes WHERE user_id = $1 AND used = false`

	var count int
	err := m.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get backup code count: %w", err)
	}

	return count, nil
}

// Helper functions

func generateRandomCode(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = digits[b[i]%byte(len(digits))]
	}
	return string(b)
}

func generateSecureCode(length int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return string(b)
}

func hashCode(code string) string {
	// In production, use bcrypt or similar
	// For now, use base32 encoding as placeholder
	return base32.StdEncoding.EncodeToString([]byte(code))
}

// ValidateTOTPSecret validates that a TOTP secret is valid
func ValidateTOTPSecret(secret string) error {
	// Check for empty secret
	if secret == "" {
		return fmt.Errorf("TOTP secret cannot be empty")
	}

	// Decode base32
	_, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("invalid base32 secret: %w", err)
	}
	return nil
}
