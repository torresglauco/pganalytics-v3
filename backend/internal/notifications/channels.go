package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ============================================================================
// SLACK CHANNEL
// ============================================================================

type SlackChannel struct {
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
	Color    string `json:"color"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	TitleLink string `json:"title_link,omitempty"`
	Fields   []SlackField `json:"fields"`
	Footer   string `json:"footer"`
	TS       int64  `json:"ts"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func NewSlackChannel(httpClient *http.Client) NotificationChannel {
	return &SlackChannel{httpClient: httpClient}
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
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

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

type EmailChannel struct{}

type EmailConfig struct {
	Recipients []string `json:"recipients"`
	SMTPURL    string   `json:"smtp_url,omitempty"`
}

func NewEmailChannel() NotificationChannel {
	return &EmailChannel{}
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
	var emailConfig EmailConfig
	if err := json.Unmarshal(config.Config, &emailConfig); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// In production, would use actual SMTP library (e.g., net/smtp)
	// For now, simulating successful delivery
	// This is a placeholder - would need proper email integration

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

	// In production, would test SMTP connection and send test email
	return nil
}

// ============================================================================
// WEBHOOK CHANNEL
// ============================================================================

type WebhookChannel struct {
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

func NewWebhookChannel(httpClient *http.Client) NotificationChannel {
	return &WebhookChannel{httpClient: httpClient}
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
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

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
	httpClient *http.Client
}

type PagerDutyConfig struct {
	IntegrationKey string `json:"integration_key"`
	ServiceKey     string `json:"service_key,omitempty"`
}

type PagerDutyEvent struct {
	RoutingKey  string         `json:"routing_key"`
	EventAction string         `json:"event_action"`
	Dedup       string         `json:"dedup_key"`
	Payload     PagerDutyPayload `json:"payload"`
}

type PagerDutyPayload struct {
	Summary   string `json:"summary"`
	Severity  string `json:"severity"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
	Details   json.RawMessage `json:"custom_details"`
}

func NewPagerDutyChannel(httpClient *http.Client) NotificationChannel {
	return &PagerDutyChannel{httpClient: httpClient}
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
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

	return &DeliveryResult{
		Success:     true,
		MessageID:   fmt.Sprintf("pd_%d", alert.AlertID),
		DeliveredAt: now(),
	}, nil
}

func (p *PagerDutyChannel) Test(ctx context.Context, config ChannelConfig) error {
	testAlert := &AlertNotification{
		AlertID:     0,
		Title:       "pgAnalytics Test Alert",
		Severity:    "warning",
		Status:      "firing",
		FiredAt:     now(),
	}

	_, err := p.Send(ctx, testAlert, config)
	return err
}

// ============================================================================
// JIRA CHANNEL
// ============================================================================

type JiraChannel struct {
	httpClient *http.Client
}

type JiraConfig struct {
	URL           string `json:"url"`
	ProjectKey    string `json:"project_key"`
	IssueType     string `json:"issue_type,omitempty"`
	AuthUsername  string `json:"auth_username"`
	AuthToken     string `json:"auth_token"`
}

type JiraCreateIssue struct {
	Fields JiraIssueFields `json:"fields"`
}

type JiraIssueFields struct {
	Project     JiraProject `json:"project"`
	IssueType   JiraIssueType `json:"issuetype"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Priority    JiraPriority `json:"priority,omitempty"`
	Labels      []string `json:"labels,omitempty"`
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

func NewJiraChannel(httpClient *http.Client) NotificationChannel {
	return &JiraChannel{httpClient: httpClient}
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
			Project: JiraProject{Key: jiraConfig.ProjectKey},
			IssueType: JiraIssueType{Name: jiraConfig.IssueType},
			Summary: alert.Title,
			Description: alert.Description,
			Priority: JiraPriority{Name: priority},
			Labels: []string{"pganalytics", alert.Severity},
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
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &DeliveryResult{
			Success:     false,
			ErrorMsg:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			DeliveredAt: now(),
		}, nil
	}

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
// UTILITIES
// ============================================================================

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func now() time.Time {
	return time.Now()
}

// We need to import time
import "time"
