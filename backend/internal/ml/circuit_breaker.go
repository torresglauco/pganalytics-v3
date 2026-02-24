package ml

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

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

// CircuitBreaker implements the circuit breaker pattern for ML service resilience
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

// NewCircuitBreaker creates a new circuit breaker
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
		return true
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
			return true
		}
		return false
	}

	// Half-open state
	return true
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
