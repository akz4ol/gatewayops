package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	size        int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Logger returns middleware that logs HTTP requests.
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := wrapResponseWriter(w)

			// Get request ID from context (set by chi middleware)
			requestID := chimiddleware.GetReqID(r.Context())

			// Log request start
			logger.Debug().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Request started")

			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			event := logger.Info()

			// Use different log levels based on status code
			if wrapped.status >= 500 {
				event = logger.Error()
			} else if wrapped.status >= 400 {
				event = logger.Warn()
			}

			event.
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.status).
				Int("size", wrapped.size).
				Dur("duration", duration).
				Str("remote_addr", r.RemoteAddr).
				Msg("Request completed")
		})
	}
}
