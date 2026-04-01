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
