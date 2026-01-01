// Package service contains business logic implementations.
package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
)

// AuditService handles audit logging operations.
type AuditService struct {
	repo   *repository.AuditRepository
	logger *slog.Logger
}

// NewAuditService creates a new audit service.
func NewAuditService(repo *repository.AuditRepository, logger *slog.Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

// LogEvent creates an audit log entry.
func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error {
	log := &domain.AuditLog{
		ID:         uuid.New(),
		OrgID:      event.OrgID,
		TeamID:     event.TeamID,
		UserID:     event.UserID,
		APIKeyID:   event.APIKeyID,
		TraceID:    event.TraceID,
		Action:     event.Action,
		Resource:   event.Resource,
		ResourceID: event.ResourceID,
		Outcome:    event.Outcome,
		Details:    event.Details,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		RequestID:  event.RequestID,
		DurationMS: event.DurationMS,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.Error("failed to create audit log",
			"action", event.Action,
			"resource", event.Resource,
			"error", err,
		)
		return err
	}

	s.logger.Debug("audit log created",
		"id", log.ID,
		"action", event.Action,
		"resource", event.Resource,
		"outcome", event.Outcome,
	)

	return nil
}

// AuditEvent represents input for creating an audit log.
type AuditEvent struct {
	OrgID      uuid.UUID
	TeamID     *uuid.UUID
	UserID     *uuid.UUID
	APIKeyID   *uuid.UUID
	TraceID    string
	Action     domain.AuditAction
	Resource   string
	ResourceID string
	Outcome    domain.AuditOutcome
	Details    map[string]interface{}
	IPAddress  string
	UserAgent  string
	RequestID  string
	DurationMS int64
}

// Get retrieves an audit log by ID.
func (s *AuditService) Get(ctx context.Context, orgID, id uuid.UUID) (*domain.AuditLog, error) {
	return s.repo.Get(ctx, orgID, id)
}

// List retrieves audit logs with filtering.
func (s *AuditService) List(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogPage, error) {
	return s.repo.List(ctx, filter)
}

// Search is an alias for List with additional search capabilities.
func (s *AuditService) Search(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogPage, error) {
	return s.repo.List(ctx, filter)
}

// Cleanup removes old audit logs based on retention policy.
func (s *AuditService) Cleanup(ctx context.Context, orgID uuid.UUID, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 90 // Default 90 days retention
	}

	before := time.Now().AddDate(0, 0, -retentionDays)
	count, err := s.repo.DeleteOlderThan(ctx, orgID, before)
	if err != nil {
		return 0, err
	}

	s.logger.Info("audit log cleanup completed",
		"org_id", orgID,
		"deleted_count", count,
		"retention_days", retentionDays,
	)

	return count, nil
}

// LogMCPToolCall logs an MCP tool call.
func (s *AuditService) LogMCPToolCall(ctx context.Context, orgID uuid.UUID, apiKeyID uuid.UUID, traceID, mcpServer, toolName string, success bool, durationMS int64, ipAddress, userAgent string) {
	outcome := domain.AuditOutcomeSuccess
	if !success {
		outcome = domain.AuditOutcomeFailure
	}

	s.LogEvent(ctx, AuditEvent{
		OrgID:      orgID,
		APIKeyID:   &apiKeyID,
		TraceID:    traceID,
		Action:     domain.AuditActionMCPToolCall,
		Resource:   mcpServer,
		ResourceID: toolName,
		Outcome:    outcome,
		DurationMS: durationMS,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Details: map[string]interface{}{
			"mcp_server": mcpServer,
			"tool_name":  toolName,
		},
	})
}

// LogAPIKeyOperation logs an API key operation.
func (s *AuditService) LogAPIKeyOperation(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, action domain.AuditAction, keyID uuid.UUID, keyName string) {
	s.LogEvent(ctx, AuditEvent{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     action,
		Resource:   "api_key",
		ResourceID: keyID.String(),
		Outcome:    domain.AuditOutcomeSuccess,
		Details: map[string]interface{}{
			"key_name": keyName,
		},
	})
}

// LogAuthEvent logs an authentication event.
func (s *AuditService) LogAuthEvent(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, action domain.AuditAction, success bool, ipAddress, userAgent string, details map[string]interface{}) {
	outcome := domain.AuditOutcomeSuccess
	if !success {
		outcome = domain.AuditOutcomeFailure
	}

	s.LogEvent(ctx, AuditEvent{
		OrgID:     orgID,
		UserID:    &userID,
		Action:    action,
		Resource:  "auth",
		Outcome:   outcome,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
	})
}

// LogRoleChange logs a role-related change.
func (s *AuditService) LogRoleChange(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, action domain.AuditAction, roleID uuid.UUID, targetUserID *uuid.UUID, details map[string]interface{}) {
	event := AuditEvent{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     action,
		Resource:   "role",
		ResourceID: roleID.String(),
		Outcome:    domain.AuditOutcomeSuccess,
		Details:    details,
	}

	if targetUserID != nil {
		if event.Details == nil {
			event.Details = make(map[string]interface{})
		}
		event.Details["target_user_id"] = targetUserID.String()
	}

	s.LogEvent(ctx, event)
}

// LogPolicyChange logs a policy-related change.
func (s *AuditService) LogPolicyChange(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, action domain.AuditAction, policyID uuid.UUID, details map[string]interface{}) {
	s.LogEvent(ctx, AuditEvent{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     action,
		Resource:   "policy",
		ResourceID: policyID.String(),
		Outcome:    domain.AuditOutcomeSuccess,
		Details:    details,
	})
}

// LogApprovalEvent logs a tool approval event.
func (s *AuditService) LogApprovalEvent(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, action domain.AuditAction, approvalID uuid.UUID, mcpServer, toolName string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["mcp_server"] = mcpServer
	details["tool_name"] = toolName

	s.LogEvent(ctx, AuditEvent{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     action,
		Resource:   "approval",
		ResourceID: approvalID.String(),
		Outcome:    domain.AuditOutcomeSuccess,
		Details:    details,
	})
}

// LogConfigChange logs a configuration change.
func (s *AuditService) LogConfigChange(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, configType string, details map[string]interface{}) {
	s.LogEvent(ctx, AuditEvent{
		OrgID:      orgID,
		UserID:     &userID,
		Action:     domain.AuditActionConfigChange,
		Resource:   "config",
		ResourceID: configType,
		Outcome:    domain.AuditOutcomeSuccess,
		Details:    details,
	})
}
