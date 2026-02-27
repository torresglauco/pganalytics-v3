package storage

import (
	"context"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// UserStoreImpl implements user data access using PostgresDB methods
type UserStoreImpl struct {
	db *PostgresDB
}

// NewUserStore creates a new user store
func NewUserStore(db *PostgresDB) *UserStoreImpl {
	return &UserStoreImpl{db: db}
}

// GetUserByUsername retrieves a user by username
func (us *UserStoreImpl) GetUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return us.db.GetUserByUsername(ctx, username)
}

// GetUserByID retrieves a user by ID
func (us *UserStoreImpl) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return us.db.GetUserByID(ctx, id)
}

// UpdateUserLastLogin updates user's last login time
func (us *UserStoreImpl) UpdateUserLastLogin(userID int, timestamp time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return us.db.UpdateUserLastLogin(ctx, userID)
}
