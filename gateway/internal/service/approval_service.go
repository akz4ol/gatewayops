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

// ApprovalService handles tool approval workflows.
type ApprovalService struct {
	toolRepo     *repository.ToolRepository
	auditService *AuditService
	logger       *slog.Logger
}

// NewApprovalService creates a new approval service.
func NewApprovalService(
	toolRepo *repository.ToolRepository,
	auditService *AuditService,
	logger *slog.Logger,
) *ApprovalService {
	return &ApprovalService{
		toolRepo:     toolRepo,
		auditService: auditService,
		logger:       logger,
	}
}

// CheckToolAccess checks if a tool requires approval and if access is granted.
func (s *ApprovalService) CheckToolAccess(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, mcpServer, toolName string) (*ToolAccessResult, error) {
	// Get tool classification
	classification, err := s.toolRepo.GetClassification(ctx, orgID, mcpServer, toolName)
	if err != nil {
		return nil, err
	}

	// If no classification, use default
	if classification == nil {
		defaultLevel := domain.GetDefaultClassification(toolName)
		classification = &domain.ToolClassification{
			MCPServer:        mcpServer,
			ToolName:         toolName,
			Classification:   defaultLevel,
			RequiresApproval: defaultLevel != domain.ToolRiskSafe,
		}
	}

	result := &ToolAccessResult{
		Classification: classification.Classification,
		Allowed:        true,
	}

	// Safe tools are always allowed
	if classification.Classification == domain.ToolRiskSafe {
		return result, nil
	}

	// Dangerous tools are blocked by default
	if classification.Classification == domain.ToolRiskDangerous {
		result.Allowed = false
		result.RequiresApproval = true
		result.Message = "This tool is classified as dangerous and requires explicit approval"
	}

	// Sensitive tools require approval
	if classification.RequiresApproval {
		// Check for active approval
		approval, err := s.toolRepo.GetActiveApproval(ctx, orgID, mcpServer, toolName, userID)
		if err != nil {
			return nil, err
		}

		if approval != nil && approval.Status == domain.ApprovalStatusApproved {
			result.Allowed = true
			result.ApprovalID = &approval.ID
		} else {
			result.Allowed = false
			result.RequiresApproval = true
			result.Message = "This tool requires approval before use"
		}
	}

	return result, nil
}

// ToolAccessResult represents the result of a tool access check.
type ToolAccessResult struct {
	Classification   domain.ToolRiskLevel `json:"classification"`
	Allowed          bool                 `json:"allowed"`
	RequiresApproval bool                 `json:"requires_approval"`
	ApprovalID       *uuid.UUID           `json:"approval_id,omitempty"`
	Message          string               `json:"message,omitempty"`
}

// RequestApproval creates a new tool approval request.
func (s *ApprovalService) RequestApproval(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, request domain.ToolApprovalRequest) (*domain.ToolApproval, error) {
	approval := &domain.ToolApproval{
		ID:          uuid.New(),
		OrgID:       orgID,
		TeamID:      request.TeamID,
		MCPServer:   request.MCPServer,
		ToolName:    request.ToolName,
		RequestedBy: userID,
		RequestedAt: time.Now(),
		Reason:      request.Reason,
		Arguments:   request.Arguments,
		Status:      domain.ApprovalStatusPending,
		TraceID:     request.TraceID,
	}

	if err := s.toolRepo.CreateApproval(ctx, approval); err != nil {
		return nil, fmt.Errorf("create approval: %w", err)
	}

	s.auditService.LogApprovalEvent(ctx, orgID, userID, domain.AuditActionApprovalCreate, approval.ID, request.MCPServer, request.ToolName, map[string]interface{}{
		"reason": request.Reason,
	})

	s.logger.Info("tool approval requested",
		"approval_id", approval.ID,
		"mcp_server", request.MCPServer,
		"tool_name", request.ToolName,
		"requested_by", userID,
	)

	return approval, nil
}

// Approve approves a tool approval request.
func (s *ApprovalService) Approve(ctx context.Context, approvalID uuid.UUID, reviewerID uuid.UUID, review domain.ToolApprovalReview) (*domain.ToolApproval, error) {
	approval, err := s.toolRepo.GetApproval(ctx, approvalID)
	if err != nil {
		return nil, err
	}
	if approval == nil {
		return nil, ErrNotFound{Resource: "approval", ID: approvalID.String()}
	}
	if approval.Status != domain.ApprovalStatusPending {
		return nil, fmt.Errorf("approval is not pending")
	}

	now := time.Now()
	approval.Status = domain.ApprovalStatusApproved
	approval.ReviewedBy = &reviewerID
	approval.ReviewedAt = &now
	approval.ReviewNote = review.ReviewNote

	// Set expiration if specified
	if review.ExpiresIn != nil && *review.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(*review.ExpiresIn) * time.Second)
		approval.ExpiresAt = &expiresAt
	}

	if err := s.toolRepo.UpdateApproval(ctx, approval); err != nil {
		return nil, fmt.Errorf("update approval: %w", err)
	}

	s.auditService.LogApprovalEvent(ctx, approval.OrgID, reviewerID, domain.AuditActionApprovalGrant, approval.ID, approval.MCPServer, approval.ToolName, map[string]interface{}{
		"review_note": review.ReviewNote,
		"expires_in":  review.ExpiresIn,
	})

	s.logger.Info("tool approval granted",
		"approval_id", approval.ID,
		"mcp_server", approval.MCPServer,
		"tool_name", approval.ToolName,
		"reviewed_by", reviewerID,
	)

	return approval, nil
}

// Deny denies a tool approval request.
func (s *ApprovalService) Deny(ctx context.Context, approvalID uuid.UUID, reviewerID uuid.UUID, review domain.ToolApprovalReview) (*domain.ToolApproval, error) {
	approval, err := s.toolRepo.GetApproval(ctx, approvalID)
	if err != nil {
		return nil, err
	}
	if approval == nil {
		return nil, ErrNotFound{Resource: "approval", ID: approvalID.String()}
	}
	if approval.Status != domain.ApprovalStatusPending {
		return nil, fmt.Errorf("approval is not pending")
	}

	now := time.Now()
	approval.Status = domain.ApprovalStatusDenied
	approval.ReviewedBy = &reviewerID
	approval.ReviewedAt = &now
	approval.ReviewNote = review.ReviewNote

	if err := s.toolRepo.UpdateApproval(ctx, approval); err != nil {
		return nil, fmt.Errorf("update approval: %w", err)
	}

	s.auditService.LogApprovalEvent(ctx, approval.OrgID, reviewerID, domain.AuditActionApprovalDeny, approval.ID, approval.MCPServer, approval.ToolName, map[string]interface{}{
		"review_note": review.ReviewNote,
	})

	s.logger.Info("tool approval denied",
		"approval_id", approval.ID,
		"mcp_server", approval.MCPServer,
		"tool_name", approval.ToolName,
		"reviewed_by", reviewerID,
	)

	return approval, nil
}

// GetApproval retrieves an approval by ID.
func (s *ApprovalService) GetApproval(ctx context.Context, id uuid.UUID) (*domain.ToolApproval, error) {
	return s.toolRepo.GetApproval(ctx, id)
}

// ListApprovals retrieves approvals with filtering.
func (s *ApprovalService) ListApprovals(ctx context.Context, filter domain.ToolApprovalFilter) (*domain.ToolApprovalPage, error) {
	return s.toolRepo.ListApprovals(ctx, filter)
}

// ListPendingApprovals retrieves pending approvals for an organization.
func (s *ApprovalService) ListPendingApprovals(ctx context.Context, orgID uuid.UUID, limit int) (*domain.ToolApprovalPage, error) {
	return s.toolRepo.ListApprovals(ctx, domain.ToolApprovalFilter{
		OrgID:    orgID,
		Statuses: []domain.ApprovalStatus{domain.ApprovalStatusPending},
		Limit:    limit,
	})
}

// ExpireApprovals marks expired approvals.
func (s *ApprovalService) ExpireApprovals(ctx context.Context) (int64, error) {
	count, err := s.toolRepo.ExpireApprovals(ctx)
	if err != nil {
		return 0, err
	}

	if count > 0 {
		s.logger.Info("expired approvals", "count", count)
	}

	return count, nil
}

// SetClassification sets or updates a tool classification.
func (s *ApprovalService) SetClassification(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, input domain.ToolClassificationInput) (*domain.ToolClassification, error) {
	classification := &domain.ToolClassification{
		ID:               uuid.New(),
		OrgID:            orgID,
		MCPServer:        input.MCPServer,
		ToolName:         input.ToolName,
		Classification:   input.Classification,
		RequiresApproval: input.RequiresApproval,
		Description:      input.Description,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        userID,
	}

	// Set defaults based on classification if not specified
	if input.Classification == domain.ToolRiskSafe {
		classification.RequiresApproval = false
	} else if input.Classification == domain.ToolRiskDangerous {
		classification.RequiresApproval = true
	}

	if err := s.toolRepo.CreateClassification(ctx, classification); err != nil {
		return nil, fmt.Errorf("create classification: %w", err)
	}

	s.logger.Info("tool classification set",
		"mcp_server", input.MCPServer,
		"tool_name", input.ToolName,
		"classification", input.Classification,
	)

	return classification, nil
}

// GetClassification retrieves a tool classification.
func (s *ApprovalService) GetClassification(ctx context.Context, orgID uuid.UUID, mcpServer, toolName string) (*domain.ToolClassification, error) {
	return s.toolRepo.GetClassification(ctx, orgID, mcpServer, toolName)
}

// ListClassifications retrieves all tool classifications.
func (s *ApprovalService) ListClassifications(ctx context.Context, orgID uuid.UUID, mcpServer string) ([]domain.ToolClassification, error) {
	return s.toolRepo.ListClassifications(ctx, orgID, mcpServer)
}

// DeleteClassification removes a tool classification.
func (s *ApprovalService) DeleteClassification(ctx context.Context, orgID uuid.UUID, mcpServer, toolName string) error {
	return s.toolRepo.DeleteClassification(ctx, orgID, mcpServer, toolName)
}
