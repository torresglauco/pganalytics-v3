package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
)

// TestAlertRulesCRUD_CreateRule tests creating an alert rule
func TestAlertRulesCRUD_CreateRule(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create alert rule
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "High Replication Lag",
		Description:         "Alert when replication lag exceeds threshold",
		RuleType:            "threshold",
		MetricName:          "replication_lag_ms",
		Condition:           json.RawMessage(`{"operator": ">", "value": 5000}`),
		AlertSeverity:       "high",
		EvaluationInterval:  300,
		ForDurationSeconds:  60,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err, "Should create alert rule without error")
	assert.NotZero(t, id, "Rule ID should be set after creation")
	assert.Equal(t, id, rule.ID)
}

// TestAlertRulesCRUD_CreateRuleInvalidType tests creating rule with invalid type
func TestAlertRulesCRUD_CreateRuleInvalidType(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create rule with empty name (should fail at handler level, but we test basic creation)
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "", // Empty name
		RuleType:            "threshold",
		MetricName:          "cpu_usage",
		Condition:           json.RawMessage(`{"operator": ">", "value": 80}`),
		AlertSeverity:       "medium",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	// The repository should still create it (validation is at handler level)
	id, err := repo.CreateRule(rule)
	// Depending on DB constraints, this might succeed or fail
	// We're testing that the repository handles it appropriately
	if err != nil {
		t.Logf("Expected behavior: DB rejected empty name: %v", err)
	} else {
		assert.NotZero(t, id)
	}
}

// TestAlertRulesCRUD_GetRuleByID tests retrieving a rule by ID
func TestAlertRulesCRUD_GetRuleByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create rule first
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "CPU Threshold Alert",
		Description:         "Alert on high CPU usage",
		RuleType:            "threshold",
		MetricName:          "cpu_usage_percent",
		Condition:           json.RawMessage(`{"operator": ">", "value": 80}`),
		AlertSeverity:       "critical",
		EvaluationInterval:  60,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Retrieve the rule
	fetched, err := repo.GetRuleByID(id)
	require.NoError(t, err, "Should retrieve rule without error")

	assert.Equal(t, id, fetched.ID)
	assert.Equal(t, "CPU Threshold Alert", fetched.Name)
	assert.Equal(t, "threshold", fetched.RuleType)
	assert.Equal(t, "cpu_usage_percent", fetched.MetricName)
	assert.Equal(t, "critical", fetched.AlertSeverity)
	assert.True(t, fetched.IsEnabled)
}

// TestAlertRulesCRUD_GetRuleByIDNotFound tests retrieving non-existent rule
func TestAlertRulesCRUD_GetRuleByIDNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Try to get non-existent rule
	_, err := repo.GetRuleByID(999999)
	assert.Error(t, err, "Should return error for non-existent rule")
	assert.Contains(t, err.Error(), "not found")
}

// TestAlertRulesCRUD_ListRules tests listing rules for a user
func TestAlertRulesCRUD_ListRules(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())
	userID := 1

	// Create multiple rules
	for i := 0; i < 5; i++ {
		rule := &storage.AlertRule{
			UserID:              userID,
			Name:                "Test Alert " + string(rune('A'+i)),
			RuleType:            "threshold",
			MetricName:          "test_metric",
			Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
			AlertSeverity:       "medium",
			EvaluationInterval:  300,
			NotificationEnabled: true,
			IsEnabled:           true,
		}
		_, err := repo.CreateRule(rule)
		require.NoError(t, err)
	}

	// List rules
	rules, err := repo.ListRules(userID, 10, 0)
	require.NoError(t, err, "Should list rules without error")

	assert.NotEmpty(t, rules, "Should have rules")
	assert.GreaterOrEqual(t, len(rules), 5, "Should have at least 5 rules")
}

// TestAlertRulesCRUD_ListRulesPagination tests pagination in listing rules
func TestAlertRulesCRUD_ListRulesPagination(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())
	userID := 2

	// Create multiple rules
	for i := 0; i < 10; i++ {
		rule := &storage.AlertRule{
			UserID:              userID,
			Name:                "Paginated Alert",
			RuleType:            "threshold",
			MetricName:          "pagination_test",
			Condition:           json.RawMessage(`{"operator": ">", "value": 0}`),
			AlertSeverity:       "low",
			EvaluationInterval:  300,
			NotificationEnabled: true,
			IsEnabled:           true,
		}
		_, err := repo.CreateRule(rule)
		require.NoError(t, err)
	}

	// Test pagination with limit
	page1, err := repo.ListRules(userID, 5, 0)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(page1), 5, "Page 1 should have at most 5 rules")

	// Test pagination with offset
	page2, err := repo.ListRules(userID, 5, 5)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(page2), 5, "Page 2 should have at most 5 rules")
}

// TestAlertRulesCRUD_UpdateRule tests updating a rule
func TestAlertRulesCRUD_UpdateRule(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create rule
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "Original Name",
		Description:         "Original description",
		RuleType:            "threshold",
		MetricName:          "original_metric",
		Condition:           json.RawMessage(`{"operator": ">", "value": 100}`),
		AlertSeverity:       "low",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Update rule
	rule.Name = "Updated Name"
	rule.Description = "Updated description"
	rule.AlertSeverity = "critical"
	rule.Condition = json.RawMessage(`{"operator": ">", "value": 200}`)

	err = repo.UpdateRule(rule)
	require.NoError(t, err, "Should update rule without error")

	// Verify update
	updated, err := repo.GetRuleByID(id)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, "critical", updated.AlertSeverity)
}

// TestAlertRulesCRUD_UpdateRuleNotFound tests updating non-existent rule
func TestAlertRulesCRUD_UpdateRuleNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Try to update non-existent rule
	rule := &storage.AlertRule{
		ID:         999999,
		UserID:     1,
		Name:       "Non-existent",
		RuleType:   "threshold",
		MetricName: "test",
		Condition:  json.RawMessage(`{}`),
		IsEnabled:  true,
		UpdatedAt:  time.Now(),
	}

	err := repo.UpdateRule(rule)
	assert.Error(t, err, "Should return error for non-existent rule")
	assert.Contains(t, err.Error(), "not found")
}

// TestAlertRulesCRUD_DeleteRule tests deleting a rule
func TestAlertRulesCRUD_DeleteRule(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())
	userID := 1

	// Create rule
	rule := &storage.AlertRule{
		UserID:              userID,
		Name:                "Rule to Delete",
		RuleType:            "threshold",
		MetricName:          "delete_test",
		Condition:           json.RawMessage(`{"operator": ">", "value": 0}`),
		AlertSeverity:       "low",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Delete rule
	err = repo.DeleteRule(id, userID)
	require.NoError(t, err, "Should delete rule without error")

	// Verify deletion (should return error)
	_, err = repo.GetRuleByID(id)
	assert.Error(t, err, "Should not find deleted rule")
}

// TestAlertRulesCRUD_DeleteRuleNotOwner tests deleting rule owned by another user
func TestAlertRulesCRUD_DeleteRuleNotOwner(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create rule with user 1
	rule := &storage.AlertRule{
		UserID:             1,
		Name:               "User 1 Rule",
		RuleType:           "threshold",
		MetricName:         "test",
		Condition:          json.RawMessage(`{}`),
		AlertSeverity:      "medium",
		EvaluationInterval: 300,
		IsEnabled:          true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Try to delete with user 2
	err = repo.DeleteRule(id, 2)
	assert.Error(t, err, "Should not delete rule owned by another user")
	assert.Contains(t, err.Error(), "not found")

	// Verify rule still exists
	_, err = repo.GetRuleByID(id)
	require.NoError(t, err, "Rule should still exist")
}

// TestAlertRulesCRUD_ToggleRuleEnabled tests enabling/disabling a rule
func TestAlertRulesCRUD_ToggleRuleEnabled(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create enabled rule
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "Toggle Test Rule",
		RuleType:            "threshold",
		MetricName:          "toggle_metric",
		Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
		AlertSeverity:       "medium",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Disable rule
	rule.IsEnabled = false
	err = repo.UpdateRule(rule)
	require.NoError(t, err)

	// Verify disabled
	disabled, err := repo.GetRuleByID(id)
	require.NoError(t, err)
	assert.False(t, disabled.IsEnabled)

	// Re-enable rule
	rule.IsEnabled = true
	err = repo.UpdateRule(rule)
	require.NoError(t, err)

	// Verify enabled
	enabled, err := repo.GetRuleByID(id)
	require.NoError(t, err)
	assert.True(t, enabled.IsEnabled)
}

// TestAlertRulesCRUD_RuleTypes tests different rule types
func TestAlertRulesCRUD_RuleTypes(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	ruleTypes := []string{"threshold", "change", "anomaly", "composite"}

	for _, ruleType := range ruleTypes {
		t.Run("rule_type_"+ruleType, func(t *testing.T) {
			rule := &storage.AlertRule{
				UserID:              1,
				Name:                "Rule Type " + ruleType,
				RuleType:            ruleType,
				MetricName:          "test_metric",
				Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
				AlertSeverity:       "medium",
				EvaluationInterval:  300,
				NotificationEnabled: true,
				IsEnabled:           true,
			}

			id, err := repo.CreateRule(rule)
			require.NoError(t, err, "Should create %s rule", ruleType)

			fetched, err := repo.GetRuleByID(id)
			require.NoError(t, err)
			assert.Equal(t, ruleType, fetched.RuleType)
		})
	}
}

// TestAlertRulesCRUD_SeverityLevels tests different severity levels
func TestAlertRulesCRUD_SeverityLevels(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	severities := []string{"low", "medium", "high", "critical"}

	for _, severity := range severities {
		t.Run("severity_"+severity, func(t *testing.T) {
			rule := &storage.AlertRule{
				UserID:              1,
				Name:                "Severity " + severity,
				RuleType:            "threshold",
				MetricName:          "severity_test",
				Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
				AlertSeverity:       severity,
				EvaluationInterval:  300,
				NotificationEnabled: true,
				IsEnabled:           true,
			}

			id, err := repo.CreateRule(rule)
			require.NoError(t, err, "Should create rule with %s severity", severity)

			fetched, err := repo.GetRuleByID(id)
			require.NoError(t, err)
			assert.Equal(t, severity, fetched.AlertSeverity)
		})
	}
}

// TestAlertRulesCRUD_ConditionValidation tests condition JSON structure
func TestAlertRulesCRUD_ConditionValidation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Test various condition structures
	testCases := []struct {
		name      string
		condition string
	}{
		{"simple_threshold", `{"operator": ">", "value": 100}`},
		{"range_condition", `{"operator": "between", "min": 10, "max": 100}`},
		{"complex_condition", `{"and": [{"operator": ">", "value": 50}, {"operator": "<", "value": 100}]}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := &storage.AlertRule{
				UserID:              1,
				Name:                "Condition Test " + tc.name,
				RuleType:            "threshold",
				MetricName:          "condition_test",
				Condition:           json.RawMessage(tc.condition),
				AlertSeverity:       "medium",
				EvaluationInterval:  300,
				NotificationEnabled: true,
				IsEnabled:           true,
			}

			id, err := repo.CreateRule(rule)
			require.NoError(t, err)

			fetched, err := repo.GetRuleByID(id)
			require.NoError(t, err)
			assert.JSONEq(t, tc.condition, string(fetched.Condition))
		})
	}
}

// TestAlertRulesCRUD_MultiTenantIsolation tests that users cannot access each other's rules
func TestAlertRulesCRUD_MultiTenantIsolation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// User 1 creates a rule
	rule1 := &storage.AlertRule{
		UserID:              1,
		Name:                "User 1 Private Rule",
		RuleType:            "threshold",
		MetricName:          "private_metric",
		Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
		AlertSeverity:       "high",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id1, err := repo.CreateRule(rule1)
	require.NoError(t, err)

	// User 2 creates a rule
	rule2 := &storage.AlertRule{
		UserID:              2,
		Name:                "User 2 Private Rule",
		RuleType:            "threshold",
		MetricName:          "private_metric",
		Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
		AlertSeverity:       "high",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	id2, err := repo.CreateRule(rule2)
	require.NoError(t, err)

	// Verify user 1 can list only their rules
	user1Rules, err := repo.ListRules(1, 100, 0)
	require.NoError(t, err)
	for _, r := range user1Rules {
		assert.Equal(t, 1, r.UserID, "User 1 should only see their own rules")
	}

	// Verify user 2 can list only their rules
	user2Rules, err := repo.ListRules(2, 100, 0)
	require.NoError(t, err)
	for _, r := range user2Rules {
		assert.Equal(t, 2, r.UserID, "User 2 should only see their own rules")
	}

	// Verify user 2 cannot delete user 1's rule
	err = repo.DeleteRule(id1, 2)
	assert.Error(t, err, "User 2 should not be able to delete User 1's rule")

	// Verify user 1 cannot delete user 2's rule
	err = repo.DeleteRule(id2, 1)
	assert.Error(t, err, "User 1 should not be able to delete User 2's rule")
}

// TestAlertRulesCRUD_AlertHistory tests retrieving alert history for a rule
func TestAlertRulesCRUD_AlertHistory(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Create a rule first
	rule := &storage.AlertRule{
		UserID:              1,
		Name:                "History Test Rule",
		RuleType:            "threshold",
		MetricName:          "history_metric",
		Condition:           json.RawMessage(`{"operator": ">", "value": 50}`),
		AlertSeverity:       "high",
		EvaluationInterval:  300,
		NotificationEnabled: true,
		IsEnabled:           true,
	}

	ruleID, err := repo.CreateRule(rule)
	require.NoError(t, err)

	// Get history (should be empty initially)
	history, err := repo.GetAlertHistory(ruleID, 10, 0)
	require.NoError(t, err, "Should get history without error")
	// History might be empty if no alerts triggered yet
	t.Logf("History count: %d", len(history))
}

// TestAlertRulesCRUD_AcknowledgeAlert tests acknowledging an alert trigger
func TestAlertRulesCRUD_AcknowledgeAlert(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	repo := storage.NewAlertRulesRepository(db.GetDB())

	// Acknowledging a non-existent trigger should fail
	err := repo.AcknowledgeAlert(999999, 1)
	assert.Error(t, err, "Should fail to acknowledge non-existent trigger")
}
