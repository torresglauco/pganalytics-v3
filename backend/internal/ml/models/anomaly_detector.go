package models

import (
	"time"
)

type AnomalyDetector struct {
	mlServiceURL string
	baseline     map[string]*MetricBaseline
}

type MetricBaseline struct {
	Mean   float64
	StdDev float64
	Min    float64
	Max    float64
}

type AnomalyAlert struct {
	MetricName   string    `json:"metric_name"`
	CurrentValue float64   `json:"current_value"`
	Baseline     float64   `json:"baseline"`
	ZScore       float64   `json:"z_score"`
	Timestamp    time.Time `json:"timestamp"`
	Severity     string    `json:"severity"` // low, medium, high
}

func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		mlServiceURL: "http://localhost:5000",
		baseline:     make(map[string]*MetricBaseline),
	}
}

func (ad *AnomalyDetector) SetBaseline(metricName string, baseline *MetricBaseline) {
	ad.baseline[metricName] = baseline
}

func (ad *AnomalyDetector) Detect(metricName string, value float64) (*AnomalyAlert, bool) {
	baseline, exists := ad.baseline[metricName]
	if !exists {
		return nil, false
	}

	// Calculate z-score
	zScore := (value - baseline.Mean) / baseline.StdDev

	// Threshold: > 2.5 std deviations = anomaly
	if zScore > 2.5 || zScore < -2.5 {
		severity := "medium"
		if zScore > 3.5 || zScore < -3.5 {
			severity = "high"
		}

		return &AnomalyAlert{
			MetricName:   metricName,
			CurrentValue: value,
			Baseline:     baseline.Mean,
			ZScore:       zScore,
			Timestamp:    time.Now(),
			Severity:     severity,
		}, true
	}

	return nil, false
}
