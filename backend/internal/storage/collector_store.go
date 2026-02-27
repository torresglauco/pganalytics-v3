package storage

import (
	"context"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/google/uuid"
)

// CollectorStoreImpl implements collector data access using PostgresDB methods
type CollectorStoreImpl struct {
	db *PostgresDB
}

// NewCollectorStore creates a new collector store
func NewCollectorStore(db *PostgresDB) *CollectorStoreImpl {
	return &CollectorStoreImpl{db: db}
}

// CreateCollector creates a new collector
func (cs *CollectorStoreImpl) CreateCollector(collector *models.Collector) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cs.db.CreateCollector(ctx, collector)
	if err != nil {
		return uuid.Nil, err
	}

	return collector.ID, nil
}

// GetCollectorByID retrieves a collector by ID
func (cs *CollectorStoreImpl) GetCollectorByID(id uuid.UUID) (*models.Collector, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cs.db.GetCollectorByID(ctx, id.String())
}

// UpdateCollectorStatus updates collector status
func (cs *CollectorStoreImpl) UpdateCollectorStatus(id uuid.UUID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cs.db.UpdateCollectorStatus(ctx, id.String(), status)
}

// UpdateCollectorCertificate updates collector certificate info
func (cs *CollectorStoreImpl) UpdateCollectorCertificate(id uuid.UUID, thumbprint string, expiresAt time.Time) error {
	// Not implemented in base PostgresDB
	// This is a no-op for now
	return nil
}

// DeleteCollector deletes a collector by ID
func (cs *CollectorStoreImpl) DeleteCollector(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return cs.db.DeleteCollector(ctx, id.String())
}
