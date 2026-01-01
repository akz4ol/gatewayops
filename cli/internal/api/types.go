package api

import "time"

// ToolDefinition represents an MCP tool.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ToolCallResult represents the result of a tool call.
type ToolCallResult struct {
	Content    interface{}            `json:"content"`
	IsError    bool                   `json:"isError"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	TraceID    string                 `json:"traceId,omitempty"`
	SpanID     string                 `json:"spanId,omitempty"`
	DurationMs int64                  `json:"durationMs,omitempty"`
	Cost       float64                `json:"cost,omitempty"`
}

// Resource represents an MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceContent represents the content of an MCP resource.
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// Prompt represents an MCP prompt.
type Prompt struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Arguments   []map[string]interface{} `json:"arguments,omitempty"`
}

// PromptMessage represents a message in a prompt response.
type PromptMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Span represents a trace span.
type Span struct {
	ID           string                 `json:"id"`
	TraceID      string                 `json:"traceId"`
	ParentSpanID string                 `json:"parentSpanId,omitempty"`
	Name         string                 `json:"name"`
	Kind         string                 `json:"kind"`
	Status       string                 `json:"status"`
	StartTime    time.Time              `json:"startTime"`
	EndTime      *time.Time             `json:"endTime,omitempty"`
	DurationMs   int64                  `json:"durationMs,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// Trace represents a distributed trace.
type Trace struct {
	ID           string     `json:"id"`
	OrgID        string     `json:"orgId"`
	APIKeyID     string     `json:"apiKeyId,omitempty"`
	MCPServer    string     `json:"mcpServer"`
	Operation    string     `json:"operation"`
	Status       string     `json:"status"`
	StartTime    time.Time  `json:"startTime"`
	EndTime      *time.Time `json:"endTime,omitempty"`
	DurationMs   int64      `json:"durationMs,omitempty"`
	Spans        []Span     `json:"spans,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
	Cost         float64    `json:"cost,omitempty"`
}

// TracePage represents a paginated list of traces.
type TracePage struct {
	Traces  []Trace `json:"traces"`
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	HasMore bool    `json:"hasMore"`
}

// CostBreakdown represents cost breakdown by dimension.
type CostBreakdown struct {
	Dimension    string  `json:"dimension"`
	Value        string  `json:"value"`
	Cost         float64 `json:"cost"`
	RequestCount int     `json:"requestCount"`
}

// CostSummary represents a cost summary.
type CostSummary struct {
	TotalCost    float64         `json:"totalCost"`
	PeriodStart  time.Time       `json:"periodStart"`
	PeriodEnd    time.Time       `json:"periodEnd"`
	RequestCount int             `json:"requestCount"`
	ByServer     []CostBreakdown `json:"byServer,omitempty"`
	ByTeam       []CostBreakdown `json:"byTeam,omitempty"`
	ByTool       []CostBreakdown `json:"byTool,omitempty"`
}

// APIKey represents an API key.
type APIKey struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"keyPrefix"`
	Environment  string     `json:"environment"`
	Permissions  string     `json:"permissions"`
	RateLimitRPM int        `json:"rateLimitRpm"`
	CreatedAt    time.Time  `json:"createdAt"`
	LastUsedAt   *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
}
