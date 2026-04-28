package api

import (
	"sync"
	"time"
)

// EndpointRateLimiter implements per-endpoint rate limiting with configurable limits
type EndpointRateLimiter struct {
	mu        sync.RWMutex
	limiters  map[string]*RateLimiter
	endpoints map[string]RateLimitConfig
}

// RateLimitConfig defines rate limit configuration for an endpoint
type RateLimitConfig struct {
	RequestsPerMinute int           // Max requests per minute
	BurstSize         int           // Burst allowance (tokens to add immediately)
	CleanupInterval   time.Duration // How often to cleanup inactive buckets
}

// NewEndpointRateLimiter creates a new endpoint rate limiter with default configurations
func NewEndpointRateLimiter() *EndpointRateLimiter {
	erl := &EndpointRateLimiter{
		limiters:  make(map[string]*RateLimiter),
		endpoints: make(map[string]RateLimitConfig),
	}

	// Configure rate limits for different endpoint categories
	erl.RegisterEndpoint("/api/v1/metrics/push", RateLimitConfig{
		RequestsPerMinute: 10000, // High volume for collector metrics
		BurstSize:         500,
		CleanupInterval:   5 * time.Minute,
	})

	erl.RegisterEndpoint("/api/v1/config/refresh", RateLimitConfig{
		RequestsPerMinute: 500, // Moderate for config refreshes
		BurstSize:         50,
		CleanupInterval:   5 * time.Minute,
	})

	erl.RegisterEndpoint("/api/v1/collectors/register", RateLimitConfig{
		RequestsPerMinute: 100, // Low for registrations
		BurstSize:         10,
		CleanupInterval:   10 * time.Minute,
	})

	erl.RegisterEndpoint("/api/v1/collectors/refresh-token", RateLimitConfig{
		RequestsPerMinute: 500, // Moderate for token refreshes
		BurstSize:         50,
		CleanupInterval:   5 * time.Minute,
	})

	erl.RegisterEndpoint("/api/v1/auth/*", RateLimitConfig{
		RequestsPerMinute: 1000, // Higher for auth endpoints
		BurstSize:         100,
		CleanupInterval:   5 * time.Minute,
	})

	erl.RegisterEndpoint("default", RateLimitConfig{
		RequestsPerMinute: 1000, // Default for other endpoints
		BurstSize:         100,
		CleanupInterval:   5 * time.Minute,
	})

	return erl
}

// RegisterEndpoint registers rate limit configuration for an endpoint
func (erl *EndpointRateLimiter) RegisterEndpoint(endpoint string, config RateLimitConfig) {
	erl.mu.Lock()
	defer erl.mu.Unlock()
	erl.endpoints[endpoint] = config
	if _, exists := erl.limiters[endpoint]; !exists {
		erl.limiters[endpoint] = NewRateLimiter(config.RequestsPerMinute)
	}
}

// Allow checks if a request is allowed for the given endpoint and client
func (erl *EndpointRateLimiter) Allow(endpoint, clientID string) bool {
	erl.mu.RLock()

	// Find matching endpoint configuration
	limiter, exists := erl.limiters[endpoint]
	if !exists {
		// Try default limiter
		limiter, exists = erl.limiters["default"]
		if !exists {
			erl.mu.RUnlock()
			// If no limiter exists, allow by default (fail open)
			return true
		}
	}
	erl.mu.RUnlock()

	return limiter.Allow(clientID)
}

// GetStats returns current rate limiter statistics for monitoring
func (erl *EndpointRateLimiter) GetStats(endpoint, clientID string) map[string]interface{} {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	limiter, exists := erl.limiters[endpoint]
	if !exists {
		limiter, exists = erl.limiters["default"]
		if !exists {
			return nil
		}
	}

	limiter.mu.RLock()
	defer limiter.mu.RUnlock()

	bucket, exists := limiter.buckets[clientID]
	if !exists {
		return map[string]interface{}{
			"tokens":            float64(limiter.capacity),
			"capacity":          limiter.capacity,
			"client_id":         clientID,
			"endpoint":          endpoint,
			"refill_per_second": float64(limiter.refill),
		}
	}

	return map[string]interface{}{
		"tokens":            bucket.tokens,
		"capacity":          bucket.capacity,
		"client_id":         clientID,
		"endpoint":          endpoint,
		"last_refill":       bucket.lastRefill,
		"refill_per_second": float64(limiter.refill),
	}
}

// Cleanup removes inactive buckets from all limiters to prevent memory leaks
func (erl *EndpointRateLimiter) Cleanup(maxInactivityDuration time.Duration) {
	erl.mu.RLock()
	limiters := erl.limiters
	erl.mu.RUnlock()

	cutoffTime := time.Now().Add(-maxInactivityDuration)

	for _, limiter := range limiters {
		limiter.mu.Lock()
		for clientID, bucket := range limiter.buckets {
			if bucket.lastRefill.Before(cutoffTime) {
				delete(limiter.buckets, clientID)
			}
		}
		limiter.mu.Unlock()
	}
}

// CollectorRateLimitKey generates a rate limit key for a collector
// Groups by collector ID for rate limiting per collector
func CollectorRateLimitKey(collectorID string) string {
	return "collector:" + collectorID
}

// UserRateLimitKey generates a rate limit key for a user
// Groups by user ID for rate limiting per user
func UserRateLimitKey(userID string) string {
	return "user:" + userID
}

// IPRateLimitKey generates a rate limit key for an IP address
// Groups by IP for rate limiting per IP
func IPRateLimitKey(ipAddress string) string {
	return "ip:" + ipAddress
}
