package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// InjectionMiddleware provides prompt injection detection.
type InjectionMiddleware struct {
	injectionService *service.InjectionService
	logger           *slog.Logger
}

// NewInjectionMiddleware creates a new injection detection middleware.
func NewInjectionMiddleware(injectionService *service.InjectionService, logger *slog.Logger) *InjectionMiddleware {
	return &InjectionMiddleware{
		injectionService: injectionService,
		logger:           logger,
	}
}

// Detect checks requests for prompt injection patterns.
func (m *InjectionMiddleware) Detect() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check MCP tool call requests
			if !strings.Contains(r.URL.Path, "/tools/call") {
				next.ServeHTTP(w, r)
				return
			}

			// Get organization ID from auth context
			authInfo := GetAuthInfo(r.Context())
			if authInfo == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Read and buffer the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				m.logger.Error("failed to read request body", "error", err)
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Extract MCP server from URL
			mcpServer := chi.URLParam(r, "server")
			if mcpServer == "" {
				mcpServer = "unknown"
			}

			// Extract input to scan
			input := m.extractInput(body)
			if input == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Perform detection
			result, err := m.injectionService.Detect(r.Context(), authInfo.OrgID, mcpServer, input)
			if err != nil {
				m.logger.Error("injection detection failed",
					"error", err,
					"mcp_server", mcpServer,
				)
				// Continue on error - don't block requests if detection fails
				next.ServeHTTP(w, r)
				return
			}

			if result.Detected {
				// Record the detection
				detection := &domain.InjectionDetection{
					OrgID:          authInfo.OrgID,
					TraceID:        GetTraceID(r.Context()),
					Type:           result.Type,
					Severity:       result.Severity,
					PatternMatched: result.PatternMatched,
					Input:          input,
					ActionTaken:    result.Action,
					MCPServer:      mcpServer,
					ToolName:       m.extractToolName(body),
					APIKeyID:       authInfo.APIKeyID,
					IPAddress:      GetClientIP(r),
					CreatedAt:      time.Now(),
				}

				go func() {
					if err := m.injectionService.RecordDetection(r.Context(), detection); err != nil {
						m.logger.Error("failed to record detection", "error", err)
					}
				}()

				// Handle based on action
				switch result.Action {
				case domain.SafetyModeBlock:
					m.logger.Warn("injection blocked",
						"type", result.Type,
						"severity", result.Severity,
						"pattern", result.PatternMatched,
						"mcp_server", mcpServer,
					)
					handler.WriteError(w, http.StatusBadRequest, "injection_detected",
						"Request blocked: potential prompt injection detected")
					return

				case domain.SafetyModeWarn:
					m.logger.Warn("injection detected (warn mode)",
						"type", result.Type,
						"severity", result.Severity,
						"pattern", result.PatternMatched,
						"mcp_server", mcpServer,
					)
					// Add warning header and continue
					w.Header().Set("X-GatewayOps-Warning", "Potential prompt injection detected")

				case domain.SafetyModeLog:
					m.logger.Info("injection detected (log mode)",
						"type", result.Type,
						"severity", result.Severity,
						"pattern", result.PatternMatched,
						"mcp_server", mcpServer,
					)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractInput extracts the text input from a tool call request body.
func (m *InjectionMiddleware) extractInput(body []byte) string {
	var request struct {
		Tool      string                 `json:"tool"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		return ""
	}

	// Collect all string values from arguments
	var inputs []string
	m.collectStrings(request.Arguments, &inputs)

	return strings.Join(inputs, " ")
}

// collectStrings recursively collects all string values from a map.
func (m *InjectionMiddleware) collectStrings(data map[string]interface{}, result *[]string) {
	for _, v := range data {
		switch val := v.(type) {
		case string:
			if len(val) > 0 {
				*result = append(*result, val)
			}
		case map[string]interface{}:
			m.collectStrings(val, result)
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok && len(s) > 0 {
					*result = append(*result, s)
				}
				if m2, ok := item.(map[string]interface{}); ok {
					m.collectStrings(m2, result)
				}
			}
		}
	}
}

// extractToolName extracts the tool name from a tool call request body.
func (m *InjectionMiddleware) extractToolName(body []byte) string {
	var request struct {
		Tool string `json:"tool"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		return ""
	}

	return request.Tool
}

// DetectWithApproval combines injection detection with tool approval checking.
func (m *InjectionMiddleware) DetectWithApproval(approvalService *service.ApprovalService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check MCP tool call requests
			if !strings.Contains(r.URL.Path, "/tools/call") {
				next.ServeHTTP(w, r)
				return
			}

			authInfo := GetAuthInfo(r.Context())
			if authInfo == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			mcpServer := chi.URLParam(r, "server")
			toolName := m.extractToolName(body)

			if mcpServer == "" || toolName == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check tool approval
			userID := GetUserID(r.Context())
			if userID == nil {
				// If no user ID, use a placeholder for API key-based calls
				placeholder := uuid.Nil
				userID = &placeholder
			}

			accessResult, err := approvalService.CheckToolAccess(r.Context(), authInfo.OrgID, *userID, mcpServer, toolName)
			if err != nil {
				m.logger.Error("failed to check tool access", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if !accessResult.Allowed {
				handler.WriteError(w, http.StatusForbidden, "tool_access_denied", accessResult.Message)
				return
			}

			// Then check for injection
			input := m.extractInput(body)
			if input != "" {
				result, err := m.injectionService.Detect(r.Context(), authInfo.OrgID, mcpServer, input)
				if err == nil && result.Detected && result.Action == domain.SafetyModeBlock {
					handler.WriteError(w, http.StatusBadRequest, "injection_detected",
						"Request blocked: potential prompt injection detected")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP from a request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
