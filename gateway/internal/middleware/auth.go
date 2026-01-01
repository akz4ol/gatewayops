package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/response"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuthInfo contains authenticated API key information.
type AuthInfo struct {
	KeyID       string
	APIKeyID    uuid.UUID
	UserID      uuid.UUID
	OrgID       uuid.UUID
	TeamID      uuid.UUID
	Environment string
	Permissions []string
	RateLimit   int
}

// Context key for auth info.
const AuthInfoKey contextKey = "auth_info"

// AuthStore defines the interface for API key validation.
type AuthStore interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (*AuthInfo, error)
}

// Auth returns middleware that validates API keys.
func Auth(store AuthStore, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, http.StatusUnauthorized, "missing_auth", "Authorization header is required")
				return
			}

			// Expect "Bearer <api_key>" format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.WriteError(w, http.StatusUnauthorized, "invalid_auth", "Authorization header must be in format: Bearer <api_key>")
				return
			}

			apiKey := parts[1]

			// Validate API key format: gwo_{env}_{32chars}
			if !isValidAPIKeyFormat(apiKey) {
				response.WriteError(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key format")
				return
			}

			// Validate against store
			authInfo, err := store.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				logger.Warn().
					Err(err).
					Str("api_key_prefix", apiKey[:12]+"...").
					Msg("API key validation failed")
				response.WriteError(w, http.StatusUnauthorized, "invalid_api_key", "Invalid or expired API key")
				return
			}

			// Add auth info to context
			ctx := context.WithValue(r.Context(), AuthInfoKey, authInfo)

			logger.Debug().
				Str("key_id", authInfo.KeyID).
				Str("org_id", authInfo.OrgID.String()).
				Str("env", authInfo.Environment).
				Msg("Request authenticated")

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isValidAPIKeyFormat checks if API key matches expected format.
// Format: gwo_{env}_{32chars} where env is 'dev', 'stg', or 'prd'
func isValidAPIKeyFormat(key string) bool {
	if len(key) < 40 {
		return false
	}

	if !strings.HasPrefix(key, "gwo_") {
		return false
	}

	// Check environment prefix
	rest := key[4:]
	validEnvs := []string{"dev_", "stg_", "prd_"}
	hasValidEnv := false
	for _, env := range validEnvs {
		if strings.HasPrefix(rest, env) {
			hasValidEnv = true
			break
		}
	}

	return hasValidEnv
}

// GetAuthInfo extracts auth info from context.
func GetAuthInfo(ctx context.Context) *AuthInfo {
	if info, ok := ctx.Value(AuthInfoKey).(*AuthInfo); ok {
		return info
	}
	return nil
}
