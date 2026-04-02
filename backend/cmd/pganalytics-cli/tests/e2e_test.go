package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"pganalytics-cli/commands"
)

// ============================================================================
// End-to-End (E2E) Tests
// ============================================================================

// TestCLIFullWorkflow tests complete workflow: setup -> usage -> verification
func TestCLIFullWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	_ = filepath.Join(tmpDir, "pganalytics.conf")

	// Step 1: Setup - Create config
	setup := commands.NewRootCmd("0.1.0")
	setup.SetArgs([]string{"config", "set", "server", "http://localhost:8080"})

	var setupOut bytes.Buffer
	setup.SetOut(&setupOut)

	if err := setup.Execute(); err != nil {
		t.Fatalf("Setup failed - config set: %v", err)
	}

	setupOutput := setupOut.String()
	if !bytes.Contains(setupOut.Bytes(), []byte("server")) {
		t.Errorf("Setup output missing server config: %s", setupOutput)
	}

	// Step 2: Execute - Run query list
	execute := commands.NewRootCmd("0.1.0")
	execute.SetArgs([]string{"query", "list"})

	var executeOut bytes.Buffer
	execute.SetOut(&executeOut)

	if err := execute.Execute(); err != nil {
		t.Fatalf("Execute failed - query list: %v", err)
	}

	executeOutput := executeOut.String()
	if !bytes.Contains(executeOut.Bytes(), []byte("Top Queries")) {
		t.Errorf("Execute output missing query list: %s", executeOutput)
	}

	// Step 3: Verify - Check index suggestions
	verify := commands.NewRootCmd("0.1.0")
	verify.SetArgs([]string{"index", "suggest"})

	var verifyOut bytes.Buffer
	verify.SetOut(&verifyOut)

	if err := verify.Execute(); err != nil {
		t.Fatalf("Verify failed - index suggest: %v", err)
	}

	verifyOutput := verifyOut.String()
	if !bytes.Contains(verifyOut.Bytes(), []byte("Recommendations")) {
		t.Errorf("Verify output missing recommendations: %s", verifyOutput)
	}
}

// TestErrorHandling tests CLI error handling for various error conditions
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "missing subcommand argument",
			args:      []string{"query", "analyze"},
			expectErr: true,
			errorMsg:  "required argument",
		},
		{
			name:      "invalid command",
			args:      []string{"invalid", "command"},
			expectErr: true,
			errorMsg:  "unknown command",
		},
		{
			name:      "missing required flag",
			args:      []string{"index", "create", "--table", "users"},
			expectErr: true,
			errorMsg:  "required",
		},
		{
			name:      "config get nonexistent",
			args:      []string{"config", "get", "nonexistent_key_xyz"},
			expectErr: true,
			errorMsg:  "key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(tt.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error %v, got %v", tt.expectErr, err != nil)
			}
		})
	}
}

// TestHelpFlags tests help functionality for all commands
func TestHelpFlags(t *testing.T) {
	commands_list := [][]string{
		{"help"},
		{"--help"},
		{"query", "--help"},
		{"config", "--help"},
		{"index", "--help"},
		{"vacuum", "--help"},
		{"mcp", "--help"},
	}

	for _, args := range commands_list {
		t.Run("help for "+args[0], func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			// Help should not error, or we get a help error that's not fatal
			if out.Len() == 0 && err == nil {
				t.Errorf("Expected help output")
			}
		})
	}
}

// TestConfigWorkflow tests complete config management workflow
func TestConfigWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	_ = tmpDir // Temp directory for config

	// Setup: Set multiple config values
	setupConfigs := map[string]string{
		"server":   "http://localhost:8080",
		"api_key":  "sk_test_123",
		"database": "mydb",
		"username": "postgres",
		"timeout":  "30s",
	}

	for key, value := range setupConfigs {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"config", "set", key, value})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to set config %s: %v", key, err)
		}
	}

	// Verification: Get each config value
	for key := range setupConfigs {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"config", "get", key})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to get config %s: %v", key, err)
		}

		if out.Len() == 0 {
			t.Errorf("Expected output for config %s", key)
		}
	}

	// List all configs
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"config", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config list failed: %v", err)
	}
	if out.Len() == 0 {
		t.Errorf("Expected list output")
	}
}

// TestQueryAnalysisWorkflow tests complete query analysis workflow
func TestQueryAnalysisWorkflow(t *testing.T) {
	// Workflow: List -> Analyze -> Explain

	// 1. List top queries
	list := commands.NewRootCmd("0.1.0")
	list.SetArgs([]string{"query", "list"})

	var listOut bytes.Buffer
	list.SetOut(&listOut)

	if err := list.Execute(); err != nil {
		t.Fatalf("query list failed: %v", err)
	}

	if !bytes.Contains(listOut.Bytes(), []byte("Top Queries")) {
		t.Fatal("Expected query list output")
	}

	// 2. Analyze specific query
	analyze := commands.NewRootCmd("0.1.0")
	analyze.SetArgs([]string{"query", "analyze", "1"})

	var analyzeOut bytes.Buffer
	analyze.SetOut(&analyzeOut)

	if err := analyze.Execute(); err != nil {
		t.Fatalf("query analyze failed: %v", err)
	}

	if !bytes.Contains(analyzeOut.Bytes(), []byte("Query Analysis")) {
		t.Fatal("Expected analysis output")
	}

	// 3. Explain query execution
	explain := commands.NewRootCmd("0.1.0")
	explain.SetArgs([]string{"query", "explain", "SELECT * FROM users WHERE id = 1"})

	var explainOut bytes.Buffer
	explain.SetOut(&explainOut)

	if err := explain.Execute(); err != nil {
		t.Fatalf("query explain failed: %v", err)
	}

	if !bytes.Contains(explainOut.Bytes(), []byte("EXPLAIN")) {
		t.Fatal("Expected explain output")
	}
}

// TestIndexOptimizationWorkflow tests complete index optimization workflow
func TestIndexOptimizationWorkflow(t *testing.T) {
	// Workflow: Suggest -> Create -> Check

	// 1. Get suggestions
	suggest := commands.NewRootCmd("0.1.0")
	suggest.SetArgs([]string{"index", "suggest"})

	var suggestOut bytes.Buffer
	suggest.SetOut(&suggestOut)

	if err := suggest.Execute(); err != nil {
		t.Fatalf("index suggest failed: %v", err)
	}

	if !bytes.Contains(suggestOut.Bytes(), []byte("Recommendations")) {
		t.Fatal("Expected suggestions output")
	}

	// 2. Create index
	create := commands.NewRootCmd("0.1.0")
	create.SetArgs([]string{"index", "create", "--table", "users", "--columns", "email"})

	var createOut bytes.Buffer
	create.SetOut(&createOut)

	if err := create.Execute(); err != nil {
		t.Fatalf("index create failed: %v", err)
	}

	if !bytes.Contains(createOut.Bytes(), []byte("Creating")) {
		t.Fatal("Expected creation output")
	}

	// 3. Check index health
	check := commands.NewRootCmd("0.1.0")
	check.SetArgs([]string{"index", "check"})

	var checkOut bytes.Buffer
	check.SetOut(&checkOut)

	if err := check.Execute(); err != nil {
		t.Fatalf("index check failed: %v", err)
	}

	if !bytes.Contains(checkOut.Bytes(), []byte("Health")) {
		t.Fatal("Expected health check output")
	}
}

// TestVacuumMaintenanceWorkflow tests complete VACUUM workflow
func TestVacuumMaintenanceWorkflow(t *testing.T) {
	// Workflow: Status -> Tune -> Estimate

	// 1. Check status
	status := commands.NewRootCmd("0.1.0")
	status.SetArgs([]string{"vacuum", "status"})

	var statusOut bytes.Buffer
	status.SetOut(&statusOut)

	if err := status.Execute(); err != nil {
		t.Fatalf("vacuum status failed: %v", err)
	}

	if !bytes.Contains(statusOut.Bytes(), []byte("Status")) {
		t.Fatal("Expected status output")
	}

	// 2. Get tuning recommendations (dry-run)
	tune := commands.NewRootCmd("0.1.0")
	tune.SetArgs([]string{"vacuum", "tune", "--dry-run"})

	var tuneOut bytes.Buffer
	tune.SetOut(&tuneOut)

	if err := tune.Execute(); err != nil {
		t.Fatalf("vacuum tune failed: %v", err)
	}

	if !bytes.Contains(tuneOut.Bytes(), []byte("Tuning")) {
		t.Fatal("Expected tuning output")
	}

	// 3. Estimate duration
	estimate := commands.NewRootCmd("0.1.0")
	estimate.SetArgs([]string{"vacuum", "estimate"})

	var estimateOut bytes.Buffer
	estimate.SetOut(&estimateOut)

	if err := estimate.Execute(); err != nil {
		t.Fatalf("vacuum estimate failed: %v", err)
	}

	if !bytes.Contains(estimateOut.Bytes(), []byte("Estimates")) {
		t.Fatal("Expected estimate output")
	}
}

// TestDatabaseConnectionSetup tests setup of database connection config
func TestDatabaseConnectionSetup(t *testing.T) {
	dbConfigs := []struct {
		name  string
		key   string
		value string
	}{
		{"host", "db_host", "localhost"},
		{"port", "db_port", "5432"},
		{"database", "db_name", "myapp_db"},
		{"username", "db_user", "postgres"},
		{"password", "db_pass", "secretpass"},
	}

	for _, cfg := range dbConfigs {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"config", "set", cfg.key, cfg.value})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to set %s: %v", cfg.name, err)
		}

		// Verify it was set
		getCmd := commands.NewRootCmd("0.1.0")
		getCmd.SetArgs([]string{"config", "get", cfg.key})

		var getOut bytes.Buffer
		getCmd.SetOut(&getOut)

		if err := getCmd.Execute(); err != nil {
			t.Fatalf("Failed to get %s: %v", cfg.name, err)
		}
	}
}

// TestQueryFiltering tests query list with different filters
func TestQueryFiltering(t *testing.T) {
	filters := [][]string{
		{"query", "list", "--limit", "5"},
		{"query", "list", "--limit", "20"},
		{"query", "list", "--sort", "latency"},
		{"query", "list", "--sort", "calls"},
	}

	for _, args := range filters {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs(args)

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("query list with filters %v failed: %v", args, err)
		}

		if !bytes.Contains(out.Bytes(), []byte("Top Queries")) {
			t.Errorf("Expected query output for filters %v", args)
		}
	}
}

// TestIndexCreationVariations tests various index creation scenarios
func TestIndexCreationVariations(t *testing.T) {
	scenarios := []struct {
		name string
		args []string
	}{
		{
			name: "single column index",
			args: []string{"index", "create", "--table", "users", "--columns", "email"},
		},
		{
			name: "multi-column index",
			args: []string{"index", "create", "--table", "orders", "--columns", "user_id,created_at"},
		},
		{
			name: "concurrent creation",
			args: []string{"index", "create", "--table", "products", "--columns", "category", "--concurrent"},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(scenario.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Failed: %v", err)
			}

			if !bytes.Contains(out.Bytes(), []byte("Creating")) {
				t.Errorf("Expected creation output")
			}
		})
	}
}

// TestCompleteUserJourney simulates a real user journey
func TestCompleteUserJourney(t *testing.T) {
	// User journey:
	// 1. Initial setup - configure server
	// 2. Check query performance
	// 3. Analyze slow queries
	// 4. Get index recommendations
	// 5. Create indexes
	// 6. Monitor VACUUM status
	// 7. Tune autovacuum

	steps := []struct {
		name string
		args []string
	}{
		{"Setup server", []string{"config", "set", "server", "http://localhost:8080"}},
		{"Check queries", []string{"query", "list"}},
		{"Analyze slow query", []string{"query", "analyze", "42"}},
		{"Get index suggestions", []string{"index", "suggest"}},
		{"Create recommended index", []string{"index", "create", "--table", "users", "--columns", "email"}},
		{"Check VACUUM status", []string{"vacuum", "status"}},
		{"Tune autovacuum", []string{"vacuum", "tune", "--dry-run"}},
	}

	for i, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(step.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Step %d (%s) failed: %v", i+1, step.name, err)
			}

			if out.Len() == 0 {
				t.Errorf("Step %d (%s) produced no output", i+1, step.name)
			}
		})
	}
}

// TestEnvironmentVariables tests using environment variables
func TestEnvironmentVariables(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)

	// Commands should work with temporary HOME
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"config", "set", "test_var", "test_value"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed with temporary HOME: %v", err)
	}
}
