package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestCircuitBreakerBasicState tests the basic state transitions
func TestCircuitBreakerBasicState(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	if cb.State() != string(StateClosed) {
		t.Errorf("Expected initial state to be closed, got %s", cb.State())
	}

	if cb.IsOpen() {
		t.Errorf("Expected circuit to be closed (IsOpen=false), got open")
	}
}

// TestCircuitBreakerFailureThreshold tests opening circuit on failures
func TestCircuitBreakerFailureThreshold(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Record 4 failures - should not open yet
	for i := 0; i < 4; i++ {
		cb.RecordFailure()
		if cb.IsOpen() {
			t.Errorf("Circuit should not open until 5 failures, but opened after %d", i+1)
		}
	}

	// 5th failure should open the circuit
	cb.RecordFailure()
	if !cb.IsOpen() {
		t.Errorf("Circuit should be open after 5 failures")
	}
}

// TestCircuitBreakerHalfOpenRecovery tests recovery from open state
func TestCircuitBreakerHalfOpenRecovery(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Errorf("Circuit should be open")
	}

	// Set timeout to 0 for testing
	cb.timeout = 0 * time.Second
	cb.lastFailureTime = time.Now().Add(-1 * time.Second)

	// Next call should transition to half-open
	cb.IsOpen()
	if cb.State() != string(StateHalfOpen) {
		t.Errorf("Expected state to be half-open after timeout, got %s", cb.State())
	}
}

// TestCircuitBreakerSuccessesInHalfOpen tests closing circuit after successes in half-open
func TestCircuitBreakerSuccessesInHalfOpen(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Transition to half-open
	cb.timeout = 0 * time.Second
	cb.lastFailureTime = time.Now().Add(-1 * time.Second)
	cb.IsOpen()

	if cb.State() != string(StateHalfOpen) {
		t.Errorf("Expected state to be half-open, got %s", cb.State())
	}

	// Record 2 successes - should not close yet
	for i := 0; i < 2; i++ {
		cb.RecordSuccess()
		if cb.State() == string(StateClosed) {
			t.Errorf("Circuit should not close until 3 successes, but closed after %d", i+1)
		}
	}

	// 3rd success should close the circuit
	cb.RecordSuccess()
	if cb.State() != string(StateClosed) {
		t.Errorf("Circuit should be closed after 3 successes in half-open state")
	}
}

// TestCircuitBreakerFailureInHalfOpen tests re-opening circuit on failure in half-open
func TestCircuitBreakerFailureInHalfOpen(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	// Transition to half-open
	cb.timeout = 0 * time.Second
	cb.lastFailureTime = time.Now().Add(-1 * time.Second)
	cb.IsOpen()

	if cb.State() != string(StateHalfOpen) {
		t.Errorf("Expected state to be half-open, got %s", cb.State())
	}

	// Record a failure - should re-open
	cb.RecordFailure()
	if cb.State() != string(StateOpen) {
		t.Errorf("Circuit should be re-opened after failure in half-open state")
	}
}

// TestCircuitBreakerReset tests reset functionality
func TestCircuitBreakerReset(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Open the circuit
	for i := 0; i < 5; i++ {
		cb.RecordFailure()
	}

	if !cb.IsOpen() {
		t.Errorf("Circuit should be open")
	}

	// Reset
	cb.Reset()

	if cb.State() != string(StateClosed) {
		t.Errorf("Expected state to be closed after reset, got %s", cb.State())
	}

	if cb.IsOpen() {
		t.Errorf("Expected IsOpen() to return false after reset")
	}
}

// TestSlackChannelCircuitBreaker tests Slack channel circuit breaker
func TestSlackChannelCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	channel := NewSlackChannel(httpClient, logger, 10*time.Second).(*SlackChannel)

	// Create config with test server URL
	slackConfig := SlackConfig{
		WebhookURL: server.URL,
	}
	configJSON, _ := json.Marshal(slackConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "slack",
		Config: configJSON,
	}

	// Create test alert
	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Record failures until circuit opens
	for i := 0; i < 5; i++ {
		result, _ := channel.Send(context.Background(), alert, config)
		if i < 4 && !channel.circuitBreaker.IsOpen() {
			// Before opening, should return error from server
			if result.Success {
				t.Errorf("Expected failure on attempt %d", i+1)
			}
		}
	}

	// Next call should return circuit open error without calling server
	result, _ := channel.Send(context.Background(), alert, config)
	if result.Success || !contains(result.ErrorMsg, "circuit open") {
		t.Errorf("Expected circuit open error, got: %s", result.ErrorMsg)
	}
}

// TestEmailChannelCircuitBreaker tests Email channel circuit breaker
func TestEmailChannelCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewProduction()
	channel := NewEmailChannel(logger, 10*time.Second).(*EmailChannel)

	// Create config
	emailConfig := EmailConfig{
		Recipients: []string{"test@example.com"},
	}
	configJSON, _ := json.Marshal(emailConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "email",
		Config: configJSON,
	}

	// Create test alert
	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Email channel should succeed (placeholder implementation)
	result, _ := channel.Send(context.Background(), alert, config)
	if !result.Success {
		t.Errorf("Email channel should succeed with valid config")
	}

	if channel.circuitBreaker.State() != string(StateClosed) {
		t.Errorf("Circuit breaker should remain closed after success")
	}
}

// TestWebhookChannelCircuitBreaker tests Webhook channel circuit breaker
func TestWebhookChannelCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create a test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	channel := NewWebhookChannel(httpClient, logger, 10*time.Second).(*WebhookChannel)

	// Create config with test server URL
	webhookConfig := WebhookConfig{
		URL:    server.URL,
		Method: "POST",
	}
	configJSON, _ := json.Marshal(webhookConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "webhook",
		Config: configJSON,
	}

	// Create test alert
	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Record failures
	for i := 0; i < 5; i++ {
		result, _ := channel.Send(context.Background(), alert, config)
		if i < 4 && !channel.circuitBreaker.IsOpen() {
			if result.Success {
				t.Errorf("Expected failure on attempt %d", i+1)
			}
		}
	}

	// Verify circuit is open
	if !channel.circuitBreaker.IsOpen() {
		t.Errorf("Circuit breaker should be open after 5 failures")
	}
}

// TestPagerDutyChannelCircuitBreaker tests PagerDuty channel circuit breaker
func TestPagerDutyChannelCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewProduction()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	channel := NewPagerDutyChannel(httpClient, logger, 10*time.Second).(*PagerDutyChannel)

	// Initial state should be closed
	if channel.circuitBreaker.IsOpen() {
		t.Errorf("Circuit should initially be closed")
	}

	// Verify we can check state before open
	state := channel.circuitBreaker.State()
	if state != string(StateClosed) {
		t.Errorf("Expected closed state, got %s", state)
	}
}

// TestJiraChannelCircuitBreaker tests Jira channel circuit breaker
func TestJiraChannelCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create a test server that times out
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 100 * time.Millisecond}
	channel := NewJiraChannel(httpClient, logger, 100*time.Millisecond).(*JiraChannel)

	// Create config with test server URL
	jiraConfig := JiraConfig{
		URL:          server.URL,
		ProjectKey:   "TEST",
		AuthUsername: "user",
		AuthToken:    "token",
	}
	configJSON, _ := json.Marshal(jiraConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "jira",
		Config: configJSON,
	}

	// Create test alert
	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Record timeouts
	for i := 0; i < 5; i++ {
		result, _ := channel.Send(context.Background(), alert, config)
		if i < 4 && !channel.circuitBreaker.IsOpen() {
			if result.Success {
				t.Errorf("Expected timeout on attempt %d", i+1)
			}
		}
	}

	// Verify circuit is open
	if !channel.circuitBreaker.IsOpen() {
		t.Errorf("Circuit breaker should be open after 5 failures")
	}
}

// TestCircuitBreakerMetrics tests metrics collection
func TestCircuitBreakerMetrics(t *testing.T) {
	logger, _ := zap.NewProduction()
	cb := NewCircuitBreaker(logger)

	// Get initial metrics
	metrics := cb.GetMetrics()

	if state, ok := metrics["state"]; !ok || state != string(StateClosed) {
		t.Errorf("Expected state in metrics")
	}

	if failCount, ok := metrics["failure_count"]; !ok || failCount != 0 {
		t.Errorf("Expected failure_count in metrics")
	}

	// Record some failures
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	metrics = cb.GetMetrics()
	if failCount := metrics["failure_count"]; failCount != 3 {
		t.Errorf("Expected 3 failures in metrics, got %v", failCount)
	}
}

// TestChannelTimeoutContext tests that context timeout is applied
func TestChannelTimeoutContext(t *testing.T) {
	logger, _ := zap.NewProduction()

	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	channel := NewSlackChannel(httpClient, logger, 100*time.Millisecond).(*SlackChannel)

	slackConfig := SlackConfig{
		WebhookURL: server.URL,
	}
	configJSON, _ := json.Marshal(slackConfig)
	config := ChannelConfig{
		ID:     1,
		Type:   "slack",
		Config: configJSON,
	}

	alert := &AlertNotification{
		AlertID:     1,
		Title:       "Test Alert",
		Description: "Test",
		Severity:    "high",
		Status:      "firing",
		FiredAt:     time.Now(),
	}

	// Send should timeout due to channel timeout
	ctx := context.Background()
	startTime := time.Now()
	result, _ := channel.Send(ctx, alert, config)
	elapsed := time.Since(startTime)

	// Should timeout quickly (within 500ms, well before the 5 second server delay)
	if elapsed > 500*time.Millisecond {
		t.Errorf("Expected timeout to occur within 500ms, took %v", elapsed)
	}

	if result.Success {
		t.Errorf("Expected failure due to timeout")
	}

	// Circuit should have recorded the failure
	if !channel.circuitBreaker.IsOpen() {
		// After one failure, circuit should still be closed
		// After 5 failures, it should be open
		metrics := channel.circuitBreaker.GetMetrics()
		if failCount := metrics["failure_count"]; failCount != 1 {
			t.Errorf("Expected 1 failure recorded, got %v", failCount)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
