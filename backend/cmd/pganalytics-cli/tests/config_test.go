package tests

import (
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
