// Package rbac provides role-based access control functionality.
package rbac

import (
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service manages roles and permissions.
type Service struct {
	logger      zerolog.Logger
	roles       map[uuid.UUID]*domain.Role
	assignments map[uuid.UUID][]domain.RoleAssignment // keyed by user ID
	mu          sync.RWMutex
}

// NewService creates a new RBAC service.
func NewService(logger zerolog.Logger) *Service {
	s := &Service{
		logger:      logger,
		roles:       make(map[uuid.UUID]*domain.Role),
		assignments: make(map[uuid.UUID][]domain.RoleAssignment),
	}

	// Initialize built-in roles
	s.initBuiltinRoles()

	// Create demo user with admin role
	s.createDemoAssignment()

	logger.Info().Msg("RBAC service initialized")
	return s
}

func (s *Service) initBuiltinRoles() {
	for _, r := range domain.BuiltinRoles {
		role := r // Copy
		role.ID = uuid.New()
		role.CreatedAt = time.Now()
		role.UpdatedAt = time.Now()
		s.roles[role.ID] = &role
	}
}

func (s *Service) createDemoAssignment() {
	demoUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Find admin role
	var adminRoleID uuid.UUID
	for id, role := range s.roles {
		if role.Name == "admin" {
			adminRoleID = id
			break
		}
	}

	if adminRoleID != uuid.Nil {
		assignment := domain.RoleAssignment{
			ID:        uuid.New(),
			UserID:    demoUserID,
			RoleID:    adminRoleID,
			CreatedAt: time.Now(),
			CreatedBy: demoUserID,
		}
		s.assignments[demoUserID] = []domain.RoleAssignment{assignment}
	}
}

// CreateRole creates a new custom role.
func (s *Service) CreateRole(input domain.RoleInput, orgID uuid.UUID) *domain.Role {
	s.mu.Lock()
	defer s.mu.Unlock()

	role := &domain.Role{
		ID:          uuid.New(),
		OrgID:       &orgID,
		Name:        input.Name,
		Description: input.Description,
		Permissions: input.Permissions,
		IsBuiltin:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.roles[role.ID] = role

	s.logger.Info().
		Str("role_id", role.ID.String()).
		Str("name", role.Name).
		Int("permissions", len(role.Permissions)).
		Msg("Role created")

	return role
}

// GetRole returns a role by ID.
func (s *Service) GetRole(id uuid.UUID) *domain.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.roles[id]
}

// GetRoleByName returns a role by name.
func (s *Service) GetRoleByName(name string) *domain.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, role := range s.roles {
		if role.Name == name {
			return role
		}
	}
	return nil
}

// ListRoles returns all roles.
func (s *Service) ListRoles(includeBuiltin bool) []domain.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]domain.Role, 0, len(s.roles))
	for _, r := range s.roles {
		if !includeBuiltin && r.IsBuiltin {
			continue
		}
		roles = append(roles, *r)
	}
	return roles
}

// UpdateRole updates an existing role.
func (s *Service) UpdateRole(id uuid.UUID, input domain.RoleInput) *domain.Role {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[id]
	if !exists {
		return nil
	}

	// Cannot update built-in roles
	if role.IsBuiltin {
		return nil
	}

	role.Name = input.Name
	role.Description = input.Description
	role.Permissions = input.Permissions
	role.UpdatedAt = time.Now()

	s.logger.Info().
		Str("role_id", id.String()).
		Str("name", role.Name).
		Msg("Role updated")

	return role
}

// DeleteRole deletes a custom role.
func (s *Service) DeleteRole(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[id]
	if !exists {
		return false
	}

	// Cannot delete built-in roles
	if role.IsBuiltin {
		return false
	}

	delete(s.roles, id)

	// Remove all assignments for this role
	for userID, assignments := range s.assignments {
		filtered := make([]domain.RoleAssignment, 0)
		for _, a := range assignments {
			if a.RoleID != id {
				filtered = append(filtered, a)
			}
		}
		s.assignments[userID] = filtered
	}

	s.logger.Info().
		Str("role_id", id.String()).
		Msg("Role deleted")

	return true
}

// AssignRole assigns a role to a user.
func (s *Service) AssignRole(userID uuid.UUID, input domain.RoleAssignmentInput, assignedBy uuid.UUID) *domain.RoleAssignment {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify role exists
	if _, exists := s.roles[input.RoleID]; !exists {
		return nil
	}

	// Check if already assigned
	for _, a := range s.assignments[userID] {
		if a.RoleID == input.RoleID && a.ScopeType == input.ScopeType {
			if input.ScopeID == nil && a.ScopeID == nil {
				return &a // Already assigned
			}
			if input.ScopeID != nil && a.ScopeID != nil && *input.ScopeID == *a.ScopeID {
				return &a // Already assigned
			}
		}
	}

	assignment := domain.RoleAssignment{
		ID:        uuid.New(),
		UserID:    userID,
		RoleID:    input.RoleID,
		ScopeType: input.ScopeType,
		ScopeID:   input.ScopeID,
		CreatedAt: time.Now(),
		CreatedBy: assignedBy,
	}

	s.assignments[userID] = append(s.assignments[userID], assignment)

	s.logger.Info().
		Str("user_id", userID.String()).
		Str("role_id", input.RoleID.String()).
		Str("scope_type", string(input.ScopeType)).
		Msg("Role assigned")

	return &assignment
}

// RevokeRole removes a role assignment.
func (s *Service) RevokeRole(userID uuid.UUID, assignmentID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	assignments, exists := s.assignments[userID]
	if !exists {
		return false
	}

	for i, a := range assignments {
		if a.ID == assignmentID {
			s.assignments[userID] = append(assignments[:i], assignments[i+1:]...)

			s.logger.Info().
				Str("user_id", userID.String()).
				Str("assignment_id", assignmentID.String()).
				Msg("Role revoked")

			return true
		}
	}

	return false
}

// GetUserRoles returns all role assignments for a user.
func (s *Service) GetUserRoles(userID uuid.UUID) []domain.RoleAssignment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := s.assignments[userID]
	if assignments == nil {
		return []domain.RoleAssignment{}
	}
	return assignments
}

// GetUserPermissions returns all effective permissions for a user.
func (s *Service) GetUserPermissions(userID uuid.UUID) []domain.Permission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	permSet := make(map[domain.Permission]bool)
	assignments := s.assignments[userID]

	for _, a := range assignments {
		role, exists := s.roles[a.RoleID]
		if !exists {
			continue
		}
		for _, p := range role.Permissions {
			permSet[p] = true
		}
	}

	perms := make([]domain.Permission, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return perms
}

// HasPermission checks if a user has a specific permission.
func (s *Service) HasPermission(userID uuid.UUID, permission domain.Permission, scopeType domain.ScopeType, scopeID *uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assignments := s.assignments[userID]
	for _, a := range assignments {
		// Check scope match
		if scopeType != "" && a.ScopeType != "" && a.ScopeType != scopeType {
			continue
		}
		if scopeID != nil && a.ScopeID != nil && *scopeID != *a.ScopeID {
			continue
		}

		role, exists := s.roles[a.RoleID]
		if !exists {
			continue
		}

		if role.HasPermission(permission) {
			return true
		}
	}

	return false
}

// CheckPermission performs a permission check and returns detailed result.
func (s *Service) CheckPermission(check domain.PermissionCheck) PermissionResult {
	hasPermission := s.HasPermission(check.UserID, check.Permission, check.ScopeType, check.ScopeID)

	return PermissionResult{
		Allowed:    hasPermission,
		Permission: check.Permission,
		UserID:     check.UserID,
		ScopeType:  check.ScopeType,
		ScopeID:    check.ScopeID,
	}
}

// PermissionResult represents the result of a permission check.
type PermissionResult struct {
	Allowed    bool              `json:"allowed"`
	Permission domain.Permission `json:"permission"`
	UserID     uuid.UUID         `json:"user_id"`
	ScopeType  domain.ScopeType  `json:"scope_type,omitempty"`
	ScopeID    *uuid.UUID        `json:"scope_id,omitempty"`
}

// GetRoleUsers returns all users with a specific role.
func (s *Service) GetRoleUsers(roleID uuid.UUID) []uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]uuid.UUID, 0)
	for userID, assignments := range s.assignments {
		for _, a := range assignments {
			if a.RoleID == roleID {
				users = append(users, userID)
				break
			}
		}
	}
	return users
}

// GetAllPermissions returns all available permissions.
func (s *Service) GetAllPermissions() []PermissionInfo {
	return []PermissionInfo{
		{Permission: domain.PermissionAll, Category: "admin", Description: "Full access to all resources"},
		{Permission: domain.PermissionMCPRead, Category: "mcp", Description: "Read MCP server info and tool listings"},
		{Permission: domain.PermissionMCPWrite, Category: "mcp", Description: "Modify MCP configurations"},
		{Permission: domain.PermissionMCPCall, Category: "mcp", Description: "Call MCP tools"},
		{Permission: domain.PermissionTracesRead, Category: "traces", Description: "Read all traces"},
		{Permission: domain.PermissionTracesReadOwn, Category: "traces", Description: "Read own traces only"},
		{Permission: domain.PermissionTracesExport, Category: "traces", Description: "Export trace data"},
		{Permission: domain.PermissionCostsRead, Category: "costs", Description: "Read all cost data"},
		{Permission: domain.PermissionCostsReadTeam, Category: "costs", Description: "Read team cost data only"},
		{Permission: domain.PermissionCostsExport, Category: "costs", Description: "Export cost data"},
		{Permission: domain.PermissionAuditRead, Category: "audit", Description: "Read audit logs"},
		{Permission: domain.PermissionAuditExport, Category: "audit", Description: "Export audit logs"},
		{Permission: domain.PermissionKeysRead, Category: "keys", Description: "View API keys"},
		{Permission: domain.PermissionKeysCreate, Category: "keys", Description: "Create API keys"},
		{Permission: domain.PermissionKeysRevoke, Category: "keys", Description: "Revoke API keys"},
		{Permission: domain.PermissionKeysRotate, Category: "keys", Description: "Rotate API keys"},
		{Permission: domain.PermissionRBACRead, Category: "rbac", Description: "View roles and permissions"},
		{Permission: domain.PermissionRBACAdmin, Category: "rbac", Description: "Manage roles and permissions"},
		{Permission: domain.PermissionUsersRead, Category: "users", Description: "View user information"},
		{Permission: domain.PermissionUsersAdmin, Category: "users", Description: "Manage users"},
		{Permission: domain.PermissionTeamsRead, Category: "teams", Description: "View team information"},
		{Permission: domain.PermissionTeamsAdmin, Category: "teams", Description: "Manage teams"},
		{Permission: domain.PermissionPoliciesRead, Category: "policies", Description: "View safety policies"},
		{Permission: domain.PermissionPoliciesAdmin, Category: "policies", Description: "Manage safety policies"},
		{Permission: domain.PermissionApprovalsRead, Category: "approvals", Description: "View approval requests"},
		{Permission: domain.PermissionApprovalsRequest, Category: "approvals", Description: "Submit approval requests"},
		{Permission: domain.PermissionApprovalsReview, Category: "approvals", Description: "Review and approve requests"},
		{Permission: domain.PermissionAlertsRead, Category: "alerts", Description: "View alerts"},
		{Permission: domain.PermissionAlertsAdmin, Category: "alerts", Description: "Manage alert rules"},
		{Permission: domain.PermissionSettingsRead, Category: "settings", Description: "View settings"},
		{Permission: domain.PermissionSettingsAdmin, Category: "settings", Description: "Manage settings"},
	}
}

// PermissionInfo provides information about a permission.
type PermissionInfo struct {
	Permission  domain.Permission `json:"permission"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
}
