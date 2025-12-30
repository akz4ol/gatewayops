// Package ratelimit provides rate limiting using Redis.
package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Limiter implements middleware.RateLimiter using Redis sliding window.
type Limiter struct {
	client *redis.Client
	logger zerolog.Logger
	window time.Duration
}

// NewLimiter creates a new Redis-based rate limiter.
func NewLimiter(client *redis.Client, logger zerolog.Logger) *Limiter {
	return &Limiter{
		client: client,
		logger: logger,
		window: time.Minute, // 1 minute sliding window
	}
}

// Allow checks if a request is allowed under the rate limit.
// Uses Redis sliding window algorithm for accurate rate limiting.
// Returns (allowed, remaining, resetSeconds, error)
func (l *Limiter) Allow(ctx context.Context, key string, limit int) (bool, int, int, error) {
	now := time.Now()
	windowStart := now.Add(-l.window)

	// Redis key for this rate limit window
	redisKey := "ratelimit:" + key

	// Use a Lua script for atomic operations
	script := redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local window_start = now - window

		-- Remove old entries outside the window
		redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)

		-- Count current requests in window
		local count = redis.call('ZCARD', key)

		if count < limit then
			-- Add current request
			redis.call('ZADD', key, now, now .. ':' .. math.random())
			-- Set expiry on the key
			redis.call('PEXPIRE', key, window)
			return {1, limit - count - 1, window / 1000}
		else
			-- Get oldest entry to calculate reset time
			local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
			local reset = 0
			if #oldest > 0 then
				reset = math.ceil((oldest[2] + window - now) / 1000)
			end
			return {0, 0, reset}
		end
	`)

	result, err := script.Run(ctx, l.client, []string{redisKey},
		now.UnixMilli(),
		l.window.Milliseconds(),
		limit,
	).Int64Slice()

	if err != nil {
		l.logger.Error().Err(err).Str("key", key).Msg("Rate limiter script error")
		// On error, allow the request (fail open)
		return true, limit, 60, nil
	}

	allowed := result[0] == 1
	remaining := int(result[1])
	resetSeconds := int(result[2])

	return allowed, remaining, resetSeconds, nil
}

// SimpleAllow provides a simpler rate limiting using INCR with expiry.
// Less accurate but lower Redis overhead.
func (l *Limiter) SimpleAllow(ctx context.Context, key string, limit int) (bool, int, int, error) {
	redisKey := "ratelimit:" + key

	pipe := l.client.Pipeline()
	incrCmd := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, l.window)
	_, err := pipe.Exec(ctx)

	if err != nil {
		l.logger.Error().Err(err).Str("key", key).Msg("Rate limiter error")
		return true, limit, 60, nil
	}

	count := int(incrCmd.Val())
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	allowed := count <= limit
	resetSeconds := int(l.window.Seconds())

	return allowed, remaining, resetSeconds, nil
}

// GetUsage returns current usage for a key.
func (l *Limiter) GetUsage(ctx context.Context, key string) (int, error) {
	redisKey := "ratelimit:" + key

	now := time.Now()
	windowStart := now.Add(-l.window)

	// Remove old entries and count
	pipe := l.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(windowStart.UnixMilli(), 10))
	countCmd := pipe.ZCard(ctx, redisKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return 0, err
	}

	return int(countCmd.Val()), nil
}

// Reset clears the rate limit for a key.
func (l *Limiter) Reset(ctx context.Context, key string) error {
	redisKey := "ratelimit:" + key
	return l.client.Del(ctx, redisKey).Err()
}

// Health checks if Redis is healthy.
func (l *Limiter) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return l.client.Ping(ctx).Err() == nil
}

// Ready checks if Redis is ready to accept requests.
func (l *Limiter) Ready() bool {
	return l.Health()
}
