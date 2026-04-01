package models

import (
	"crypto/md5"
	"fmt"
)

type QueryFeatures struct {
	JoinCount      int     `json:"join_count"`
	ScanType       string  `json:"scan_type"`       // seq_scan, index_scan, bitmap_scan
	RowCount       int     `json:"row_count"`
	FilterCount    int     `json:"filter_count"`
	SubqueryCount  int     `json:"subquery_count"`
	AggregateType  string  `json:"aggregate_type"`  // none, sum, count, group_by
	ExecutionTimeMs float64 `json:"execution_time_ms"`
}

func (qf *QueryFeatures) Fingerprint() string {
	// Create normalized fingerprint for aggregation
	data := fmt.Sprintf("%d|%s|%d|%d|%d|%s",
		qf.JoinCount,
		qf.ScanType,
		qf.RowCount/1000,  // Bucketing
		qf.FilterCount,
		qf.SubqueryCount,
		qf.AggregateType,
	)

	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:16]
}

func (qf *QueryFeatures) Vector() []float64 {
	scanTypeVal := 0.0
	switch qf.ScanType {
	case "seq_scan":
		scanTypeVal = 0.0
	case "index_scan":
		scanTypeVal = 1.0
	case "bitmap_scan":
		scanTypeVal = 2.0
	}

	aggregateVal := 0.0
	switch qf.AggregateType {
	case "none":
		aggregateVal = 0.0
	case "sum", "count":
		aggregateVal = 1.0
	case "group_by":
		aggregateVal = 2.0
	}

	return []float64{
		float64(qf.JoinCount),
		scanTypeVal,
		float64(qf.RowCount) / 1000,  // Normalize
		float64(qf.FilterCount),
		float64(qf.SubqueryCount),
		aggregateVal,
	}
}
