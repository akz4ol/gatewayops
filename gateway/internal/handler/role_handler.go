package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// RoleHandler handles role and RBAC endpoints.
type RoleHandler struct {
	rbacService *service.RBACService
}

// NewRoleHandler creates a new role handler.
func NewRoleHandler(rbacService *service.RBACService) *RoleHandler {
	return &RoleHandler{
		rbacService: rbacService,
	}
}

// ListRoles lists all roles for an organization.
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	roles, err := h.rbacService.ListRoles(r.Context(), authInfo.OrgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list roles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"roles": roles,
	})
}

// GetRole retrieves a role by ID.
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	role, err := h.rbacService.GetRole(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get role")
		return
	}

	if role == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Role not found")
		return
	}

	WriteSuccess(w, role)
}

// CreateRole creates a new custom role.
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	var input domain.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "missing_name", "Role name is required")
		return
	}

	if len(input.Permissions) == 0 {
		WriteError(w, http.StatusBadRequest, "missing_permissions", "At least one permission is required")
		return
	}

	role, err := h.rbacService.CreateRole(r.Context(), authInfo.OrgID, *userID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create role")
		return
	}

	WriteJSON(w, http.StatusCreated, role)
}

// UpdateRole updates a custom role.
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	var input domain.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	role, err := h.rbacService.UpdateRole(r.Context(), id, *userID, input)
	if err != nil {
		if _, ok := err.(service.ErrPermissionDenied); ok {
			WriteError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update role")
		return
	}

	WriteSuccess(w, role)
}

// DeleteRole deletes a custom role.
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	if err := h.rbacService.DeleteRole(r.Context(), id, *userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete role")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}

// AssignRole assigns a role to a user.
func (h *RoleHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	assignedBy := middleware.GetUserID(r.Context())
	if assignedBy == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var input domain.RoleAssignmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.RoleID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "missing_role_id", "Role ID is required")
		return
	}

	assignment, err := h.rbacService.AssignRole(r.Context(), userID, *assignedBy, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to assign role")
		return
	}

	WriteJSON(w, http.StatusCreated, assignment)
}

// RevokeRole revokes a role from a user.
func (h *RoleHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	revokedBy := middleware.GetUserID(r.Context())
	if revokedBy == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	roleID, err := uuid.Parse(chi.URLParam(r, "roleId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_role_id", "Invalid role ID")
		return
	}

	if err := h.rbacService.RevokeRole(r.Context(), userID, roleID, *revokedBy); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke role")
		return
	}

	WriteSuccess(w, map[string]string{"status": "revoked"})
}

// GetUserRoles retrieves all role assignments for a user.
func (h *RoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	roles, err := h.rbacService.GetUserRoles(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get user roles")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"roles": roles,
	})
}

// GetUserPermissions retrieves all permissions for a user.
func (h *RoleHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	permissions, err := h.rbacService.GetUserPermissions(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get user permissions")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"permissions": permissions,
	})
}

// CheckPermission checks if a user has a specific permission.
func (h *RoleHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	permission := domain.Permission(r.URL.Query().Get("permission"))
	if permission == "" {
		WriteError(w, http.StatusBadRequest, "missing_permission", "Permission is required")
		return
	}

	hasPermission, err := h.rbacService.HasPermissionSimple(r.Context(), userID, permission)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to check permission")
		return
	}

	WriteSuccess(w, map[string]bool{
		"has_permission": hasPermission,
	})
}
