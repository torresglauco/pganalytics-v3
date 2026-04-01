package tests

import (
	"testing"
	"pganalytics-cli/commands"
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
