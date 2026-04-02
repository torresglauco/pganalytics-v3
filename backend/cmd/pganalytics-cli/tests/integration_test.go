package tests

import (
	"bytes"
	"path/filepath"
	"testing"
	"pganalytics-cli/commands"
	"pganalytics-cli/internal/config"
)

// ============================================================================
// Integration Tests
// ============================================================================

// TestQueryCommandIntegration tests full query command workflow
func TestQueryCommandIntegration(t *testing.T) {
	// Test: List queries -> Analyze specific query
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("query list failed: %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("Top Queries")) {
		t.Errorf("Expected query list output")
	}
}

// TestQueryAnalyzeIntegration tests query analyze with various query IDs
func TestQueryAnalyzeIntegration(t *testing.T) {
	queryIDs := []string{"1", "42", "999"}

	for _, qID := range queryIDs {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"query", "analyze", qID})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("query analyze %s failed: %v", qID, err)
		}

		if !bytes.Contains(out.Bytes(), []byte(qID)) {
			t.Errorf("Expected query ID %s in output", qID)
		}
	}
}

// TestQueryExplainIntegration tests query explain with different SQL queries
func TestQueryExplainIntegration(t *testing.T) {
	queries := []string{
		"SELECT * FROM users",
		"SELECT id, name FROM users WHERE id = 1",
		"SELECT COUNT(*) FROM orders GROUP BY user_id",
	}

	for _, sql := range queries {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"query", "explain", sql})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("query explain failed for: %s, error: %v", sql, err)
		}

		if !bytes.Contains(out.Bytes(), []byte("EXPLAIN")) {
			t.Errorf("Expected EXPLAIN output for query: %s", sql)
		}
	}
}

// TestIndexCommandIntegration tests full index command workflow
func TestIndexCommandIntegration(t *testing.T) {
	// Test: Suggest -> Create -> Check workflow

	// Step 1: Suggest indexes
	cmd1 := commands.NewRootCmd("0.1.0")
	cmd1.SetArgs([]string{"index", "suggest"})

	var out1 bytes.Buffer
	cmd1.SetOut(&out1)

	if err := cmd1.Execute(); err != nil {
		t.Fatalf("index suggest failed: %v", err)
	}

	// Step 2: Create index
	cmd2 := commands.NewRootCmd("0.1.0")
	cmd2.SetArgs([]string{"index", "create", "--table", "users", "--columns", "email"})

	var out2 bytes.Buffer
	cmd2.SetOut(&out2)

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("index create failed: %v", err)
	}

	if !bytes.Contains(out2.Bytes(), []byte("Creating")) {
		t.Errorf("Expected index creation output")
	}

	// Step 3: Check indexes
	cmd3 := commands.NewRootCmd("0.1.0")
	cmd3.SetArgs([]string{"index", "check"})

	var out3 bytes.Buffer
	cmd3.SetOut(&out3)

	if err := cmd3.Execute(); err != nil {
		t.Fatalf("index check failed: %v", err)
	}
	if !bytes.Contains(out3.Bytes(), []byte("Health")) {
		t.Errorf("Expected health report")
	}
}

// TestIndexSuggestWithTableIntegration tests index suggest for specific tables
func TestIndexSuggestWithTableIntegration(t *testing.T) {
	tables := []string{"users", "orders", "products"}

	for _, table := range tables {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"index", "suggest", table})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("index suggest for table %s failed: %v", table, err)
		}
		if !bytes.Contains(out.Bytes(), []byte("Recommendations")) {
			t.Errorf("Expected recommendations for table %s", table)
		}
	}
}

// TestIndexCreateWithOptionsIntegration tests index create with various options
func TestIndexCreateWithOptionsIntegration(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic create",
			args:    []string{"index", "create", "--table", "users", "--columns", "email"},
			wantErr: false,
		},
		{
			name:    "create with concurrent",
			args:    []string{"index", "create", "--table", "orders", "--columns", "user_id", "--concurrent"},
			wantErr: false,
		},
		{
			name:    "missing table",
			args:    []string{"index", "create", "--columns", "email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(tt.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("got error %v, want error %v", err != nil, tt.wantErr)
			}
		})
	}
}

// TestVacuumCommandIntegration tests full vacuum workflow
func TestVacuumCommandIntegration(t *testing.T) {
	// Test: Status -> Tune -> Estimate workflow

	// Step 1: Check status
	cmd1 := commands.NewRootCmd("0.1.0")
	cmd1.SetArgs([]string{"vacuum", "status"})

	var out1 bytes.Buffer
	cmd1.SetOut(&out1)

	if err := cmd1.Execute(); err != nil {
		t.Fatalf("vacuum status failed: %v", err)
	}
	if !bytes.Contains(out1.Bytes(), []byte("Status")) {
		t.Errorf("Expected status output")
	}

	// Step 2: Get tune recommendations
	cmd2 := commands.NewRootCmd("0.1.0")
	cmd2.SetArgs([]string{"vacuum", "tune"})

	var out2 bytes.Buffer
	cmd2.SetOut(&out2)

	if err := cmd2.Execute(); err != nil {
		t.Fatalf("vacuum tune failed: %v", err)
	}
	if !bytes.Contains(out2.Bytes(), []byte("Tuning")) {
		t.Errorf("Expected tuning recommendations")
	}

	// Step 3: Estimate duration
	cmd3 := commands.NewRootCmd("0.1.0")
	cmd3.SetArgs([]string{"vacuum", "estimate"})

	var out3 bytes.Buffer
	cmd3.SetOut(&out3)

	if err := cmd3.Execute(); err != nil {
		t.Fatalf("vacuum estimate failed: %v", err)
	}
	if !bytes.Contains(out3.Bytes(), []byte("Estimates")) {
		t.Errorf("Expected estimates")
	}
}

// TestVacuumTuneOptionsIntegration tests vacuum tune with different options
func TestVacuumTuneOptionsIntegration(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "dry-run (default)",
			args: []string{"vacuum", "tune"},
		},
		{
			name: "with apply",
			args: []string{"vacuum", "tune", "--apply"},
		},
		{
			name: "explicit dry-run",
			args: []string{"vacuum", "tune", "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(tt.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("vacuum tune failed: %v", err)
			}
			if !bytes.Contains(out.Bytes(), []byte("Autovacuum")) {
				t.Errorf("Expected output")
			}
		})
	}
}

// TestConfigCommandIntegration tests full config workflow
func TestConfigCommandIntegration(t *testing.T) {
	// Test: Set -> Get -> List workflow

	// Step 1: Set multiple config values
	configValues := map[string]string{
		"server":  "http://localhost:8080",
		"api-key": "sk_test_123",
		"timeout": "30s",
	}

	for key, value := range configValues {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"config", "set", key, value})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("config set %s=%s failed: %v", key, value, err)
		}
	}

	// Step 2: Get individual values
	for key := range configValues {
		cmd := commands.NewRootCmd("0.1.0")
		cmd.SetArgs([]string{"config", "get", key})

		var out bytes.Buffer
		cmd.SetOut(&out)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("config get %s failed: %v", key, err)
		}
	}

	// Step 3: List all values
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"config", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config list failed: %v", err)
	}
}

// TestConfigPersistenceIntegration tests config persistence across commands
func TestConfigPersistenceIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	// Create a persistent store
	store := config.NewFileStore(cfgPath)

	// Set some values
	store.Set("db_host", "localhost")
	store.Set("db_port", "5432")
	store.Set("db_name", "mydb")

	// Create another store instance and verify values persist
	store2 := config.NewFileStore(cfgPath)

	host, _ := store2.Get("db_host")
	if host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}

	port, _ := store2.Get("db_port")
	if port != "5432" {
		t.Errorf("Expected '5432', got '%s'", port)
	}

	name, _ := store2.Get("db_name")
	if name != "mydb" {
		t.Errorf("Expected 'mydb', got '%s'", name)
	}
}

// TestMCPCommandIntegration tests MCP command integration
func TestMCPCommandIntegration(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"mcp", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mcp status failed: %v", err)
	}

	// MCP status currently outputs to stdout directly, test passes if no error
}

// TestCLIWithGlobalFlags tests commands with global flags
func TestCLIWithGlobalFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "query with server flag",
			args: []string{"--server", "http://api.example.com", "query", "list"},
		},
		{
			name: "query with api-key flag",
			args: []string{"--api-key", "sk_test_123", "query", "list"},
		},
		{
			name: "query with format flag",
			args: []string{"--format", "json", "query", "list"},
		},
		{
			name: "all global flags",
			args: []string{"--server", "http://api.example.com", "--api-key", "key123", "--format", "csv", "query", "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := commands.NewRootCmd("0.1.0")
			cmd.SetArgs(tt.args)

			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("command failed: %v", err)
			}
		})
	}
}

// TestMultipleCommandSequence tests running multiple commands in sequence
func TestMultipleCommandSequence(t *testing.T) {
	// Simulate: config setup -> query list -> index suggest

	// 1. Set config
	cmd1 := commands.NewRootCmd("0.1.0")
	cmd1.SetArgs([]string{"config", "set", "db_host", "localhost"})
	var out1 bytes.Buffer
	cmd1.SetOut(&out1)
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	// 2. Query list
	cmd2 := commands.NewRootCmd("0.1.0")
	cmd2.SetArgs([]string{"query", "list"})
	var out2 bytes.Buffer
	cmd2.SetOut(&out2)
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("query list failed: %v", err)
	}

	// 3. Index suggest
	cmd3 := commands.NewRootCmd("0.1.0")
	cmd3.SetArgs([]string{"index", "suggest"})
	var out3 bytes.Buffer
	cmd3.SetOut(&out3)
	if err := cmd3.Execute(); err != nil {
		t.Fatalf("index suggest failed: %v", err)
	}

	// All should have produced output
	if out1.String() == "" || out2.String() == "" || out3.String() == "" {
		t.Fatal("Expected output from all commands")
	}
}
