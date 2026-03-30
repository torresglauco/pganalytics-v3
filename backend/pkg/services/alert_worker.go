// backend/pkg/services/alert_worker.go
package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// AlertWorker evaluates alert rules periodically
type AlertWorker struct {
	db        *storage.PostgresDB
	wsManager *ConnectionManager
	logger    *zap.Logger
	ticker    *time.Ticker
	done      chan bool
}

// NewAlertWorker creates a new alert worker
func NewAlertWorker(db *storage.PostgresDB, wsManager *ConnectionManager, logger *zap.Logger) *AlertWorker {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return &AlertWorker{
		db:        db,
		wsManager: wsManager,
		logger:    logger,
		ticker:    time.NewTicker(60 * time.Second),
		done:      make(chan bool),
	}
}

// Start begins the alert evaluation loop
func (aw *AlertWorker) Start(ctx context.Context) {
	go func() {
		// Run immediately on start
		aw.evaluateAlerts(ctx)

		for {
			select {
			case <-aw.ticker.C:
				aw.evaluateAlerts(ctx)
			case <-aw.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the alert worker
func (aw *AlertWorker) Stop() {
	aw.ticker.Stop()
	close(aw.done)
}

// evaluateAlerts checks all active alerts and creates triggers if conditions met
func (aw *AlertWorker) evaluateAlerts(ctx context.Context) {
	aw.logger.Info("Starting alert evaluation...")

	// Fetch all active alert rules from database
	alerts, err := aw.db.GetActiveAlertRules(ctx)
	if err != nil {
		aw.logger.Error("Error fetching alert rules", zap.Error(err))
		return
	}
	aw.logger.Info("Fetched alert rules for evaluation", zap.Int("count", len(alerts)))

	for _, alert := range alerts {
		// Check if already triggered recently (within 5 minutes)
		if aw.recentlyTriggered(ctx, int64(alert.ID)) {
			aw.logger.Debug("Alert triggered recently, skipping", zap.Int("alert_id", alert.ID))
			continue
		}

		// Get instances for each alert to evaluate
		instances, err := aw.db.GetAlertInstances(ctx, int64(alert.ID))
		if err != nil {
			aw.logger.Error("Error fetching instances", zap.Error(err), zap.Int("alert_id", alert.ID))
			continue
		}

		for _, instance := range instances {
			// Evaluate alert condition for this instance
			shouldTrigger, currentValue, err := aw.evaluateConditions(ctx, alert, instance)
			if err != nil {
				aw.logger.Error("Error evaluating condition", zap.Error(err), zap.Int("alert_id", alert.ID), zap.Int("instance_id", instance.ID))
				continue
			}

			if shouldTrigger {
				// Create alert trigger
				trigger := &models.AlertTrigger{
					AlertID:     int64(alert.ID),
					InstanceID:  instance.ID,
					TriggeredAt: time.Now(),
					CreatedAt:   time.Now(),
				}

				// Insert trigger into database
				triggerID, err := aw.db.CreateAlertTrigger(ctx, trigger)
				if err != nil {
					aw.logger.Error("Error creating trigger", zap.Error(err), zap.Int64("alert_id", trigger.AlertID), zap.Int("instance_id", trigger.InstanceID))
					continue
				}

				trigger.ID = triggerID
				aw.logger.Info("Alert trigger created", zap.Int64("trigger_id", triggerID), zap.Int("alert_id", alert.ID), zap.Int("instance_id", instance.ID))

				// Create notifications for all channels
				if err := aw.createNotifications(ctx, trigger); err != nil {
					aw.logger.Error("Error creating notifications", zap.Error(err), zap.Int64("trigger_id", triggerID))
				}

				// Broadcast WebSocket event
				aw.wsManager.BroadcastAlertEvent(map[string]interface{}{
					"alert_id":      alert.ID,
					"alert_name":    alert.Name,
					"instance_id":   instance.ID,
					"triggered_at":  trigger.TriggeredAt,
					"current_value": currentValue,
				}, instance.ID)
			}
		}
	}

	aw.logger.Info("Alert evaluation complete")
}

// recentlyTriggered checks if alert was triggered in the last 5 minutes
func (aw *AlertWorker) recentlyTriggered(ctx context.Context, alertID int64) bool {
	// Query database for recent triggers
	// For simplicity, check if the alert was triggered in the last 5 minutes
	trigger, err := aw.db.GetMostRecentTrigger(ctx, alertID, 0) // 0 means any instance
	if err != nil {
		aw.logger.Debug("Error checking recent triggers", zap.Error(err))
		return false
	}

	if trigger == nil {
		return false
	}

	// Don't trigger again if triggered within 5 minutes
	return time.Since(trigger.TriggeredAt) < 5*time.Minute
}

// evaluateConditions evaluates if alert conditions are met
// Returns (shouldTrigger, currentValue, error)
func (aw *AlertWorker) evaluateConditions(ctx context.Context, alert *models.AlertRule, instance *models.PostgreSQLInstance) (bool, float64, error) {
	// Fetch latest metric for this instance
	// Use metric_type from the alert rule
	metric, err := aw.db.GetLatestMetric(ctx, int64(instance.ID), alert.MetricType)
	if err != nil {
		return false, 0, fmt.Errorf("error fetching latest metric: %w", err)
	}

	if metric == nil {
		return false, 0, nil // No metric data yet
	}

	// Parse threshold value from ConditionValue
	threshold := 0.0
	if alert.ConditionValue != "" {
		thresholdVal, err := strconv.ParseFloat(alert.ConditionValue, 64)
		if err != nil {
			return false, metric.Value, fmt.Errorf("invalid threshold value: %s", alert.ConditionValue)
		}
		threshold = thresholdVal
	}

	// Evaluate condition based on condition type
	// ConditionType is stored as VARCHAR(50) in alert_rules table
	// Common conditions: "gt" (greater than), "lt" (less than), "eq" (equal), etc.
	switch alert.ConditionType {
	case "gt", "greater_than":
		// Trigger if metric value is greater than threshold
		if metric.Value > threshold {
			return true, metric.Value, nil
		}
	case "lt", "less_than":
		// Trigger if metric value is less than threshold
		if metric.Value < threshold {
			return true, metric.Value, nil
		}
	case "gte", "greater_than_or_equal":
		// Trigger if metric value is greater than or equal to threshold
		if metric.Value >= threshold {
			return true, metric.Value, nil
		}
	case "lte", "less_than_or_equal":
		// Trigger if metric value is less than or equal to threshold
		if metric.Value <= threshold {
			return true, metric.Value, nil
		}
	case "eq", "equal":
		// Trigger if metric value equals threshold
		if metric.Value == threshold {
			return true, metric.Value, nil
		}
	case "ne", "not_equal":
		// Trigger if metric value does not equal threshold
		if metric.Value != threshold {
			return true, metric.Value, nil
		}
	default:
		return false, metric.Value, fmt.Errorf("unknown condition type: %s", alert.ConditionType)
	}

	return false, metric.Value, nil
}

// createNotifications creates notification records for all alert channels
func (aw *AlertWorker) createNotifications(ctx context.Context, trigger *models.AlertTrigger) error {
	// Fetch all notification channels for this alert
	channels, err := aw.db.GetAlertChannels(ctx, trigger.AlertID)
	if err != nil {
		aw.logger.Error("Error fetching channels", zap.Error(err), zap.Int64("alert_id", trigger.AlertID))
		return err
	}

	if len(channels) == 0 {
		aw.logger.Debug("No notification channels configured for alert", zap.Int64("alert_id", trigger.AlertID))
		return nil
	}

	for _, channel := range channels {
		notif := &models.Notification{
			AlertTriggerID: trigger.ID,
			ChannelID:      channel.ID,
			Status:         "pending",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		notifID, err := aw.db.CreateNotification(ctx, notif)
		if err != nil {
			aw.logger.Error("Error creating notification", zap.Error(err), zap.Int64("trigger_id", trigger.ID), zap.Int64("channel_id", channel.ID))
			continue // Continue with other channels even if one fails
		}

		aw.logger.Debug("Notification created", zap.Int64("notification_id", notifID), zap.Int64("trigger_id", trigger.ID), zap.Int64("channel_id", channel.ID))
	}

	return nil
}
