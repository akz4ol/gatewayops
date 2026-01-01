package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// SSOMiddleware provides SSO session authentication.
type SSOMiddleware struct {
	authService *service.AuthService
	logger      *slog.Logger
}

// NewSSOMiddleware creates a new SSO middleware.
func NewSSOMiddleware(authService *service.AuthService, logger *slog.Logger) *SSOMiddleware {
	return &SSOMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// SessionContextKey is the context key for the user session.
type SessionContextKey struct{}

// UserContextKey is the context key for the authenticated user.
type UserContextKey struct{}

// Authenticate validates SSO sessions from cookies or Bearer tokens.
func (m *SSOMiddleware) Authenticate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to get session from Authorization header first
			sessionID := m.extractSessionID(r)
			if sessionID == uuid.Nil {
				// Fall back to cookie
				sessionID = m.extractSessionFromCookie(r)
			}

			if sessionID == uuid.Nil {
				// No session found - continue without user context
				// API key auth will be tried by another middleware
				next.ServeHTTP(w, r)
				return
			}

			// Validate session
			session, err := m.authService.GetSession(r.Context(), sessionID)
			if err != nil {
				m.logger.Error("failed to get session", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if session == nil {
				// Session not found or expired
				next.ServeHTTP(w, r)
				return
			}

			// Update session activity
			go m.authService.UpdateSessionActivity(context.Background(), sessionID)

			// Add session and user info to context
			ctx := context.WithValue(r.Context(), SessionContextKey{}, session)

			// Set auth info for compatibility with existing handlers
			authInfo := &AuthInfo{
				OrgID:    session.OrgID,
				UserID:   &session.UserID,
				AuthType: "session",
			}
			ctx = context.WithValue(ctx, AuthInfoKey{}, authInfo)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireSession ensures a valid session exists.
func (m *SSOMiddleware) RequireSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSession(r.Context())
			if session == nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized", "Valid session required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractSessionID extracts the session ID from the Authorization header.
func (m *SSOMiddleware) extractSessionID(r *http.Request) uuid.UUID {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return uuid.Nil
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return uuid.Nil
	}

	token := parts[1]

	// Check if it's a session token (starts with "gwo_session_")
	if !strings.HasPrefix(token, "gwo_session_") {
		return uuid.Nil
	}

	// Extract UUID from token
	uuidStr := strings.TrimPrefix(token, "gwo_session_")
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.Nil
	}

	return id
}

// extractSessionFromCookie extracts the session ID from a cookie.
func (m *SSOMiddleware) extractSessionFromCookie(r *http.Request) uuid.UUID {
	cookie, err := r.Cookie("gwo_session")
	if err != nil {
		return uuid.Nil
	}

	id, err := uuid.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil
	}

	return id
}

// GetSession retrieves the session from context.
func GetSession(ctx context.Context) *domain.UserSession {
	session, ok := ctx.Value(SessionContextKey{}).(*domain.UserSession)
	if !ok {
		return nil
	}
	return session
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) *uuid.UUID {
	session := GetSession(ctx)
	if session != nil {
		return &session.UserID
	}

	authInfo := GetAuthInfo(ctx)
	if authInfo != nil {
		return authInfo.UserID
	}

	return nil
}
