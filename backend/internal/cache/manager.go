package cache

import (
	"time"

	"go.uber.org/zap"
)

// Manager coordinates all caches for the application
type Manager struct {
	featureCache    *Cache[string, interface{}]
	predictionCache *Cache[string, interface{}]
	fingerprintCache *Cache[string, interface{}]
	explainPlanCache *Cache[string, interface{}]
	anomalyCache    *Cache[string, interface{}]
	logger          *zap.Logger
}

// NewManager creates a new cache manager with configured TTLs
func NewManager(
	maxSize int,
	featureCacheTTL time.Duration,
	predictionCacheTTL time.Duration,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		featureCache:    NewCache[string, interface{}](featureCacheTTL, maxSize),
		predictionCache: NewCache[string, interface{}](predictionCacheTTL, maxSize),
		fingerprintCache: NewCache[string, interface{}](10 * time.Minute, maxSize),
		explainPlanCache: NewCache[string, interface{}](30 * time.Minute, maxSize),
		anomalyCache:    NewCache[string, interface{}](5 * time.Minute, maxSize),
		logger:          logger,
	}
}

// GetFeatures retrieves cached query features
func (m *Manager) GetFeatures(key string) (interface{}, bool) {
	return m.featureCache.Get(key)
}

// SetFeatures caches query features
func (m *Manager) SetFeatures(key string, features interface{}) {
	m.featureCache.Set(key, features)
}

// ClearFeatures removes feature cache entry
func (m *Manager) ClearFeatures(key string) {
	m.featureCache.Delete(key)
}

// GetPrediction retrieves cached prediction
func (m *Manager) GetPrediction(key string) (interface{}, bool) {
	return m.predictionCache.Get(key)
}

// SetPrediction caches a prediction
func (m *Manager) SetPrediction(key string, prediction interface{}) {
	m.predictionCache.Set(key, prediction)
}

// ClearPrediction removes prediction cache entry
func (m *Manager) ClearPrediction(key string) {
	m.predictionCache.Delete(key)
}

// GetFingerprint retrieves cached query fingerprints
func (m *Manager) GetFingerprint(key string) (interface{}, bool) {
	return m.fingerprintCache.Get(key)
}

// SetFingerprint caches query fingerprints
func (m *Manager) SetFingerprint(key string, fingerprint interface{}) {
	m.fingerprintCache.Set(key, fingerprint)
}

// ClearFingerprint removes fingerprint cache entry
func (m *Manager) ClearFingerprint(key string) {
	m.fingerprintCache.Delete(key)
}

// GetExplainPlan retrieves cached EXPLAIN plan
func (m *Manager) GetExplainPlan(key string) (interface{}, bool) {
	return m.explainPlanCache.Get(key)
}

// SetExplainPlan caches an EXPLAIN plan
func (m *Manager) SetExplainPlan(key string, plan interface{}) {
	m.explainPlanCache.Set(key, plan)
}

// ClearExplainPlan removes EXPLAIN plan cache entry
func (m *Manager) ClearExplainPlan(key string) {
	m.explainPlanCache.Delete(key)
}

// GetAnomalies retrieves cached anomalies
func (m *Manager) GetAnomalies(key string) (interface{}, bool) {
	return m.anomalyCache.Get(key)
}

// SetAnomalies caches anomalies
func (m *Manager) SetAnomalies(key string, anomalies interface{}) {
	m.anomalyCache.Set(key, anomalies)
}

// ClearAnomalies removes anomaly cache entry
func (m *Manager) ClearAnomalies(key string) {
	m.anomalyCache.Delete(key)
}

// GetMetrics returns combined cache metrics
type ManagerMetrics struct {
	FeatureCacheMetrics    CacheMetrics
	PredictionCacheMetrics CacheMetrics
	FingerprintCacheMetrics CacheMetrics
	ExplainPlanCacheMetrics CacheMetrics
	AnomalyCacheMetrics    CacheMetrics
}

// GetMetrics returns metrics for all caches
func (m *Manager) GetMetrics() ManagerMetrics {
	return ManagerMetrics{
		FeatureCacheMetrics:    m.featureCache.GetMetrics(),
		PredictionCacheMetrics: m.predictionCache.GetMetrics(),
		FingerprintCacheMetrics: m.fingerprintCache.GetMetrics(),
		ExplainPlanCacheMetrics: m.explainPlanCache.GetMetrics(),
		AnomalyCacheMetrics:    m.anomalyCache.GetMetrics(),
	}
}

// Clear clears all caches
func (m *Manager) Clear() {
	m.featureCache.Clear()
	m.predictionCache.Clear()
	m.fingerprintCache.Clear()
	m.explainPlanCache.Clear()
	m.anomalyCache.Clear()
	m.logger.Info("All caches cleared")
}

// Close closes all caches
func (m *Manager) Close() error {
	_ = m.featureCache.Close()
	_ = m.predictionCache.Close()
	_ = m.fingerprintCache.Close()
	_ = m.explainPlanCache.Close()
	_ = m.anomalyCache.Close()
	m.logger.Info("Cache manager closed")
	return nil
}
