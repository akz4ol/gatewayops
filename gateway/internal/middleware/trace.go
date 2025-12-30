package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// Context keys for trace information.
type contextKey string

const (
	TraceIDKey  contextKey = "trace_id"
	SpanIDKey   contextKey = "span_id"
	StartTimeKey contextKey = "start_time"
)

// TraceID format: tr_{timestamp_hex}_{random_hex}
// SpanID format: sp_{random_hex}

// Trace returns middleware that adds trace context to requests.
func Trace() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for existing trace ID in header (for distributed tracing)
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateTraceID()
			}

			// Always generate new span ID for this request
			spanID := generateSpanID()

			// Add trace info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, TraceIDKey, traceID)
			ctx = context.WithValue(ctx, SpanIDKey, spanID)
			ctx = context.WithValue(ctx, StartTimeKey, time.Now())

			// Add trace headers to response
			w.Header().Set("X-Trace-ID", traceID)
			w.Header().Set("X-Span-ID", spanID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// generateTraceID creates a trace ID in format: tr_{timestamp}_{random}
func generateTraceID() string {
	timestamp := time.Now().UnixMilli()
	random := make([]byte, 8)
	rand.Read(random)
	return "tr_" + hex.EncodeToString([]byte{
		byte(timestamp >> 40),
		byte(timestamp >> 32),
		byte(timestamp >> 24),
		byte(timestamp >> 16),
		byte(timestamp >> 8),
		byte(timestamp),
	}) + "_" + hex.EncodeToString(random)
}

// generateSpanID creates a span ID in format: sp_{random}
func generateSpanID() string {
	random := make([]byte, 8)
	rand.Read(random)
	return "sp_" + hex.EncodeToString(random)
}

// GetTraceID extracts trace ID from context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(TraceIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSpanID extracts span ID from context.
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(SpanIDKey).(string); ok {
		return id
	}
	return ""
}

// GetStartTime extracts request start time from context.
func GetStartTime(ctx context.Context) time.Time {
	if t, ok := ctx.Value(StartTimeKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}
