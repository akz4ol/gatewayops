package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/rs/zerolog"
)

// RateLimiter defines the interface for rate limiting.
type RateLimiter interface {
	// Allow checks if a request is allowed under the rate limit.
	// Returns (allowed, remaining, resetSeconds, error)
	Allow(ctx context.Context, key string, limit int) (bool, int, int, error)
}

// RateLimit returns middleware that enforces rate limits.
func RateLimit(limiter RateLimiter, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get auth info for rate limit key
			authInfo := GetAuthInfo(r.Context())
			if authInfo == nil {
				// No auth info means no rate limiting (shouldn't happen if auth middleware runs first)
				next.ServeHTTP(w, r)
				return
			}

			// Rate limit key: org_id:key_id
			key := fmt.Sprintf("%s:%s", authInfo.OrgID, authInfo.KeyID)
			limit := authInfo.RateLimit
			if limit == 0 {
				limit = 1000 // Default 1000 requests per minute
			}

			allowed, remaining, resetSeconds, err := limiter.Allow(r.Context(), key, limit)
			if err != nil {
				logger.Error().
					Err(err).
					Str("rate_limit_key", key).
					Msg("Rate limiter error")
				// On error, allow the request but log it
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(resetSeconds))

			if !allowed {
				logger.Warn().
					Str("rate_limit_key", key).
					Int("limit", limit).
					Msg("Rate limit exceeded")

				w.Header().Set("Retry-After", strconv.Itoa(resetSeconds))
				handler.WriteError(w, http.StatusTooManyRequests, "rate_limit_exceeded",
					fmt.Sprintf("Rate limit exceeded. Try again in %d seconds", resetSeconds))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
