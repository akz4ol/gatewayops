package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/rbac"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// RBACHandler handles RBAC-related HTTP requests.
type RBACHandler struct {
	logger  zerolog.Logger
	service *rbac.Service
}

// NewRBACHandler creates a new RBAC handler.
func NewRBACHandler(logger zerolog.Logger, service *rbac.Service) *RBACHandler {
	return &RBACHandler{
		logger:  logger,
		service: service,
	}
}

// ListRoles returns all roles.
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	includeBuiltin := r.URL.Query().Get("include_builtin") != "false"
	roles := h.service.ListRoles(includeBuiltin)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"roles": roles,
		"total": len(roles),
	})
}

// GetRole returns a specific role.
func (h *RBACHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "roleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	role := h.service.GetRole(id)
	if role == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Role not found")
		return
	}

	// Include user count
	users := h.service.GetRoleUsers(id)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"role":       role,
		"user_count": len(users),
	})
}

// CreateRole creates a new custom role.
func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var input domain.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if len(input.Permissions) == 0 {
		WriteError(w, http.StatusBadRequest, "validation_error", "At least one permission is required")
		return
	}

	// Check for duplicate name
	if existing := h.service.GetRoleByName(input.Name); existing != nil {
		WriteError(w, http.StatusConflict, "duplicate_name", "A role with this name already exists")
		return
	}

	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	role := h.service.CreateRole(input, orgID)
	WriteJSON(w, http.StatusCreated, role)
}

// UpdateRole updates an existing role.
func (h *RBACHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "roleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	var input domain.RoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	// Check if built-in
	existing := h.service.GetRole(id)
	if existing != nil && existing.IsBuiltin {
		WriteError(w, http.StatusForbidden, "builtin_role", "Built-in roles cannot be modified")
		return
	}

	role := h.service.UpdateRole(id, input)
	if role == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Role not found")
		return
	}

	WriteJSON(w, http.StatusOK, role)
}

// DeleteRole deletes a custom role.
func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "roleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	// Check if built-in
	existing := h.service.GetRole(id)
	if existing != nil && existing.IsBuiltin {
		WriteError(w, http.StatusForbidden, "builtin_role", "Built-in roles cannot be deleted")
		return
	}

	if !h.service.DeleteRole(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Role not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetUserRoles returns roles assigned to a user.
func (h *RBACHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	assignments := h.service.GetUserRoles(id)
	permissions := h.service.GetUserPermissions(id)

	// Enrich assignments with role details
	enriched := make([]map[string]interface{}, 0, len(assignments))
	for _, a := range assignments {
		role := h.service.GetRole(a.RoleID)
		enriched = append(enriched, map[string]interface{}{
			"assignment": a,
			"role":       role,
		})
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":     id,
		"assignments": enriched,
		"permissions": permissions,
	})
}

// AssignRole assigns a role to a user.
func (h *RBACHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	var input domain.RoleAssignmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.RoleID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Role ID is required")
		return
	}

	// Verify role exists
	if h.service.GetRole(input.RoleID) == nil {
		WriteError(w, http.StatusNotFound, "role_not_found", "Role not found")
		return
	}

	// Demo assigner
	assignedBy := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assignment := h.service.AssignRole(userID, input, assignedBy)
	if assignment == nil {
		WriteError(w, http.StatusBadRequest, "assignment_failed", "Failed to assign role")
		return
	}

	WriteJSON(w, http.StatusCreated, assignment)
}

// RevokeRole removes a role assignment from a user.
func (h *RBACHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	assignmentIDStr := chi.URLParam(r, "assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid assignment ID")
		return
	}

	if !h.service.RevokeRole(userID, assignmentID) {
		WriteError(w, http.StatusNotFound, "not_found", "Assignment not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// CheckPermission checks if a user has a specific permission.
func (h *RBACHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	permission := r.URL.Query().Get("permission")

	if userIDStr == "" {
		// Use demo user
		userIDStr = "00000000-0000-0000-0000-000000000001"
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	if permission == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Permission is required")
		return
	}

	check := domain.PermissionCheck{
		UserID:     userID,
		Permission: domain.Permission(permission),
	}

	// Optional scope
	if scopeType := r.URL.Query().Get("scope_type"); scopeType != "" {
		check.ScopeType = domain.ScopeType(scopeType)
	}
	if scopeIDStr := r.URL.Query().Get("scope_id"); scopeIDStr != "" {
		if scopeID, err := uuid.Parse(scopeIDStr); err == nil {
			check.ScopeID = &scopeID
		}
	}

	result := h.service.CheckPermission(check)
	WriteJSON(w, http.StatusOK, result)
}

// ListPermissions returns all available permissions.
func (h *RBACHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions := h.service.GetAllPermissions()

	// Group by category
	categories := make(map[string][]rbac.PermissionInfo)
	for _, p := range permissions {
		categories[p.Category] = append(categories[p.Category], p)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"permissions": permissions,
		"categories":  categories,
		"total":       len(permissions),
	})
}

// GetMyPermissions returns permissions for the current user.
func (h *RBACHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	// Demo user
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	assignments := h.service.GetUserRoles(userID)
	permissions := h.service.GetUserPermissions(userID)

	// Enrich assignments with role details
	roles := make([]domain.Role, 0)
	for _, a := range assignments {
		if role := h.service.GetRole(a.RoleID); role != nil {
			roles = append(roles, *role)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"roles":       roles,
		"permissions": permissions,
	})
}

// GetRoleUsers returns users assigned to a role.
func (h *RBACHandler) GetRoleUsers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "roleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID")
		return
	}

	// Parse pagination
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users := h.service.GetRoleUsers(id)
	total := len(users)

	// Apply pagination
	start := offset
	if start > len(users) {
		start = len(users)
	}
	end := start + limit
	if end > len(users) {
		end = len(users)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"role_id":  id,
		"user_ids": users[start:end],
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"has_more": end < len(users),
	})
}
