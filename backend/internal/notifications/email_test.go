package notifications

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestEmailChannel_Validate tests EmailChannel config validation
func TestEmailChannel_Validate(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second)

	tests := []struct {
		name    string
		config  ChannelConfig
		wantErr bool
	}{
		{
			name: "valid config with recipients",
			config: ChannelConfig{
				Config: json.RawMessage(`{"recipients":["test@example.com","admin@example.com"]}`),
			},
			wantErr: false,
		},
		{
			name: "missing recipients",
			config: ChannelConfig{
				Config: json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "empty recipients array",
			config: ChannelConfig{
				Config: json.RawMessage(`{"recipients":[]}`),
			},
			wantErr: true,
		},
		{
			name: "invalid JSON",
			config: ChannelConfig{
				Config: json.RawMessage(`{invalid json}`),
			},
			wantErr: true,
		},
		{
			name: "config with optional fields",
			config: ChannelConfig{
				Config: json.RawMessage(`{"recipients":["test@example.com"],"smtp_url":"mail.example.com:587","from":"alerts@example.com"}`),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := channel.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailChannel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestEmailChannel_Test_MissingSMTP tests SMTP config validation in Test method
func TestEmailChannel_Test_MissingSMTP(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second)

	// Create valid email config
	emailConfig := EmailConfig{
		Recipients: []string{"test@example.com"},
	}
	configJSON, _ := json.Marshal(emailConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "email",
		Config: configJSON,
	}

	// Clear all SMTP environment variables
	originalHost := os.Getenv("SMTP_HOST")
	originalUser := os.Getenv("SMTP_USER")
	originalPassword := os.Getenv("SMTP_PASSWORD")
	originalFrom := os.Getenv("SMTP_FROM")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("SMTP_FROM")
	defer func() {
		if originalHost != "" {
			os.Setenv("SMTP_HOST", originalHost)
		}
		if originalUser != "" {
			os.Setenv("SMTP_USER", originalUser)
		}
		if originalPassword != "" {
			os.Setenv("SMTP_PASSWORD", originalPassword)
		}
		if originalFrom != "" {
			os.Setenv("SMTP_FROM", originalFrom)
		}
	}()

	// Test should fail with missing SMTP_HOST
	err := channel.Test(context.Background(), config)
	if err == nil {
		t.Error("Expected error for missing SMTP_HOST, got nil")
	}
	if err != nil && err.Error() != "SMTP_HOST not configured" {
		t.Errorf("Expected 'SMTP_HOST not configured' error, got: %v", err)
	}

	// Set SMTP_HOST but miss other vars
	os.Setenv("SMTP_HOST", "mail.example.com")
	err = channel.Test(context.Background(), config)
	if err == nil {
		t.Error("Expected error for missing SMTP_USER, got nil")
	}

	// Set SMTP_USER
	os.Setenv("SMTP_USER", "user@example.com")
	err = channel.Test(context.Background(), config)
	if err == nil {
		t.Error("Expected error for missing SMTP_PASSWORD, got nil")
	}

	// Set SMTP_PASSWORD
	os.Setenv("SMTP_PASSWORD", "password123")
	err = channel.Test(context.Background(), config)
	if err == nil {
		t.Error("Expected error for missing SMTP_FROM, got nil")
	}

	// Set SMTP_FROM - now should pass
	os.Setenv("SMTP_FROM", "alerts@example.com")
	err = channel.Test(context.Background(), config)
	if err != nil {
		t.Errorf("Expected no error with all SMTP vars set, got: %v", err)
	}
}

// TestEmailChannel_Send_CircuitBreakerOpen tests circuit breaker behavior
func TestEmailChannel_Send_CircuitBreakerOpen(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second).(*EmailChannel)

	// Open the circuit breaker manually
	for i := 0; i < 5; i++ {
		channel.circuitBreaker.RecordFailure()
	}

	if !channel.circuitBreaker.IsOpen() {
		t.Error("Circuit breaker should be open")
	}

	// Create valid config
	emailConfig := EmailConfig{
		Recipients: []string{"test@example.com"},
	}
	configJSON, _ := json.Marshal(emailConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "email",
		Config: configJSON,
	}

	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test Description",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Send should return circuit open error without attempting SMTP
	result, err := channel.Send(context.Background(), alert, config)
	if err != nil {
		t.Errorf("Expected no error from Send, got: %v", err)
	}
	if result.Success {
		t.Error("Expected failure due to circuit breaker being open")
	}
	if result.ErrorMsg == "" || !contains(result.ErrorMsg, "circuit open") {
		t.Errorf("Expected circuit open error message, got: %s", result.ErrorMsg)
	}
}

// TestEmailChannel_Send_MissingRecipients tests recipient validation
func TestEmailChannel_Send_MissingRecipients(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second).(*EmailChannel)

	// Create config with no recipients
	emailConfig := EmailConfig{
		Recipients: []string{},
	}
	configJSON, _ := json.Marshal(emailConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "email",
		Config: configJSON,
	}

	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test Description",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Send should return error for missing recipients
	result, err := channel.Send(context.Background(), alert, config)
	if err != nil {
		t.Errorf("Expected no error from Send, got: %v", err)
	}
	if result.Success {
		t.Error("Expected failure due to missing recipients")
	}
	if result.ErrorMsg == "" {
		t.Error("Expected error message for missing recipients")
	}
}

// TestEmailChannel_Send_MissingSMTPConfig tests SMTP configuration validation
func TestEmailChannel_Send_MissingSMTPConfig(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second).(*EmailChannel)

	// Clear all SMTP environment variables
	originalHost := os.Getenv("SMTP_HOST")
	originalUser := os.Getenv("SMTP_USER")
	originalPassword := os.Getenv("SMTP_PASSWORD")
	originalFrom := os.Getenv("SMTP_FROM")
	os.Unsetenv("SMTP_HOST")
	os.Unsetenv("SMTP_USER")
	os.Unsetenv("SMTP_PASSWORD")
	os.Unsetenv("SMTP_FROM")
	defer func() {
		if originalHost != "" {
			os.Setenv("SMTP_HOST", originalHost)
		}
		if originalUser != "" {
			os.Setenv("SMTP_USER", originalUser)
		}
		if originalPassword != "" {
			os.Setenv("SMTP_PASSWORD", originalPassword)
		}
		if originalFrom != "" {
			os.Setenv("SMTP_FROM", originalFrom)
		}
	}()

	// Create valid config
	emailConfig := EmailConfig{
		Recipients: []string{"test@example.com"},
	}
	configJSON, _ := json.Marshal(emailConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "email",
		Config: configJSON,
	}

	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test Description",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Send should return error for missing SMTP config
	result, err := channel.Send(context.Background(), alert, config)
	if err != nil {
		t.Errorf("Expected no error from Send, got: %v", err)
	}
	if result.Success {
		t.Error("Expected failure due to missing SMTP config")
	}
	if result.ErrorMsg != "SMTP not configured" {
		t.Errorf("Expected 'SMTP not configured' error, got: %s", result.ErrorMsg)
	}
}

// TestEmailChannel_Type tests channel type
func TestEmailChannel_Type(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second)

	if channel.Type() != "email" {
		t.Errorf("Expected channel type 'email', got: %s", channel.Type())
	}
}

// TestEmailConfig_OptionalFields tests that optional fields are properly parsed
func TestEmailConfig_OptionalFields(t *testing.T) {
	tests := []struct {
		name          string
		configJSON    string
		wantSMTPURL   string
		wantFrom      string
		wantRecipients int
	}{
		{
			name:          "basic config",
			configJSON:    `{"recipients":["test@example.com"]}`,
			wantSMTPURL:   "",
			wantFrom:      "",
			wantRecipients: 1,
		},
		{
			name:          "with optional smtp_url",
			configJSON:    `{"recipients":["test@example.com"],"smtp_url":"mail.custom.com:587"}`,
			wantSMTPURL:   "mail.custom.com:587",
			wantFrom:      "",
			wantRecipients: 1,
		},
		{
			name:          "with optional from",
			configJSON:    `{"recipients":["test@example.com"],"from":"custom@example.com"}`,
			wantSMTPURL:   "",
			wantFrom:      "custom@example.com",
			wantRecipients: 1,
		},
		{
			name:          "with all optional fields",
			configJSON:    `{"recipients":["test@example.com","admin@example.com"],"smtp_url":"mail.custom.com:587","from":"custom@example.com"}`,
			wantSMTPURL:   "mail.custom.com:587",
			wantFrom:      "custom@example.com",
			wantRecipients: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config EmailConfig
			if err := json.Unmarshal(json.RawMessage(tt.configJSON), &config); err != nil {
				t.Fatalf("Failed to parse config: %v", err)
			}

			if config.SMTPURL != tt.wantSMTPURL {
				t.Errorf("SMTPURL = %q, want %q", config.SMTPURL, tt.wantSMTPURL)
			}
			if config.From != tt.wantFrom {
				t.Errorf("From = %q, want %q", config.From, tt.wantFrom)
			}
			if len(config.Recipients) != tt.wantRecipients {
				t.Errorf("Recipients count = %d, want %d", len(config.Recipients), tt.wantRecipients)
			}
		})
	}
}

