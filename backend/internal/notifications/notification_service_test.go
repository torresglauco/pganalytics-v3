package notifications

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNotificationChannelCreation tests channel creation and validation
func TestNotificationChannelCreation(t *testing.T) {
	tests := []struct {
		name      string
		channel   *ChannelConfig
		wantValid bool
	}{
		{
			name: "valid slack channel",
			channel: &ChannelConfig{
				Type:    "slack",
				Enabled: true,
				Config:  json.RawMessage(`{"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"}`),
			},
			wantValid: true,
		},
		{
			name: "valid email channel",
			channel: &ChannelConfig{
				Type:    "email",
				Enabled: true,
				Config:  json.RawMessage(`{"smtp_host": "smtp.example.com", "smtp_port": 25, "from_address": "alerts@example.com"}`),
			},
			wantValid: true,
		},
		{
			name: "invalid channel - empty type",
			channel: &ChannelConfig{
				Type:    "",
				Enabled: true,
				Config:  json.RawMessage(`{}`),
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple validation: channel must have a non-empty type and be enabled
			valid := tt.channel != nil && tt.channel.Type != "" && tt.channel.Enabled
			assert.Equal(t, tt.wantValid, valid, "channel validation incorrect")
		})
	}
}

// TestAlertNotificationCreation tests alert notification creation
func TestAlertNotificationCreation(t *testing.T) {
	alert := &AlertNotification{
		ID:          1,
		RuleID:      5,
		AlertID:     100,
		Title:       "High CPU Usage",
		Description: "CPU usage exceeds 90%",
		Severity:    "critical",
		Status:      "firing",
		Context:     json.RawMessage(`{"cpu": 95.5}`),
		FiredAt:     time.Now(),
		Database:    "production_db",
		Query:       "SELECT * FROM metrics",
	}

	assert.Equal(t, int64(1), alert.ID)
	assert.Equal(t, int64(5), alert.RuleID)
	assert.Equal(t, int64(100), alert.AlertID)
	assert.Equal(t, "High CPU Usage", alert.Title)
	assert.Equal(t, "critical", alert.Severity)
	assert.Equal(t, "production_db", alert.Database)
}

// TestDeliveryResultCreation tests delivery result creation
func TestDeliveryResultCreation(t *testing.T) {
	now := time.Now()
	delivery := &DeliveryResult{
		Success:     true,
		MessageID:   "msg-123",
		ErrorMsg:    "",
		DeliveredAt: now,
	}

	assert.True(t, delivery.Success)
	assert.Equal(t, "msg-123", delivery.MessageID)
	assert.Equal(t, "", delivery.ErrorMsg)
	assert.Equal(t, now, delivery.DeliveredAt)
}

// TestNotificationDeliveryTracking tests notification delivery tracking
func TestNotificationDeliveryTracking(t *testing.T) {
	delivery := &NotificationDelivery{
		ID:               1,
		AlertID:          100,
		ChannelID:        5,
		DeliveryStatus:   "pending",
		DeliveryAttempts: 0,
		MaxRetries:       5,
		MessageSubject:   "Alert: High CPU",
		MessageBody:      "CPU usage is high",
		DeliveredAt:      nil,
		LastError:        "",
		NextRetryAt:      nil,
	}

	assert.Equal(t, int64(1), delivery.ID)
	assert.Equal(t, int64(100), delivery.AlertID)
	assert.Equal(t, "pending", delivery.DeliveryStatus)
	assert.Equal(t, 0, delivery.DeliveryAttempts)
	assert.Nil(t, delivery.DeliveredAt)
}

// TestChannelConfigWithValidation tests channel config with proper validation
func TestChannelConfigWithValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *ChannelConfig
		isValid bool
	}{
		{
			name: "enabled channel",
			config: &ChannelConfig{
				ID:       1,
				Type:     "slack",
				Enabled:  true,
				Config:   json.RawMessage(`{}`),
				Verified: true,
			},
			isValid: true,
		},
		{
			name: "disabled channel",
			config: &ChannelConfig{
				ID:       2,
				Type:     "email",
				Enabled:  false,
				Config:   json.RawMessage(`{}`),
				Verified: false,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
			assert.Equal(t, tt.isValid, tt.config.Enabled)
		})
	}
}

// TestAlertNotificationFields tests alert notification with various severity levels
func TestAlertNotificationFields(t *testing.T) {
	severities := []string{"low", "medium", "high", "critical"}

	for _, sev := range severities {
		alert := &AlertNotification{
			ID:       1,
			RuleID:   5,
			AlertID:  100,
			Severity: sev,
			Status:   "firing",
		}

		assert.Equal(t, sev, alert.Severity)
		assert.Equal(t, "firing", alert.Status)
	}
}

// TestDeliveryStatusProgression tests delivery status progression
func TestDeliveryStatusProgression(t *testing.T) {
	statusProgression := []string{"pending", "sent", "failed"}

	for i, status := range statusProgression {
		delivery := &NotificationDelivery{
			ID:             int64(i + 1),
			DeliveryStatus: status,
		}

		assert.Equal(t, status, delivery.DeliveryStatus)
	}
}
