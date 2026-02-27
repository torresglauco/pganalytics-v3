package storage

import (
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// TokenStoreImpl implements token data access
type TokenStoreImpl struct {
	db *PostgresDB
}

// NewTokenStore creates a new token store
func NewTokenStore(db *PostgresDB) *TokenStoreImpl {
	return &TokenStoreImpl{db: db}
}

// CreateAPIToken creates a new API token (not implemented yet)
func (ts *TokenStoreImpl) CreateAPIToken(token *models.APIToken) (int, error) {
	// Not implemented - tokens are handled by JWT manager
	return 0, nil
}

// GetAPITokenByHash retrieves a token by hash (not implemented yet)
func (ts *TokenStoreImpl) GetAPITokenByHash(hash string) (*models.APIToken, error) {
	// Not implemented - tokens are handled by JWT manager
	return nil, nil
}

// UpdateAPITokenLastUsed updates when the token was last used (not implemented yet)
func (ts *TokenStoreImpl) UpdateAPITokenLastUsed(id int, timestamp time.Time) error {
	// Not implemented - tokens are handled by JWT manager
	return nil
}
