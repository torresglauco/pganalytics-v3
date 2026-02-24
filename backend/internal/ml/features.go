package ml

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// IFeatureExtractor defines the interface for feature extraction
type IFeatureExtractor interface {
	ExtractQueryFeatures(ctx context.Context, queryHash int64) (*QueryFeatures, error)
	ExtractBatchQueryFeatures(ctx context.Context, queryHashes []int64) (map[int64]*QueryFeatures, error)
}

// FeatureExtractor extracts features for ML predictions
type FeatureExtractor struct {
	postgres *storage.PostgresDB
	logger   *zap.Logger
}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor(postgres *storage.PostgresDB, logger *zap.Logger) *FeatureExtractor {
	return &FeatureExtractor{
		postgres: postgres,
		logger:   logger,
	}
}

// QueryFeatures represents extracted features for a query
type QueryFeatures struct {
	QueryHash               int64                  `json:"query_hash"`
	FingerprintHash         int64                  `json:"fingerprint_hash,omitempty"`
	MeanExecutionTimeMs     float64                `json:"mean_execution_time_ms"`
	StddevExecutionTimeMs   float64                `json:"stddev_execution_time_ms"`
	MinExecutionTimeMs      float64                `json:"min_execution_time_ms"`
	MaxExecutionTimeMs      float64                `json:"max_execution_time_ms"`
	CallsPerMinute          float64                `json:"calls_per_minute"`
	IndexCount              int                    `json:"index_count"`
	ScanType                string                 `json:"scan_type"`
	TableRowCount           int64                  `json:"table_row_count,omitempty"`
	MeanTableSizeMB         float64                `json:"mean_table_size_mb,omitempty"`
	SequentialScans         int                    `json:"sequential_scans,omitempty"`
	IndexScans              int                    `json:"index_scans,omitempty"`
	LastSeen                *string                `json:"last_seen,omitempty"`
	ExecutionComplexity     float64                `json:"execution_complexity"`     // Derived: stddev/mean
	VolumeImpact            float64                `json:"volume_impact"`            // Derived: mean * calls_per_minute
	OptimizationOpportunity float64                `json:"optimization_opportunity"` // Derived: score 0-1
	FeatureMap              map[string]interface{} `json:"-"`                        // For ML service
}

// ExtractQueryFeatures extracts ML features from query statistics
func (fe *FeatureExtractor) ExtractQueryFeatures(ctx context.Context, queryHash int64) (*QueryFeatures, error) {
	// Get latest query statistics
	queryStats, err := fe.postgres.GetQueryTimeline(ctx, queryHash, time.Now().Add(-24*time.Hour))
	if err != nil {
		fe.logger.Error("Failed to get query statistics", zap.Error(err), zap.Int64("query_hash", queryHash))
		return nil, fmt.Errorf("failed to get query statistics: %w", err)
	}

	if len(queryStats) == 0 {
		return nil, fmt.Errorf("query not found: %d", queryHash)
	}

	// Use the most recent stats
	latest := queryStats[len(queryStats)-1]

	// Calculate calls per minute
	callsPerMin := float64(latest.Calls) / (24.0 * 60.0) // Assuming 24-hour data

	lastSeenStr := latest.Time.Format(time.RFC3339)
	features := &QueryFeatures{
		QueryHash:             queryHash,
		MeanExecutionTimeMs:   latest.MeanTime,
		StddevExecutionTimeMs: latest.StddevTime,
		MinExecutionTimeMs:    latest.MinTime,
		MaxExecutionTimeMs:    latest.MaxTime,
		CallsPerMinute:        callsPerMin,
		IndexCount:            1,            // Default; would be enhanced with EXPLAIN data
		ScanType:              "Index Scan", // Default; would be from EXPLAIN data
		TableRowCount:         latest.Rows,
		MeanTableSizeMB:       0, // Would be from table statistics
		LastSeen:              &lastSeenStr,
	}

	// Calculate derived features
	features.ExecutionComplexity = calculateExecutionComplexity(
		features.MeanExecutionTimeMs,
		features.StddevExecutionTimeMs,
	)

	features.VolumeImpact = features.MeanExecutionTimeMs * features.CallsPerMinute

	features.OptimizationOpportunity = calculateOptimizationOpportunity(
		features.MeanExecutionTimeMs,
		features.CallsPerMinute,
		features.IndexCount,
		features.ScanType,
	)

	// Count scan types
	if features.ScanType == "Sequential Scan" {
		features.SequentialScans = 1
	} else if features.ScanType == "Index Scan" {
		features.IndexScans = 1
	}

	// Build feature map for ML service
	features.FeatureMap = fe.buildFeatureMap(features)

	return features, nil
}

// ExtractBatchFeatures extracts features for multiple queries
func (fe *FeatureExtractor) ExtractBatchFeatures(ctx context.Context, queryHashes []int64) (map[int64]*QueryFeatures, error) {
	result := make(map[int64]*QueryFeatures)

	for _, queryHash := range queryHashes {
		features, err := fe.ExtractQueryFeatures(ctx, queryHash)
		if err != nil {
			fe.logger.Warn("Failed to extract features for query",
				zap.Int64("query_hash", queryHash),
				zap.Error(err))
			continue
		}
		result[queryHash] = features
	}

	return result, nil
}

// ExtractBatchQueryFeatures is an alias for ExtractBatchFeatures to implement IFeatureExtractor
func (fe *FeatureExtractor) ExtractBatchQueryFeatures(ctx context.Context, queryHashes []int64) (map[int64]*QueryFeatures, error) {
	return fe.ExtractBatchFeatures(ctx, queryHashes)
}

// buildFeatureMap constructs a map for ML service consumption
func (fe *FeatureExtractor) buildFeatureMap(features *QueryFeatures) map[string]interface{} {
	return map[string]interface{}{
		"query_hash":               features.QueryHash,
		"mean_execution_time_ms":   features.MeanExecutionTimeMs,
		"stddev_execution_time_ms": features.StddevExecutionTimeMs,
		"min_execution_time_ms":    features.MinExecutionTimeMs,
		"max_execution_time_ms":    features.MaxExecutionTimeMs,
		"calls_per_minute":         features.CallsPerMinute,
		"index_count":              features.IndexCount,
		"scan_type":                features.ScanType,
		"table_row_count":          features.TableRowCount,
		"mean_table_size_mb":       features.MeanTableSizeMB,
		"sequential_scans":         features.SequentialScans,
		"index_scans":              features.IndexScans,
		"execution_complexity":     features.ExecutionComplexity,
		"volume_impact":            features.VolumeImpact,
		"optimization_opportunity": features.OptimizationOpportunity,
	}
}

// calculateExecutionComplexity calculates the variance in execution time
// High complexity indicates unpredictable performance
func calculateExecutionComplexity(meanMs, stddevMs float64) float64 {
	if meanMs == 0 {
		return 0
	}
	// Coefficient of variation: stddev / mean
	cv := stddevMs / meanMs
	// Cap at 1.0 for reasonable score
	if cv > 1.0 {
		return 1.0
	}
	return cv
}

// calculateOptimizationOpportunity calculates a score 0-1 for optimization potential
func calculateOptimizationOpportunity(meanMs, callsPerMin float64, indexCount int, scanType string) float64 {
	// Base score from volume impact
	volumeScore := math.Min(meanMs*callsPerMin/10000, 1.0) // Normalize to 0-1

	// Sequential scans get higher priority (no index)
	scanScore := 0.5
	if scanType == "Sequential Scan" {
		scanScore = 1.0
	} else if scanType == "Index Scan" && indexCount < 2 {
		scanScore = 0.7
	}

	// Combine scores
	opportunity := (volumeScore * 0.6) + (scanScore * 0.4)

	// Cap at 1.0
	if opportunity > 1.0 {
		return 1.0
	}
	return opportunity
}

// NormalizeFeatures normalizes features for ML model input
func (fe *FeatureExtractor) NormalizeFeatures(features *QueryFeatures, stats *NormalizationStats) map[string]interface{} {
	normalized := make(map[string]interface{})

	// Normalize numeric features using z-score normalization
	normalized["mean_execution_time_ms"] = normalizeValue(
		features.MeanExecutionTimeMs,
		stats.MeanExecutionTimeMean,
		stats.MeanExecutionTimeStddev,
	)

	normalized["calls_per_minute"] = normalizeValue(
		features.CallsPerMinute,
		stats.CallsPerMinuteMean,
		stats.CallsPerMinuteStddev,
	)

	normalized["index_count"] = normalizeValue(
		float64(features.IndexCount),
		stats.IndexCountMean,
		stats.IndexCountStddev,
	)

	normalized["table_row_count"] = normalizeValue(
		float64(features.TableRowCount),
		stats.TableRowCountMean,
		stats.TableRowCountStddev,
	)

	// Categorical features (one-hot encoding)
	normalized["scan_type_sequential"] = 0.0
	normalized["scan_type_index"] = 0.0
	normalized["scan_type_bitmap"] = 0.0

	if features.ScanType == "Sequential Scan" {
		normalized["scan_type_sequential"] = 1.0
	} else if features.ScanType == "Index Scan" {
		normalized["scan_type_index"] = 1.0
	} else if features.ScanType == "Bitmap Scan" {
		normalized["scan_type_bitmap"] = 1.0
	}

	// Derived features (already normalized)
	normalized["execution_complexity"] = features.ExecutionComplexity
	normalized["volume_impact"] = math.Min(features.VolumeImpact/10000, 1.0)

	return normalized
}

// NormalizationStats holds statistics for feature normalization
type NormalizationStats struct {
	MeanExecutionTimeMean   float64
	MeanExecutionTimeStddev float64
	CallsPerMinuteMean      float64
	CallsPerMinuteStddev    float64
	IndexCountMean          float64
	IndexCountStddev        float64
	TableRowCountMean       float64
	TableRowCountStddev     float64
}

// normalizeValue performs z-score normalization
func normalizeValue(value, mean, stddev float64) float64 {
	if stddev == 0 {
		return 0
	}
	return (value - mean) / stddev
}
