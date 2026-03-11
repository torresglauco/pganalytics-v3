package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ConfigCache caches collector and query configurations with versioning
type ConfigCache struct {
	mu               sync.RWMutex
	cache            map[string]*CachedConfig
	versionMap       map[string]int // Map of config key to version
	ttl              time.Duration
	defaultTTL       time.Duration
	maxSize          int
	hitCount         int64
	missCount        int64
	evictionCount    int64
	logger           *zap.Logger
}

// CachedConfig represents a cached configuration entry
type CachedConfig struct {
	Key          string          `json:"key"`
	Version      int             `json:"version"`
	Data         json.RawMessage `json:"data"`
	Hash         string          `json:"hash"`
	ExpiresAt    time.Time       `json:"expires_at"`
	CreatedAt    time.Time       `json:"created_at"`
	LastAccessed time.Time       `json:"last_accessed"`
	AccessCount  int64           `json:"access_count"`
}

// ConfigChangeEvent represents a configuration change
type ConfigChangeEvent struct {
	Key     string    `json:"key"`
	Version int       `json:"version"`
	OldHash string    `json:"old_hash"`
	NewHash string    `json:"new_hash"`
	Time    time.Time `json:"time"`
}

// NewConfigCache creates a new configuration cache
func NewConfigCache(ttl time.Duration, maxSize int, logger *zap.Logger) *ConfigCache {
	cc := &ConfigCache{
		cache:      make(map[string]*CachedConfig),
		versionMap: make(map[string]int),
		ttl:        ttl,
		defaultTTL: ttl,
		maxSize:    maxSize,
		logger:     logger,
	}

	// Start cleanup goroutine
	go cc.cleanupExpired()

	return cc
}

// Set stores or updates a configuration in the cache
func (cc *ConfigCache) Set(key string, data json.RawMessage) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot cache empty data for key: %s", key)
	}

	hash := cc.hashData(data)

	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Check if we're at capacity and need to evict
	if len(cc.cache) >= cc.maxSize {
		if oldKey := cc.evictLRU(); oldKey != "" {
			cc.evictionCount++
		}
	}

	oldVersion := cc.versionMap[key]
	newVersion := oldVersion + 1
	cc.versionMap[key] = newVersion

	now := time.Now()
	cc.cache[key] = &CachedConfig{
		Key:          key,
		Version:      newVersion,
		Data:         data,
		Hash:         hash,
		ExpiresAt:    now.Add(cc.ttl),
		CreatedAt:    now,
		LastAccessed: now,
		AccessCount:  0,
	}

	cc.logger.Debug(
		"Config cached",
		zap.String("key", key),
		zap.Int("version", newVersion),
		zap.String("hash", hash),
	)

	return nil
}

// Get retrieves a configuration from the cache
// Returns (data, version, found, error)
func (cc *ConfigCache) Get(key string) (json.RawMessage, int, bool) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	config, exists := cc.cache[key]
	if !exists {
		cc.missCount++
		return nil, 0, false
	}

	// Check if expired
	if time.Now().After(config.ExpiresAt) {
		delete(cc.cache, key)
		cc.missCount++
		return nil, 0, false
	}

	// Update access stats
	config.LastAccessed = time.Now()
	config.AccessCount++
	cc.hitCount++

	return config.Data, config.Version, true
}

// GetVersion returns only the version for a key
func (cc *ConfigCache) GetVersion(key string) int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.versionMap[key]
}

// InvalidateKey removes a key from the cache
func (cc *ConfigCache) InvalidateKey(key string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if _, exists := cc.cache[key]; exists {
		delete(cc.cache, key)
		cc.logger.Debug("Config invalidated", zap.String("key", key))
	}
}

// InvalidatePattern removes all keys matching a pattern
func (cc *ConfigCache) InvalidatePattern(pattern string) int {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	count := 0
	for key := range cc.cache {
		// Simple prefix matching
		if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
			delete(cc.cache, key)
			count++
		}
	}

	cc.logger.Debug(
		"Config pattern invalidated",
		zap.String("pattern", pattern),
		zap.Int("count", count),
	)

	return count
}

// Clear empties the entire cache
func (cc *ConfigCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	oldSize := len(cc.cache)
	cc.cache = make(map[string]*CachedConfig)
	cc.versionMap = make(map[string]int)

	cc.logger.Debug("Config cache cleared", zap.Int("entries_removed", oldSize))
}

// GetStats returns cache statistics
func (cc *ConfigCache) GetStats() map[string]interface{} {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	total := cc.hitCount + cc.missCount
	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(cc.hitCount) / float64(total) * 100
	}

	return map[string]interface{}{
		"size":              len(cc.cache),
		"max_size":          cc.maxSize,
		"version_map_size":  len(cc.versionMap),
		"hits":              cc.hitCount,
		"misses":            cc.missCount,
		"total_requests":    total,
		"hit_rate_percent":  fmt.Sprintf("%.2f", hitRate),
		"evictions":         cc.evictionCount,
		"ttl_seconds":       int(cc.ttl.Seconds()),
	}
}

// SetTTL updates the TTL for the cache
func (cc *ConfigCache) SetTTL(ttl time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.ttl = ttl
	cc.logger.Debug("Config cache TTL updated", zap.Duration("ttl", ttl))
}

// evictLRU evicts the least recently used entry (must be called with lock held)
func (cc *ConfigCache) evictLRU() string {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, config := range cc.cache {
		if config.LastAccessed.Before(oldestTime) {
			oldestTime = config.LastAccessed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(cc.cache, oldestKey)
		cc.logger.Debug("Cache entry evicted (LRU)", zap.String("key", oldestKey))
	}

	return oldestKey
}

// cleanupExpired periodically removes expired entries
func (cc *ConfigCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cc.mu.Lock()

		now := time.Now()
		expiredCount := 0

		for key, config := range cc.cache {
			if now.After(config.ExpiresAt) {
				delete(cc.cache, key)
				expiredCount++
			}
		}

		cc.mu.Unlock()

		if expiredCount > 0 {
			cc.logger.Debug(
				"Expired cache entries removed",
				zap.Int("count", expiredCount),
			)
		}
	}
}

// hashData computes a SHA256 hash of the data
func (cc *ConfigCache) hashData(data json.RawMessage) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// CollectorConfigKey generates a cache key for a collector's configuration
func CollectorConfigKey(collectorID string) string {
	return fmt.Sprintf("collector:config:%s", collectorID)
}

// QueryConfigKey generates a cache key for a query configuration
func QueryConfigKey(queryID string) string {
	return fmt.Sprintf("query:config:%s", queryID)
}

// DatabaseConfigKey generates a cache key for a database configuration
func DatabaseConfigKey(databaseID string) string {
	return fmt.Sprintf("database:config:%s", databaseID)
}

// InvalidateCollectorConfig invalidates cache for a specific collector
func (cc *ConfigCache) InvalidateCollectorConfig(collectorID string) {
	cc.InvalidateKey(CollectorConfigKey(collectorID))
	// Also invalidate related query configs
	cc.InvalidatePattern(fmt.Sprintf("query:config:%s:", collectorID))
}

// InvalidateDatabaseConfig invalidates cache for a specific database
func (cc *ConfigCache) InvalidateDatabaseConfig(databaseID string) {
	cc.InvalidateKey(DatabaseConfigKey(databaseID))
	// Also invalidate related query configs
	cc.InvalidatePattern(fmt.Sprintf("query:config:*:%s", databaseID))
}

// Close cleans up resources (no-op for now, but useful for future extensions)
func (cc *ConfigCache) Close() error {
	cc.Clear()
	return nil
}
