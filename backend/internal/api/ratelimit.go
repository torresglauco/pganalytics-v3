package api

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiting algorithm
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	capacity int       // tokens per minute
	refill   int       // tokens added per second
	interval time.Duration
}

// TokenBucket represents a single rate limit bucket for a client
type TokenBucket struct {
	tokens    float64
	lastRefill time.Time
	capacity  int
}

// NewRateLimiter creates a new rate limiter
// capacity: max requests per minute
func NewRateLimiter(capacity int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		refill:   capacity / 60, // Distribute tokens evenly over 60 seconds
		interval: time.Second,
	}
	return rl
}

// Allow checks if a request from the given client is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientID]
	if !exists {
		bucket = &TokenBucket{
			tokens:    float64(rl.capacity),
			lastRefill: time.Now(),
			capacity:  rl.capacity,
		}
		rl.buckets[clientID] = bucket
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	tokensToAdd := elapsed * float64(rl.refill)
	bucket.tokens = min(bucket.tokens+tokensToAdd, float64(bucket.capacity))
	bucket.lastRefill = now

	// Check if we have tokens available
	if bucket.tokens >= 1.0 {
		bucket.tokens--
		return true
	}

	return false
}

// Reset clears all buckets (for testing)
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.buckets = make(map[string]*TokenBucket)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
