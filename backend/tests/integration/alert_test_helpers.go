package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"go.uber.org/zap"
)

// setupTestDB creates and initializes a test database connection
func setupTestDB(t *testing.T) *storage.PostgresDB {
	// Use test database connection string from environment or use default
	connStr := os.Getenv("TEST_DB_CONNECTION")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/pganalytics_test?sslmode=disable"
	}

	db, err := storage.NewPostgresDB(connStr)
	if err != nil {
		t.Skipf("Could not connect to test database: %v", err)
	}

	return db
}

// setupTestAlert creates a complete alert setup with rule, instance, and channel
// Returns (alertRule, instance, ruleID)
func setupTestAlert(t *testing.T, db *storage.PostgresDB) (*models.AlertRule, *models.PostgreSQLInstance, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create test server first
	serverID := createTestDBServer(t, db)

	// Create test PostgreSQL instance
	instanceID := createTestInstance(t, db, serverID)
	instance := &models.PostgreSQLInstance{
		ID:                  instanceID,
		ServerID:            serverID,
		Name:                "test-instance",
		Port:                5432,
		MaintenanceDatabase: "postgres",
		MonitoringRole:      "pganalytics",
		IsActive:            true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Create alert rule
	rule := &models.AlertRule{
		Name:               "Test Alert",
		Enabled:            true,
		MetricType:         "cpu_usage",
		ConditionType:      "gt",
		ConditionValue:     "50.0",
		Severity:           "critical",
		EvaluationInterval: 60,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Insert alert rule into database
	query := `
		INSERT INTO pganalytics.alert_rules (
			name, enabled, metric_type, condition_type, condition_value,
			severity, evaluation_interval, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var ruleID int
	err := db.QueryRowContext(ctx, query,
		rule.Name, rule.Enabled, rule.MetricType, rule.ConditionType,
		rule.ConditionValue, rule.Severity, rule.EvaluationInterval,
		rule.CreatedAt, rule.UpdatedAt,
	).Scan(&ruleID)
	require.NoError(t, err, "Failed to create alert rule")

	rule.ID = ruleID

	return rule, instance, ruleID
}

// createTestDBServer creates a test database server for instances
func createTestDBServer(t *testing.T, db *storage.PostgresDB) int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO pganalytics.servers (
			collector_id, hostname, ip_address, port, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var serverID int
	collectorID := uuid.New()
	err := db.QueryRowContext(ctx, query,
		collectorID, "test-server", "127.0.0.1", 5432, true, time.Now(), time.Now(),
	).Scan(&serverID)
	require.NoError(t, err, "Failed to create test server")

	return serverID
}

// createTestInstance creates a test PostgreSQL instance
func createTestInstance(t *testing.T, db *storage.PostgresDB, serverID int) int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO pganalytics.postgresql_instances (
			server_id, name, port, maintenance_database, monitoring_role,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var instanceID int
	err := db.QueryRowContext(ctx, query,
		serverID, "test-instance", 5432, "postgres", "pganalytics",
		true, time.Now(), time.Now(),
	).Scan(&instanceID)
	require.NoError(t, err, "Failed to create test instance")

	return instanceID
}

// setupTestNotificationChannel creates a test notification channel
func setupTestNotificationChannel(t *testing.T, db *storage.PostgresDB, alertID int, channelType string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := map[string]interface{}{
		"webhook_url": "https://hooks.example.com/services/test",
	}

	channel := &models.NotificationChannel{
		AlertID:   alertID,
		Type:      channelType,
		Config:    config,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO pganalytics.notification_channels (
			alert_id, type, config, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var channelID int64
	configJSON := "{\"webhook_url\": \"https://hooks.example.com/services/test\"}"
	err := db.QueryRowContext(ctx, query,
		channel.AlertID, channel.Type, configJSON, channel.IsActive,
		channel.CreatedAt, channel.UpdatedAt,
	).Scan(&channelID)
	require.NoError(t, err, "Failed to create notification channel")

	return channelID
}

// createTestMetric inserts a metric for an instance
func createTestMetric(t *testing.T, db *storage.PostgresDB, instanceID int, metricName string, value float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO pganalytics.metrics (
			instance_id, metric_name, value, timestamp, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now()
	_, err := db.ExecContext(ctx, query,
		instanceID, metricName, value, now, now,
	)
	require.NoError(t, err, "Failed to create metric")
}

// getAlertTriggers retrieves all triggers for an alert rule
func getAlertTriggers(t *testing.T, db *storage.PostgresDB, alertID int64) []*models.AlertTrigger {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT id, alert_id, instance_id, triggered_at, created_at
		FROM pganalytics.alert_triggers
		WHERE alert_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, alertID)
	require.NoError(t, err, "Failed to query alert triggers")
	defer rows.Close()

	var triggers []*models.AlertTrigger
	for rows.Next() {
		trigger := &models.AlertTrigger{}
		err := rows.Scan(
			&trigger.ID, &trigger.AlertID, &trigger.InstanceID,
			&trigger.TriggeredAt, &trigger.CreatedAt,
		)
		require.NoError(t, err, "Failed to scan alert trigger")
		triggers = append(triggers, trigger)
	}

	return triggers
}

// getNotifications retrieves all notifications for an alert
func getNotifications(t *testing.T, db *storage.PostgresDB, alertID int64) []*models.Notification {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT n.id, n.channel_id, n.alert_trigger_id, n.status, n.retry_count,
		       n.last_retry_at, n.sent_at, n.created_at, n.updated_at
		FROM pganalytics.notifications n
		JOIN pganalytics.alert_triggers t ON n.alert_trigger_id = t.id
		WHERE t.alert_id = $1
		ORDER BY n.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, alertID)
	require.NoError(t, err, "Failed to query notifications")
	defer rows.Close()

	var notifs []*models.Notification
	for rows.Next() {
		notif := &models.Notification{}
		err := rows.Scan(
			&notif.ID, &notif.ChannelID, &notif.AlertTriggerID, &notif.Status,
			&notif.RetryCount, &notif.LastRetryAt, &notif.SentAt, &notif.CreatedAt,
			&notif.UpdatedAt,
		)
		require.NoError(t, err, "Failed to scan notification")
		notifs = append(notifs, notif)
	}

	return notifs
}

// getNotification retrieves a single notification by ID
func getNotification(t *testing.T, db *storage.PostgresDB, notifID int64) *models.Notification {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT id, channel_id, alert_trigger_id, status, retry_count,
		       last_retry_at, sent_at, created_at, updated_at
		FROM pganalytics.notifications
		WHERE id = $1
	`

	notif := &models.Notification{}
	err := db.QueryRowContext(ctx, query, notifID).Scan(
		&notif.ID, &notif.ChannelID, &notif.AlertTriggerID, &notif.Status,
		&notif.RetryCount, &notif.LastRetryAt, &notif.SentAt, &notif.CreatedAt,
		&notif.UpdatedAt,
	)
	require.NoError(t, err, "Failed to query notification")

	return notif
}

// createTestAlertSilence creates a silence rule for an alert
func createTestAlertSilence(t *testing.T, db *storage.PostgresDB, alertID int, instanceID int, duration time.Duration) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		INSERT INTO pganalytics.alert_silences (
			alert_rule_id, instance_id, silenced_until, silence_type, created_at
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var silenceID int64
	silencedUntil := time.Now().Add(duration)
	err := db.QueryRowContext(ctx, query,
		alertID, instanceID, silencedUntil, "temporary", time.Now(),
	).Scan(&silenceID)
	require.NoError(t, err, "Failed to create alert silence")

	return silenceID
}

// isSilenced checks if an alert is silenced for a given instance
func isSilenced(t *testing.T, db *storage.PostgresDB, alertID int, instanceID int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT EXISTS (
			SELECT 1 FROM pganalytics.alert_silences
			WHERE alert_rule_id = $1 AND instance_id = $2 AND silenced_until > NOW()
		)
	`

	var exists bool
	err := db.QueryRowContext(ctx, query, alertID, instanceID).Scan(&exists)
	require.NoError(t, err, "Failed to check if alert is silenced")

	return exists
}

// MockHTTPClient creates a mock HTTP client that returns predefined responses
type MockHTTPClient struct {
	responses    map[string]*http.Response
	requestCount int
	failCount    int
	maxFails     int
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses:    make(map[string]*http.Response),
		requestCount: 0,
		failCount:    0,
		maxFails:     0,
	}
}

// SetFailCount sets how many requests should fail before succeeding
func (m *MockHTTPClient) SetFailCount(count int) {
	m.maxFails = count
}

// Do executes a mock HTTP request
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.requestCount++

	// Check if we should fail this request
	if m.failCount < m.maxFails {
		m.failCount++
		return &http.Response{
			StatusCode: 500,
			Status:     "500 Internal Server Error",
			Header:     http.Header{},
			Body:       httptest.NewRecorder().Result().Body,
		}, nil
	}

	// Return success
	return &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     http.Header{},
		Body:       httptest.NewRecorder().Result().Body,
	}, nil
}

// getTestLogger creates a test logger
func getTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewProduction()
	require.NoError(t, err)
	return logger
}

// TestWebhookServer creates a test HTTP server for webhook testing
type TestWebhookServer struct {
	server        *httptest.Server
	receivedCalls int
	lastPayload   []byte
}

// NewTestWebhookServer creates a new test webhook server
func NewTestWebhookServer() *TestWebhookServer {
	tw := &TestWebhookServer{}

	tw.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw.receivedCalls++

		// Read and store the payload
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		tw.lastPayload = buf[:n]

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))

	return tw
}

// GetURL returns the webhook server URL
func (tw *TestWebhookServer) GetURL() string {
	return tw.server.URL
}

// GetReceivedCalls returns the number of calls received
func (tw *TestWebhookServer) GetReceivedCalls() int {
	return tw.receivedCalls
}

// GetLastPayload returns the last received payload
func (tw *TestWebhookServer) GetLastPayload() []byte {
	return tw.lastPayload
}

// Close closes the test server
func (tw *TestWebhookServer) Close() {
	tw.server.Close()
}
