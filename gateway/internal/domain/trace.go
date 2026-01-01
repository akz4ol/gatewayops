package domain

import (
	"time"

	"github.com/google/uuid"
)

// Trace represents a single MCP request trace.
type Trace struct {
	ID          uuid.UUID         `json:"id"`
	TraceID     string            `json:"trace_id"`
	SpanID      string            `json:"span_id"`
	ParentID    string            `json:"parent_id,omitempty"`
	OrgID       uuid.UUID         `json:"org_id"`
	TeamID      *uuid.UUID        `json:"team_id,omitempty"`
	APIKeyID    uuid.UUID         `json:"api_key_id"`
	MCPServer   string            `json:"mcp_server"`
	Operation   string            `json:"operation"` // tools/call, resources/read, etc.
	ToolName    string            `json:"tool_name,omitempty"`
	Status      string            `json:"status"` // success, error, timeout
	StatusCode  int               `json:"status_code"`
	DurationMs  int64             `json:"duration_ms"`
	RequestSize int               `json:"request_size"`
	ResponseSize int              `json:"response_size"`
	Cost        float64           `json:"cost"`
	ErrorMsg    string            `json:"error_msg,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// TraceSpan represents a span within a trace.
type TraceSpan struct {
	ID         uuid.UUID         `json:"id"`
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Name       string            `json:"name"`
	Kind       string            `json:"kind"` // client, server, internal
	Status     string            `json:"status"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	DurationMs int64             `json:"duration_ms"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// TraceDetail includes a trace with all its spans.
type TraceDetail struct {
	Trace Trace       `json:"trace"`
	Spans []TraceSpan `json:"spans"`
}

// TraceFilter represents filters for querying traces.
type TraceFilter struct {
	OrgID     uuid.UUID  `json:"org_id"`
	TeamID    *uuid.UUID `json:"team_id,omitempty"`
	MCPServer string     `json:"mcp_server,omitempty"`
	Operation string     `json:"operation,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// TraceStats represents aggregated trace statistics.
type TraceStats struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessCount    int64   `json:"success_count"`
	ErrorCount      int64   `json:"error_count"`
	AvgDurationMs   float64 `json:"avg_duration_ms"`
	P50DurationMs   float64 `json:"p50_duration_ms"`
	P95DurationMs   float64 `json:"p95_duration_ms"`
	P99DurationMs   float64 `json:"p99_duration_ms"`
	TotalCost       float64 `json:"total_cost"`
	ErrorRate       float64 `json:"error_rate"`
}
