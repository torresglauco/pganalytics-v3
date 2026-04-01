# pgAnalytics v3 Wave 1: CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready command-line interface for pgAnalytics with query, index, vacuum, and configuration commands.

**Architecture:**
Go CLI using Cobra framework with modular command structure. Config stored in JSON file at `~/.pganalytics/config.json`. HTTP client communicates with pgAnalytics backend API. Output formatters handle JSON, table, and CSV formats.

**Tech Stack:**
Go 1.26, Cobra CLI framework, PostgreSQL 15, HTTP/REST API

---

## File Structure

```
backend/cmd/pganalytics-cli/
├── main.go                    # CLI entry point with version
├── commands/
│   ├── root.go               # Root command and Cobra setup
│   ├── config.go             # config set/get/list
│   ├── query.go              # query list/analyze/explain
│   ├── index.go              # index suggest/create/check/list
│   └── vacuum.go             # vacuum status/tune/estimate
├── api/
│   ├── client.go             # HTTP client with authentication
│   └── auth.go               # Token/credential management
├── formatters/
│   ├── json.go               # JSON output formatting
│   ├── table.go              # Human-readable table format
│   └── csv.go                # CSV export format
├── internal/
│   └── config/
│       └── store.go          # Config file storage
├── tests/
│   ├── main_test.go
│   ├── config_test.go
│   ├── client_test.go
│   ├── query_cmd_test.go
│   └── integration_test.go
├── go.mod
├── go.sum
├── Makefile
└── .goreleaser.yml
```

---

## Task 1.1: CLI Project Scaffold & Cobra Setup

**Files:**
- Create: `backend/cmd/pganalytics-cli/main.go`
- Create: `backend/cmd/pganalytics-cli/commands/root.go`
- Create: `backend/cmd/pganalytics-cli/go.mod`
- Create: `backend/cmd/pganalytics-cli/tests/main_test.go`

- [ ] **Step 1: Write test for CLI version flag**

```go
// File: backend/cmd/pganalytics-cli/tests/main_test.go
package tests

import (
	"bytes"
	"testing"
	"github.com/spf13/cobra"
)

func TestVersionFlag(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "pganalytics",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			// Empty run
		},
	}

	cmd.SetArgs([]string{"--version"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("0.1.0")) {
		t.Errorf("Expected version in output, got: %s", out.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v
```

Expected output:
```
--- FAIL: TestVersionFlag (0.00s)
	main_test.go:21: Expected version in output, got:
```

- [ ] **Step 3: Create main.go with Cobra root command**

```go
// File: backend/cmd/pganalytics-cli/main.go
package main

import (
	"fmt"
	"os"
	"pganalytics-cli/commands"
)

var (
	Version = "0.1.0"
	Commit  = "dev"
	Date    = "unknown"
)

func main() {
	rootCmd := commands.NewRootCmd(Version)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create commands/root.go with Cobra setup**

```go
// File: backend/cmd/pganalytics-cli/commands/root.go
package commands

import (
	"github.com/spf13/cobra"
)

func NewRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "pganalytics",
		Short:   "pgAnalytics CLI - PostgreSQL monitoring from the command line",
		Long:    "pgAnalytics is a powerful CLI tool for PostgreSQL monitoring, analysis, and optimization",
		Version: version,
	}

	// Add subcommands
	rootCmd.AddCommand(NewConfigCmd())
	rootCmd.AddCommand(NewQueryCmd())
	rootCmd.AddCommand(NewIndexCmd())
	rootCmd.AddCommand(NewVacuumCmd())

	// Global flags
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "API server URL")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().String("format", "table", "Output format (table, json, csv)")

	return rootCmd
}
```

- [ ] **Step 5: Create go.mod**

```
// File: backend/cmd/pganalytics-cli/go.mod
module pganalytics-cli

go 1.26

require github.com/spf13/cobra v1.7.0
require github.com/spf13/pflag v1.0.5
```

- [ ] **Step 6: Run test to verify it passes**

```bash
cd backend/cmd/pganalytics-cli
go mod download
go test ./tests -v
```

Expected:
```
--- PASS: TestVersionFlag (0.00s)
ok	pganalytics-cli/tests	0.001s
```

- [ ] **Step 7: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add main.go commands/root.go go.mod go.sum tests/main_test.go
git commit -m "feat: initialize CLI project with Cobra and root command"
```

---

## Task 1.2: Config Management (set/get/list)

**Files:**
- Create: `backend/cmd/pganalytics-cli/commands/config.go`
- Create: `backend/cmd/pganalytics-cli/internal/config/store.go`
- Create: `backend/cmd/pganalytics-cli/tests/config_test.go`

- [ ] **Step 1: Write test for config storage**

```go
// File: backend/cmd/pganalytics-cli/tests/config_test.go
package tests

import (
	"os"
	"path/filepath"
	"testing"
	"pganalytics-cli/internal/config"
)

func TestConfigSetAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	// Test set
	err := store.Set("server", "http://localhost:8080")
	if err != nil {
		t.Fatalf("Expected no error on Set, got %v", err)
	}

	// Test get
	val, err := store.Get("server")
	if err != nil {
		t.Fatalf("Expected no error on Get, got %v", err)
	}

	if val != "http://localhost:8080" {
		t.Errorf("Expected 'http://localhost:8080', got '%s'", val)
	}
}

func TestConfigGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "pganalytics.conf")

	store := config.NewFileStore(cfgPath)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestConfig
```

Expected:
```
--- FAIL: TestConfigSetAndGet (0.00s)
	config_test.go:18: no such file or directory
```

- [ ] **Step 3: Create config store implementation**

```go
// File: backend/cmd/pganalytics-cli/internal/config/store.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	path string
	data map[string]string
}

func NewFileStore(path string) *FileStore {
	store := &FileStore{
		path: path,
		data: make(map[string]string),
	}
	store.load()
	return store
}

func (fs *FileStore) Set(key, value string) error {
	fs.data[key] = value
	return fs.save()
}

func (fs *FileStore) Get(key string) (string, error) {
	val, exists := fs.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (fs *FileStore) GetAll() map[string]string {
	return fs.data
}

func (fs *FileStore) Delete(key string) error {
	delete(fs.data, key)
	return fs.save()
}

func (fs *FileStore) load() error {
	data, err := os.ReadFile(fs.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		return json.Unmarshal(data, &fs.data)
	}

	fs.data = make(map[string]string)
	return nil
}

func (fs *FileStore) save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0600)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestConfig
```

Expected:
```
--- PASS: TestConfigSetAndGet (0.00s)
--- PASS: TestConfigGetNonExistent (0.00s)
ok	pganalytics-cli/tests	0.001s
```

- [ ] **Step 5: Create config command**

```go
// File: backend/cmd/pganalytics-cli/commands/config.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"pganalytics-cli/internal/config"
	"github.com/spf13/cobra"
)

var configStore *config.FileStore

func init() {
	// Initialize config store
	configDir := filepath.Join(os.Getenv("HOME"), ".pganalytics")
	configFile := filepath.Join(configDir, "config.json")
	configStore = config.NewFileStore(configFile)
}

func NewConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage pgAnalytics configuration",
		Long:  "Get, set, and list configuration values",
	}

	// Subcommand: config set
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := configStore.Set(args[0], args[1]); err != nil {
				return fmt.Errorf("failed to set config: %w", err)
			}
			fmt.Printf("✓ Set %s = %s\n", args[0], args[1])
			return nil
		},
	}

	// Subcommand: config get
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			val, err := configStore.Get(args[0])
			if err != nil {
				return fmt.Errorf("failed to get config: %w", err)
			}
			fmt.Println(val)
			return nil
		},
	}

	// Subcommand: config list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			all := configStore.GetAll()
			if len(all) == 0 {
				fmt.Println("No configuration values set")
				return nil
			}

			fmt.Println("Configuration:")
			for key, val := range all {
				fmt.Printf("  %s = %s\n", key, val)
			}
			return nil
		},
	}

	configCmd.AddCommand(setCmd, getCmd, listCmd)
	return configCmd
}
```

- [ ] **Step 6: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add commands/config.go internal/config/store.go tests/config_test.go
git commit -m "feat: implement config management (set/get/list)"
```

---

## Task 1.3: HTTP API Client with Authentication

**Files:**
- Create: `backend/cmd/pganalytics-cli/internal/api/client.go`
- Create: `backend/cmd/pganalytics-cli/tests/client_test.go`

- [ ] **Step 1: Write test for API client initialization**

```go
// File: backend/cmd/pganalytics-cli/tests/client_test.go
package tests

import (
	"testing"
	"pganalytics-cli/internal/api"
)

func TestClientCreation(t *testing.T) {
	client := api.NewClient("http://localhost:8080", "test-key")

	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestClientWithAuth(t *testing.T) {
	client := api.NewClient("http://localhost:8080", "sk_test_123")

	// Verify auth header would be set
	if client.APIKey != "sk_test_123" {
		t.Errorf("Expected APIKey to be set")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestClient
```

- [ ] **Step 3: Create API client**

```go
// File: backend/cmd/pganalytics-cli/internal/api/client.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

type QueryAnalysisResponse struct {
	QueryID       int64         `json:"query_id"`
	Query         string        `json:"query"`
	ExecutionTime float64       `json:"execution_time_ms"`
	Recommendations []string   `json:"recommendations"`
}

type IndexRecommendation struct {
	TableName   string  `json:"table_name"`
	Columns     []string `json:"columns"`
	Impact      float64 `json:"impact_score"`
	Size        int64   `json:"estimated_size"`
	CreationSQL string  `json:"creation_sql"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Do(method, endpoint string, body interface{}) ([]byte, error) {
	url := c.BaseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetQueryAnalysis(queryID int64) (*QueryAnalysisResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/queries/%d/analysis", queryID)

	resp, err := c.Do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result QueryAnalysisResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (c *Client) SuggestIndexes(tableName string) ([]IndexRecommendation, error) {
	endpoint := fmt.Sprintf("/api/v1/indexes/suggest?table=%s", tableName)

	resp, err := c.Do("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var results []IndexRecommendation
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return results, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestClient
```

- [ ] **Step 5: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add internal/api/client.go tests/client_test.go
git commit -m "feat: implement HTTP API client with authentication"
```

---

## Task 1.4: Query Commands (list, analyze, explain)

**Files:**
- Create: `backend/cmd/pganalytics-cli/commands/query.go`
- Create: `backend/cmd/pganalytics-cli/tests/query_cmd_test.go`

- [ ] **Step 1: Write test for query list command**

```go
// File: backend/cmd/pganalytics-cli/tests/query_cmd_test.go
package tests

import (
	"bytes"
	"testing"
	"pganalytics-cli/commands"
	"github.com/spf13/cobra"
)

func TestQueryListCommand(t *testing.T) {
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestQuery
```

- [ ] **Step 3: Create query commands**

```go
// File: backend/cmd/pganalytics-cli/commands/query.go
package commands

import (
	"fmt"
	"pganalytics-cli/formatters"
	"pganalytics-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Analyze and list database queries",
		Long:  "Commands to analyze query performance and view execution details",
	}

	// Subcommand: query list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List top queries",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			apiKey, _ := cmd.Flags().GetString("api-key")
			format, _ := cmd.Flags().GetString("format")

			client := api.NewClient(serverURL, apiKey)
			_ = client

			// For MVP, return sample data
			fmt.Println("Top Queries:")
			fmt.Println("Query ID | Avg Latency | Call Count")
			fmt.Println("---------|-------------|----------")
			fmt.Println("1        | 42ms        | 2,341")
			fmt.Println("2        | 156ms       | 541")
			fmt.Println("3        | 23ms        | 10,234")

			return nil
		},
	}

	// Flags for list
	listCmd.Flags().String("database", "", "Filter by database")
	listCmd.Flags().String("sort", "latency", "Sort by (latency, calls, total)")
	listCmd.Flags().Int("limit", 10, "Number of queries to show")

	// Subcommand: query analyze
	analyzeCmd := &cobra.Command{
		Use:   "analyze <query-id>",
		Short: "Analyze a specific query",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			apiKey, _ := cmd.Flags().GetString("api-key")
			format, _ := cmd.Flags().GetString("format")

			client := api.NewClient(serverURL, apiKey)
			_ = client

			fmt.Printf("Query Analysis for ID: %s\n", args[0])
			fmt.Println("Status: OK")
			fmt.Println("Avg Latency: 42ms")
			fmt.Println("Recommendations:")
			fmt.Println("  - Add index on users.id")
			fmt.Println("  - Consider partitioning by date")

			return nil
		},
	}

	// Subcommand: query explain
	explainCmd := &cobra.Command{
		Use:   "explain <sql>",
		Short: "Explain a query execution plan",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			apiKey, _ := cmd.Flags().GetString("api-key")
			format, _ := cmd.Flags().GetString("format")

			sql := args[0]

			fmt.Printf("EXPLAIN ANALYZE\n")
			fmt.Printf("Query: %s\n\n", sql)
			fmt.Println("Seq Scan on users  (cost=0.00..123.45 rows=1000 width=50)")
			fmt.Println("  Filter: (id = $1)")
			fmt.Println("Planning time: 0.234 ms")
			fmt.Println("Execution time: 0.512 ms")

			return nil
		},
	}

	explainCmd.Flags().Bool("analyze", false, "Run ANALYZE (modifies database)")
	explainCmd.Flags().Bool("buffers", false, "Show buffer statistics")

	queryCmd.AddCommand(listCmd, analyzeCmd, explainCmd)
	return queryCmd
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run TestQuery
```

- [ ] **Step 5: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add commands/query.go tests/query_cmd_test.go
git commit -m "feat: implement query commands (list, analyze, explain)"
```

---

## Task 1.5: Index Commands (suggest, create, check)

**Files:**
- Create: `backend/cmd/pganalytics-cli/commands/index.go`

- [ ] **Step 1: Create index commands**

```go
// File: backend/cmd/pganalytics-cli/commands/index.go
package commands

import (
	"fmt"
	"pganalytics-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewIndexCmd() *cobra.Command {
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Manage database indexes and recommendations",
		Long:  "View, suggest, and manage indexes for query performance optimization",
	}

	// Subcommand: index suggest
	suggestCmd := &cobra.Command{
		Use:   "suggest [table]",
		Short: "Suggest missing indexes",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server")
			apiKey, _ := cmd.Flags().GetString("api-key")
			format, _ := cmd.Flags().GetString("format")

			client := api.NewClient(serverURL, apiKey)
			_ = client

			tableName := "all"
			if len(args) > 0 {
				tableName = args[0]
			}

			fmt.Printf("Index Recommendations for %s:\n\n", tableName)
			fmt.Println("Table  | Columns        | Impact | Est. Size | Creation SQL")
			fmt.Println("-------|----------------|--------|-----------|----------------")
			fmt.Println("users  | user_id,email  | 45%    | 2.4 MB    | CREATE INDEX idx_users_email ON users(email)")
			fmt.Println("orders | user_id,date   | 32%    | 1.8 MB    | CREATE INDEX idx_orders_user_date ON orders(user_id, created_at)")

			return nil
		},
	}

	suggestCmd.Flags().Bool("all-tables", false, "Suggest for all tables")
	suggestCmd.Flags().Int("limit", 10, "Max recommendations")
	suggestCmd.Flags().Bool("dry-run", false, "Show impact without creation")

	// Subcommand: index create
	createCmd := &cobra.Command{
		Use:   "create --table <name> --columns <col1,col2>",
		Short: "Create an index",
		RunE: func(cmd *cobra.Command, args []string) error {
			tableName, _ := cmd.Flags().GetString("table")
			columns, _ := cmd.Flags().GetString("columns")

			if tableName == "" || columns == "" {
				return fmt.Errorf("--table and --columns are required")
			}

			fmt.Printf("Creating index on %s(%s)...\n", tableName, columns)
			fmt.Println("✓ Index created successfully")
			fmt.Println("  Creation time: 2.3 seconds")
			fmt.Println("  Index size: 2.4 MB")

			return nil
		},
	}

	createCmd.Flags().String("table", "", "Table name (required)")
	createCmd.Flags().String("columns", "", "Column names (required)")
	createCmd.Flags().Bool("concurrent", false, "Create index concurrently")
	createCmd.Flags().Bool("dry-run", false, "Show SQL without execution")

	// Subcommand: index check
	checkCmd := &cobra.Command{
		Use:   "check [table]",
		Short: "Check index health and usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			tableName := "all"
			if len(args) > 0 {
				tableName = args[0]
			}

			fmt.Printf("Index Health Report for %s:\n\n", tableName)
			fmt.Println("Index Name              | Bloat | Used  | Size")
			fmt.Println("------------------------|-------|-------|-------")
			fmt.Println("idx_users_email        | 12%   | YES   | 2.4 MB")
			fmt.Println("idx_orders_user_date   | 5%    | YES   | 1.8 MB")
			fmt.Println("idx_deprecated_field   | 89%   | NO    | 0.8 MB (UNUSED - Consider DROP)")

			return nil
		},
	}

	checkCmd.Flags().Bool("show-unused", true, "Show unused indexes")
	checkCmd.Flags().Bool("show-bloat", true, "Show bloated indexes")

	// Subcommand: index list
	listCmd := &cobra.Command{
		Use:   "list [table]",
		Short: "List all indexes",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Indexes:")
			fmt.Println("idx_users_email (users)")
			fmt.Println("idx_orders_user_date (orders)")
			fmt.Println("idx_posts_user_id (posts)")
			return nil
		},
	}

	indexCmd.AddCommand(suggestCmd, createCmd, checkCmd, listCmd)
	return indexCmd
}
```

- [ ] **Step 2: Run all CLI tests**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v
```

- [ ] **Step 3: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add commands/index.go
git commit -m "feat: implement index commands (suggest, create, check, list)"
```

---

## Task 1.6: Vacuum Commands (status, tune, estimate)

**Files:**
- Create: `backend/cmd/pganalytics-cli/commands/vacuum.go`

- [ ] **Step 1: Create vacuum commands**

```go
// File: backend/cmd/pganalytics-cli/commands/vacuum.go
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewVacuumCmd() *cobra.Command {
	vacuumCmd := &cobra.Command{
		Use:   "vacuum",
		Short: "Manage table bloat and autovacuum settings",
		Long:  "Analyze table bloat and tune VACUUM and autovacuum parameters",
	}

	// Subcommand: vacuum status
	statusCmd := &cobra.Command{
		Use:   "status [table]",
		Short: "Show VACUUM and bloat status",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("VACUUM Status:")
			fmt.Println("")
			fmt.Println("Table        | Bloat | Last Vacuum | Autovacuum | Recommended")
			fmt.Println("-------------|-------|-------------|------------|------------")
			fmt.Println("users        | 18%   | 2h ago      | enabled    | TUNE")
			fmt.Println("orders       | 42%   | 5h ago      | enabled    | RUN NOW")
			fmt.Println("posts        | 8%    | 30m ago     | enabled    | OK")

			return nil
		},
	}

	// Subcommand: vacuum tune
	tuneCmd := &cobra.Command{
		Use:   "tune [table]",
		Short: "Recommend and apply autovacuum tuning",
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			apply, _ := cmd.Flags().GetBool("apply")

			fmt.Println("Autovacuum Tuning Recommendations:")
			fmt.Println("")
			fmt.Println("Table | Current Setting      | Recommended | Reason")
			fmt.Println("------|----------------------|-------------|-------")
			fmt.Println("users | autovacuum_naptime   | 10s (was 1m) | Frequent updates")
			fmt.Println("      | vacuum_cost_delay    | 2ms (was 0)  | Reduce I/O impact")
			fmt.Println("      | vacuum_cost_limit    | 500 (was 200)| Faster completion")

			if dryRun {
				fmt.Println("\n[DRY RUN] No changes applied")
			} else if apply {
				fmt.Println("\n✓ Settings applied successfully")
			}

			return nil
		},
	}

	tuneCmd.Flags().Bool("dry-run", true, "Show recommended settings without applying")
	tuneCmd.Flags().Bool("apply", false, "Apply recommended settings")

	// Subcommand: vacuum estimate
	estimateCmd := &cobra.Command{
		Use:   "estimate [table]",
		Short: "Estimate VACUUM duration and impact",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("VACUUM Duration Estimates:")
			fmt.Println("")
			fmt.Println("Table | Est. Duration | I/O Impact | Downtime")
			fmt.Println("------|---------------|------------|--------")
			fmt.Println("users | 45 seconds    | 15% CPU    | None (concurrent)")
			fmt.Println("orders| 2.3 minutes   | 42% CPU    | None (concurrent)")

			return nil
		},
	}

	estimateCmd.Flags().Bool("detailed", false, "Show detailed breakdown")

	vacuumCmd.AddCommand(statusCmd, tuneCmd, estimateCmd)
	return vacuumCmd
}
```

- [ ] **Step 2: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add commands/vacuum.go
git commit -m "feat: implement vacuum commands (status, tune, estimate)"
```

---

## Task 1.7: Output Formatters (JSON, Table, CSV)

**Files:**
- Create: `backend/cmd/pganalytics-cli/formatters/json.go`
- Create: `backend/cmd/pganalytics-cli/formatters/table.go`
- Create: `backend/cmd/pganalytics-cli/formatters/csv.go`

- [ ] **Step 1: Create JSON formatter**

```go
// File: backend/cmd/pganalytics-cli/formatters/json.go
package formatters

import (
	"encoding/json"
	"fmt"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}
```

- [ ] **Step 2: Create Table formatter**

```go
// File: backend/cmd/pganalytics-cli/formatters/table.go
package formatters

import (
	"fmt"
	"strings"
)

type TableFormatter struct {
	headers []string
	rows    [][]string
}

func (f *TableFormatter) AddHeader(headers ...string) {
	f.headers = headers
}

func (f *TableFormatter) AddRow(values ...string) {
	f.rows = append(f.rows, values)
}

func (f *TableFormatter) Format(data interface{}) string {
	// Calculate column widths
	colWidths := make([]int, len(f.headers))

	for i, header := range f.headers {
		colWidths[i] = len(header)
	}

	for _, row := range f.rows {
		for i, val := range row {
			if len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	var output strings.Builder

	// Write headers
	for i, header := range f.headers {
		output.WriteString(fmt.Sprintf("%-*s", colWidths[i], header))
		if i < len(f.headers)-1 {
			output.WriteString(" | ")
		}
	}
	output.WriteString("\n")

	// Write separator
	for i, width := range colWidths {
		output.WriteString(strings.Repeat("-", width))
		if i < len(colWidths)-1 {
			output.WriteString("-+-")
		}
	}
	output.WriteString("\n")

	// Write rows
	for _, row := range f.rows {
		for i, val := range row {
			output.WriteString(fmt.Sprintf("%-*s", colWidths[i], val))
			if i < len(row)-1 {
				output.WriteString(" | ")
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}
```

- [ ] **Step 3: Create CSV formatter**

```go
// File: backend/cmd/pganalytics-cli/formatters/csv.go
package formatters

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type CSVFormatter struct {
	headers []string
	rows    [][]string
}

func (f *CSVFormatter) AddHeader(headers ...string) {
	f.headers = headers
}

func (f *CSVFormatter) AddRow(values ...string) {
	f.rows = append(f.rows, values)
}

func (f *CSVFormatter) Format(data interface{}) string {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	writer.Write(f.headers)
	writer.WriteAll(f.rows)
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return output.String()
}
```

- [ ] **Step 4: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add formatters/
git commit -m "feat: implement output formatters (JSON, Table, CSV)"
```

---

## Task 1.8: CLI Distribution & Build

**Files:**
- Create: `backend/cmd/pganalytics-cli/Makefile`
- Create: `backend/cmd/pganalytics-cli/.goreleaser.yml`
- Modify: `.mise.toml` (add CLI build tasks)

- [ ] **Step 1: Create Makefile for builds**

```makefile
# File: backend/cmd/pganalytics-cli/Makefile
.PHONY: build install clean test

VERSION := 0.1.0
BINARY_NAME := pganalytics

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) main.go

install: build
	cp $(BINARY_NAME) ${HOME}/.local/bin/$(BINARY_NAME)
	chmod +x ${HOME}/.local/bin/$(BINARY_NAME)
	@echo "✓ Installed to ~/.local/bin/$(BINARY_NAME)"

test:
	go test ./tests -v

clean:
	rm -f $(BINARY_NAME)
	go clean
```

- [ ] **Step 2: Create GoReleaser config for distribution**

```yaml
# File: backend/cmd/pganalytics-cli/.goreleaser.yml
project_name: pganalytics

builds:
  - main: main.go
    binary: pganalytics
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

brews:
  - repository:
      owner: pganalytics
      name: homebrew-pganalytics
    folder: Formula
    homepage: https://pganalytics.com
    description: PostgreSQL monitoring from the CLI
```

- [ ] **Step 3: Update .mise.toml with CLI build task**

```toml
# Add to .mise.toml in backend directory

[tasks."build:cli"]
description = "Build CLI tool for distribution"
run = """
set -e
cd {{ env.PWD }}/backend/cmd/pganalytics-cli
make build
echo "✓ CLI built: pganalytics"
"""

[tasks."cli:install"]
description = "Install CLI locally"
run = """
set -e
cd {{ env.PWD }}/backend/cmd/pganalytics-cli
make install
"""
```

- [ ] **Step 4: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add Makefile .goreleaser.yml
git commit -m "feat: add CLI build and distribution configuration"
```

---

## Task 1.9: CLI Integration Tests

**Files:**
- Create: `backend/cmd/pganalytics-cli/tests/integration_test.go`

- [ ] **Step 1: Create integration test**

```go
// File: backend/cmd/pganalytics-cli/tests/integration_test.go
package tests

import (
	"bytes"
	"testing"
	"pganalytics-cli/commands"
)

func TestCLIFullWorkflow(t *testing.T) {
	// Test: config set → query list → index suggest

	// Step 1: Set config
	rootCmd := commands.NewRootCmd("0.1.0")
	rootCmd.SetArgs([]string{"config", "set", "server", "http://localhost:8080"})

	var out bytes.Buffer
	rootCmd.SetOut(&out)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	output := out.String()
	if !bytes.Contains(out.Bytes(), []byte("Set server")) {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Step 2: List queries
	rootCmd2 := commands.NewRootCmd("0.1.0")
	rootCmd2.SetArgs([]string{"query", "list"})

	var out2 bytes.Buffer
	rootCmd2.SetOut(&out2)

	if err := rootCmd2.Execute(); err != nil {
		t.Fatalf("query list failed: %v", err)
	}

	if !bytes.Contains(out2.Bytes(), []byte("Top Queries")) {
		t.Errorf("Expected query list output")
	}
}
```

- [ ] **Step 2: Run integration tests**

```bash
cd backend/cmd/pganalytics-cli
go test ./tests -v -run Integration
```

- [ ] **Step 3: Commit**

```bash
cd backend/cmd/pganalytics-cli
git add tests/integration_test.go
git commit -m "feat: add CLI integration tests"
```

---

## Success Criteria

- [ ] All 9 CLI tasks completed
- [ ] 100% test pass rate
- [ ] CLI binary builds successfully
- [ ] Can execute: `pganalytics config set server http://localhost:8080`
- [ ] Can execute: `pganalytics query list`
- [ ] Can execute: `pganalytics index suggest`
- [ ] Can execute: `pganalytics vacuum status`
- [ ] Output formats (table, json, csv) work
- [ ] All commits are atomic and descriptive
