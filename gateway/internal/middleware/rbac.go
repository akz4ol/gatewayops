package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/handler"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// RBACMiddleware provides role-based access control.
type RBACMiddleware struct {
	rbacService *service.RBACService
	logger      *slog.Logger
}

// NewRBACMiddleware creates a new RBAC middleware.
func NewRBACMiddleware(rbacService *service.RBACService, logger *slog.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		rbacService: rbacService,
		logger:      logger,
	}
}

// RequirePermission requires the user to have a specific permission.
func (m *RBACMiddleware) RequirePermission(permission domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == nil {
				// No user in context - check if we have API key with permission
				authInfo := GetAuthInfo(r.Context())
				if authInfo != nil && m.hasAPIKeyPermission(authInfo, permission) {
					next.ServeHTTP(w, r)
					return
				}

				handler.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Check user permission
			hasPermission, err := m.rbacService.HasPermissionSimple(r.Context(), *userID, permission)
			if err != nil {
				m.logger.Error("failed to check permission",
					"user_id", userID,
					"permission", permission,
					"error", err,
				)
				handler.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to verify permissions")
				return
			}

			if !hasPermission {
				m.logger.Warn("permission denied",
					"user_id", userID,
					"permission", permission,
				)
				handler.WriteError(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissionScoped requires the user to have a permission within a scope.
func (m *RBACMiddleware) RequirePermissionScoped(permission domain.Permission, scopeType domain.ScopeType, scopeIDParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Get scope ID from URL if specified
			var scopeID *uuid.UUID
			if scopeIDParam != "" {
				// This would typically come from chi.URLParam or similar
				scopeIDStr := r.URL.Query().Get(scopeIDParam)
				if scopeIDStr != "" {
					id, err := uuid.Parse(scopeIDStr)
					if err == nil {
						scopeID = &id
					}
				}
			}

			// Check scoped permission
			hasPermission, err := m.rbacService.HasPermission(r.Context(), *userID, permission, scopeType, scopeID)
			if err != nil {
				m.logger.Error("failed to check scoped permission",
					"user_id", userID,
					"permission", permission,
					"scope_type", scopeType,
					"scope_id", scopeID,
					"error", err,
				)
				handler.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to verify permissions")
				return
			}

			if !hasPermission {
				handler.WriteError(w, http.StatusForbidden, "forbidden", "Insufficient permissions for this resource")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission requires the user to have at least one of the specified permissions.
func (m *RBACMiddleware) RequireAnyPermission(permissions ...domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Check each permission
			for _, permission := range permissions {
				hasPermission, err := m.rbacService.HasPermissionSimple(r.Context(), *userID, permission)
				if err != nil {
					continue
				}
				if hasPermission {
					next.ServeHTTP(w, r)
					return
				}
			}

			handler.WriteError(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
		})
	}
}

// RequireAllPermissions requires the user to have all of the specified permissions.
func (m *RBACMiddleware) RequireAllPermissions(permissions ...domain.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == nil {
				handler.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
				return
			}

			// Check all permissions
			for _, permission := range permissions {
				hasPermission, err := m.rbacService.HasPermissionSimple(r.Context(), *userID, permission)
				if err != nil {
					m.logger.Error("failed to check permission",
						"user_id", userID,
						"permission", permission,
						"error", err,
					)
					handler.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to verify permissions")
					return
				}
				if !hasPermission {
					handler.WriteError(w, http.StatusForbidden, "forbidden", "Insufficient permissions")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin requires the user to have admin permission.
func (m *RBACMiddleware) RequireAdmin() func(http.Handler) http.Handler {
	return m.RequirePermission(domain.PermissionAll)
}

// hasAPIKeyPermission checks if an API key has a specific permission based on its permissions string.
func (m *RBACMiddleware) hasAPIKeyPermission(authInfo *AuthInfo, permission domain.Permission) bool {
	// API keys have their permissions stored in the auth info
	// This is a simplified check - in production, this would be more sophisticated
	if authInfo.Permissions == "" {
		return false
	}

	// Check for wildcard permission
	if authInfo.Permissions == "*" {
		return true
	}

	// Check for specific permission
	permStr := string(permission)
	permissions := authInfo.Permissions

	// Simple contains check - in production, use proper permission parsing
	return containsPermission(permissions, permStr)
}

// containsPermission checks if a comma-separated permission string contains a specific permission.
func containsPermission(permissions, target string) bool {
	// Split by comma and check each
	for _, p := range splitPermissions(permissions) {
		if p == target || p == "*" {
			return true
		}
		// Check for prefix match (e.g., "mcp:*" matches "mcp:read")
		if len(p) > 0 && p[len(p)-1] == '*' {
			prefix := p[:len(p)-1]
			if len(target) >= len(prefix) && target[:len(prefix)] == prefix {
				return true
			}
		}
	}
	return false
}

// splitPermissions splits a comma-separated permission string.
func splitPermissions(permissions string) []string {
	var result []string
	current := ""
	for _, c := range permissions {
		if c == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else if c != ' ' {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
