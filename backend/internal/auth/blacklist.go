package auth

import (
	"context"
	"time"
)

// TokenBlacklist interface defines token revocation operations
type TokenBlacklist interface {
	// RevokeToken adds a token to the blacklist until expiration
	RevokeToken(ctx context.Context, token string, expiresAt time.Time) error

	// IsBlacklisted checks if a token is revoked
	IsBlacklisted(ctx context.Context, token string) (bool, error)
}

// InMemoryBlacklist is a simple in-memory implementation for development
type InMemoryBlacklist struct {
	revokedTokens map[string]time.Time
}

// NewInMemoryBlacklist creates a new in-memory token blacklist
func NewInMemoryBlacklist() *InMemoryBlacklist {
	return &InMemoryBlacklist{
		revokedTokens: make(map[string]time.Time),
	}
}

// RevokeToken adds a token to the blacklist until expiration
func (b *InMemoryBlacklist) RevokeToken(ctx context.Context, token string, expiresAt time.Time) error {
	b.revokedTokens[token] = expiresAt
	return nil
}

// IsBlacklisted checks if a token is revoked
func (b *InMemoryBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if expiresAt, exists := b.revokedTokens[token]; exists {
		// Token is blacklisted if expiration time hasn't passed yet
		if time.Now().Before(expiresAt) {
			return true, nil
		}
		// Clean up expired entries
		delete(b.revokedTokens, token)
	}
	return false, nil
}

// RedisBlacklist is a Redis-backed implementation for production
// Note: Full Redis implementation requires redis package
// This is a stub showing the interface
type RedisBlacklist struct {
	// redisClient *redis.Client
}

// NewRedisBlacklist creates a new Redis token blacklist
func NewRedisBlacklist() *RedisBlacklist {
	// Would initialize Redis connection here
	return &RedisBlacklist{}
}

// RevokeToken adds a token to the Redis blacklist
func (b *RedisBlacklist) RevokeToken(ctx context.Context, token string, expiresAt time.Time) error {
	// Implementation would use Redis SET with EX (expiration)
	// ttl := time.Until(expiresAt)
	// return b.redisClient.Set(ctx, "blacklist:"+token, "true", ttl).Err()
	return nil
}

// IsBlacklisted checks if a token is revoked in Redis
func (b *RedisBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	// Implementation would use Redis GET
	// val, err := b.redisClient.Get(ctx, "blacklist:"+token).Result()
	// if err == redis.Nil {
	//     return false, nil
	// }
	// return val == "true", err
	return false, nil
}
