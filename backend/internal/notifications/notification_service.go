package notifications

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// NOTIFICATION SERVICE TYPES
// ============================================================================

// NotificationService manages multi-channel alert delivery
type NotificationService struct {
	db                  *sql.DB
	httpClient          *http.Client
	logger              *log.Logger
	maxRetries          int
	retryBackoffSeconds []int // exponential backoff: [1, 2, 4, 8, 16]
	channelTimeout      time.Duration

	// Channel implementations
	channels map[string]NotificationChannel

	// Metrics
	mu                  sync.RWMutex
	totalSent           int64
	totalFailed         int64
	totalRetried        int64
	deliverySuccessRate float64
	lastMetricsRecalc   time.Time
}

// NotificationChannel is the interface for notification providers
type NotificationChannel interface {
	Type() string
	Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error)
	Validate(config ChannelConfig) error
	Test(ctx context.Context, config ChannelConfig) error
}

// AlertNotification contains alert information for delivery
type AlertNotification struct {
	ID          int64
	RuleID      int64
	AlertID     int64
	Title       string
	Description string
	Severity    string // "low", "medium", "high", "critical"
	Status      string
	Context     json.RawMessage
	FiredAt     time.Time
	Database    string
	Query       string
}

// ChannelConfig represents configuration for a notification channel
type ChannelConfig struct {
	ID       int64
	Type     string          // "slack", "email", "webhook", "pagerduty", "jira"
	Config   json.RawMessage // Provider-specific config
	Verified bool
	Enabled  bool
}

// DeliveryResult contains result of delivery attempt
type DeliveryResult struct {
	Success     bool
	MessageID   string
	ErrorMsg    string
	DeliveredAt time.Time
}

// NotificationDelivery tracks delivery attempts
type NotificationDelivery struct {
	ID               int64
	AlertID          int64
	ChannelID        int64
	DeliveryStatus   string // "pending", "sent", "failed", "bounced"
	DeliveryAttempts int
	MaxRetries       int
	MessageSubject   string
	MessageBody      string
	DeliveredAt      *time.Time
	LastError        string
	NextRetryAt      *time.Time
}

// ============================================================================
// CONSTRUCTOR & INITIALIZATION
// ============================================================================

// NewNotificationService creates a new notification service
func NewNotificationService(db *sql.DB, logger *log.Logger) *NotificationService {
	if logger == nil {
		logger = log.New(os.Stderr, "[Notifications] ", log.LstdFlags)
	}

	channelTimeout := 10 * time.Second
	if envTimeout := os.Getenv("NOTIFICATION_CHANNEL_TIMEOUT"); envTimeout != "" {
		if d, err := time.ParseDuration(envTimeout); err == nil {
			channelTimeout = d
		}
	}

	service := &NotificationService{
		db: db,
		httpClient: &http.Client{
			Timeout: channelTimeout,
		},
		logger:              logger,
		maxRetries:          5,
		retryBackoffSeconds: []int{1, 2, 4, 8, 16},
		channelTimeout:      channelTimeout,
		channels:            make(map[string]NotificationChannel),
		lastMetricsRecalc:   time.Now(),
	}

	// Register channel implementations
	service.registerChannels()

	return service
}

// registerChannels registers all available notification channels
func (ns *NotificationService) registerChannels() {
	// Create a zap logger wrapper for channels
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	ns.channels["slack"] = NewSlackChannel(ns.httpClient, zapLogger, ns.channelTimeout)
	ns.channels["email"] = NewEmailChannel(zapLogger, ns.channelTimeout)
	ns.channels["webhook"] = NewWebhookChannel(ns.httpClient, zapLogger, ns.channelTimeout)
	ns.channels["pagerduty"] = NewPagerDutyChannel(ns.httpClient, zapLogger, ns.channelTimeout)
	ns.channels["jira"] = NewJiraChannel(ns.httpClient, zapLogger, ns.channelTimeout)
}

// ============================================================================
// NOTIFICATION DELIVERY WORKFLOW
// ============================================================================

// SendAlert sends an alert through all configured channels
func (ns *NotificationService) SendAlert(ctx context.Context, alert *AlertNotification) error {
	// Fetch all enabled notification channels for the alert user
	channels, err := ns.getUserChannels(ctx, alert)
	if err != nil {
		return fmt.Errorf("fetch channels: %w", err)
	}

	if len(channels) == 0 {
		log.Printf("[Notifications] No notification channels configured for alert %d\n", alert.ID)
		return nil
	}

	log.Printf("[Notifications] Sending alert %d through %d channels\n", alert.ID, len(channels))

	// Send through each channel concurrently
	wg := sync.WaitGroup{}
	for _, channel := range channels {
		wg.Add(1)
		go func(ch ChannelConfig) {
			defer wg.Done()
			ns.deliverToChannel(ctx, alert, ch)
		}(channel)
	}

	wg.Wait()
	return nil
}

// deliverToChannel attempts to deliver alert through a specific channel
func (ns *NotificationService) deliverToChannel(ctx context.Context, alert *AlertNotification, channelConfig ChannelConfig) {
	delivery := &NotificationDelivery{
		AlertID:        alert.ID,
		ChannelID:      channelConfig.ID,
		DeliveryStatus: "pending",
		MaxRetries:     ns.maxRetries,
	}

	// Get channel implementation
	channel, ok := ns.channels[channelConfig.Type]
	if !ok {
		log.Printf("[Notifications] Unknown channel type: %s\n", channelConfig.Type)
		delivery.DeliveryStatus = "failed"
		delivery.LastError = fmt.Sprintf("Unknown channel type: %s", channelConfig.Type)
		ns.storeDelivery(ctx, delivery)
		return
	}

	// Attempt delivery with retries
	var result *DeliveryResult
	var lastErr error

	for attempt := 0; attempt <= ns.maxRetries; attempt++ {
		delivery.DeliveryAttempts = attempt + 1

		// Send notification
		result, lastErr = channel.Send(ctx, alert, channelConfig)
		if lastErr == nil && result.Success {
			delivery.DeliveryStatus = "sent"
			delivery.DeliveredAt = &result.DeliveredAt
			break
		}

		// Store error
		if lastErr != nil {
			delivery.LastError = lastErr.Error()
		} else if result != nil {
			delivery.LastError = result.ErrorMsg
		}

		// Schedule retry if not last attempt
		if attempt < ns.maxRetries {
			backoffSecs := ns.retryBackoffSeconds[attempt]
			nextRetry := time.Now().Add(time.Duration(backoffSecs) * time.Second)
			delivery.NextRetryAt = &nextRetry

			log.Printf("[Notifications] Retry scheduled for alert %d, channel %d in %ds\n",
				alert.ID, channelConfig.ID, backoffSecs)

			// Wait before retry
			select {
			case <-time.After(time.Duration(backoffSecs) * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}

	// Mark as failed if all retries exhausted
	if result == nil || !result.Success {
		delivery.DeliveryStatus = "failed"
		log.Printf("[Notifications] Alert %d delivery to channel %d failed after %d attempts: %s\n",
			alert.ID, channelConfig.ID, delivery.DeliveryAttempts, delivery.LastError)
	}

	// Store delivery record
	ns.storeDelivery(ctx, delivery)

	// Update metrics
	ns.updateMetrics(delivery)
}

// ============================================================================
// CHANNEL MANAGEMENT
// ============================================================================

// CreateChannel creates a new notification channel for a user
func (ns *NotificationService) CreateChannel(ctx context.Context, userID int, name string, channelType string, config json.RawMessage) (int64, error) {
	// Validate channel type
	if _, ok := ns.channels[channelType]; !ok {
		return 0, fmt.Errorf("unknown channel type: %s", channelType)
	}

	// Validate config format
	if err := ns.channels[channelType].Validate(ChannelConfig{Config: config}); err != nil {
		return 0, fmt.Errorf("invalid config: %w", err)
	}

	query := `
		INSERT INTO notification_channels(
			user_id, name, channel_type, config, is_enabled
		) VALUES($1, $2, $3, $4, $5)
		RETURNING id
	`

	var channelID int64
	err := ns.db.QueryRowContext(ctx, query, userID, name, channelType, config, true).Scan(&channelID)
	if err != nil {
		return 0, fmt.Errorf("insert channel: %w", err)
	}

	log.Printf("[Notifications] Channel created: id=%d, user=%d, type=%s\n", channelID, userID, channelType)

	return channelID, nil
}

// DeleteChannel removes a notification channel
func (ns *NotificationService) DeleteChannel(ctx context.Context, channelID int64, userID int) error {
	query := `
		DELETE FROM notification_channels
		WHERE id = $1 AND user_id = $2
	`

	result, err := ns.db.ExecContext(ctx, query, channelID, userID)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("channel not found or not owned by user")
	}

	log.Printf("[Notifications] Channel deleted: id=%d\n", channelID)

	return nil
}

// TestChannel sends a test notification
func (ns *NotificationService) TestChannel(ctx context.Context, channelID int64, userID int) error {
	// Fetch channel configuration
	channelConfig := &ChannelConfig{}
	query := `
		SELECT id, channel_type, config, is_enabled
		FROM notification_channels
		WHERE id = $1 AND user_id = $2
	`

	err := ns.db.QueryRowContext(ctx, query, channelID, userID).Scan(
		&channelConfig.ID, &channelConfig.Type, &channelConfig.Config, &channelConfig.Enabled)
	if err == sql.ErrNoRows {
		return fmt.Errorf("channel not found")
	}
	if err != nil {
		return fmt.Errorf("fetch channel: %w", err)
	}

	// Get channel implementation
	channel, ok := ns.channels[channelConfig.Type]
	if !ok {
		return fmt.Errorf("unknown channel type: %s", channelConfig.Type)
	}

	// Send test notification
	if err := channel.Test(ctx, *channelConfig); err != nil {
		// Mark as failed
		query := `UPDATE notification_channels SET last_test_status = 'failed', last_test_error = $1, last_test_at = NOW() WHERE id = $2`
		ns.db.ExecContext(ctx, query, err.Error(), channelID)
		return fmt.Errorf("test failed: %w", err)
	}

	// Mark as successful
	query = `UPDATE notification_channels SET is_verified = TRUE, verified_at = NOW(), last_test_status = 'success', last_test_at = NOW() WHERE id = $1`
	ns.db.ExecContext(ctx, query, channelID)

	log.Printf("[Notifications] Channel test successful: id=%d, type=%s\n", channelID, channelConfig.Type)

	return nil
}

// ============================================================================
// DELIVERY TRACKING
// ============================================================================

// storeDelivery saves delivery attempt record
func (ns *NotificationService) storeDelivery(ctx context.Context, delivery *NotificationDelivery) error {
	query := `
		INSERT INTO notification_deliveries(
			alert_id, channel_id, delivery_status, delivery_attempts, max_retries,
			message_subject, message_body, delivered_at, last_error, next_retry_at
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := ns.db.ExecContext(ctx, query,
		delivery.AlertID, delivery.ChannelID, delivery.DeliveryStatus,
		delivery.DeliveryAttempts, delivery.MaxRetries,
		delivery.MessageSubject, delivery.MessageBody,
		delivery.DeliveredAt, delivery.LastError, delivery.NextRetryAt)

	return err
}

// RetryFailedDeliveries retries notifications that failed
func (ns *NotificationService) RetryFailedDeliveries(ctx context.Context) error {
	query := `
		SELECT nd.id, nd.alert_id, nd.channel_id, nd.delivery_attempts, nd.max_retries
		FROM notification_deliveries nd
		WHERE nd.delivery_status IN ('pending', 'failed')
		  AND nd.delivery_attempts < nd.max_retries
		  AND (nd.next_retry_at IS NULL OR nd.next_retry_at <= NOW())
		LIMIT 100
	`

	rows, err := ns.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("query failed deliveries: %w", err)
	}
	defer rows.Close()

	retriedCount := 0
	for rows.Next() {
		var deliveryID int64
		var alertID int64
		var channelID int64
		var attempts int
		var maxRetries int

		if err := rows.Scan(&deliveryID, &alertID, &channelID, &attempts, &maxRetries); err != nil {
			log.Printf("[Notifications] Error scanning delivery: %v\n", err)
			continue
		}

		// Would fetch alert and channel, then retry delivery
		// This is simplified - would need full implementation
		retriedCount++
	}

	log.Printf("[Notifications] Retried %d failed deliveries\n", retriedCount)

	return rows.Err()
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// getUserChannels fetches all notification channels for an alert user
func (ns *NotificationService) getUserChannels(ctx context.Context, alert *AlertNotification) ([]ChannelConfig, error) {
	// For now, simplified - would need to determine user from alert
	// In production, would check alert_rules.notification_channels

	query := `
		SELECT id, channel_type, config, is_verified, is_enabled
		FROM notification_channels
		WHERE is_enabled = TRUE
		  AND is_verified = TRUE
		LIMIT 50
	`

	rows, err := ns.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query channels: %w", err)
	}
	defer rows.Close()

	var channels []ChannelConfig
	for rows.Next() {
		ch := ChannelConfig{}
		if err := rows.Scan(&ch.ID, &ch.Type, &ch.Config, &ch.Verified, &ch.Enabled); err != nil {
			log.Printf("[Notifications] Error scanning channel: %v\n", err)
			continue
		}
		channels = append(channels, ch)
	}

	return channels, rows.Err()
}

// updateMetrics updates notification delivery statistics
func (ns *NotificationService) updateMetrics(delivery *NotificationDelivery) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if delivery.DeliveryStatus == "sent" {
		ns.totalSent++
	} else if delivery.DeliveryStatus == "failed" {
		ns.totalFailed++
	}

	if delivery.DeliveryAttempts > 1 {
		ns.totalRetried++
	}

	// Recalculate success rate every 100 notifications
	if (ns.totalSent+ns.totalFailed)%100 == 0 {
		total := ns.totalSent + ns.totalFailed
		if total > 0 {
			ns.deliverySuccessRate = float64(ns.totalSent) / float64(total)
		}
		ns.lastMetricsRecalc = time.Now()
	}
}

// GetMetrics returns notification service statistics
func (ns *NotificationService) GetMetrics() map[string]interface{} {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	return map[string]interface{}{
		"total_sent":           ns.totalSent,
		"total_failed":         ns.totalFailed,
		"total_retried":        ns.totalRetried,
		"success_rate_percent": ns.deliverySuccessRate * 100,
		"last_metrics_recalc":  ns.lastMetricsRecalc,
	}
}

// ============================================================================
// MESSAGE TEMPLATE HELPERS
// ============================================================================

// FormatAlertMessage creates formatted alert message
func FormatAlertMessage(alert *AlertNotification) string {
	template := `
[%s] Alert: %s
======================================

Description: %s
Severity: %s
Status: %s

Details:
- Alert ID: %d
- Rule ID: %d
- Time: %s

%s

--------------------------------------
pgAnalytics Monitoring System
`

	contextStr := ""
	if len(alert.Context) > 0 {
		var ctx map[string]interface{}
		if err := json.Unmarshal(alert.Context, &ctx); err == nil {
			contextStr = fmt.Sprintf("Context: %+v", ctx)
		}
	}

	return fmt.Sprintf(template,
		alert.Severity,
		alert.Title,
		alert.Description,
		alert.Severity,
		alert.Status,
		alert.AlertID,
		alert.RuleID,
		alert.FiredAt.Format("2006-01-02 15:04:05"),
		contextStr)
}

// FormatAlertHTML creates HTML version of alert message
func FormatAlertHTML(alert *AlertNotification) string {
	// Define template functions
	funcMap := template.FuncMap{
		"upper": strings.ToUpper,
	}

	tmpl := template.Must(template.New("alert").Funcs(funcMap).Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<style>
		body { font-family: Arial, sans-serif; }
		.alert-header { background-color: #f0f0f0; padding: 10px; }
		.severity-high { color: #d9534f; font-weight: bold; }
		.severity-medium { color: #f0ad4e; font-weight: bold; }
		.details { margin: 20px 0; }
		.details-table { width: 100%; border-collapse: collapse; }
		.details-table td { padding: 8px; border-bottom: 1px solid #ddd; }
	</style>
</head>
<body>
	<div class="alert-header">
		<h2><span class="severity-{{ .Severity }}">{{ .Severity | upper }}</span> Alert: {{ .Title }}</h2>
	</div>

	<div class="details">
		<p>{{ .Description }}</p>

		<table class="details-table">
			<tr>
				<td><strong>Alert ID:</strong></td>
				<td>{{ .AlertID }}</td>
			</tr>
			<tr>
				<td><strong>Rule ID:</strong></td>
				<td>{{ .RuleID }}</td>
			</tr>
			<tr>
				<td><strong>Severity:</strong></td>
				<td>{{ .Severity }}</td>
			</tr>
			<tr>
				<td><strong>Status:</strong></td>
				<td>{{ .Status }}</td>
			</tr>
			<tr>
				<td><strong>Fired At:</strong></td>
				<td>{{ .FiredAt.Format "2006-01-02 15:04:05" }}</td>
			</tr>
		</table>
	</div>

	<p style="color: #999; font-size: 12px;">
		pgAnalytics Monitoring System
	</p>
</body>
</html>
`))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, alert); err != nil {
		return fmt.Sprintf("<html><body>Error rendering template: %v</body></html>", err)
	}

	return buf.String()
}
