// Package agent provides agent platform connection management.
package agent

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ConnectionState represents the state of an agent connection.
type ConnectionState string

const (
	StateConnecting  ConnectionState = "connecting"
	StateConnected   ConnectionState = "connected"
	StateDisconnected ConnectionState = "disconnected"
)

// Transport represents the connection transport type.
type Transport string

const (
	TransportHTTP      Transport = "http"
	TransportWebSocket Transport = "websocket"
	TransportSSE       Transport = "sse"
)

// Connection represents an active agent platform connection.
type Connection struct {
	ID           uuid.UUID       `json:"id"`
	AgentID      string          `json:"agent_id"`
	Platform     string          `json:"platform"`
	OrgID        uuid.UUID       `json:"org_id"`
	UserID       uuid.UUID       `json:"user_id"`
	Transport    Transport       `json:"transport"`
	State        ConnectionState `json:"state"`
	Capabilities []string        `json:"capabilities"`
	CallbackURL  string          `json:"callback_url,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	LastActiveAt time.Time       `json:"last_active_at"`

	// WebSocket connection (not serialized)
	ws     *websocket.Conn
	mu     sync.Mutex
	sendCh chan []byte
	done   chan struct{}
}

// ConnectRequest represents a request to establish an agent connection.
type ConnectRequest struct {
	AgentID      string         `json:"agent_id"`
	Platform     string         `json:"platform"`
	Capabilities []string       `json:"capabilities"`
	Transport    Transport      `json:"transport"`
	CallbackURL  string         `json:"callback_url,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// ConnectResponse represents the response to a connection request.
type ConnectResponse struct {
	ConnectionID     uuid.UUID        `json:"connection_id"`
	GatewayURL       string           `json:"gateway_url"`
	AvailableServers []ServerInfo     `json:"available_servers"`
	RateLimits       RateLimitInfo    `json:"rate_limits"`
}

// ServerInfo provides information about an available MCP server.
type ServerInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ToolCount   int    `json:"tools"`
	ResourceCount int  `json:"resources"`
}

// RateLimitInfo provides rate limit configuration.
type RateLimitInfo struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	TokensPerMinute   int `json:"tokens_per_minute,omitempty"`
}

// ToolCall represents a single tool call in a batch.
type ToolCall struct {
	ID        string         `json:"id"`
	Server    string         `json:"server"`
	Tool      string         `json:"tool"`
	Arguments map[string]any `json:"arguments"`
}

// ExecuteRequest represents a batch tool execution request.
type ExecuteRequest struct {
	ConnectionID  uuid.UUID  `json:"connection_id,omitempty"`
	Calls         []ToolCall `json:"calls"`
	ExecutionMode string     `json:"execution_mode"` // "parallel" or "sequential"
	TimeoutMs     int        `json:"timeout_ms,omitempty"`
}

// ToolResult represents the result of a single tool call.
type ToolResult struct {
	ID         string         `json:"id"`
	Status     string         `json:"status"` // "success", "error", "timeout"
	Content    []ContentBlock `json:"content,omitempty"`
	Error      *ErrorInfo     `json:"error,omitempty"`
	DurationMs int            `json:"duration_ms"`
	Cost       float64        `json:"cost"`
}

// ContentBlock represents a content block in a tool result.
type ContentBlock struct {
	Type string `json:"type"` // "text", "image", "resource"
	Text string `json:"text,omitempty"`
	Data string `json:"data,omitempty"` // base64 for binary
	URI  string `json:"uri,omitempty"`
}

// ErrorInfo provides error details.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ExecuteResponse represents the response to a batch execution request.
type ExecuteResponse struct {
	Results   []ToolResult `json:"results"`
	TraceID   string       `json:"trace_id"`
	TotalCost float64      `json:"total_cost"`
}

// WSMessage represents a WebSocket message.
type WSMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// WSMessageType constants
const (
	WSTypeToolCall   = "tool_call"
	WSTypeToolResult = "tool_result"
	WSTypeProgress   = "progress"
	WSTypeError      = "error"
	WSTypeCancel     = "cancel"
	WSTypePing       = "ping"
	WSTypePong       = "pong"
)

// ProgressPayload represents progress update data.
type ProgressPayload struct {
	Progress float64 `json:"progress"`
	Message  string  `json:"message,omitempty"`
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// SSEEventType constants
const (
	SSEEventStart    = "start"
	SSEEventProgress = "progress"
	SSEEventChunk    = "chunk"
	SSEEventComplete = "complete"
	SSEEventError    = "error"
	SSEEventDone     = "done"
)
