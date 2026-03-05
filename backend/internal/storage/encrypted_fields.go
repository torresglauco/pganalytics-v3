package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/torresglauco/pganalytics-v3/backend/internal/crypto"
)

// EncryptedField represents an encrypted database field with support for value scanning and driving
type EncryptedField struct {
	plaintext  string
	ciphertext string
	encrypted  bool
	version    int
	encryptor  *crypto.ColumnEncryption
}

// NewEncryptedField creates a new encrypted field
func NewEncryptedField(encryptor *crypto.ColumnEncryption, ciphertext string) *EncryptedField {
	return &EncryptedField{
		ciphertext: ciphertext,
		encrypted:  true,
		encryptor:  encryptor,
	}
}

// NewPlaintextField creates a new plaintext field that will be encrypted
func NewPlaintextField(encryptor *crypto.ColumnEncryption, plaintext string) *EncryptedField {
	return &EncryptedField{
		plaintext: plaintext,
		encrypted: false,
		encryptor: encryptor,
	}
}

// GetDecrypted returns the decrypted plaintext
func (ef *EncryptedField) GetDecrypted() (string, error) {
	if ef.plaintext != "" {
		return ef.plaintext, nil
	}

	if ef.ciphertext == "" {
		return "", nil
	}

	if ef.encryptor == nil {
		return "", fmt.Errorf("encryptor not configured")
	}

	plaintext, err := ef.encryptor.DecryptString(nil, ef.ciphertext)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	ef.plaintext = plaintext
	return plaintext, nil
}

// GetCiphertext returns the encrypted ciphertext (encrypts if needed)
func (ef *EncryptedField) GetCiphertext() (string, error) {
	if ef.ciphertext != "" {
		return ef.ciphertext, nil
	}

	if ef.plaintext == "" {
		return "", nil
	}

	if ef.encryptor == nil {
		return "", fmt.Errorf("encryptor not configured")
	}

	ciphertext, _, err := ef.encryptor.EncryptString(nil, ef.plaintext)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	ef.ciphertext = ciphertext
	return ciphertext, nil
}

// Scan implements sql.Scanner for reading from database
func (ef *EncryptedField) Scan(value interface{}) error {
	if value == nil {
		ef.ciphertext = ""
		ef.plaintext = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		ef.ciphertext = string(v)
	case string:
		ef.ciphertext = v
	default:
		return fmt.Errorf("cannot scan %T into EncryptedField", value)
	}

	ef.plaintext = "" // Clear plaintext to force decryption on next read
	return nil
}

// Value implements driver.Valuer for writing to database
func (ef *EncryptedField) Value() (driver.Value, error) {
	ciphertext, err := ef.GetCiphertext()
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptedFieldRegistry manages which fields should be encrypted
type EncryptedFieldRegistry struct {
	fields map[string]*FieldEncryptionConfig
}

// FieldEncryptionConfig describes how a field should be encrypted
type FieldEncryptionConfig struct {
	TableName  string
	ColumnName string
	Encrypted  bool
	Algorithm  string
}

// NewEncryptedFieldRegistry creates a new registry
func NewEncryptedFieldRegistry() *EncryptedFieldRegistry {
	return &EncryptedFieldRegistry{
		fields: make(map[string]*FieldEncryptionConfig),
	}
}

// Register registers a field for encryption
func (efr *EncryptedFieldRegistry) Register(tableName, columnName string) {
	key := fmt.Sprintf("%s.%s", tableName, columnName)
	efr.fields[key] = &FieldEncryptionConfig{
		TableName:  tableName,
		ColumnName: columnName,
		Encrypted:  true,
		Algorithm:  "aes-256-gcm",
	}
}

// IsEncrypted checks if a field should be encrypted
func (efr *EncryptedFieldRegistry) IsEncrypted(tableName, columnName string) bool {
	key := fmt.Sprintf("%s.%s", tableName, columnName)
	config, exists := efr.fields[key]
	return exists && config.Encrypted
}

// GetConfig gets the encryption config for a field
func (efr *EncryptedFieldRegistry) GetConfig(tableName, columnName string) *FieldEncryptionConfig {
	key := fmt.Sprintf("%s.%s", tableName, columnName)
	return efr.fields[key]
}

// EncryptionHooks provides hooks for transparent encryption/decryption
type EncryptionHooks struct {
	columnEncryption *crypto.ColumnEncryption
	registry         *EncryptedFieldRegistry
}

// NewEncryptionHooks creates new encryption hooks
func NewEncryptionHooks(columnEncryption *crypto.ColumnEncryption, registry *EncryptedFieldRegistry) *EncryptionHooks {
	return &EncryptionHooks{
		columnEncryption: columnEncryption,
		registry:         registry,
	}
}

// BeforeScan applies decryption before scanning from database
func (eh *EncryptionHooks) BeforeScan(tableName, columnName string, value interface{}) (interface{}, error) {
	if !eh.registry.IsEncrypted(tableName, columnName) {
		return value, nil
	}

	// For encrypted fields, wrap in EncryptedField
	if str, ok := value.(string); ok {
		ef := NewEncryptedField(eh.columnEncryption, str)
		return ef, nil
	}

	return value, nil
}

// BeforeWrite applies encryption before writing to database
func (eh *EncryptionHooks) BeforeWrite(tableName, columnName string, value interface{}) (interface{}, error) {
	if !eh.registry.IsEncrypted(tableName, columnName) {
		return value, nil
	}

	// Convert to string if needed
	var plaintext string
	switch v := value.(type) {
	case string:
		plaintext = v
	case *string:
		if v != nil {
			plaintext = *v
		}
	case []byte:
		plaintext = string(v)
	case json.Marshaler:
		data, _ := json.Marshal(v)
		plaintext = string(data)
	default:
		return value, nil
	}

	if plaintext == "" {
		return nil, nil
	}

	// Encrypt the plaintext
	ciphertext, _, err := eh.columnEncryption.EncryptString(nil, plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt field %s.%s: %w", tableName, columnName, err)
	}

	return ciphertext, nil
}

// EncryptionConfig represents configuration for transparent encryption
type EncryptionConfig struct {
	Enabled   bool
	Algorithm string
	Fields    map[string]bool // table.column -> should encrypt
}

// InitializeEncryption initializes the encryption system with sensitive field registry
func InitializeEncryption(registry *EncryptedFieldRegistry) {
	// Register standard sensitive fields
	registry.Register("users", "email")
	registry.Register("users", "password_hash")
	registry.Register("registration_secrets", "secret_value")
	registry.Register("postgresql_instances", "connection_string")
	registry.Register("api_tokens", "token_hash")
	registry.Register("oauth_providers", "client_id_encrypted")
	registry.Register("oauth_providers", "client_secret_encrypted")
	registry.Register("ldap_config", "server_url_encrypted")
	registry.Register("ldap_config", "bind_dn_encrypted")
	registry.Register("ldap_config", "bind_password_encrypted")
	registry.Register("saml_config", "cert_encrypted")
	registry.Register("saml_config", "key_encrypted")
}

// DataMigrationHelper helps migrate plaintext data to encrypted format
type DataMigrationHelper struct {
	db               *sql.DB
	columnEncryption *crypto.ColumnEncryption
	registry         *EncryptedFieldRegistry
}

// NewDataMigrationHelper creates a new migration helper
func NewDataMigrationHelper(db *sql.DB, columnEncryption *crypto.ColumnEncryption, registry *EncryptedFieldRegistry) *DataMigrationHelper {
	return &DataMigrationHelper{
		db:               db,
		columnEncryption: columnEncryption,
		registry:         registry,
	}
}

// MigrateField migrates plaintext column to encrypted format
func (dmh *DataMigrationHelper) MigrateField(tableName, plainColumnName, encryptedColumnName string) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET %s = (SELECT encrypted FROM (
			SELECT %s AS plaintext, (
				SELECT %s
			) AS encrypted
		) AS subquery)
		WHERE %s IS NOT NULL AND %s IS NULL
	`, tableName, encryptedColumnName, plainColumnName, plainColumnName, plainColumnName, encryptedColumnName)

	_, err := dmh.db.Exec(query)
	return err
}
