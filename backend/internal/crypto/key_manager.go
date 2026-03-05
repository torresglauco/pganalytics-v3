package crypto

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// KeyBackend defines different key storage backends
type KeyBackend string

const (
	KeyBackendLocal KeyBackend = "local"
	KeyBackendAWS   KeyBackend = "aws"
	KeyBackendVault KeyBackend = "vault"
	KeyBackendGCP   KeyBackend = "gcp"
)

// KeyVersion represents a versioned encryption key
type KeyVersion struct {
	Version     int
	CreatedAt   time.Time
	RetiredAt   *time.Time
	KeyMaterial []byte // Should be stored encrypted in database
	Algorithm   string
	Active      bool
}

// KeyManager handles key generation, rotation, and versioning
type KeyManager interface {
	GetCurrentKey(ctx context.Context) ([]byte, int, error)
	GetKeyByVersion(ctx context.Context, version int) ([]byte, error)
	RotateKey(ctx context.Context) (int, error)
	RetireKey(ctx context.Context, version int) error
	GetAllVersions(ctx context.Context) ([]KeyVersion, error)
}

// LocalKeyManager stores keys locally (for development)
type LocalKeyManager struct {
	keys map[int]KeyVersion
	mu   sync.RWMutex
}

// NewLocalKeyManager creates a new local key manager
func NewLocalKeyManager() *LocalKeyManager {
	lkm := &LocalKeyManager{
		keys: make(map[int]KeyVersion),
	}

	// Initialize with v1
	initialKey := make([]byte, 32)
	rand.Read(initialKey)
	lkm.keys[1] = KeyVersion{
		Version:     1,
		CreatedAt:   time.Now(),
		KeyMaterial: initialKey,
		Algorithm:   "aes-256-gcm",
		Active:      true,
	}

	return lkm
}

// GetCurrentKey returns the current active key
func (lkm *LocalKeyManager) GetCurrentKey(ctx context.Context) ([]byte, int, error) {
	lkm.mu.RLock()
	defer lkm.mu.RUnlock()

	// Find active key with highest version
	var activeKey KeyVersion
	var maxVersion int

	for _, kv := range lkm.keys {
		if kv.Active && kv.Version > maxVersion {
			activeKey = kv
			maxVersion = kv.Version
		}
	}

	if maxVersion == 0 {
		return nil, 0, fmt.Errorf("no active key found")
	}

	return activeKey.KeyMaterial, activeKey.Version, nil
}

// GetKeyByVersion returns a key by version
func (lkm *LocalKeyManager) GetKeyByVersion(ctx context.Context, version int) ([]byte, error) {
	lkm.mu.RLock()
	defer lkm.mu.RUnlock()

	kv, ok := lkm.keys[version]
	if !ok {
		return nil, fmt.Errorf("key version %d not found", version)
	}

	return kv.KeyMaterial, nil
}

// RotateKey rotates to a new key version
func (lkm *LocalKeyManager) RotateKey(ctx context.Context) (int, error) {
	lkm.mu.Lock()
	defer lkm.mu.Unlock()

	// Find max version
	maxVersion := 0
	for v := range lkm.keys {
		if v > maxVersion {
			maxVersion = v
		}
	}

	newVersion := maxVersion + 1

	// Generate new key
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return 0, fmt.Errorf("failed to generate key: %w", err)
	}

	// Mark old active key as retired
	for i := range lkm.keys {
		if lkm.keys[i].Active {
			kv := lkm.keys[i]
			now := time.Now()
			kv.RetiredAt = &now
			kv.Active = false
			lkm.keys[i] = kv
		}
	}

	// Add new key
	lkm.keys[newVersion] = KeyVersion{
		Version:     newVersion,
		CreatedAt:   time.Now(),
		KeyMaterial: newKey,
		Algorithm:   "aes-256-gcm",
		Active:      true,
	}

	return newVersion, nil
}

// RetireKey marks a key as retired
func (lkm *LocalKeyManager) RetireKey(ctx context.Context, version int) error {
	lkm.mu.Lock()
	defer lkm.mu.Unlock()

	kv, ok := lkm.keys[version]
	if !ok {
		return fmt.Errorf("key version %d not found", version)
	}

	now := time.Now()
	kv.RetiredAt = &now
	kv.Active = false
	lkm.keys[version] = kv

	return nil
}

// GetAllVersions returns all key versions
func (lkm *LocalKeyManager) GetAllVersions(ctx context.Context) ([]KeyVersion, error) {
	lkm.mu.RLock()
	defer lkm.mu.RUnlock()

	versions := make([]KeyVersion, 0, len(lkm.keys))
	for _, kv := range lkm.keys {
		versions = append(versions, kv)
	}

	return versions, nil
}

// TODO: Implement AWS Secrets Manager backend
// type AWSKeyManager struct {
// 	client *secretsmanager.Client
// 	arn    string
// }

// TODO: Implement HashiCorp Vault backend
// type VaultKeyManager struct {
// 	client *vault.Client
// 	path   string
// }

// TODO: Implement GCP KMS backend
// type GCPKeyManager struct {
// 	client *kms.KeyManagementServiceClient
// 	name   string
// }

// KeyRotationScheduler schedules periodic key rotation
type KeyRotationScheduler struct {
	km       KeyManager
	interval time.Duration
	stopCh   chan struct{}
}

// NewKeyRotationScheduler creates a new key rotation scheduler
func NewKeyRotationScheduler(km KeyManager, interval time.Duration) *KeyRotationScheduler {
	return &KeyRotationScheduler{
		km:       km,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the rotation scheduler
func (krs *KeyRotationScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(krs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-krs.stopCh:
			return
		case <-ticker.C:
			newVersion, err := krs.km.RotateKey(ctx)
			if err != nil {
				// Log error and continue
				fmt.Printf("failed to rotate key: %v\n", err)
			} else {
				fmt.Printf("successfully rotated key to version %d\n", newVersion)
			}
		}
	}
}

// Stop stops the rotation scheduler
func (krs *KeyRotationScheduler) Stop() {
	close(krs.stopCh)
}
