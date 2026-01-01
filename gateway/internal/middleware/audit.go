package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/audit"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuditLogger defines the interface for audit logging.
type AuditLogger interface {
	LogEvent(ctx context.Context, event audit.Event)
}

// responseWriter wraps http.ResponseWriter to capture status code.
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *auditResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

// Audit returns middleware that logs all requests to the audit log.
func Audit(logger AuditLogger, zlog zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Determine the action based on the path
			action, resource, resourceID := determineAction(r)
			if action == "" {
				// Skip auditing for non-auditable endpoints
				next.ServeHTTP(w, r)
				return
			}

			// Read body for details (for POST/PUT requests)
			var details map[string]interface{}
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				body, err := io.ReadAll(r.Body)
				if err == nil && len(body) > 0 {
					r.Body = io.NopCloser(bytes.NewBuffer(body))
					details = extractAuditDetails(body, action)
				}
			}

			// Wrap response writer to capture status
			wrapped := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Execute handler
			next.ServeHTTP(wrapped, r)

			// Determine outcome based on status code
			outcome := determineOutcome(wrapped.statusCode)

			// Get auth info
			var userID, apiKeyID *uuid.UUID
			orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Demo org
			if authInfo := GetAuthInfo(r.Context()); authInfo != nil {
				apiKeyID = &authInfo.APIKeyID
				orgID = authInfo.OrgID
				userID = &authInfo.UserID
			}

			// Log the event
			event := audit.Event{
				OrgID:      orgID,
				UserID:     userID,
				APIKeyID:   apiKeyID,
				TraceID:    chimiddleware.GetReqID(r.Context()),
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				Outcome:    outcome,
				Details:    details,
				IPAddress:  r.RemoteAddr,
				UserAgent:  r.UserAgent(),
				RequestID:  chimiddleware.GetReqID(r.Context()),
				DurationMS: time.Since(start).Milliseconds(),
			}

			logger.LogEvent(r.Context(), event)
		})
	}
}

// determineAction determines the audit action from the request.
func determineAction(r *http.Request) (domain.AuditAction, string, string) {
	path := r.URL.Path

	// MCP endpoints
	if strings.Contains(path, "/mcp/") {
		server := chi.URLParam(r, "server")
		if server == "" {
			// Try to extract from path
			parts := strings.Split(path, "/mcp/")
			if len(parts) > 1 {
				serverParts := strings.Split(parts[1], "/")
				if len(serverParts) > 0 {
					server = serverParts[0]
				}
			}
		}

		if strings.Contains(path, "/tools/call") {
			return domain.AuditActionMCPToolCall, "mcp:" + server, ""
		}
		if strings.Contains(path, "/tools/list") {
			return domain.AuditActionMCPToolList, "mcp:" + server, ""
		}
		if strings.Contains(path, "/resources/") {
			return domain.AuditActionMCPResourceGet, "mcp:" + server, ""
		}
	}

	// API key endpoints
	if strings.Contains(path, "/api-keys") {
		keyID := chi.URLParam(r, "keyID")
		switch r.Method {
		case http.MethodPost:
			if strings.Contains(path, "/rotate") {
				return domain.AuditActionAPIKeyRotate, "api_key", keyID
			}
			return domain.AuditActionAPIKeyCreate, "api_key", ""
		case http.MethodDelete:
			return domain.AuditActionAPIKeyRevoke, "api_key", keyID
		}
	}

	// Safety policy endpoints
	if strings.Contains(path, "/safety/policies") {
		policyID := chi.URLParam(r, "policyID")
		switch r.Method {
		case http.MethodPost:
			return domain.AuditActionPolicyCreate, "safety_policy", ""
		case http.MethodPut:
			return domain.AuditActionPolicyUpdate, "safety_policy", policyID
		case http.MethodDelete:
			return domain.AuditActionPolicyDelete, "safety_policy", policyID
		}
	}

	// Role endpoints
	if strings.Contains(path, "/roles") {
		roleID := chi.URLParam(r, "roleID")
		switch r.Method {
		case http.MethodPost:
			return domain.AuditActionRoleCreate, "role", ""
		case http.MethodPut:
			return domain.AuditActionRoleUpdate, "role", roleID
		case http.MethodDelete:
			return domain.AuditActionRoleDelete, "role", roleID
		}
	}

	return "", "", ""
}

// determineOutcome determines the audit outcome from status code.
func determineOutcome(statusCode int) domain.AuditOutcome {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return domain.AuditOutcomeSuccess
	case statusCode == 400 && statusCode == 403:
		return domain.AuditOutcomeBlocked
	default:
		return domain.AuditOutcomeFailure
	}
}

// extractAuditDetails extracts relevant details from the request body.
func extractAuditDetails(body []byte, action domain.AuditAction) map[string]interface{} {
	details := make(map[string]interface{})

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return details
	}

	switch action {
	case domain.AuditActionMCPToolCall:
		if name, ok := data["name"].(string); ok {
			details["tool_name"] = name
		}
	case domain.AuditActionAPIKeyCreate:
		if name, ok := data["name"].(string); ok {
			details["key_name"] = name
		}
		if env, ok := data["environment"].(string); ok {
			details["environment"] = env
		}
	case domain.AuditActionPolicyCreate, domain.AuditActionPolicyUpdate:
		if name, ok := data["name"].(string); ok {
			details["policy_name"] = name
		}
		if mode, ok := data["mode"].(string); ok {
			details["mode"] = mode
		}
	case domain.AuditActionRoleCreate, domain.AuditActionRoleUpdate:
		if name, ok := data["name"].(string); ok {
			details["role_name"] = name
		}
	}

	return details
}
