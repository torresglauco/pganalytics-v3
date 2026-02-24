package ml

import (
	"context"
	"fmt"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"go.uber.org/zap"
)

// CachedFeatureExtractor wraps FeatureExtractor with caching
type CachedFeatureExtractor struct {
	extractor *FeatureExtractor
	cache     *cache.Cache[string, *QueryFeatures]
	logger    *zap.Logger
}

// NewCachedFeatureExtractor creates a new cached feature extractor
func NewCachedFeatureExtractor(
	extractor *FeatureExtractor,
	cacheTTL time.Duration,
	cacheMaxSize int,
	logger *zap.Logger,
) *CachedFeatureExtractor {
	return &CachedFeatureExtractor{
		extractor: extractor,
		cache:     cache.NewCache[string, *QueryFeatures](cacheTTL, cacheMaxSize),
		logger:    logger,
	}
}

// ExtractQueryFeatures extracts features with caching
func (c *CachedFeatureExtractor) ExtractQueryFeatures(
	ctx context.Context,
	queryHash int64,
) (*QueryFeatures, error) {
	// Generate cache key
	cacheKey := fmt.Sprintf("features:%d", queryHash)

	// Check cache
	if cached, found := c.cache.Get(cacheKey); found {
		c.logger.Debug("Feature cache hit",
			zap.Int64("query_hash", queryHash),
		)
		return cached, nil
	}

	// Extract features
	c.logger.Debug("Feature cache miss, extracting",
		zap.Int64("query_hash", queryHash),
	)

	features, err := c.extractor.ExtractQueryFeatures(ctx, queryHash)
	if err != nil {
		return nil, err
	}

	// Cache the result
	c.cache.Set(cacheKey, features)

	return features, nil
}

// ExtractBatchQueryFeatures extracts features for multiple queries with caching
func (c *CachedFeatureExtractor) ExtractBatchQueryFeatures(
	ctx context.Context,
	queryHashes []int64,
) (map[int64]*QueryFeatures, error) {
	results := make(map[int64]*QueryFeatures)
	var uncachedHashes []int64

	// Check cache for each query
	for _, hash := range queryHashes {
		cacheKey := fmt.Sprintf("features:%d", hash)
		if cached, found := c.cache.Get(cacheKey); found {
			results[hash] = cached
		} else {
			uncachedHashes = append(uncachedHashes, hash)
		}
	}

	c.logger.Debug("Batch feature extraction",
		zap.Int("total_requested", len(queryHashes)),
		zap.Int("cache_hits", len(queryHashes)-len(uncachedHashes)),
		zap.Int("cache_misses", len(uncachedHashes)),
	)

	// Extract uncached features
	if len(uncachedHashes) > 0 {
		for _, hash := range uncachedHashes {
			features, err := c.extractor.ExtractQueryFeatures(ctx, hash)
			if err != nil {
				c.logger.Warn("Failed to extract features",
					zap.Int64("query_hash", hash),
					zap.Error(err),
				)
				continue
			}

			// Cache the result
			cacheKey := fmt.Sprintf("features:%d", hash)
			c.cache.Set(cacheKey, features)
			results[hash] = features
		}
	}

	return results, nil
}

// ClearFeatureCache removes a feature from cache
func (c *CachedFeatureExtractor) ClearFeatureCache(queryHash int64) {
	cacheKey := fmt.Sprintf("features:%d", queryHash)
	c.cache.Delete(cacheKey)
	c.logger.Debug("Feature cache cleared",
		zap.Int64("query_hash", queryHash),
	)
}

// GetCacheMetrics returns feature cache metrics
func (c *CachedFeatureExtractor) GetCacheMetrics() cache.CacheMetrics {
	return c.cache.GetMetrics()
}

// Close closes the cache
func (c *CachedFeatureExtractor) Close() error {
	return c.cache.Close()
}
