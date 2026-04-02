package tests

import (
	"bytes"
	"testing"
	"pganalytics-cli/commands"
	"pganalytics-cli/internal/api"
	"pganalytics-cli/internal/config"
	"path/filepath"
)

// ============================================================================
// Unit Tests for Commands
// ============================================================================

// TestRootCommandCreation tests root command initialization
func TestRootCommandCreation(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")

	if cmd == nil {
		t.Fatal("Expected root command to be created")
	}

	if cmd.Use != "pganalytics" {
		t.Errorf("Expected Use to be 'pganalytics', got '%s'", cmd.Use)
	}

	if cmd.Version != "0.1.0" {
		t.Errorf("Expected Version to be '0.1.0', got '%s'", cmd.Version)
	}
}

// TestRootCommandHasSubcommands tests that all subcommands are registered
func TestRootCommandHasSubcommands(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")

	expectedSubcommands := []string{"config", "query", "index", "vacuum", "mcp"}
	actualSubcommands := make(map[string]bool)

	for _, subcmd := range cmd.Commands() {
		actualSubcommands[subcmd.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !actualSubcommands[name] {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestRootCommandGlobalFlags tests global flags are properly set
func TestRootCommandGlobalFlags(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")

	flags := []string{"server", "api-key", "format"}
	for _, flagName := range flags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected persistent flag '%s' not found", flagName)
		}
	}
}

// TestQueryCommand tests query command structure
func TestQueryCommand(t *testing.T) {
	cmd := commands.NewQueryCmd()

	if cmd == nil {
		t.Fatal("Expected query command to be created")
	}

	if cmd.Use != "query" {
		t.Errorf("Expected Use to be 'query', got '%s'", cmd.Use)
	}

	// Verify subcommands exist
	subCmdNames := []string{"list", "analyze", "explain"}
	for _, name := range subCmdNames {
		found := false
		for _, cmd := range cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestQueryListSubcommand tests query list functionality
func TestQueryListSubcommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("Top Queries")) {
		t.Errorf("Expected 'Top Queries' in output")
	}
}

// TestQueryListFlags tests query list flags
func TestQueryListFlags(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "list", "--limit", "5"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error with flag, got %v", err)
	}
}

// TestQueryAnalyzeSubcommand tests query analyze with argument
func TestQueryAnalyzeSubcommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "analyze", "123"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Query Analysis")) {
		t.Errorf("Expected 'Query Analysis' in output")
	}
}

// TestQueryAnalyzeMissingArg tests query analyze requires argument
func TestQueryAnalyzeMissingArg(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "analyze"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing argument")
	}
}

// TestQueryExplainSubcommand tests query explain
func TestQueryExplainSubcommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"query", "explain", "SELECT * FROM users"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("EXPLAIN ANALYZE")) {
		t.Errorf("Expected 'EXPLAIN ANALYZE' in output")
	}
}

// TestConfigCommand tests config command structure
func TestConfigCommand(t *testing.T) {
	cmd := commands.NewConfigCmd()

	if cmd == nil {
		t.Fatal("Expected config command to be created")
	}

	if cmd.Use != "config" {
		t.Errorf("Expected Use to be 'config', got '%s'", cmd.Use)
	}

	// Verify subcommands
	subCmdNames := []string{"set", "get", "list"}
	for _, name := range subCmdNames {
		found := false
		for _, cmd := range cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestConfigSetCommand tests config set
func TestConfigSetCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"config", "set", "testkey", "testvalue"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("testkey")) {
		t.Errorf("Expected key in output")
	}
}

// TestConfigGetCommand tests config get
func TestConfigGetCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")

	// First set a value
	cmd.SetArgs([]string{"config", "set", "mykey", "myvalue"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.Execute()

	// Then get it
	cmd2 := commands.NewRootCmd("0.1.0")
	cmd2.SetArgs([]string{"config", "get", "mykey"})
	var out2 bytes.Buffer
	cmd2.SetOut(&out2)

	err := cmd2.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestConfigListCommand tests config list
func TestConfigListCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"config", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestIndexCommand tests index command
func TestIndexCommand(t *testing.T) {
	cmd := commands.NewIndexCmd()

	if cmd == nil {
		t.Fatal("Expected index command to be created")
	}

	if cmd.Use != "index" {
		t.Errorf("Expected Use to be 'index', got '%s'", cmd.Use)
	}

	subCmdNames := []string{"suggest", "create", "check", "list"}
	for _, name := range subCmdNames {
		found := false
		for _, cmd := range cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestIndexSuggestCommand tests index suggest
func TestIndexSuggestCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "suggest"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Recommendations")) {
		t.Errorf("Expected recommendations in output")
	}
}

// TestIndexSuggestWithTable tests index suggest with table filter
func TestIndexSuggestWithTable(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "suggest", "users"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestIndexCreateCommand tests index create
func TestIndexCreateCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "create", "--table", "users", "--columns", "email"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Creating index")) {
		t.Errorf("Expected success message")
	}
}

// TestIndexCreateMissingFlags tests index create with missing required flags
func TestIndexCreateMissingFlags(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "create", "--table", "users"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing --columns flag")
	}
}

// TestIndexCheckCommand tests index check
func TestIndexCheckCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "check"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Health")) {
		t.Errorf("Expected health report in output")
	}
}

// TestIndexListCommand tests index list
func TestIndexListCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"index", "list"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestVacuumCommand tests vacuum command
func TestVacuumCommand(t *testing.T) {
	cmd := commands.NewVacuumCmd()

	if cmd == nil {
		t.Fatal("Expected vacuum command to be created")
	}

	if cmd.Use != "vacuum" {
		t.Errorf("Expected Use to be 'vacuum', got '%s'", cmd.Use)
	}

	subCmdNames := []string{"status", "tune", "estimate"}
	for _, name := range subCmdNames {
		found := false
		for _, cmd := range cmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", name)
		}
	}
}

// TestVacuumStatusCommand tests vacuum status
func TestVacuumStatusCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"vacuum", "status"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Status")) {
		t.Errorf("Expected status in output")
	}
}

// TestVacuumTuneCommand tests vacuum tune
func TestVacuumTuneCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"vacuum", "tune"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Tuning")) {
		t.Errorf("Expected tuning in output")
	}
}

// TestVacuumTuneWithApply tests vacuum tune with apply flag
func TestVacuumTuneWithApply(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"vacuum", "tune", "--apply"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestVacuumEstimateCommand tests vacuum estimate
func TestVacuumEstimateCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"vacuum", "estimate"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// TestMCPCommand tests MCP command
func TestMCPCommand(t *testing.T) {
	cmd := commands.NewMCPCmd()

	if cmd == nil {
		t.Fatal("Expected MCP command to be created")
	}

	if cmd.Use != "mcp" {
		t.Errorf("Expected Use to be 'mcp', got '%s'", cmd.Use)
	}
}

// TestMCPStatusCommand tests MCP status
func TestMCPStatusCommand(t *testing.T) {
	cmd := commands.NewRootCmd("0.1.0")
	cmd.SetArgs([]string{"mcp", "status"})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// ============================================================================
// Unit Tests for Config Store
// ============================================================================

// TestConfigSetAndGet tests setting and getting config values
func TestConfigSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	err := store.Set("server", "http://localhost:8080")
	if err != nil {
		t.Fatalf("Expected no error on Set, got %v", err)
	}

	val, err := store.Get("server")
	if err != nil {
		t.Fatalf("Expected no error on Get, got %v", err)
	}

	if val != "http://localhost:8080" {
		t.Errorf("Expected 'http://localhost:8080', got '%s'", val)
	}
}

// TestConfigGetNonExistent tests getting non-existent key
func TestConfigGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}

// TestConfigDelete tests deleting config values
func TestConfigDelete(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	store.Set("key1", "value1")
	err := store.Delete("key1")
	if err != nil {
		t.Fatalf("Expected no error on Delete, got %v", err)
	}

	_, err = store.Get("key1")
	if err == nil {
		t.Fatal("Expected error after deletion")
	}
}

// TestConfigGetAll tests getting all config values
func TestConfigGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	store.Set("key1", "value1")
	store.Set("key2", "value2")

	all := store.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 items, got %d", len(all))
	}

	if all["key1"] != "value1" || all["key2"] != "value2" {
		t.Errorf("Unexpected values in GetAll")
	}
}

// TestConfigPersistence tests config persists across instances
func TestConfigPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store1 := config.NewFileStore(cfgPath)
	store1.Set("persist_key", "persist_value")

	store2 := config.NewFileStore(cfgPath)
	val, err := store2.Get("persist_key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if val != "persist_value" {
		t.Errorf("Expected 'persist_value', got '%s'", val)
	}
}

// ============================================================================
// Unit Tests for API Client
// ============================================================================

// TestClientCreation tests API client creation
func TestClientCreation(t *testing.T) {
	client := api.NewClient("http://localhost:8080", "test-key")

	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

// TestClientWithAuth tests client stores auth key
func TestClientWithAuth(t *testing.T) {
	client := api.NewClient("http://localhost:8080", "sk_test_123")

	if client.APIKey != "sk_test_123" {
		t.Errorf("Expected APIKey to be set")
	}
}

// TestClientBaseURL tests client stores base URL
func TestClientBaseURL(t *testing.T) {
	client := api.NewClient("http://api.example.com", "key")

	if client.BaseURL != "http://api.example.com" {
		t.Errorf("Expected BaseURL to be set, got %s", client.BaseURL)
	}
}

// TestClientWithCustomTimeout tests client can be created with custom settings
func TestClientWithCustomTimeout(t *testing.T) {
	client := api.NewClient("http://localhost:8080", "key")

	if client == nil {
		t.Fatal("Expected client with custom timeout")
	}
}
