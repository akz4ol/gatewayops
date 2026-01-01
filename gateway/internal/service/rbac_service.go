package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
)

// RBACService handles role-based access control.
type RBACService struct {
	roleRepo     *repository.RoleRepository
	userRepo     *repository.UserRepository
	auditService *AuditService
	logger       *slog.Logger
}

// NewRBACService creates a new RBAC service.
func NewRBACService(
	roleRepo *repository.RoleRepository,
	userRepo *repository.UserRepository,
	auditService *AuditService,
	logger *slog.Logger,
) *RBACService {
	return &RBACService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		auditService: auditService,
		logger:       logger,
	}
}

// CreateRole creates a new custom role.
func (s *RBACService) CreateRole(ctx context.Context, orgID uuid.UUID, createdBy uuid.UUID, input domain.RoleInput) (*domain.Role, error) {
	// Validate permissions
	if len(input.Permissions) == 0 {
		return nil, fmt.Errorf("at least one permission is required")
	}

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

	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	s.auditService.LogRoleChange(ctx, orgID, createdBy, domain.AuditActionRoleCreate, role.ID, nil, map[string]interface{}{
		"role_name":   role.Name,
		"permissions": role.Permissions,
	})

	s.logger.Info("role created",
		"role_id", role.ID,
		"name", role.Name,
		"org_id", orgID,
	)

	return role, nil
}

// GetRole retrieves a role by ID.
func (s *RBACService) GetRole(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetRole(ctx, id)
}

// GetRoleByName retrieves a role by name.
func (s *RBACService) GetRoleByName(ctx context.Context, orgID *uuid.UUID, name string) (*domain.Role, error) {
	return s.roleRepo.GetRoleByName(ctx, orgID, name)
}

// ListRoles retrieves all roles for an organization.
func (s *RBACService) ListRoles(ctx context.Context, orgID uuid.UUID) ([]domain.Role, error) {
	return s.roleRepo.ListRoles(ctx, orgID)
}

// UpdateRole updates a custom role.
func (s *RBACService) UpdateRole(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID, input domain.RoleInput) (*domain.Role, error) {
	role, err := s.roleRepo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	if role.IsBuiltin {
		return nil, fmt.Errorf("cannot modify builtin role")
	}

	role.Name = input.Name
	role.Description = input.Description
	role.Permissions = input.Permissions
	role.UpdatedAt = time.Now()

	if err := s.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	if role.OrgID != nil {
		s.auditService.LogRoleChange(ctx, *role.OrgID, updatedBy, domain.AuditActionRoleUpdate, role.ID, nil, map[string]interface{}{
			"role_name":   role.Name,
			"permissions": role.Permissions,
		})
	}

	return role, nil
}

// DeleteRole deletes a custom role.
func (s *RBACService) DeleteRole(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	role, err := s.roleRepo.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}
	if role.IsBuiltin {
		return fmt.Errorf("cannot delete builtin role")
	}

	if err := s.roleRepo.DeleteRole(ctx, id); err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	if role.OrgID != nil {
		s.auditService.LogRoleChange(ctx, *role.OrgID, deletedBy, domain.AuditActionRoleDelete, role.ID, nil, map[string]interface{}{
			"role_name": role.Name,
		})
	}

	s.logger.Info("role deleted",
		"role_id", id,
		"name", role.Name,
	)

	return nil
}

// AssignRole assigns a role to a user.
func (s *RBACService) AssignRole(ctx context.Context, userID uuid.UUID, assignedBy uuid.UUID, input domain.RoleAssignmentInput) (*domain.RoleAssignment, error) {
	// Verify role exists
	role, err := s.roleRepo.GetRole(ctx, input.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}

	// Verify user exists
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	assignment := &domain.RoleAssignment{
		ID:        uuid.New(),
		UserID:    userID,
		RoleID:    input.RoleID,
		ScopeType: input.ScopeType,
		ScopeID:   input.ScopeID,
		CreatedAt: time.Now(),
		CreatedBy: assignedBy,
	}

	if err := s.roleRepo.CreateRoleAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("create role assignment: %w", err)
	}

	s.auditService.LogRoleChange(ctx, user.OrgID, assignedBy, domain.AuditActionRoleAssign, input.RoleID, &userID, map[string]interface{}{
		"role_name":  role.Name,
		"scope_type": input.ScopeType,
		"scope_id":   input.ScopeID,
	})

	s.logger.Info("role assigned",
		"user_id", userID,
		"role_id", input.RoleID,
		"role_name", role.Name,
		"scope_type", input.ScopeType,
	)

	return assignment, nil
}

// RevokeRole removes a role from a user.
func (s *RBACService) RevokeRole(ctx context.Context, userID, roleID uuid.UUID, revokedBy uuid.UUID) error {
	// Verify user exists
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Get role for audit
	role, _ := s.roleRepo.GetRole(ctx, roleID)

	if err := s.roleRepo.DeleteRoleAssignment(ctx, userID, roleID); err != nil {
		return fmt.Errorf("delete role assignment: %w", err)
	}

	roleName := ""
	if role != nil {
		roleName = role.Name
	}

	s.auditService.LogRoleChange(ctx, user.OrgID, revokedBy, domain.AuditActionRoleRevoke, roleID, &userID, map[string]interface{}{
		"role_name": roleName,
	})

	s.logger.Info("role revoked",
		"user_id", userID,
		"role_id", roleID,
	)

	return nil
}

// GetUserRoles retrieves all role assignments for a user.
func (s *RBACService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]domain.RoleAssignment, error) {
	return s.roleRepo.ListUserRoles(ctx, userID)
}

// GetUserPermissions retrieves all permissions for a user.
func (s *RBACService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]domain.Permission, error) {
	return s.roleRepo.GetUserPermissions(ctx, userID)
}

// HasPermission checks if a user has a specific permission.
func (s *RBACService) HasPermission(ctx context.Context, userID uuid.UUID, permission domain.Permission, scopeType domain.ScopeType, scopeID *uuid.UUID) (bool, error) {
	return s.roleRepo.HasPermission(ctx, userID, permission, scopeType, scopeID)
}

// HasPermissionSimple checks if a user has a permission without scope.
func (s *RBACService) HasPermissionSimple(ctx context.Context, userID uuid.UUID, permission domain.Permission) (bool, error) {
	return s.roleRepo.HasPermission(ctx, userID, permission, "", nil)
}

// RequirePermission checks if a user has a permission and returns an error if not.
func (s *RBACService) RequirePermission(ctx context.Context, userID uuid.UUID, permission domain.Permission) error {
	has, err := s.HasPermissionSimple(ctx, userID, permission)
	if err != nil {
		return err
	}
	if !has {
		return ErrPermissionDenied{Permission: permission}
	}
	return nil
}

// ErrPermissionDenied represents a permission denied error.
type ErrPermissionDenied struct {
	Permission domain.Permission
}

func (e ErrPermissionDenied) Error() string {
	return fmt.Sprintf("permission denied: %s", e.Permission)
}

// SeedBuiltinRoles ensures builtin roles exist.
func (s *RBACService) SeedBuiltinRoles(ctx context.Context) error {
	return s.roleRepo.SeedBuiltinRoles(ctx)
}

// GetUserWithRoles retrieves a user with their role assignments.
func (s *RBACService) GetUserWithRoles(ctx context.Context, userID uuid.UUID) (*domain.UserWithRoles, error) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	roles, err := s.roleRepo.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserWithRoles{
		User:  *user,
		Roles: roles,
	}, nil
}

// GetUser retrieves a user by ID.
func (s *RBACService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetUser(ctx, userID)
}

// ListUsers retrieves all users in an organization.
func (s *RBACService) ListUsers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.User, int64, error) {
	return s.userRepo.ListUsersByOrg(ctx, orgID, limit, offset)
}
