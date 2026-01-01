package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
)

// RoleRepository handles role and permission persistence.
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository.
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// CreateRole inserts a new role.
func (r *RoleRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	permissions, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}

	query := `
		INSERT INTO roles (
			id, org_id, name, description, permissions, is_builtin, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = r.db.ExecContext(ctx, query,
		role.ID, role.OrgID, role.Name, role.Description, permissions,
		role.IsBuiltin, role.CreatedAt, role.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert role: %w", err)
	}

	return nil
}

// GetRole retrieves a role by ID.
func (r *RoleRepository) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	query := `
		SELECT id, org_id, name, description, permissions, is_builtin, created_at, updated_at
		FROM roles
		WHERE id = $1`

	var role domain.Role
	var orgID sql.NullString
	var permissions []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &orgID, &role.Name, &role.Description, &permissions,
		&role.IsBuiltin, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query role: %w", err)
	}

	if orgID.Valid {
		oid, _ := uuid.Parse(orgID.String)
		role.OrgID = &oid
	}

	if err := json.Unmarshal(permissions, &role.Permissions); err != nil {
		return nil, fmt.Errorf("unmarshal permissions: %w", err)
	}

	return &role, nil
}

// GetRoleByName retrieves a role by name within an organization (or builtin).
func (r *RoleRepository) GetRoleByName(ctx context.Context, orgID *uuid.UUID, name string) (*domain.Role, error) {
	var query string
	var args []interface{}

	if orgID != nil {
		query = `
			SELECT id, org_id, name, description, permissions, is_builtin, created_at, updated_at
			FROM roles
			WHERE (org_id = $1 OR org_id IS NULL) AND name = $2
			ORDER BY org_id DESC NULLS LAST
			LIMIT 1`
		args = []interface{}{*orgID, name}
	} else {
		query = `
			SELECT id, org_id, name, description, permissions, is_builtin, created_at, updated_at
			FROM roles
			WHERE org_id IS NULL AND name = $1`
		args = []interface{}{name}
	}

	var role domain.Role
	var oid sql.NullString
	var permissions []byte

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&role.ID, &oid, &role.Name, &role.Description, &permissions,
		&role.IsBuiltin, &role.CreatedAt, &role.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query role by name: %w", err)
	}

	if oid.Valid {
		ouid, _ := uuid.Parse(oid.String)
		role.OrgID = &ouid
	}

	if err := json.Unmarshal(permissions, &role.Permissions); err != nil {
		return nil, fmt.Errorf("unmarshal permissions: %w", err)
	}

	return &role, nil
}

// ListRoles retrieves all roles for an organization (including builtins).
func (r *RoleRepository) ListRoles(ctx context.Context, orgID uuid.UUID) ([]domain.Role, error) {
	query := `
		SELECT id, org_id, name, description, permissions, is_builtin, created_at, updated_at
		FROM roles
		WHERE org_id = $1 OR org_id IS NULL
		ORDER BY is_builtin DESC, name ASC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		var oid sql.NullString
		var permissions []byte

		err := rows.Scan(
			&role.ID, &oid, &role.Name, &role.Description, &permissions,
			&role.IsBuiltin, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}

		if oid.Valid {
			ouid, _ := uuid.Parse(oid.String)
			role.OrgID = &ouid
		}

		json.Unmarshal(permissions, &role.Permissions)
		roles = append(roles, role)
	}

	return roles, nil
}

// UpdateRole updates an existing role.
func (r *RoleRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	if role.IsBuiltin {
		return fmt.Errorf("cannot update builtin role")
	}

	permissions, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}

	query := `
		UPDATE roles SET
			name = $2, description = $3, permissions = $4, updated_at = $5
		WHERE id = $1 AND is_builtin = false`

	result, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, permissions, role.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("role not found or is builtin")
	}

	return nil
}

// DeleteRole deletes a custom role.
func (r *RoleRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM roles WHERE id = $1 AND is_builtin = false", id,
	)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	count, _ := result.RowsAffected()
	if count == 0 {
		return fmt.Errorf("role not found or is builtin")
	}

	return nil
}

// CreateRoleAssignment assigns a role to a user.
func (r *RoleRepository) CreateRoleAssignment(ctx context.Context, assignment *domain.RoleAssignment) error {
	query := `
		INSERT INTO user_roles (
			id, user_id, role_id, scope_type, scope_id, created_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	scopeType := string(assignment.ScopeType)
	if scopeType == "" {
		scopeType = "global"
	}

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.UserID, assignment.RoleID,
		scopeType, assignment.ScopeID, assignment.CreatedAt, assignment.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert role assignment: %w", err)
	}

	return nil
}

// GetRoleAssignment retrieves a role assignment by ID.
func (r *RoleRepository) GetRoleAssignment(ctx context.Context, id uuid.UUID) (*domain.RoleAssignment, error) {
	query := `
		SELECT id, user_id, role_id, scope_type, scope_id, created_at, created_by
		FROM user_roles
		WHERE id = $1`

	var assignment domain.RoleAssignment
	var scopeType string
	var scopeID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&assignment.ID, &assignment.UserID, &assignment.RoleID,
		&scopeType, &scopeID, &assignment.CreatedAt, &assignment.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query role assignment: %w", err)
	}

	if scopeType != "global" {
		assignment.ScopeType = domain.ScopeType(scopeType)
	}
	if scopeID.Valid {
		sid, _ := uuid.Parse(scopeID.String)
		assignment.ScopeID = &sid
	}

	return &assignment, nil
}

// ListUserRoles retrieves all role assignments for a user.
func (r *RoleRepository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.RoleAssignment, error) {
	query := `
		SELECT id, user_id, role_id, scope_type, scope_id, created_at, created_by
		FROM user_roles
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user roles: %w", err)
	}
	defer rows.Close()

	var assignments []domain.RoleAssignment
	for rows.Next() {
		var assignment domain.RoleAssignment
		var scopeType string
		var scopeID sql.NullString

		err := rows.Scan(
			&assignment.ID, &assignment.UserID, &assignment.RoleID,
			&scopeType, &scopeID, &assignment.CreatedAt, &assignment.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan role assignment: %w", err)
		}

		if scopeType != "global" {
			assignment.ScopeType = domain.ScopeType(scopeType)
		}
		if scopeID.Valid {
			sid, _ := uuid.Parse(scopeID.String)
			assignment.ScopeID = &sid
		}

		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// DeleteRoleAssignment removes a role assignment.
func (r *RoleRepository) DeleteRoleAssignment(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
		userID, roleID,
	)
	if err != nil {
		return fmt.Errorf("delete role assignment: %w", err)
	}

	return nil
}

// GetUserPermissions retrieves all permissions for a user (resolved from roles).
func (r *RoleRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	query := `
		SELECT DISTINCT r.permissions
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user permissions: %w", err)
	}
	defer rows.Close()

	permissionSet := make(map[domain.Permission]bool)
	for rows.Next() {
		var permBytes []byte
		if err := rows.Scan(&permBytes); err != nil {
			return nil, fmt.Errorf("scan permissions: %w", err)
		}

		var perms []domain.Permission
		if err := json.Unmarshal(permBytes, &perms); err != nil {
			continue
		}

		for _, p := range perms {
			permissionSet[p] = true
		}
	}

	permissions := make([]domain.Permission, 0, len(permissionSet))
	for p := range permissionSet {
		permissions = append(permissions, p)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission.
func (r *RoleRepository) HasPermission(ctx context.Context, userID uuid.UUID, permission domain.Permission, scopeType domain.ScopeType, scopeID *uuid.UUID) (bool, error) {
	// Get all roles for the user
	var query string
	var args []interface{}

	if scopeType == "" || scopeType == domain.ScopeTypeGlobal {
		query = `
			SELECT r.permissions
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1 AND (ur.scope_type = 'global' OR ur.scope_type = '')`
		args = []interface{}{userID}
	} else {
		query = `
			SELECT r.permissions
			FROM user_roles ur
			JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = $1 AND (
				(ur.scope_type = 'global' OR ur.scope_type = '') OR
				(ur.scope_type = $2 AND ur.scope_id = $3)
			)`
		args = []interface{}{userID, scopeType, scopeID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("query permissions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var permBytes []byte
		if err := rows.Scan(&permBytes); err != nil {
			continue
		}

		var perms []domain.Permission
		if err := json.Unmarshal(permBytes, &perms); err != nil {
			continue
		}

		for _, p := range perms {
			if p == domain.PermissionAll || p == permission {
				return true, nil
			}
		}
	}

	return false, nil
}

// SeedBuiltinRoles ensures builtin roles exist in the database.
func (r *RoleRepository) SeedBuiltinRoles(ctx context.Context) error {
	for _, role := range domain.BuiltinRoles {
		permissions, _ := json.Marshal(role.Permissions)

		query := `
			INSERT INTO roles (id, org_id, name, description, permissions, is_builtin, created_at, updated_at)
			VALUES ($1, NULL, $2, $3, $4, true, NOW(), NOW())
			ON CONFLICT (name) WHERE org_id IS NULL AND is_builtin = true
			DO NOTHING`

		_, err := r.db.ExecContext(ctx, query,
			uuid.New(), role.Name, role.Description, permissions,
		)
		if err != nil {
			return fmt.Errorf("seed builtin role %s: %w", role.Name, err)
		}
	}

	return nil
}
