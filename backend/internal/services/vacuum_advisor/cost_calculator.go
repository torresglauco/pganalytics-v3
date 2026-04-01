package vacuum_advisor

import (
	"math"
)

// CostCalculator provides cost analysis for VACUUM operations
type CostCalculator struct {
	// VACUUM operation costs (milliseconds per operation)
	SeqScanCostPerPage    float64 // 1.0 - cost to read a page sequentially
	RandomAccessCostPerTuple float64 // 4.0 - cost to access a tuple randomly

	// PostgreSQL configuration defaults
	RandomPageCost float64 // 4.0 - relative cost of random I/O vs sequential
	CpuTupleOperationCost float64 // 0.01 - cost per tuple operation
	CpuIndexTupleOperationCost float64 // 0.005 - cost per index operation

	// Table bloat metrics
	AveragePageSize int64 // 8192 - PostgreSQL page size in bytes
	AverageTupleSize int64 // 100-200 bytes typical
}

// NewCostCalculator creates a new cost calculator with PostgreSQL defaults
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{
		SeqScanCostPerPage:        1.0,
		RandomAccessCostPerTuple:  4.0,
		RandomPageCost:            4.0,
		CpuTupleOperationCost:     0.01,
		CpuIndexTupleOperationCost: 0.005,
		AveragePageSize:           8192,
		AverageTupleSize:          150,
	}
}

// EstimateVacuumDuration estimates how long a VACUUM operation will take
func (cc *CostCalculator) EstimateVacuumDuration(tableSize int64, deadTuples int64) float64 {
	if tableSize <= 0 {
		return 0.0
	}

	// Calculate number of pages to scan
	pages := float64(tableSize) / float64(cc.AveragePageSize)

	// Sequential scan cost (dominant factor)
	scanCost := pages * cc.SeqScanCostPerPage

	// Index cleanup cost (if any)
	indexCost := float64(deadTuples) * cc.CpuIndexTupleOperationCost

	// Tuple processing cost
	tupleCost := float64(deadTuples) * cc.CpuTupleOperationCost

	// Total cost in milliseconds (rough estimate)
	totalCost := scanCost + indexCost + tupleCost

	// Convert to seconds
	return totalCost / 1000.0
}

// EstimateVacuumImpact estimates the impact of VACUUM on system performance
func (cc *CostCalculator) EstimateVacuumImpact(databaseSize int64, tableSize int64) VacuumImpactMetrics {
	impact := VacuumImpactMetrics{
		TableSize:         tableSize,
		DatabaseSize:      databaseSize,
		BlocketTables:     0,
		QuerySlowdownFactor: 1.0,
		DiskIOIncrease:    1.0,
	}

	// Percentage of database this VACUUM will process
	percentageOfDB := float64(tableSize) / float64(databaseSize)

	if percentageOfDB > 0.1 {
		// Large table - significant disk I/O
		impact.DiskIOIncrease = 2.5
		impact.QuerySlowdownFactor = 1.2
	} else if percentageOfDB > 0.05 {
		// Medium table
		impact.DiskIOIncrease = 1.8
		impact.QuerySlowdownFactor = 1.1
	} else {
		// Small table - minimal impact
		impact.DiskIOIncrease = 1.2
		impact.QuerySlowdownFactor = 1.02
	}

	return impact
}

// CalculateOptimalSchedule determines the best time window for VACUUM
func (cc *CostCalculator) CalculateOptimalSchedule(tableSize int64, peakHours int) ScheduleRecommendation {
	duration := cc.EstimateVacuumDuration(tableSize, tableSize/100) // Assume 1% bloat

	rec := ScheduleRecommendation{
		EstimatedDuration: duration,
		RecommendedWindow: "maintenance window",
		Rationale:         "VACUUM should run during low-traffic periods",
	}

	// Determine optimal schedule based on table size and impact
	if duration < 1.0 {
		// Very fast VACUUM - can run anytime
		rec.RecommendedWindow = "can run anytime"
		rec.Rationale = "Short duration allows execution during business hours"
	} else if duration < 60.0 {
		// Moderate duration - prefer off-peak
		rec.RecommendedWindow = "early morning or evening"
		rec.Rationale = "Medium duration suggests off-peak execution"
	} else {
		// Long duration - requires dedicated maintenance window
		rec.RecommendedWindow = "dedicated maintenance window"
		rec.Rationale = "Long duration requires dedicated low-traffic window"
	}

	return rec
}

// CalculateRecoverableSpace estimates how much disk space can be recovered
func (cc *CostCalculator) CalculateRecoverableSpace(tableSize int64, deadTuplesRatio float64) int64 {
	if tableSize <= 0 || deadTuplesRatio <= 0 {
		return 0
	}

	// Dead space that could be recovered
	deadSpace := float64(tableSize) * (deadTuplesRatio / 100.0)

	// Apply recovery factor based on bloat level
	recoveryFactor := 0.85
	if deadTuplesRatio < 10.0 {
		recoveryFactor = 0.7 // Less fragmented tables recover less
	} else if deadTuplesRatio > 30.0 {
		recoveryFactor = 0.95 // Highly fragmented tables recover more
	}

	recovered := deadSpace * recoveryFactor

	return int64(math.Max(recovered, 0.0))
}

// CalculateIndexBlowup estimates index bloat from dead tuples
func (cc *CostCalculator) CalculateIndexBlowup(tableSize int64, deadTuples int64, indexCount int) int64 {
	if tableSize <= 0 || deadTuples <= 0 || indexCount <= 0 {
		return 0
	}

	// Rough estimation: each dead tuple wastes space in each index
	// Index bloat is typically 0.5-1.5x the table bloat
	spacePerDeadTuple := int64(50) // bytes in index per dead tuple
	indexBloat := int64(deadTuples) * spacePerDeadTuple * int64(indexCount)

	return indexBloat
}

// CalculateAutovacuumEfficiency calculates how well autovacuum is working
func (cc *CostCalculator) CalculateAutovacuumEfficiency(
	tableSize int64,
	deadTuplesRatio float64,
	lastVacuumDaysAgo float64,
	churnRateTuplesPerDay int64,
) float64 {
	if tableSize <= 0 {
		return 100.0 // Perfect score for empty table
	}

	// Calculate expected dead tuple ratio if autovacuum worked perfectly
	expectedDead := float64(churnRateTuplesPerDay) * lastVacuumDaysAgo
	expectedRatio := (expectedDead / float64(tableSize)) * 100.0

	// Efficiency = (expected ratio / actual ratio) * 100
	// Less than 100% means autovacuum is falling behind
	if expectedRatio <= 0 || deadTuplesRatio <= 0 {
		return 100.0
	}

	efficiency := (expectedRatio / deadTuplesRatio) * 100.0

	// Clamp between 0-200% (anything over 100% means autovacuum is ahead)
	return math.Min(efficiency, 200.0)
}

// VacuumImpactMetrics represents the impact of a VACUUM operation
type VacuumImpactMetrics struct {
	TableSize           int64
	DatabaseSize        int64
	BlocketTables       int
	QuerySlowdownFactor float64
	DiskIOIncrease      float64
}

// ScheduleRecommendation represents when a VACUUM should be scheduled
type ScheduleRecommendation struct {
	EstimatedDuration float64
	RecommendedWindow string
	Rationale         string
}
