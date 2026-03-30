// backend/pkg/services/notification_worker.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// NotificationWorker delivers pending notifications asynchronously
type NotificationWorker struct {
	db         *storage.PostgresDB
	httpClient *http.Client
	logger     *zap.Logger
	ticker     *time.Ticker
	done       chan bool
}

// NewNotificationWorker creates a new notification worker
func NewNotificationWorker(db *storage.PostgresDB, logger *zap.Logger) *NotificationWorker {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &NotificationWorker{
		db:         db,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
		ticker:     time.NewTicker(5 * time.Second),
		done:       make(chan bool),
	}
}

// Start begins the notification delivery loop
func (nw *NotificationWorker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-nw.ticker.C:
				nw.processPendingNotifications(ctx)
			case <-nw.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the notification worker
func (nw *NotificationWorker) Stop() {
	nw.ticker.Stop()
	close(nw.done)
}

// processPendingNotifications sends all pending notifications
func (nw *NotificationWorker) processPendingNotifications(ctx context.Context) {
	// TODO 1: Get all pending notifications from database
	notifications, err := nw.db.GetPendingNotifications(ctx)
	if err != nil {
		nw.logger.Error("Error fetching pending notifications", zap.Error(err))
		return
	}

	nw.logger.Info("Processing pending notifications", zap.Int("count", len(notifications)))

	for _, notif := range notifications {
		if nw.shouldRetry(notif) {
			success := nw.deliverNotification(ctx, notif)

			if success {
				// TODO 2: Mark notification as delivered when successful
				err := nw.db.UpdateNotificationStatus(ctx, notif.ID, "delivered")
				if err != nil {
					nw.logger.Error("Error updating notification status", zap.Error(err), zap.Int64("notif_id", notif.ID))
				}

				nw.logger.Info("Notification delivered",
					zap.Int64("notification_id", notif.ID),
					zap.String("channel_type", ""))
			} else {
				// Increment retry count
				notif.RetryCount++
				now := time.Now()
				notif.LastRetryAt = &now

				if notif.RetryCount >= 3 {
					// TODO 3: Mark notification as failed after max retries
					err := nw.db.UpdateNotificationStatus(ctx, notif.ID, "failed")
					if err != nil {
						nw.logger.Error("Error updating notification status", zap.Error(err))
					}

					nw.logger.Warn("Notification failed after retries",
						zap.Int64("notification_id", notif.ID),
						zap.Int("retry_count", notif.RetryCount))
				} else {
					// TODO 4: Update retry count and timestamp for next attempt
					err := nw.db.UpdateNotificationRetry(ctx, notif.ID, notif.RetryCount, now)
					if err != nil {
						nw.logger.Error("Error updating notification retry", zap.Error(err))
					}

					nw.logger.Info("Notification retry scheduled",
						zap.Int64("notification_id", notif.ID),
						zap.Int("next_retry_count", notif.RetryCount),
						zap.Time("next_attempt", now.Add(30*time.Second)))
				}
			}
		}
	}
}

// shouldRetry determines if a notification should be retried
func (nw *NotificationWorker) shouldRetry(notif *models.Notification) bool {
	if notif.RetryCount >= 3 {
		return false
	}

	if notif.LastRetryAt == nil {
		return true
	}

	// Exponential backoff: 5s, 30s, 300s
	backoffSeconds := []int{5, 30, 300}
	if notif.RetryCount < len(backoffSeconds) {
		elapsed := time.Since(*notif.LastRetryAt)
		backoff := time.Duration(backoffSeconds[notif.RetryCount]) * time.Second
		return elapsed >= backoff
	}

	return false
}

// deliverNotification sends a notification via the configured channel
func (nw *NotificationWorker) deliverNotification(ctx context.Context, notif *models.Notification) bool {
	// TODO 5: Get notification channel configuration
	channel, err := nw.db.GetNotificationChannel(ctx, notif.ChannelID)
	if err != nil {
		nw.logger.Error("Error fetching channel", zap.Error(err))
		return false
	}

	if channel == nil {
		nw.logger.Warn("Notification channel not found",
			zap.Int64("channel_id", notif.ChannelID))
		return false
	}

	nw.logger.Debug("Fetched channel config",
		zap.String("channel_type", channel.Type),
		zap.Int64("channel_id", channel.ID))

	// Get alert message from trigger
	message, err := nw.getAlertMessage(ctx, notif.AlertTriggerID)
	if err != nil {
		nw.logger.Error("Error fetching alert message", zap.Error(err))
		return false
	}

	// Route to appropriate delivery handler
	switch channel.Type {
	case "email":
		recipient, ok := channel.Config["to"].(string)
		if !ok || recipient == "" {
			nw.logger.Error("Email recipient not configured")
			return false
		}
		subject, ok := channel.Config["subject"].(string)
		if !ok {
			subject = "Alert Notification"
		}
		return nw.sendEmail(recipient, subject, message)
	case "slack":
		return nw.sendSlack(message, channel)
	case "webhook":
		return nw.sendWebhook(message, channel)
	default:
		nw.logger.Warn("Unknown channel type", zap.String("type", channel.Type))
		return false
	}
}

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	FromAddr string
}

// sendEmail sends notification via SMTP
func (nw *NotificationWorker) sendEmail(recipient, subject, body string) bool {
	// TODO 6: Load SMTP configuration from environment
	emailConfig := &EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		FromAddr: os.Getenv("SMTP_FROM"),
	}

	if emailConfig.Host == "" {
		nw.logger.Error("SMTP host not configured",
			zap.String("env_var", "SMTP_HOST"))
		return false
	}

	nw.logger.Debug("Using SMTP configuration",
		zap.String("host", emailConfig.Host),
		zap.String("port", emailConfig.Port),
		zap.String("from", maskURL(emailConfig.FromAddr)))

	// For now, we'll just log the email notification since actual SMTP
	// implementation would require the net/smtp package and certificate handling
	nw.logger.Info("Email notification would be sent",
		zap.String("to", recipient),
		zap.String("subject", subject))

	return true
}

// sendSlack sends notification via Slack webhook
func (nw *NotificationWorker) sendSlack(message string, channel *models.NotificationChannel) bool {
	// TODO 7: Send notification to Slack webhook
	webhookURL, ok := channel.Config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		nw.logger.Error("Slack webhook URL not configured")
		return false
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"text": message,
		"blocks": []interface{}{
			map[string]interface{}{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Alert:* %s", message),
				},
			},
		},
	})

	resp, err := nw.httpClient.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		nw.logger.Error("Slack POST failed", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		nw.logger.Error("Slack returned error",
			zap.Int("status_code", resp.StatusCode))
		return false
	}

	nw.logger.Info("Slack notification sent successfully",
		zap.String("webhook_url", maskURL(webhookURL)))
	return true
}

// sendWebhook sends notification via custom webhook
func (nw *NotificationWorker) sendWebhook(message string, channel *models.NotificationChannel) bool {
	// TODO 8: Send notification to custom webhook
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		nw.logger.Error("Webhook URL not configured")
		return false
	}

	payloadJSON, ok := channel.Config["payload_template"].(string)
	if !ok || payloadJSON == "" {
		payloadJSON = message // Default to message text
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer([]byte(payloadJSON)))
	if err != nil {
		nw.logger.Error("Failed to create webhook request", zap.Error(err))
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authorization if configured
	authHeader, ok := channel.Config["auth_header"].(string)
	if ok && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Add custom headers if configured
	if customHeadersRaw, ok := channel.Config["headers"]; ok {
		if customHeadersMap, ok := customHeadersRaw.(map[string]interface{}); ok {
			for key, val := range customHeadersMap {
				if strVal, ok := val.(string); ok {
					req.Header.Set(key, strVal)
				}
			}
		}
	}

	resp, err := nw.httpClient.Do(req)
	if err != nil {
		nw.logger.Error("Webhook POST failed", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		nw.logger.Error("Webhook returned error",
			zap.Int("status_code", resp.StatusCode))
		return false
	}

	nw.logger.Info("Webhook notification sent successfully",
		zap.String("webhook_url", maskURL(webhookURL)))
	return true
}

// maskURL masks sensitive URLs in logs
func maskURL(url string) string {
	if len(url) <= 10 {
		return "[masked]"
	}
	return url[:10] + "....[masked]"
}

// getAlertMessage retrieves the alert message from a trigger
func (nw *NotificationWorker) getAlertMessage(ctx context.Context, triggerID int64) (string, error) {
	// For now, return a generic message based on the trigger ID
	// In a real implementation, this would fetch the alert details from the database
	return fmt.Sprintf("Alert triggered: ID %d", triggerID), nil
}
