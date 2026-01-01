package domain

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a single permission in the system.
type Permission string

const (
	// Wildcard permission (admin only)
	PermissionAll Permission = "*"

	// MCP permissions
	PermissionMCPRead  Permission = "mcp:read"
	PermissionMCPWrite Permission = "mcp:write"
	PermissionMCPCall  Permission = "mcp:call"

	// Trace permissions
	PermissionTracesRead    Permission = "traces:read"
	PermissionTracesReadOwn Permission = "traces:read:own"
	PermissionTracesExport  Permission = "traces:export"

	// Cost permissions
	PermissionCostsRead     Permission = "costs:read"
	PermissionCostsReadTeam Permission = "costs:read:team"
	PermissionCostsExport   Permission = "costs:export"

	// Audit permissions
	PermissionAuditRead   Permission = "audit:read"
	PermissionAuditExport Permission = "audit:export"

	// API Key permissions
	PermissionKeysRead   Permission = "keys:read"
	PermissionKeysCreate Permission = "keys:create"
	PermissionKeysRevoke Permission = "keys:revoke"
	PermissionKeysRotate Permission = "keys:rotate"

	// RBAC permissions
	PermissionRBACRead  Permission = "rbac:read"
	PermissionRBACAdmin Permission = "rbac:admin"

	// User permissions
	PermissionUsersRead  Permission = "users:read"
	PermissionUsersAdmin Permission = "users:admin"

	// Team permissions
	PermissionTeamsRead  Permission = "teams:read"
	PermissionTeamsAdmin Permission = "teams:admin"

	// Safety/Policy permissions
	PermissionPoliciesRead  Permission = "policies:read"
	PermissionPoliciesAdmin Permission = "policies:admin"

	// Approval permissions
	PermissionApprovalsRead    Permission = "approvals:read"
	PermissionApprovalsRequest Permission = "approvals:request"
	PermissionApprovalsReview  Permission = "approvals:review"

	// Alert permissions
	PermissionAlertsRead  Permission = "alerts:read"
	PermissionAlertsAdmin Permission = "alerts:admin"

	// Settings permissions
	PermissionSettingsRead  Permission = "settings:read"
	PermissionSettingsAdmin Permission = "settings:admin"
)

// Role represents a role with a set of permissions.
type Role struct {
	ID          uuid.UUID    `json:"id"`
	OrgID       *uuid.UUID   `json:"org_id,omitempty"` // nil for built-in roles
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
	IsBuiltin   bool         `json:"is_builtin"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// RoleInput represents input for creating/updating a role.
type RoleInput struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Permissions []Permission `json:"permissions"`
}

// ScopeType represents the type of scope for a role assignment.
type ScopeType string

const (
	ScopeTypeGlobal    ScopeType = ""          // No scope - applies globally within org
	ScopeTypeTeam      ScopeType = "team"      // Scoped to a specific team
	ScopeTypeMCPServer ScopeType = "mcp_server" // Scoped to a specific MCP server
)

// RoleAssignment represents the assignment of a role to a user.
type RoleAssignment struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	RoleID    uuid.UUID  `json:"role_id"`
	ScopeType ScopeType  `json:"scope_type,omitempty"`
	ScopeID   *uuid.UUID `json:"scope_id,omitempty"` // Team ID or MCP Server ID
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy uuid.UUID  `json:"created_by"`
}

// RoleAssignmentInput represents input for assigning a role.
type RoleAssignmentInput struct {
	RoleID    uuid.UUID  `json:"role_id"`
	ScopeType ScopeType  `json:"scope_type,omitempty"`
	ScopeID   *uuid.UUID `json:"scope_id,omitempty"`
}

// UserWithRoles represents a user with their role assignments.
type UserWithRoles struct {
	User  User             `json:"user"`
	Roles []RoleAssignment `json:"roles"`
}

// PermissionCheck represents a permission check request.
type PermissionCheck struct {
	UserID     uuid.UUID  `json:"user_id"`
	Permission Permission `json:"permission"`
	ScopeType  ScopeType  `json:"scope_type,omitempty"`
	ScopeID    *uuid.UUID `json:"scope_id,omitempty"`
}

// BuiltinRoles defines the default built-in roles.
var BuiltinRoles = []Role{
	{
		Name:        "admin",
		Description: "Full access to all resources and settings",
		Permissions: []Permission{PermissionAll},
		IsBuiltin:   true,
	},
	{
		Name:        "developer",
		Description: "Access to MCP tools, traces, and own API keys",
		Permissions: []Permission{
			PermissionMCPRead,
			PermissionMCPCall,
			PermissionTracesRead,
			PermissionCostsReadTeam,
			PermissionKeysRead,
			PermissionKeysCreate,
		},
		IsBuiltin: true,
	},
	{
		Name:        "viewer",
		Description: "Read-only access to traces and costs",
		Permissions: []Permission{
			PermissionTracesReadOwn,
			PermissionCostsReadTeam,
		},
		IsBuiltin: true,
	},
	{
		Name:        "billing",
		Description: "Access to costs and usage data only",
		Permissions: []Permission{
			PermissionCostsRead,
			PermissionCostsExport,
		},
		IsBuiltin: true,
	},
}

// HasPermission checks if a role has a specific permission.
func (r *Role) HasPermission(perm Permission) bool {
	for _, p := range r.Permissions {
		if p == PermissionAll || p == perm {
			return true
		}
		// Check for wildcard matches (e.g., "mcp:*" matches "mcp:read")
		if len(p) > 0 && p[len(p)-1] == '*' {
			prefix := string(p[:len(p)-1])
			if len(perm) >= len(prefix) && string(perm[:len(prefix)]) == prefix {
				return true
			}
		}
	}
	return false
}
