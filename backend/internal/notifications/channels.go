package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ============================================================================
// CIRCUIT BREAKER SUPPORT
// ============================================================================

// BaseChannel provides common functionality for all notification channels
type BaseChannel struct {
	circuitBreaker *CircuitBreaker
	logger         *zap.Logger
	timeout        time.Duration
}

// NewBaseChannel creates a new base channel with circuit breaker
func NewBaseChannel(logger *zap.Logger, timeout time.Duration) *BaseChannel {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &BaseChannel{
		circuitBreaker: NewCircuitBreaker(logger),
		logger:         logger,
		timeout:        timeout,
	}
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	// StateClosed means the circuit is closed (normal operation)
	StateClosed CircuitBreakerState = "closed"
	// StateOpen means the circuit is open (service unavailable)
	StateOpen CircuitBreakerState = "open"
	// StateHalfOpen means the circuit is half-open (testing recovery)
	StateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker implements the circuit breaker pattern for notification channels
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failureCount     int
	successCount     int
	lastFailureTime  time.Time
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	logger           *zap.Logger
}

// NewCircuitBreaker creates a new circuit breaker for a notification channel
func NewCircuitBreaker(logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureCount:     0,
		successCount:     0,
		failureThreshold: 5,                // Open after 5 failures
		successThreshold: 3,                // Close after 3 successes
		timeout:          30 * time.Second, // Try recovery after 30 seconds
		logger:           logger,
	}
}

// RecordSuccess records a successful call
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Success in closed state, reset counter
		cb.failureCount = 0
		cb.successCount = 0

	case StateHalfOpen:
		// Success in half-open state, increment success counter
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
			cb.logger.Info("Circuit breaker closed - service recovered")
		}

	case StateOpen:
		// Ignore successes when open (waiting for timeout)
	}
}

// RecordFailure records a failed call
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Failure in closed state, increment counter
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
			cb.logger.Warn("Circuit breaker opened - too many failures",
				zap.Int("failure_count", cb.failureCount))
		}

	case StateHalfOpen:
		// Failure in half-open state, re-open the circuit
		cb.state = StateOpen
		cb.failureCount = 0
		cb.successCount = 0
		cb.logger.Warn("Circuit breaker reopened - failure during recovery")

	case StateOpen:
		// Already open, just update timestamp
		cb.lastFailureTime = time.Now()
	}
}

// IsOpen checks if the circuit is open (service unavailable)
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateClosed {
		return false
	}

	if cb.state == StateOpen {
		// Check if timeout has elapsed to try recovery
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Upgrade to half-open
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.failureCount = 0
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			cb.logger.Info("Circuit breaker half-open - attempting recovery")
			return false
		}
		return true
	}

	// Half-open state
	return false
}

// State returns the current state as a string
func (cb *CircuitBreaker) State() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return string(cb.state)
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailureTime = time.Time{}
	cb.logger.Info("Circuit breaker reset to closed state")
}

// GetMetrics returns the current metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":              string(cb.state),
		"failure_count":      cb.failureCount,
		"success_count":      cb.successCount,
		"failure_threshold":  cb.failureThreshold,
		"success_threshold":  cb.successThreshold,
		"last_failure_time":  cb.lastFailureTime,
		"time_since_failure": time.Since(cb.lastFailureTime).Seconds(),
	}
}

// ============================================================================
// SLACK CHANNEL
// ============================================================================

type SlackChannel struct {
	*BaseChannel
	httpClient *http.Client
}

type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
}

type SlackMessage struct {
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text"`
	Attachments []SlackAttach `json:"attachments"`
}

type SlackAttach struct {
	Color     string       `json:"color"`
	Title     string       `json:"title"`
	Text      string       `json:"text"`
	TitleLink string       `json:"title_link,omitempty"`
	Fields    []SlackField `json:"fields"`
	Footer    string       `json:"footer"`
	TS        int64        `json:"ts"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NewSlackChannel(httpClient *http.Client, logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &SlackChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
		httpClient:  httpClient,
	}
}

func (s *SlackChannel) Type() string {
	return "slack"
}

func (s *SlackChannel) Validate(config ChannelConfig) error {
	var slackConfig SlackConfig
	if err := json.Unmarshal(config.Config, &slackConfig); err != nil {
		return fmt.Errorf("unmarshal slack config: %w", err)
	}

	if slackConfig.WebhookURL == "" {
		return fmt.Errorf("webhook_url required")
	}

	return nil
}

func (s *SlackChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if s.circuitBreaker.IsOpen() {
		s.logger.Warn("Slack circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "Slack service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	var slackConfig SlackConfig
	if err := json.Unmarshal(config.Config, &slackConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Color based on severity
	color := "#36a64f" // green
	switch alert.Severity {
	case "critical":
		color = "#d9534f" // red
	case "high":
		color = "#f0ad4e" // orange
	case "medium":
		color = "#5bc0de" // blue
	}

	// Build message
	msg := SlackMessage{
		Channel:  slackConfig.Channel,
		Username: slackConfig.Username,
		Text:     fmt.Sprintf("Alert: %s", alert.Title),
		Attachments: []SlackAttach{
			{
				Color: color,
				Title: alert.Title,
				Text:  alert.Description,
				Fields: []SlackField{
					{
						Title: "Severity",
						Value: strings.ToUpper(alert.Severity),
						Short: true,
					},
					{
						Title: "Status",
						Value: alert.Status,
						Short: true,
					},
					{
						Title: "Database",
						Value: alert.Database,
						Short: true,
					},
					{
						Title: "Query",
						Value: alert.Query,
						Short: true,
					},
				},
				Footer: "pgAnalytics",
				TS:     alert.FiredAt.Unix(),
			},
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal message: %w", err)
	}

	// Send to Slack
	req, err := http.NewRequestWithContext(ctx, "POST", slackConfig.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.circuitBreaker.RecordFailure()
		s.logger.Error("Slack POST failed", zap.Error(err))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("Slack POST failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.circuitBreaker.RecordFailure()
		s.logger.Error("Slack returned error",
			zap.Int("status_code", resp.StatusCode))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	// Success
	s.circuitBreaker.RecordSuccess()
	s.logger.Info("Slack notification delivered")
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("slack_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (s *SlackChannel) Test(ctx context.Context, config ChannelConfig) error {
	var slackConfig SlackConfig
	if err := json.Unmarshal(config.Config, &slackConfig); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	testMsg := SlackMessage{
		Text: "🔔 pgAnalytics notification channel test - connection successful!",
	}

	payload, _ := json.Marshal(testMsg)

	req, err := http.NewRequestWithContext(ctx, "POST", slackConfig.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

// ============================================================================
// EMAIL CHANNEL
// ============================================================================

type EmailChannel struct {
	*BaseChannel
}

type EmailConfig struct {
	Recipients []string `json:"recipients"`
	SMTPURL    string   `json:"smtp_url,omitempty"` // Optional override for SMTP host:port
	From       string   `json:"from,omitempty"`     // Optional from address override
}

func NewEmailChannel(logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &EmailChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
	}
}

func (e *EmailChannel) Type() string {
	return "email"
}

func (e *EmailChannel) Validate(config ChannelConfig) error {
	var emailConfig EmailConfig
	if err := json.Unmarshal(config.Config, &emailConfig); err != nil {
		return fmt.Errorf("unmarshal email config: %w", err)
	}

	if len(emailConfig.Recipients) == 0 {
		return fmt.Errorf("recipients required")
	}

	return nil
}

func (e *EmailChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if e.circuitBreaker.IsOpen() {
		e.logger.Warn("Email circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "Email service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	_, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	var emailConfig EmailConfig
	if err := json.Unmarshal(config.Config, &emailConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if len(emailConfig.Recipients) == 0 {
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "Email recipients not configured",
			DeliveredAt: now(),
		}, nil
	}

	// Read SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587"
	}
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	// Allow per-channel overrides from EmailConfig
	if emailConfig.SMTPURL != "" {
		// Parse SMTPURL as host:port
		parts := strings.Split(emailConfig.SMTPURL, ":")
		if len(parts) >= 1 {
			smtpHost = parts[0]
			if len(parts) >= 2 {
				smtpPort = parts[1]
			}
		}
	}
	if emailConfig.From != "" {
		smtpFrom = emailConfig.From
	}

	// Validate required SMTP settings
	if smtpHost == "" || smtpUser == "" || smtpPassword == "" || smtpFrom == "" {
		e.logger.Warn("SMTP not configured",
			zap.Bool("has_host", smtpHost != ""),
			zap.Bool("has_user", smtpUser != ""),
			zap.Bool("has_password", smtpPassword != ""),
			zap.Bool("has_from", smtpFrom != ""))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "SMTP not configured",
			DeliveredAt: now(),
		}, nil
	}

	// Build email message with HTML content
	subject := fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Severity), alert.Title)
	body := FormatAlertHTML(alert)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		smtpFrom, strings.Join(emailConfig.Recipients, ","), subject, body)

	// Create SMTP authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// Send email using net/smtp
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, smtpFrom, emailConfig.Recipients, []byte(msg))
	if err != nil {
		e.circuitBreaker.RecordFailure()
		e.logger.Error("SMTP send failed",
			zap.Error(err),
			zap.String("host", smtpHost),
			zap.Int("recipient_count", len(emailConfig.Recipients)))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("SMTP send failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}

	// Success
	e.circuitBreaker.RecordSuccess()
	e.logger.Info("Email notification delivered",
		zap.String("host", smtpHost),
		zap.Int("recipient_count", len(emailConfig.Recipients)))
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("email_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (e *EmailChannel) Test(ctx context.Context, config ChannelConfig) error {
	var emailConfig EmailConfig
	if err := json.Unmarshal(config.Config, &emailConfig); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	// Validate email addresses format
	for _, recipient := range emailConfig.Recipients {
		if !isValidEmail(recipient) {
			return fmt.Errorf("invalid email: %s", recipient)
		}
	}

	// Verify SMTP configuration from environment
	if os.Getenv("SMTP_HOST") == "" {
		return fmt.Errorf("SMTP_HOST not configured")
	}
	if os.Getenv("SMTP_USER") == "" {
		return fmt.Errorf("SMTP_USER not configured")
	}
	if os.Getenv("SMTP_PASSWORD") == "" {
		return fmt.Errorf("SMTP_PASSWORD not configured")
	}
	if os.Getenv("SMTP_FROM") == "" {
		return fmt.Errorf("SMTP_FROM not configured")
	}

	// Note: A full SMTP connection test could be added here
	// to verify connectivity without sending an actual email

	return nil
}

// ============================================================================
// WEBHOOK CHANNEL
// ============================================================================

type WebhookChannel struct {
	*BaseChannel
	httpClient *http.Client
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Auth    *AuthConfig       `json:"auth,omitempty"`
}

type AuthConfig struct {
	Type     string `json:"type"` // "basic", "bearer"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
}

type WebhookPayload struct {
	Alert     AlertNotification `json:"alert"`
	Timestamp int64             `json:"timestamp"`
	Source    string            `json:"source"`
}

func NewWebhookChannel(httpClient *http.Client, logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &WebhookChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
		httpClient:  httpClient,
	}
}

func (w *WebhookChannel) Type() string {
	return "webhook"
}

func (w *WebhookChannel) Validate(config ChannelConfig) error {
	var webhookConfig WebhookConfig
	if err := json.Unmarshal(config.Config, &webhookConfig); err != nil {
		return fmt.Errorf("unmarshal webhook config: %w", err)
	}

	if webhookConfig.URL == "" {
		return fmt.Errorf("url required")
	}

	if webhookConfig.Method == "" {
		webhookConfig.Method = "POST"
	}

	return nil
}

func (w *WebhookChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if w.circuitBreaker.IsOpen() {
		w.logger.Warn("Webhook circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "Webhook service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	var webhookConfig WebhookConfig
	if err := json.Unmarshal(config.Config, &webhookConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if webhookConfig.Method == "" {
		webhookConfig.Method = "POST"
	}

	// Build payload
	payload := WebhookPayload{
		Alert:     *alert,
		Timestamp: now().Unix(),
		Source:    "pganalytics",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, webhookConfig.Method, webhookConfig.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if webhookConfig.Headers != nil {
		for k, v := range webhookConfig.Headers {
			req.Header.Set(k, v)
		}
	}

	// Add auth
	if webhookConfig.Auth != nil {
		switch webhookConfig.Auth.Type {
		case "basic":
			req.SetBasicAuth(webhookConfig.Auth.Username, webhookConfig.Auth.Password)
		case "bearer":
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", webhookConfig.Auth.Token))
		}
	}

	// Send request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.circuitBreaker.RecordFailure()
		w.logger.Error("Webhook POST failed", zap.Error(err))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("Webhook POST failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		w.circuitBreaker.RecordFailure()
		w.logger.Error("Webhook returned error",
			zap.Int("status_code", resp.StatusCode))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	w.circuitBreaker.RecordSuccess()
	w.logger.Info("Webhook notification delivered")
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("webhook_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (w *WebhookChannel) Test(ctx context.Context, config ChannelConfig) error {
	var webhookConfig WebhookConfig
	if err := json.Unmarshal(config.Config, &webhookConfig); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	testAlert := &AlertNotification{
		Title:       "Test Alert",
		Description: "This is a test notification from pgAnalytics",
		Severity:    "medium",
		Status:      "firing",
		FiredAt:     now(),
	}

	_, err := w.Send(ctx, testAlert, config)
	return err
}

// ============================================================================
// PAGERDUTY CHANNEL
// ============================================================================

type PagerDutyChannel struct {
	*BaseChannel
	httpClient *http.Client
}

type PagerDutyConfig struct {
	IntegrationKey string `json:"integration_key"`
	ServiceKey     string `json:"service_key,omitempty"`
}

type PagerDutyEvent struct {
	RoutingKey  string           `json:"routing_key"`
	EventAction string           `json:"event_action"`
	Dedup       string           `json:"dedup_key"`
	Payload     PagerDutyPayload `json:"payload"`
}

type PagerDutyPayload struct {
	Summary   string          `json:"summary"`
	Severity  string          `json:"severity"`
	Source    string          `json:"source"`
	Timestamp string          `json:"timestamp"`
	Details   json.RawMessage `json:"custom_details"`
}

func NewPagerDutyChannel(httpClient *http.Client, logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &PagerDutyChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
		httpClient:  httpClient,
	}
}

func (p *PagerDutyChannel) Type() string {
	return "pagerduty"
}

func (p *PagerDutyChannel) Validate(config ChannelConfig) error {
	var pdConfig PagerDutyConfig
	if err := json.Unmarshal(config.Config, &pdConfig); err != nil {
		return fmt.Errorf("unmarshal pagerduty config: %w", err)
	}

	if pdConfig.IntegrationKey == "" {
		return fmt.Errorf("integration_key required")
	}

	return nil
}

func (p *PagerDutyChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if p.circuitBreaker.IsOpen() {
		p.logger.Warn("PagerDuty circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "PagerDuty service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	var pdConfig PagerDutyConfig
	if err := json.Unmarshal(config.Config, &pdConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Map severity
	pdSeverity := "error"
	switch alert.Severity {
	case "critical":
		pdSeverity = "critical"
	case "high":
		pdSeverity = "error"
	case "medium":
		pdSeverity = "warning"
	case "low":
		pdSeverity = "info"
	}

	// Build event
	event := PagerDutyEvent{
		RoutingKey:  pdConfig.IntegrationKey,
		EventAction: "trigger",
		Dedup:       fmt.Sprintf("pganalytics_%d", alert.AlertID),
		Payload: PagerDutyPayload{
			Summary:   alert.Title,
			Severity:  pdSeverity,
			Source:    "pgAnalytics",
			Timestamp: now().Format("2006-01-02T15:04:05Z07:00"),
			Details:   alert.Context,
		},
	}

	body, _ := json.Marshal(event)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.circuitBreaker.RecordFailure()
		p.logger.Error("PagerDuty POST failed", zap.Error(err))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("PagerDuty POST failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		p.circuitBreaker.RecordFailure()
		p.logger.Error("PagerDuty returned error",
			zap.Int("status_code", resp.StatusCode))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	p.circuitBreaker.RecordSuccess()
	p.logger.Info("PagerDuty notification delivered")
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("pd_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (p *PagerDutyChannel) Test(ctx context.Context, config ChannelConfig) error {
	testAlert := &AlertNotification{
		AlertID:  0,
		Title:    "pgAnalytics Test Alert",
		Severity: "warning",
		Status:   "firing",
		FiredAt:  now(),
	}

	_, err := p.Send(ctx, testAlert, config)
	return err
}

// ============================================================================
// JIRA CHANNEL
// ============================================================================

type JiraChannel struct {
	*BaseChannel
	httpClient *http.Client
}

type JiraConfig struct {
	URL          string `json:"url"`
	ProjectKey   string `json:"project_key"`
	IssueType    string `json:"issue_type,omitempty"`
	AuthUsername string `json:"auth_username"`
	AuthToken    string `json:"auth_token"`
}

type JiraCreateIssue struct {
	Fields JiraIssueFields `json:"fields"`
}

type JiraIssueFields struct {
	Project     JiraProject   `json:"project"`
	IssueType   JiraIssueType `json:"issuetype"`
	Summary     string        `json:"summary"`
	Description string        `json:"description"`
	Priority    JiraPriority  `json:"priority,omitempty"`
	Labels      []string      `json:"labels,omitempty"`
}

type JiraProject struct {
	Key string `json:"key"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraPriority struct {
	Name string `json:"name"`
}

func NewJiraChannel(httpClient *http.Client, logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &JiraChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
		httpClient:  httpClient,
	}
}

func (j *JiraChannel) Type() string {
	return "jira"
}

func (j *JiraChannel) Validate(config ChannelConfig) error {
	var jiraConfig JiraConfig
	if err := json.Unmarshal(config.Config, &jiraConfig); err != nil {
		return fmt.Errorf("unmarshal jira config: %w", err)
	}

	if jiraConfig.URL == "" || jiraConfig.ProjectKey == "" || jiraConfig.AuthToken == "" {
		return fmt.Errorf("url, project_key, and auth_token required")
	}

	return nil
}

func (j *JiraChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if j.circuitBreaker.IsOpen() {
		j.logger.Warn("Jira circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "Jira service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, j.timeout)
	defer cancel()

	var jiraConfig JiraConfig
	if err := json.Unmarshal(config.Config, &jiraConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if jiraConfig.IssueType == "" {
		jiraConfig.IssueType = "Bug"
	}

	// Map severity to priority
	priority := "Medium"
	switch alert.Severity {
	case "critical":
		priority = "Highest"
	case "high":
		priority = "High"
	case "medium":
		priority = "Medium"
	case "low":
		priority = "Low"
	}

	// Build issue
	issue := JiraCreateIssue{
		Fields: JiraIssueFields{
			Project:     JiraProject{Key: jiraConfig.ProjectKey},
			IssueType:   JiraIssueType{Name: jiraConfig.IssueType},
			Summary:     alert.Title,
			Description: alert.Description,
			Priority:    JiraPriority{Name: priority},
			Labels:      []string{"pganalytics", alert.Severity},
		},
	}

	body, _ := json.Marshal(issue)

	url := strings.TrimRight(jiraConfig.URL, "/") + "/rest/api/3/issues"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(jiraConfig.AuthUsername, jiraConfig.AuthToken)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		j.circuitBreaker.RecordFailure()
		j.logger.Error("Jira POST failed", zap.Error(err))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("Jira POST failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		j.circuitBreaker.RecordFailure()
		j.logger.Error("Jira returned error",
			zap.Int("status_code", resp.StatusCode))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	j.circuitBreaker.RecordSuccess()
	j.logger.Info("Jira notification delivered")
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("jira_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (j *JiraChannel) Test(ctx context.Context, config ChannelConfig) error {
	testAlert := &AlertNotification{
		AlertID:     0,
		Title:       "pgAnalytics Test Issue",
		Description: "Test notification from pgAnalytics",
		Severity:    "medium",
		FiredAt:     now(),
	}

	_, err := j.Send(ctx, testAlert, config)
	return err
}

// ============================================================================
// OPSGENIE CHANNEL
// ============================================================================

type OpsGenieChannel struct {
	*BaseChannel
	httpClient *http.Client
}

type OpsGenieConfig struct {
	APIKey string `json:"api_key"`
	Region string `json:"region,omitempty"` // "us" or "eu"
	TeamID string `json:"team_id,omitempty"`
}

type OpsGenieAlert struct {
	Message     string            `json:"message"`
	Alias       string            `json:"alias"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags"`
	Details     map[string]string `json:"details,omitempty"`
}

type OpsGenieResponse struct {
	Result  string `json:"result"`
	AlertID string `json:"alertId,omitempty"`
}

func NewOpsGenieChannel(httpClient *http.Client, logger *zap.Logger, timeout time.Duration) NotificationChannel {
	return &OpsGenieChannel{
		BaseChannel: NewBaseChannel(logger, timeout),
		httpClient:  httpClient,
	}
}

func (o *OpsGenieChannel) Type() string {
	return "opsgenie"
}

func (o *OpsGenieChannel) Validate(config ChannelConfig) error {
	var ogConfig OpsGenieConfig
	if err := json.Unmarshal(config.Config, &ogConfig); err != nil {
		return fmt.Errorf("unmarshal opsgenie config: %w", err)
	}

	if ogConfig.APIKey == "" {
		return fmt.Errorf("api_key required")
	}

	// Validate region if provided
	if ogConfig.Region != "" && ogConfig.Region != "us" && ogConfig.Region != "eu" {
		return fmt.Errorf("invalid region '%s'. Valid values: us, eu", ogConfig.Region)
	}

	return nil
}

func (o *OpsGenieChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
	// Check circuit breaker
	if o.circuitBreaker.IsOpen() {
		o.logger.Warn("OpsGenie circuit breaker is open")
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    "OpsGenie service temporarily unavailable (circuit open)",
			DeliveredAt: now(),
		}, nil
	}

	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	var ogConfig OpsGenieConfig
	if err := json.Unmarshal(config.Config, &ogConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Determine API URL based on region
	baseURL := "https://api.opsgenie.com"
	if ogConfig.Region == "eu" {
		baseURL = "https://api.eu.opsgenie.com"
	}

	// Map severity to OpsGenie priority
	priority := "P3"
	switch alert.Severity {
	case "critical":
		priority = "P1"
	case "high":
		priority = "P2"
	case "medium":
		priority = "P3"
	case "low":
		priority = "P4"
	}

	// Build alert payload
	opsgenieAlert := OpsGenieAlert{
		Message:     alert.Title,
		Alias:       fmt.Sprintf("pganalytics_%d", alert.AlertID),
		Description: alert.Description,
		Priority:    priority,
		Tags:        []string{"pganalytics", alert.Severity},
	}

	// Add details from context if available
	if len(alert.Context) > 0 {
		var ctxMap map[string]interface{}
		if err := json.Unmarshal(alert.Context, &ctxMap); err == nil {
			details := make(map[string]string)
			for k, v := range ctxMap {
				details[k] = fmt.Sprintf("%v", v)
			}
			opsgenieAlert.Details = details
		}
	}

	// Add database and query info if available
	if alert.Database != "" {
		opsgenieAlert.Details["database"] = alert.Database
	}
	if alert.Query != "" {
		opsgenieAlert.Details["query"] = alert.Query
	}

	body, err := json.Marshal(opsgenieAlert)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v2/alerts", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+ogConfig.APIKey)

	// Send request
	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.circuitBreaker.RecordFailure()
		o.logger.Error("OpsGenie POST failed", zap.Error(err))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("OpsGenie POST failed: %v", err),
			DeliveredAt: now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		o.circuitBreaker.RecordFailure()
		o.logger.Error("OpsGenie returned error",
			zap.Int("status_code", resp.StatusCode))
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	// Parse response to get alert ID
	var ogResp OpsGenieResponse
	if err := json.NewDecoder(resp.Body).Decode(&ogResp); err == nil && ogResp.AlertID != "" {
		o.circuitBreaker.RecordSuccess()
		o.logger.Info("OpsGenie notification delivered",
			zap.String("alert_id", ogResp.AlertID))
		return &DeliveryResult{
			Success:     true,
			MessageID:   ogResp.AlertID,
			DeliveredAt: now(),
		}, nil
	}

	o.circuitBreaker.RecordSuccess()
	o.logger.Info("OpsGenie notification delivered")
	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("opsgenie_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (o *OpsGenieChannel) Test(ctx context.Context, config ChannelConfig) error {
	testAlert := &AlertNotification{
		AlertID:     0,
		Title:       "pgAnalytics Test Alert",
		Description: "Test notification from pgAnalytics - connection successful!",
		Severity:    "low",
		Status:      "firing",
		FiredAt:     now(),
	}

	_, err := o.Send(ctx, testAlert, config)
	return err
}

// ============================================================================
// UTILITIES
// ============================================================================

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func now() time.Time {
	return time.Now()
}
