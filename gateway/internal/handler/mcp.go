package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/config"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// MCPClient defines the interface for MCP server communication.
type MCPClient interface {
	Forward(ctx context.Context, server, endpoint string, body []byte) ([]byte, int, error)
}

// MCPHandler handles MCP proxy requests.
type MCPHandler struct {
	config     *config.Config
	logger     zerolog.Logger
	httpClient *http.Client
}

// NewMCPHandler creates a new MCP handler.
func NewMCPHandler(cfg *config.Config, logger zerolog.Logger) *MCPHandler {
	return &MCPHandler{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MCPRequest represents a generic MCP request.
type MCPRequest struct {
	Tool      string                 `json:"tool,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	URI       string                 `json:"uri,omitempty"`
	Name      string                 `json:"name,omitempty"`
}

// ToolsCall handles POST /v1/mcp/{server}/tools/call
func (h *MCPHandler) ToolsCall(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/tools/call")
}

// ToolsList handles POST /v1/mcp/{server}/tools/list
func (h *MCPHandler) ToolsList(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/tools/list")
}

// ResourcesRead handles POST /v1/mcp/{server}/resources/read
func (h *MCPHandler) ResourcesRead(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/resources/read")
}

// ResourcesList handles POST /v1/mcp/{server}/resources/list
func (h *MCPHandler) ResourcesList(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/resources/list")
}

// PromptsGet handles POST /v1/mcp/{server}/prompts/get
func (h *MCPHandler) PromptsGet(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/prompts/get")
}

// PromptsList handles POST /v1/mcp/{server}/prompts/list
func (h *MCPHandler) PromptsList(w http.ResponseWriter, r *http.Request) {
	h.proxyRequest(w, r, "/prompts/list")
}

// proxyRequest forwards the request to the target MCP server.
func (h *MCPHandler) proxyRequest(w http.ResponseWriter, r *http.Request, endpoint string) {
	serverName := chi.URLParam(r, "server")
	if serverName == "" {
		WriteError(w, http.StatusBadRequest, "missing_server", "Server name is required")
		return
	}

	// Look up server configuration
	serverConfig, ok := h.config.MCPServers[serverName]
	if !ok {
		WriteError(w, http.StatusNotFound, "server_not_found", fmt.Sprintf("MCP server '%s' not found", serverName))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Get trace info for logging
	traceID := middleware.GetTraceID(r.Context())
	spanID := middleware.GetSpanID(r.Context())

	// Get auth info
	authInfo := middleware.GetAuthInfo(r.Context())

	h.logger.Info().
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Str("server", serverName).
		Str("endpoint", endpoint).
		Str("org_id", authInfo.OrgID.String()).
		Int("body_size", len(body)).
		Msg("Proxying MCP request")

	start := time.Now()

	// Build target URL
	targetURL := serverConfig.URL + endpoint

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create proxy request")
		WriteError(w, http.StatusInternalServerError, "proxy_error", "Failed to create proxy request")
		return
	}

	// Copy relevant headers
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-Trace-ID", traceID)
	proxyReq.Header.Set("X-Span-ID", spanID)
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

	// Set timeout from config
	ctx, cancel := context.WithTimeout(r.Context(), serverConfig.Timeout)
	defer cancel()
	proxyReq = proxyReq.WithContext(ctx)

	// Send request to MCP server
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		duration := time.Since(start)
		h.logger.Error().
			Err(err).
			Str("trace_id", traceID).
			Dur("duration", duration).
			Str("target_url", targetURL).
			Msg("MCP server request failed")
		WriteError(w, http.StatusBadGateway, "upstream_error", "Failed to reach MCP server")
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to read MCP server response")
		WriteError(w, http.StatusBadGateway, "upstream_error", "Failed to read MCP server response")
		return
	}

	duration := time.Since(start)

	h.logger.Info().
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Str("server", serverName).
		Str("endpoint", endpoint).
		Int("status", resp.StatusCode).
		Int("response_size", len(respBody)).
		Dur("duration", duration).
		Msg("MCP request completed")

	// Calculate cost (simple per-call pricing for now)
	cost := serverConfig.Pricing.PerCall

	// TODO: Publish trace span to Redis Streams for async processing
	// TODO: Publish cost event to ClickHouse

	// Forward response to client
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-MCP-Server", serverName)
	w.Header().Set("X-MCP-Duration-Ms", fmt.Sprintf("%d", duration.Milliseconds()))
	w.Header().Set("X-MCP-Cost", fmt.Sprintf("%.6f", cost))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// validateMCPRequest validates the MCP request body.
func validateMCPRequest(body []byte, endpoint string) error {
	var req MCPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	switch endpoint {
	case "/tools/call":
		if req.Tool == "" {
			return fmt.Errorf("tool name is required")
		}
	case "/resources/read":
		if req.URI == "" {
			return fmt.Errorf("resource URI is required")
		}
	case "/prompts/get":
		if req.Name == "" {
			return fmt.Errorf("prompt name is required")
		}
	}

	return nil
}
