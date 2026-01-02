// Package ratelimit provides rate limiting using Redis.
package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/database"
	"github.com/rs/zerolog"
)

// Limiter implements rate limiting using Redis.
type Limiter struct {
	redis  *database.Redis
	logger zerolog.Logger
	window time.Duration
}

// NewLimiter creates a new Redis-backed rate limiter.
func NewLimiter(redis *database.Redis, logger zerolog.Logger) *Limiter {
	logger.Info().Msg("Rate limiter initialized with Redis backend")
	return &Limiter{
		redis:  redis,
		logger: logger,
		window: time.Minute,
	}
}

// Allow checks if a request is allowed under the rate limit.
// Returns: allowed, remaining, reset (seconds), error
func (l *Limiter) Allow(ctx context.Context, key string, limit int) (bool, int, int, error) {
	if l.redis == nil || l.redis.Client == nil {
		// Fallback: allow all requests if Redis is unavailable
		l.logger.Warn().Msg("Redis unavailable, allowing request")
		return true, limit, 60, nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	now := time.Now()
	windowStart := now.Truncate(l.window)
	windowEnd := windowStart.Add(l.window)
	resetSeconds := int(windowEnd.Sub(now).Seconds())

	// Use Redis INCR with expiration for sliding window
	count, err := l.redis.Incr(ctx, redisKey)
	if err != nil {
		l.logger.Error().Err(err).Str("key", key).Msg("Failed to increment rate limit counter")
		// Fallback: allow request on error
		return true, limit, resetSeconds, nil
	}

	// Set expiration on first request
	if count == 1 {
		if err := l.redis.Expire(ctx, redisKey, l.window); err != nil {
			l.logger.Error().Err(err).Str("key", key).Msg("Failed to set expiration on rate limit key")
		}
	}

	// Get TTL for reset time
	ttl, err := l.redis.TTL(ctx, redisKey)
	if err == nil && ttl > 0 {
		resetSeconds = int(ttl.Seconds())
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	if int(count) > limit {
		return false, 0, resetSeconds, nil
	}

	return true, remaining, resetSeconds, nil
}

// AllowWithBurst checks rate limit with burst capacity.
func (l *Limiter) AllowWithBurst(ctx context.Context, key string, limit, burst int) (bool, int, int, error) {
	// For burst, we use a token bucket algorithm
	// This is a simplified version using the same sliding window
	return l.Allow(ctx, key, limit+burst)
}

// GetUsage returns current usage for a key.
func (l *Limiter) GetUsage(ctx context.Context, key string) (int, error) {
	if l.redis == nil || l.redis.Client == nil {
		return 0, nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	val, err := l.redis.Get(ctx, redisKey)
	if err != nil {
		return 0, nil // Key doesn't exist yet
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Reset clears the rate limit for a key.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	if l.redis == nil || l.redis.Client == nil {
		return nil
	}

	redisKey := fmt.Sprintf("ratelimit:%s", key)
	return l.redis.Del(ctx, redisKey)
}

// Health checks if rate limiter is healthy.
func (l *Limiter) Health() bool {
	if l.redis == nil {
		return false
	}
	return l.redis.Health()
}

// Ready checks if rate limiter is ready.
func (l *Limiter) Ready() bool {
	if l.redis == nil {
		return false
	}
	return l.redis.Ready()
}
