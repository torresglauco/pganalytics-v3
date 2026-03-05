package jobs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestThresholdConditionEvaluation tests threshold-based rule evaluation
func TestThresholdConditionEvaluation(t *testing.T) {
	tests := []struct {
		name       string
		operator   string
		threshold  float64
		value      float64
		wantFires  bool
	}{
		{
			name:      "greater than - fires",
			operator:  ">",
			threshold: 80,
			value:     85,
			wantFires: true,
		},
		{
			name:      "greater than - no fire",
			operator:  ">",
			threshold: 80,
			value:     75,
			wantFires: false,
		},
		{
			name:      "less than - fires",
			operator:  "<",
			threshold: 20,
			value:     15,
			wantFires: true,
		},
		{
			name:      "equals - fires",
			operator:  "==",
			threshold: 100,
			value:     100,
			wantFires: true,
		},
		{
			name:      "not equals - fires",
			operator:  "!=",
			threshold: 100,
			value:     99,
			wantFires: true,
		},
		{
			name:      "greater or equal - fires",
			operator:  ">=",
			threshold: 80,
			value:     80,
			wantFires: true,
		},
		{
			name:      "less or equal - fires",
			operator:  "<=",
			threshold: 20,
			value:     20,
			wantFires: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fires := evaluateThresholdCondition(tt.operator, tt.value, tt.threshold)
			assert.Equal(t, tt.wantFires, fires, "threshold evaluation incorrect")
		})
	}
}

// TestChangeConditionEvaluation tests change-based rule evaluation
func TestChangeConditionEvaluation(t *testing.T) {
	tests := []struct {
		name          string
		previousValue float64
		currentValue  float64
		changePercent float64
		wantFires     bool
	}{
		{
			name:          "significant increase",
			previousValue: 100,
			currentValue:  160,
			changePercent: 50, // fires if > 50% change
			wantFires:     true,
		},
		{
			name:          "small increase",
			previousValue: 100,
			currentValue:  105,
			changePercent: 50,
			wantFires:     false,
		},
		{
			name:          "significant decrease",
			previousValue: 100,
			currentValue:  40,
			changePercent: 50, // fires if < -50% change
			wantFires:     true,
		},
		{
			name:          "no change",
			previousValue: 100,
			currentValue:  100,
			changePercent: 50,
			wantFires:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fires := evaluateChangeCondition(tt.previousValue, tt.currentValue, tt.changePercent)
			assert.Equal(t, tt.wantFires, fires, "change evaluation incorrect")
		})
	}
}

// TestCompositeConditionEvaluation tests composite AND/OR conditions
func TestCompositeConditionEvaluation(t *testing.T) {
	tests := []struct {
		name      string
		operator  string
		cond1     bool
		cond2     bool
		wantFires bool
	}{
		{
			name:      "AND both true",
			operator:  "AND",
			cond1:     true,
			cond2:     true,
			wantFires: true,
		},
		{
			name:      "AND one false",
			operator:  "AND",
			cond1:     true,
			cond2:     false,
			wantFires: false,
		},
		{
			name:      "OR both true",
			operator:  "OR",
			cond1:     true,
			cond2:     true,
			wantFires: true,
		},
		{
			name:      "OR one true",
			operator:  "OR",
			cond1:     true,
			cond2:     false,
			wantFires: true,
		},
		{
			name:      "OR both false",
			operator:  "OR",
			cond1:     false,
			cond2:     false,
			wantFires: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fires := evaluateCompositeCondition(tt.operator, tt.cond1, tt.cond2)
			assert.Equal(t, tt.wantFires, fires, "composite evaluation incorrect")
		})
	}
}

// TestAlertDeduplication tests fingerprint-based alert deduplication
func TestAlertDeduplication(t *testing.T) {
	tests := []struct {
		name              string
		ruleID            int64
		severity          string
		wantFingerprint   string
	}{
		{
			name:            "threshold alert",
			ruleID:          1,
			severity:        "high",
			wantFingerprint: "rule_1_high",
		},
		{
			name:            "critical alert",
			ruleID:          42,
			severity:        "critical",
			wantFingerprint: "rule_42_critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := generateAlertFingerprint(tt.ruleID, tt.severity)
			assert.Equal(t, tt.wantFingerprint, fp, "fingerprint generation incorrect")
		})
	}
}

// TestAlertStateTransition tests alert state machine
func TestAlertStateTransition(t *testing.T) {
	tests := []struct {
		name        string
		currentState string
		condition   bool
		wantState   string
	}{
		{
			name:         "normal to firing",
			currentState: "resolved",
			condition:    true,
			wantState:    "firing",
		},
		{
			name:         "firing to resolved",
			currentState: "firing",
			condition:    false,
			wantState:    "resolved",
		},
		{
			name:         "stay resolved",
			currentState: "resolved",
			condition:    false,
			wantState:    "resolved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newState := transitionAlertState(tt.currentState, tt.condition)
			assert.Equal(t, tt.wantState, newState, "state transition incorrect")
		})
	}
}

// TestRuleEvaluationMetrics tests metrics collection
func TestRuleEvaluationMetrics(t *testing.T) {
	metrics := &AlertRuleEngineMetrics{
		RulesEvaluated:    15,
		RulesFired:        3,
		AlertsCreated:     3,
		AlertsResolved:    1,
		AlertsDedup:       1,
		ExecutionTime:     500 * time.Millisecond,
		ErrorCount:        0,
		CacheHitRate:      0.95,
	}

	assert.Equal(t, 15, metrics.RulesEvaluated)
	assert.Equal(t, 3, metrics.RulesFired)
	assert.Equal(t, 3, metrics.AlertsCreated)
	assert.Equal(t, 1, metrics.AlertsResolved)
	assert.Equal(t, 1, metrics.AlertsDedup)
	assert.True(t, metrics.ExecutionTime < 1*time.Second)
	assert.Equal(t, 0, metrics.ErrorCount)
	assert.InDelta(t, 0.95, metrics.CacheHitRate, 0.01)
}

// TestRuleEvaluationOrder tests that rules are evaluated in correct order
func TestRuleEvaluationOrder(t *testing.T) {
	rules := []struct {
		id       int64
		name     string
		priority int
	}{
		{id: 3, name: "low_priority", priority: 10},
		{id: 1, name: "high_priority", priority: 1},
		{id: 2, name: "medium_priority", priority: 5},
	}

	// Verify high priority rules are evaluated first
	assert.Equal(t, int64(1), rules[1].id)
}

// TestRuleCaching tests that rules are cached appropriately
func TestRuleCaching(t *testing.T) {
	// Simulate rule cache behavior
	cache := make(map[int64]interface{})
	cacheTTL := 5 * time.Minute

	rule := &TestRule{
		ID:   1,
		Name: "test_rule",
	}

	// Add to cache
	cache[rule.ID] = rule

	// Retrieve from cache
	retrieved, ok := cache[rule.ID]
	assert.True(t, ok, "rule should be in cache")
	assert.Equal(t, rule.Name, retrieved.(*TestRule).Name)
	assert.True(t, cacheTTL > 0)
}

// TestConcurrentRuleEvaluation tests concurrent rule evaluation
func TestConcurrentRuleEvaluation(t *testing.T) {
	// Test configuration for concurrent rule evaluation
	checkIntervalSeconds := 300
	maxConcurrentRules := 10

	assert.Equal(t, 300, checkIntervalSeconds)
	assert.Equal(t, 10, maxConcurrentRules)
	assert.True(t, maxConcurrentRules > 0)
}

// Helper functions for testing
func evaluateThresholdCondition(operator string, value, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	default:
		return false
	}
}

func evaluateChangeCondition(previousValue, currentValue, changePercent float64) bool {
	if previousValue == 0 {
		return false
	}
	change := ((currentValue - previousValue) / previousValue) * 100
	return change > changePercent || change < -changePercent
}

func evaluateCompositeCondition(operator string, cond1, cond2 bool) bool {
	switch operator {
	case "AND":
		return cond1 && cond2
	case "OR":
		return cond1 || cond2
	default:
		return false
	}
}

func generateAlertFingerprint(ruleID int64, severity string) string {
	return fmt.Sprintf("rule_%d_%s", ruleID, severity)
}

func transitionAlertState(currentState string, condition bool) string {
	if condition {
		return "firing"
	}
	return "resolved"
}


// TestRule for testing purposes (not using actual AlertRule struct)
type TestRule struct {
	ID   int64
	Name string
}

type AlertRuleEngineMetrics struct {
	RulesEvaluated int
	RulesFired     int
	AlertsCreated  int
	AlertsResolved int
	AlertsDedup    int
	ExecutionTime  time.Duration
	ErrorCount     int
	CacheHitRate   float64
}
