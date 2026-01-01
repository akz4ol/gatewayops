package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/safety"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// InjectionDetector defines the interface for injection detection.
type InjectionDetector interface {
	Detect(input string, opts safety.DetectOptions) domain.DetectionResult
}

// Injection returns middleware that detects prompt injection attempts.
func Injection(detector InjectionDetector, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check POST requests to tools/call endpoint
			if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/tools/call") {
				next.ServeHTTP(w, r)
				return
			}

			// Read body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			// Restore body for downstream handlers
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Parse request to extract input
			var toolCall struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(body, &toolCall); err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Extract text content from arguments
			inputText := extractTextContent(toolCall.Arguments)
			if inputText == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Get context info
			mcpServer := chi.URLParam(r, "server")
			reqID := chimiddleware.GetReqID(r.Context())

			// Get auth info if available
			var apiKeyID *uuid.UUID
			var orgID uuid.UUID
			if authInfo := GetAuthInfo(r.Context()); authInfo != nil {
				apiKeyID = &authInfo.APIKeyID
				orgID = authInfo.OrgID
			} else {
				orgID = uuid.MustParse("00000000-0000-0000-0000-000000000001") // Demo org
			}

			// Detect injection
			opts := safety.DetectOptions{
				Input:     inputText,
				OrgID:     orgID,
				TraceID:   reqID,
				MCPServer: mcpServer,
				ToolName:  toolCall.Name,
				APIKeyID:  apiKeyID,
				IPAddress: r.RemoteAddr,
			}

			result := detector.Detect(inputText, opts)

			// Handle detection result
			if result.Detected {
				switch result.Action {
				case domain.SafetyModeBlock:
					logger.Warn().
						Str("severity", string(result.Severity)).
						Str("pattern", result.PatternMatched).
						Str("mcp_server", mcpServer).
						Str("tool", toolCall.Name).
						Msg("Blocked request due to prompt injection detection")

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "injection_detected",
							"message": "Request blocked: potential prompt injection detected",
							"details": map[string]interface{}{
								"severity": result.Severity,
								"type":     result.Type,
							},
						},
					})
					return

				case domain.SafetyModeWarn:
					// Add warning header and continue
					w.Header().Set("X-Safety-Warning", "potential_injection_detected")
					w.Header().Set("X-Safety-Severity", string(result.Severity))
					logger.Warn().
						Str("severity", string(result.Severity)).
						Str("pattern", result.PatternMatched).
						Msg("Warning: potential prompt injection detected (allowed)")

				case domain.SafetyModeLog:
					// Just log, no action
					logger.Debug().
						Str("severity", string(result.Severity)).
						Str("pattern", result.PatternMatched).
						Msg("Logged potential prompt injection")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTextContent extracts text content from tool arguments for analysis.
func extractTextContent(args map[string]interface{}) string {
	var texts []string

	for key, value := range args {
		switch v := value.(type) {
		case string:
			// Include string values that might contain user input
			if isUserInputField(key) {
				texts = append(texts, v)
			}
		case map[string]interface{}:
			// Recursively extract from nested objects
			texts = append(texts, extractTextContent(v))
		case []interface{}:
			// Extract from arrays
			for _, item := range v {
				if str, ok := item.(string); ok {
					texts = append(texts, str)
				} else if obj, ok := item.(map[string]interface{}); ok {
					texts = append(texts, extractTextContent(obj))
				}
			}
		}
	}

	return strings.Join(texts, " ")
}

// isUserInputField checks if a field name suggests it contains user input.
func isUserInputField(fieldName string) bool {
	// Common field names that typically contain user-provided content
	userInputFields := []string{
		"content", "text", "message", "prompt", "query", "input",
		"body", "data", "value", "description", "command", "code",
		"path", "file", "url", "name", "title", "question", "answer",
	}

	lowerField := strings.ToLower(fieldName)
	for _, field := range userInputFields {
		if strings.Contains(lowerField, field) {
			return true
		}
	}

	return true // Be conservative - check all string fields
}
