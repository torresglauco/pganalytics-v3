package index_advisor

type CostCalculator struct{}

func NewCostCalculator() *CostCalculator {
	return &CostCalculator{}
}

// CalculateImprovement calculates the percentage improvement from costWithout to costWith
func (cc *CostCalculator) CalculateImprovement(costWithout, costWith float64) float64 {
	if costWithout == 0 {
		return 0
	}
	return ((costWithout - costWith) / costWithout) * 100
}

// EstimateBenefit multiplies cost improvement by frequency to get absolute benefit
func (cc *CostCalculator) EstimateBenefit(costImprovement, frequency float64) float64 {
	return costImprovement * frequency / 100
}

// CalculateIndexMaintenanceCost estimates the write overhead cost of maintaining an index
// Each index adds approximately 2-5% write overhead, using 3% as average
func (cc *CostCalculator) CalculateIndexMaintenanceCost(tableWriteFrequency float64) float64 {
	return tableWriteFrequency * 0.03
}

// ShouldCreateIndex determines if an index should be created based on benefit vs maintenance cost
// Creates index if benefit > maintenance cost by 2x factor for good ROI
func (cc *CostCalculator) ShouldCreateIndex(benefit, maintenanceCost float64) bool {
	return benefit > (maintenanceCost * 2)
}
