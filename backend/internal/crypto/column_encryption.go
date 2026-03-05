package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// ColumnEncryption handles column-level encryption/decryption
type ColumnEncryption struct {
	keyManager KeyManager
}

// NewColumnEncryption creates a new column encryption system
func NewColumnEncryption(keyManager KeyManager) *ColumnEncryption {
	return &ColumnEncryption{
		keyManager: keyManager,
	}
}

// EncryptString encrypts a string value
func (ce *ColumnEncryption) EncryptString(ctx interface{}, plaintext string) (string, int, error) {
	if plaintext == "" {
		return "", 0, nil
	}

	// Get current key
	key, version, err := ce.keyManager.GetCurrentKey(nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt
	ciphertext, err := ce.encrypt(key, []byte(plaintext))
	if err != nil {
		return "", 0, fmt.Errorf("encryption failed: %w", err)
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	// Prepend version for key rotation support
	// Format: v{version}:{ciphertext}
	versionedCiphertext := fmt.Sprintf("v%d:%s", version, encoded)

	return versionedCiphertext, version, nil
}

// DecryptString decrypts a string value
func (ce *ColumnEncryption) DecryptString(ctx interface{}, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Parse version from ciphertext
	var version int
	var encodedCiphertext string

	// Try new format first (v{version}:{data})
	n, err := fmt.Sscanf(ciphertext, "v%d:%s", &version, &encodedCiphertext)
	if n != 2 || err != nil {
		// Fallback to old format (assume version 1)
		version = 1
		encodedCiphertext = ciphertext
	}

	// Decode from base64
	encryptedData, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Get key by version
	key, err := ce.keyManager.GetKeyByVersion(nil, version)
	if err != nil {
		return "", fmt.Errorf("failed to get decryption key: %w", err)
	}

	// Decrypt
	plaintext, err := ce.decrypt(key, encryptedData)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts binary data
func (ce *ColumnEncryption) EncryptBytes(ctx interface{}, data []byte) ([]byte, int, error) {
	if len(data) == 0 {
		return []byte{}, 0, nil
	}

	// Get current key
	key, version, err := ce.keyManager.GetCurrentKey(nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt
	ciphertext, err := ce.encrypt(key, data)
	if err != nil {
		return nil, 0, fmt.Errorf("encryption failed: %w", err)
	}

	// Prepend version (1 byte) + ciphertext
	versionedCiphertext := make([]byte, 1+len(ciphertext))
	versionedCiphertext[0] = byte(version)
	copy(versionedCiphertext[1:], ciphertext)

	return versionedCiphertext, version, nil
}

// DecryptBytes decrypts binary data
func (ce *ColumnEncryption) DecryptBytes(ctx interface{}, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return []byte{}, nil
	}

	// Extract version (first byte)
	version := int(ciphertext[0])
	encryptedData := ciphertext[1:]

	// Get key by version
	key, err := ce.keyManager.GetKeyByVersion(nil, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get decryption key: %w", err)
	}

	// Decrypt
	plaintext, err := ce.decrypt(key, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// Internal encryption/decryption functions

// encrypt encrypts data using AES-256-GCM
func (ce *ColumnEncryption) encrypt(key []byte, plaintext []byte) ([]byte, error) {
	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt: nonce + ciphertext
	// Format: [nonce (12 bytes)] + [ciphertext + tag (variable)]
	ciphertext := aead.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func (ce *ColumnEncryption) decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce from beginning of ciphertext
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// MigrationHelper helps with data migration during encryption enablement
type MigrationHelper struct {
	ce *ColumnEncryption
}

// NewMigrationHelper creates a new migration helper
func NewMigrationHelper(ce *ColumnEncryption) *MigrationHelper {
	return &MigrationHelper{
		ce: ce,
	}
}

// EncryptExistingData encrypts existing plaintext data
// This is used for one-time migration of data
func (mh *MigrationHelper) EncryptExistingData(plaintext string) (string, error) {
	encrypted, _, err := mh.ce.EncryptString(nil, plaintext)
	return encrypted, err
}

// BatchEncryptData encrypts multiple values
func (mh *MigrationHelper) BatchEncryptData(plaintexts []string) ([]string, error) {
	encrypted := make([]string, 0, len(plaintexts))

	for _, plaintext := range plaintexts {
		enc, _, err := mh.ce.EncryptString(nil, plaintext)
		if err != nil {
			return nil, fmt.Errorf("batch encryption failed: %w", err)
		}
		encrypted = append(encrypted, enc)
	}

	return encrypted, nil
}

// VerifyEncryption verifies that encrypted data can be decrypted
func (mh *MigrationHelper) VerifyEncryption(plaintext, encrypted string) error {
	decrypted, err := mh.ce.DecryptString(nil, encrypted)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	if plaintext != decrypted {
		return fmt.Errorf("decrypted data does not match original")
	}

	return nil
}

// SensitiveField represents a field that should be encrypted
type SensitiveField struct {
	TableName  string
	ColumnName string
	IsActive   bool // Whether to use encrypted column
}

// CommonSensitiveFields returns list of commonly encrypted fields
func CommonSensitiveFields() []SensitiveField {
	return []SensitiveField{
		{TableName: "users", ColumnName: "email", IsActive: true},
		{TableName: "users", ColumnName: "password_hash", IsActive: true},
		{TableName: "registration_secrets", ColumnName: "secret_value", IsActive: true},
		{TableName: "postgresql_instances", ColumnName: "connection_string", IsActive: true},
		{TableName: "api_tokens", ColumnName: "token_hash", IsActive: true},
		{TableName: "secrets", ColumnName: "secret_encrypted", IsActive: true},
	}
}
