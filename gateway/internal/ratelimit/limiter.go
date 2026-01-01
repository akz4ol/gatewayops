// Package ratelimit provides rate limiting.
package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Limiter implements middleware.RateLimiter using in-memory storage for demo mode.
type Limiter struct {
	logger zerolog.Logger
	window time.Duration
	mu     sync.Mutex
	counts map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

// NewLimiter creates a new in-memory rate limiter.
func NewLimiter(client interface{}, logger zerolog.Logger) *Limiter {
	logger.Info().Msg("Demo mode: Using in-memory rate limiter")
	return &Limiter{
		logger: logger,
		window: time.Minute,
		counts: make(map[string]*rateLimitEntry),
	}
}

// Allow checks if a request is allowed under the rate limit.
func (l *Limiter) Allow(ctx context.Context, key string, limit int) (bool, int, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	entry, exists := l.counts[key]
	if !exists || now.After(entry.resetTime) {
		// Start new window
		l.counts[key] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(l.window),
		}
		return true, limit - 1, int(l.window.Seconds()), nil
	}

	if entry.count >= limit {
		resetSeconds := int(entry.resetTime.Sub(now).Seconds())
		if resetSeconds < 0 {
			resetSeconds = 0
		}
		return false, 0, resetSeconds, nil
	}

	entry.count++
	remaining := limit - entry.count
	resetSeconds := int(entry.resetTime.Sub(now).Seconds())

	return true, remaining, resetSeconds, nil
}

// Health checks if rate limiter is healthy.
func (l *Limiter) Health() bool {
	return true // Demo mode always healthy
}

// Ready checks if rate limiter is ready.
func (l *Limiter) Ready() bool {
	return true // Demo mode always ready
}
