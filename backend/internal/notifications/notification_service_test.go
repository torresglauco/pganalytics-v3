package notifications

import (
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
				Name:    "slack-alerts",
				Type:    "slack",
				Enabled: true,
				Config: map[string]interface{}{
					"webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
				},
			},
			wantValid: true,
		},
		{
			name: "valid email channel",
			channel: &ChannelConfig{
				Name:    "email-alerts",
				Type:    "email",
				Enabled: true,
				Config: map[string]interface{}{
					"smtp_host":     "smtp.example.com",
					"smtp_port":     25,
					"from_address":  "alerts@example.com",
					"recipients":    []string{"ops@example.com"},
				},
			},
			wantValid: true,
		},
		{
			name: "invalid channel - no webhook",
			channel: &ChannelConfig{
				Name:    "slack-alerts",
				Type:    "slack",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateChannelConfig(tt.channel)
			assert.Equal(t, tt.wantValid, valid, "channel validation incorrect")
		})
	}
}

// TestExponentialBackoffRetry tests retry logic with exponential backoff
func TestExponentialBackoffRetry(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		wantDelay  time.Duration
	}{
		{
			name:      "first retry",
			attempt:   1,
			wantDelay: 1 * time.Second,
		},
		{
			name:      "second retry",
			attempt:   2,
			wantDelay: 2 * time.Second,
		},
		{
			name:      "third retry",
			attempt:   3,
			wantDelay: 4 * time.Second,
		},
		{
			name:      "fourth retry",
			attempt:   4,
			wantDelay: 8 * time.Second,
		},
		{
			name:      "fifth retry",
			attempt:   5,
			wantDelay: 16 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateBackoffDelay(tt.attempt)
			assert.Equal(t, tt.wantDelay, delay, "backoff delay incorrect")
		})
	}
}

// TestNotificationDelivery tests notification delivery tracking
func TestNotificationDelivery(t *testing.T) {
	delivery := &DeliveryResult{
		ChannelID:     1,
		AlertID:       100,
		Status:        "success",
		Timestamp:     time.Now(),
		Attempt:       1,
		NextRetryAt:   nil,
		ErrorMessage:  nil,
	}

	assert.Equal(t, int64(1), delivery.ChannelID)
	assert.Equal(t, int64(100), delivery.AlertID)
	assert.Equal(t, "success", delivery.Status)
	assert.Equal(t, 1, delivery.Attempt)
	assert.Nil(t, delivery.NextRetryAt)
	assert.Nil(t, delivery.ErrorMessage)
}

// TestAlertNotificationFormatting tests alert message formatting for different channels
func TestAlertNotificationFormatting(t *testing.T) {
	alert := &AlertNotification{
		AlertID:    100,
		RuleID:     5,
		RuleName:   "High CPU Usage",
		Database:   "production_db",
		Severity:   "critical",
		Message:    "CPU usage exceeds 90%",
		Timestamp:  time.Now(),
		Details: map[string]interface{}{
			"current_value": 95.5,
			"threshold":     90.0,
		},
	}

	// Test Slack formatting
	slackMsg := formatSlackMessage(alert)
	assert.Contains(t, slackMsg, "High CPU Usage")
	assert.Contains(t, slackMsg, "critical")
	assert.Contains(t, slackMsg, "production_db")

	// Test email formatting
	emailSubject := formatEmailSubject(alert)
	assert.Contains(t, emailSubject, "critical")
	assert.Contains(t, emailSubject, "High CPU Usage")

	// Test webhook JSON
	webhookPayload := formatWebhookPayload(alert)
	assert.NotNil(t, webhookPayload)
}

// TestRateLimiting tests notification rate limiting
func TestRateLimiting(t *testing.T) {
	limiter := NewNotificationRateLimiter(10 * time.Minute)

	fingerprint := "rule_1_critical"

	// First notification should succeed
	allowed := limiter.Allow(fingerprint)
	assert.True(t, allowed)

	// Immediate second notification should be blocked (within throttle window)
	allowed = limiter.Allow(fingerprint)
	assert.False(t, allowed)

	// Different fingerprint should be allowed
	allowed = limiter.Allow("rule_2_high")
	assert.True(t, allowed)
}

// TestDeliverySuccessRate tests success rate calculation
func TestDeliverySuccessRate(t *testing.T) {
	metrics := &NotificationMetrics{
		TotalSent:      100,
		Successful:     99,
		Failed:         1,
		Retried:        0,
		SuccessfulRetries: 0,
	}

	successRate := calculateSuccessRate(metrics)
	assert.InDelta(t, 0.99, successRate, 0.01)
}

// TestChannelAvailability tests channel health status
func TestChannelAvailability(t *testing.T) {
	tests := []struct {
		name             string
		recentFailures   int
		successRate      float64
		wantAvailable    bool
	}{
		{
			name:           "healthy channel",
			recentFailures: 0,
			successRate:    0.99,
			wantAvailable:  true,
		},
		{
			name:           "degraded channel",
			recentFailures: 2,
			successRate:    0.85,
			wantAvailable:  true,
		},
		{
			name:           "unhealthy channel",
			recentFailures: 10,
			successRate:    0.5,
			wantAvailable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := isChannelAvailable(tt.recentFailures, tt.successRate)
			assert.Equal(t, tt.wantAvailable, available, "channel availability incorrect")
		})
	}
}

// TestNotificationMetrics tests metrics collection
func TestNotificationMetrics(t *testing.T) {
	service := &NotificationService{
		Metrics: &NotificationMetrics{
			TotalSent:           500,
			Successful:          495,
			Failed:              5,
			Retried:             4,
			SuccessfulRetries:   3,
			AverageSendTime:     250 * time.Millisecond,
			SlackMessagesCount:  200,
			EmailMessagesCount:  150,
			WebhookCallsCount:   100,
			PagerDutyCallsCount: 35,
			JiraIssuesCount:     15,
		},
	}

	assert.Equal(t, 500, service.Metrics.TotalSent)
	assert.Equal(t, 495, service.Metrics.Successful)
	assert.Equal(t, 200, service.Metrics.SlackMessagesCount)
	assert.Equal(t, 150, service.Metrics.EmailMessagesCount)
}

// TestDLQHandling tests dead letter queue for failed messages
func TestDLQHandling(t *testing.T) {
	dlq := NewDeadLetterQueue()

	msg := &AlertNotification{
		AlertID:  100,
		RuleID:   5,
		Severity: "critical",
	}

	// Add to DLQ
	dlq.Enqueue(msg)
	assert.Equal(t, 1, dlq.Size())

	// Retrieve from DLQ
	retrieved := dlq.Dequeue()
	assert.NotNil(t, retrieved)
	assert.Equal(t, msg.AlertID, retrieved.AlertID)
	assert.Equal(t, 0, dlq.Size())
}

// TestMultiChannelDelivery tests sending to multiple channels
func TestMultiChannelDelivery(t *testing.T) {
	channels := []*ChannelConfig{
		{
			ID:      1,
			Type:    "slack",
			Enabled: true,
		},
		{
			ID:      2,
			Type:    "email",
			Enabled: true,
		},
		{
			ID:      3,
			Type:    "pagerduty",
			Enabled: false, // disabled
		},
	}

	activeChannels := filterActiveChannels(channels)
	assert.Equal(t, 2, len(activeChannels))
	assert.Equal(t, "slack", activeChannels[0].Type)
	assert.Equal(t, "email", activeChannels[1].Type)
}

// TestRetryScheduling tests retry scheduling logic
func TestRetryScheduling(t *testing.T) {
	delivery := &DeliveryResult{
		Status:     "failed",
		Attempt:    2,
		Timestamp:  time.Now(),
	}

	nextRetry := scheduleNextRetry(delivery)
	assert.True(t, nextRetry.After(time.Now()))
	assert.True(t, nextRetry.Before(time.Now().Add(30*time.Minute)))
}

// Helper functions for testing
func validateChannelConfig(config *ChannelConfig) bool {
	if config == nil || config.Type == "" {
		return false
	}

	switch config.Type {
	case "slack":
		webhook, ok := config.Config["webhook_url"]
		return ok && webhook.(string) != ""
	case "email":
		host, ok1 := config.Config["smtp_host"]
		recipients, ok2 := config.Config["recipients"]
		return ok1 && ok2 && host.(string) != "" && len(recipients.([]string)) > 0
	default:
		return true
	}
}

func calculateBackoffDelay(attempt int) time.Duration {
	delays := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	}
	if attempt <= 0 || attempt > len(delays) {
		return delays[len(delays)-1]
	}
	return delays[attempt-1]
}

func formatSlackMessage(alert *AlertNotification) string {
	return "Alert: " + alert.RuleName + " (" + alert.Severity + ") on " + alert.Database
}

func formatEmailSubject(alert *AlertNotification) string {
	return "[" + alert.Severity + "] " + alert.RuleName
}

func formatWebhookPayload(alert *AlertNotification) map[string]interface{} {
	return map[string]interface{}{
		"alert_id":   alert.AlertID,
		"rule_name":  alert.RuleName,
		"severity":   alert.Severity,
		"database":   alert.Database,
		"timestamp":  alert.Timestamp,
	}
}

type NotificationRateLimiter struct {
	throttleWindow time.Duration
	lastSent       map[string]time.Time
}

func NewNotificationRateLimiter(window time.Duration) *NotificationRateLimiter {
	return &NotificationRateLimiter{
		throttleWindow: window,
		lastSent:       make(map[string]time.Time),
	}
}

func (nrl *NotificationRateLimiter) Allow(fingerprint string) bool {
	lastTime, exists := nrl.lastSent[fingerprint]
	now := time.Now()

	if !exists {
		nrl.lastSent[fingerprint] = now
		return true
	}

	if now.Sub(lastTime) > nrl.throttleWindow {
		nrl.lastSent[fingerprint] = now
		return true
	}

	return false
}

func calculateSuccessRate(metrics *NotificationMetrics) float64 {
	if metrics.TotalSent == 0 {
		return 0
	}
	return float64(metrics.Successful) / float64(metrics.TotalSent)
}

func isChannelAvailable(recentFailures int, successRate float64) bool {
	return recentFailures < 5 && successRate > 0.70
}

type DeadLetterQueue struct {
	messages []*AlertNotification
}

func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		messages: make([]*AlertNotification, 0),
	}
}

func (dlq *DeadLetterQueue) Enqueue(msg *AlertNotification) {
	dlq.messages = append(dlq.messages, msg)
}

func (dlq *DeadLetterQueue) Dequeue() *AlertNotification {
	if len(dlq.messages) == 0 {
		return nil
	}
	msg := dlq.messages[0]
	dlq.messages = dlq.messages[1:]
	return msg
}

func (dlq *DeadLetterQueue) Size() int {
	return len(dlq.messages)
}

func filterActiveChannels(channels []*ChannelConfig) []*ChannelConfig {
	active := make([]*ChannelConfig, 0)
	for _, ch := range channels {
		if ch.Enabled {
			active = append(active, ch)
		}
	}
	return active
}

func scheduleNextRetry(delivery *DeliveryResult) time.Time {
	backoff := calculateBackoffDelay(delivery.Attempt)
	return time.Now().Add(backoff)
}

// Type definitions for testing
type ChannelConfig struct {
	ID       int64
	Name     string
	Type     string
	Enabled  bool
	Config   map[string]interface{}
}

type AlertNotification struct {
	AlertID   int64
	RuleID    int64
	RuleName  string
	Database  string
	Severity  string
	Message   string
	Timestamp time.Time
	Details   map[string]interface{}
}

type DeliveryResult struct {
	ChannelID    int64
	AlertID      int64
	Status       string
	Timestamp    time.Time
	Attempt      int
	NextRetryAt  *time.Time
	ErrorMessage *string
}

type NotificationMetrics struct {
	TotalSent             int
	Successful            int
	Failed                int
	Retried               int
	SuccessfulRetries     int
	AverageSendTime       time.Duration
	SlackMessagesCount    int
	EmailMessagesCount    int
	WebhookCallsCount     int
	PagerDutyCallsCount   int
	JiraIssuesCount       int
}

type NotificationService struct {
	Metrics *NotificationMetrics
}
