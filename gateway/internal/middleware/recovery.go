// Package middleware provides HTTP middleware for the gateway.
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/rs/zerolog"
)

// Recoverer returns middleware that recovers from panics.
func Recoverer(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()
					logger.Error().
						Interface("panic", rec).
						Bytes("stack", stack).
						Str("method", r.Method).
						Str("path", r.URL.Path).
						Msg("Panic recovered")

					handler.WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
