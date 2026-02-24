package metrics

import (
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
)

// CacheMetricsSnapshot represents a snapshot of cache performance metrics
type CacheMetricsSnapshot struct {
	FeatureCacheMetrics    CacheDetailedMetrics `json:"feature_cache"`
	PredictionCacheMetrics CacheDetailedMetrics `json:"prediction_cache"`
	FingerprintCacheMetrics CacheDetailedMetrics `json:"fingerprint_cache"`
	ExplainPlanCacheMetrics CacheDetailedMetrics `json:"explain_plan_cache"`
	AnomalyCacheMetrics    CacheDetailedMetrics `json:"anomaly_cache"`
	TotalCacheHits         int64                `json:"total_cache_hits"`
	TotalCacheMisses       int64                `json:"total_cache_misses"`
	OverallHitRate         float64              `json:"overall_hit_rate"`
}

// CacheDetailedMetrics represents detailed metrics for a single cache
type CacheDetailedMetrics struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	HitRate   float64 `json:"hit_rate"`
	Size      int     `json:"size"`
}

// CalculateMetricsSnapshot calculates metrics from a cache manager
func CalculateMetricsSnapshot(manager *cache.Manager) *CacheMetricsSnapshot {
	if manager == nil {
		return &CacheMetricsSnapshot{}
	}

	managerMetrics := manager.GetMetrics()

	// Calculate detailed metrics for each cache
	featureMetrics := calculateDetailedMetrics(managerMetrics.FeatureCacheMetrics)
	predictionMetrics := calculateDetailedMetrics(managerMetrics.PredictionCacheMetrics)
	fingerprintMetrics := calculateDetailedMetrics(managerMetrics.FingerprintCacheMetrics)
	explainPlanMetrics := calculateDetailedMetrics(managerMetrics.ExplainPlanCacheMetrics)
	anomalyMetrics := calculateDetailedMetrics(managerMetrics.AnomalyCacheMetrics)

	// Calculate overall metrics
	totalHits := featureMetrics.Hits + predictionMetrics.Hits + fingerprintMetrics.Hits + explainPlanMetrics.Hits + anomalyMetrics.Hits
	totalMisses := featureMetrics.Misses + predictionMetrics.Misses + fingerprintMetrics.Misses + explainPlanMetrics.Misses + anomalyMetrics.Misses
	overallHitRate := 0.0
	if totalHits+totalMisses > 0 {
		overallHitRate = float64(totalHits) / float64(totalHits+totalMisses)
	}

	return &CacheMetricsSnapshot{
		FeatureCacheMetrics:    featureMetrics,
		PredictionCacheMetrics: predictionMetrics,
		FingerprintCacheMetrics: fingerprintMetrics,
		ExplainPlanCacheMetrics: explainPlanMetrics,
		AnomalyCacheMetrics:    anomalyMetrics,
		TotalCacheHits:         totalHits,
		TotalCacheMisses:       totalMisses,
		OverallHitRate:         overallHitRate,
	}
}

// calculateDetailedMetrics calculates metrics for a single cache
func calculateDetailedMetrics(m cache.CacheMetrics) CacheDetailedMetrics {
	hitRate := 0.0
	if m.Hits+m.Misses > 0 {
		hitRate = float64(m.Hits) / float64(m.Hits+m.Misses)
	}

	return CacheDetailedMetrics{
		Hits:      m.Hits,
		Misses:    m.Misses,
		Evictions: m.Evictions,
		HitRate:   hitRate,
	}
}

// CacheStatusResponse is the response for cache status endpoint
type CacheStatusResponse struct {
	Enabled    bool                  `json:"enabled"`
	MaxSize    int                   `json:"max_size"`
	Metrics    *CacheMetricsSnapshot `json:"metrics"`
	Message    string                `json:"message"`
}
