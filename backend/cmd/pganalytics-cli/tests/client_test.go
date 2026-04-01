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
