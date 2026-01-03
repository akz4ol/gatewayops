package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/agent"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AgentHandler handles agent platform API requests.
type AgentHandler struct {
	logger  zerolog.Logger
	manager *agent.Manager
	baseURL string
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler(logger zerolog.Logger, manager *agent.Manager, baseURL string) *AgentHandler {
	return &AgentHandler{
		logger:  logger,
		manager: manager,
		baseURL: baseURL,
	}
}

// Connect establishes a new agent connection.
func (h *AgentHandler) Connect(w http.ResponseWriter, r *http.Request) {
	var req agent.ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate request
	if req.Platform == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Platform is required")
		return
	}
	if req.Transport == "" {
		req.Transport = agent.TransportHTTP
	}

	// Get auth info
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	if authInfo != nil {
		orgID = authInfo.OrgID
		userID = authInfo.UserID
	}

	// Create connection
	conn, err := h.manager.Connect(r.Context(), req, orgID, userID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create agent connection")
		WriteError(w, http.StatusInternalServerError, "connection_error", "Failed to create connection")
		return
	}

	// Build response
	resp := agent.ConnectResponse{
		ConnectionID: conn.ID,
		GatewayURL:   fmt.Sprintf("wss://%s/v1/agents/%s", h.baseURL, conn.ID),
		AvailableServers: []agent.ServerInfo{
			{Name: "filesystem", Description: "File system operations", ToolCount: 12, ResourceCount: 5},
			{Name: "github", Description: "GitHub integration", ToolCount: 8, ResourceCount: 3},
			{Name: "database", Description: "Database operations", ToolCount: 6, ResourceCount: 2},
			{Name: "slack", Description: "Slack messaging", ToolCount: 4, ResourceCount: 1},
		},
		RateLimits: agent.RateLimitInfo{
			RequestsPerMinute: 1000,
			TokensPerMinute:   100000,
		},
	}

	WriteJSON(w, http.StatusOK, resp)
}

// WebSocket handles WebSocket upgrade for agent connections.
func (h *AgentHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	connIDStr := chi.URLParam(r, "connectionID")
	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_connection_id", "Invalid connection ID")
		return
	}

	if err := h.manager.UpgradeToWebSocket(w, r, connID); err != nil {
		h.logger.Error().Err(err).Str("connection_id", connIDStr).Msg("WebSocket upgrade failed")
		// Note: Can't write error response after upgrade attempt
		return
	}
}

// Execute handles batch tool execution.
func (h *AgentHandler) Execute(w http.ResponseWriter, r *http.Request) {
	var req agent.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if len(req.Calls) == 0 {
		WriteError(w, http.StatusBadRequest, "invalid_request", "At least one call is required")
		return
	}

	if req.ExecutionMode == "" {
		req.ExecutionMode = "parallel"
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 30000
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(req.TimeoutMs)*time.Millisecond)
	defer cancel()

	var results []agent.ToolResult
	var totalCost float64
	traceID := fmt.Sprintf("tr_%s", uuid.New().String()[:8])

	if req.ExecutionMode == "parallel" {
		results, totalCost = h.executeParallel(ctx, req.Calls)
	} else {
		results, totalCost = h.executeSequential(ctx, req.Calls)
	}

	resp := agent.ExecuteResponse{
		Results:   results,
		TraceID:   traceID,
		TotalCost: totalCost,
	}

	WriteJSON(w, http.StatusOK, resp)
}

// executeParallel executes tool calls in parallel.
func (h *AgentHandler) executeParallel(ctx context.Context, calls []agent.ToolCall) ([]agent.ToolResult, float64) {
	results := make([]agent.ToolResult, len(calls))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalCost float64

	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c agent.ToolCall) {
			defer wg.Done()
			result := h.executeToolCall(ctx, c)
			
			mu.Lock()
			results[idx] = result
			totalCost += result.Cost
			mu.Unlock()
		}(i, call)
	}

	wg.Wait()
	return results, totalCost
}

// executeSequential executes tool calls sequentially.
func (h *AgentHandler) executeSequential(ctx context.Context, calls []agent.ToolCall) ([]agent.ToolResult, float64) {
	results := make([]agent.ToolResult, 0, len(calls))
	var totalCost float64

	for _, call := range calls {
		select {
		case <-ctx.Done():
			// Add timeout results for remaining calls
			for i := len(results); i < len(calls); i++ {
				results = append(results, agent.ToolResult{
					ID:     calls[i].ID,
					Status: "timeout",
					Error: &agent.ErrorInfo{
						Code:    "timeout",
						Message: "Execution timed out",
					},
				})
			}
			return results, totalCost

		default:
			result := h.executeToolCall(ctx, call)
			results = append(results, result)
			totalCost += result.Cost
		}
	}

	return results, totalCost
}

// executeToolCall executes a single tool call.
func (h *AgentHandler) executeToolCall(ctx context.Context, call agent.ToolCall) agent.ToolResult {
	start := time.Now()

	// TODO: Integrate with actual MCP handler
	// For now, return mock results
	duration := time.Since(start)

	// Simulate some processing time
	time.Sleep(20 * time.Millisecond)

	return agent.ToolResult{
		ID:     call.ID,
		Status: "success",
		Content: []agent.ContentBlock{
			{
				Type: "text",
				Text: fmt.Sprintf("Tool %s.%s executed successfully", call.Server, call.Tool),
			},
		},
		DurationMs: int(duration.Milliseconds()) + 20,
		Cost:       0.0001,
	}
}

// ExecuteStream handles SSE streaming tool execution.
func (h *AgentHandler) ExecuteStream(w http.ResponseWriter, r *http.Request) {
	var req agent.ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if len(req.Calls) == 0 {
		WriteError(w, http.StatusBadRequest, "invalid_request", "At least one call is required")
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		WriteError(w, http.StatusInternalServerError, "streaming_not_supported", "Streaming not supported")
		return
	}

	traceID := fmt.Sprintf("tr_%s", uuid.New().String()[:8])
	var totalCost float64

	for _, call := range req.Calls {
		// Send start event
		h.sendSSE(w, flusher, agent.SSEEventStart, map[string]any{
			"call_id": call.ID,
			"server":  call.Server,
			"tool":    call.Tool,
		})

		// Send progress
		h.sendSSE(w, flusher, agent.SSEEventProgress, map[string]any{
			"call_id":  call.ID,
			"progress": 0.5,
			"message":  fmt.Sprintf("Executing %s.%s...", call.Server, call.Tool),
		})

		// Simulate execution
		time.Sleep(50 * time.Millisecond)

		// Send complete
		cost := 0.0001
		totalCost += cost
		h.sendSSE(w, flusher, agent.SSEEventComplete, map[string]any{
			"call_id":     call.ID,
			"status":      "success",
			"duration_ms": 50,
			"cost":        cost,
			"content": []map[string]any{
				{"type": "text", "text": fmt.Sprintf("Tool %s.%s executed", call.Server, call.Tool)},
			},
		})
	}

	// Send done event
	h.sendSSE(w, flusher, agent.SSEEventDone, map[string]any{
		"trace_id":   traceID,
		"total_cost": totalCost,
	})
}

// sendSSE sends a Server-Sent Event.
func (h *AgentHandler) sendSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonData)
	flusher.Flush()
}

// GetConnection returns information about a specific connection.
func (h *AgentHandler) GetConnection(w http.ResponseWriter, r *http.Request) {
	connIDStr := chi.URLParam(r, "connectionID")
	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_connection_id", "Invalid connection ID")
		return
	}

	conn, exists := h.manager.GetConnection(connID)
	if !exists {
		WriteError(w, http.StatusNotFound, "not_found", "Connection not found")
		return
	}

	WriteJSON(w, http.StatusOK, conn)
}

// Disconnect closes an agent connection.
func (h *AgentHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	connIDStr := chi.URLParam(r, "connectionID")
	connID, err := uuid.Parse(connIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_connection_id", "Invalid connection ID")
		return
	}

	h.manager.Disconnect(connID)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

// GetStats returns agent connection statistics.
func (h *AgentHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.manager.GetStats()
	WriteJSON(w, http.StatusOK, stats)
}

// ListTools returns all available tools across MCP servers (OpenAI format).
func (h *AgentHandler) ListTools(w http.ResponseWriter, r *http.Request) {
	// Return tools in OpenAI function calling format
	tools := []map[string]any{
		{
			"type": "function",
			"function": map[string]any{
				"name":        "filesystem__read_file",
				"description": "Read contents of a file from the filesystem",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "Path to the file to read",
						},
					},
					"required": []string{"path"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]any{
				"name":        "filesystem__write_file",
				"description": "Write contents to a file",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"path": map[string]any{
							"type":        "string",
							"description": "Path to write to",
						},
						"content": map[string]any{
							"type":        "string",
							"description": "Content to write",
						},
					},
					"required": []string{"path", "content"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]any{
				"name":        "github__create_issue",
				"description": "Create a GitHub issue",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"repo": map[string]any{
							"type":        "string",
							"description": "Repository (owner/repo)",
						},
						"title": map[string]any{
							"type":        "string",
							"description": "Issue title",
						},
						"body": map[string]any{
							"type":        "string",
							"description": "Issue body",
						},
					},
					"required": []string{"repo", "title"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]any{
				"name":        "github__search_code",
				"description": "Search code on GitHub",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search query",
						},
						"repo": map[string]any{
							"type":        "string",
							"description": "Repository to search in (optional)",
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]any{
				"name":        "database__query",
				"description": "Execute a database query",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"sql": map[string]any{
							"type":        "string",
							"description": "SQL query to execute",
						},
					},
					"required": []string{"sql"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]any{
				"name":        "slack__send_message",
				"description": "Send a Slack message",
				"parameters": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"channel": map[string]any{
							"type":        "string",
							"description": "Channel to send to",
						},
						"message": map[string]any{
							"type":        "string",
							"description": "Message content",
						},
					},
					"required": []string{"channel", "message"},
				},
			},
		},
	}

	WriteJSON(w, http.StatusOK, map[string]any{"tools": tools})
}
