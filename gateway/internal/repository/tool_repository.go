package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
)

// ToolRepository handles tool classification and approval persistence.
type ToolRepository struct {
	db *sql.DB
}

// NewToolRepository creates a new tool repository.
func NewToolRepository(db *sql.DB) *ToolRepository {
	return &ToolRepository{db: db}
}

// CreateClassification inserts a new tool classification.
func (r *ToolRepository) CreateClassification(ctx context.Context, classification *domain.ToolClassification) error {
	query := `
		INSERT INTO tool_classifications (
			id, org_id, mcp_server, tool_name, classification,
			requires_approval, description, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (org_id, mcp_server, tool_name)
		DO UPDATE SET
			classification = EXCLUDED.classification,
			requires_approval = EXCLUDED.requires_approval,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at`

	_, err := r.db.ExecContext(ctx, query,
		classification.ID, classification.OrgID, classification.MCPServer,
		classification.ToolName, classification.Classification, classification.RequiresApproval,
		classification.Description, classification.CreatedAt, classification.UpdatedAt, classification.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert tool classification: %w", err)
	}

	return nil
}

// GetClassification retrieves a tool classification.
func (r *ToolRepository) GetClassification(ctx context.Context, orgID uuid.UUID, mcpServer, toolName string) (*domain.ToolClassification, error) {
	query := `
		SELECT id, org_id, mcp_server, tool_name, classification,
			   requires_approval, description, created_at, updated_at, created_by
		FROM tool_classifications
		WHERE org_id = $1 AND mcp_server = $2 AND tool_name = $3`

	var classification domain.ToolClassification
	err := r.db.QueryRowContext(ctx, query, orgID, mcpServer, toolName).Scan(
		&classification.ID, &classification.OrgID, &classification.MCPServer,
		&classification.ToolName, &classification.Classification, &classification.RequiresApproval,
		&classification.Description, &classification.CreatedAt, &classification.UpdatedAt, &classification.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query tool classification: %w", err)
	}

	return &classification, nil
}

// ListClassifications retrieves all tool classifications for an organization.
func (r *ToolRepository) ListClassifications(ctx context.Context, orgID uuid.UUID, mcpServer string) ([]domain.ToolClassification, error) {
	var query string
	var args []interface{}

	if mcpServer != "" {
		query = `
			SELECT id, org_id, mcp_server, tool_name, classification,
				   requires_approval, description, created_at, updated_at, created_by
			FROM tool_classifications
			WHERE org_id = $1 AND mcp_server = $2
			ORDER BY mcp_server, tool_name`
		args = []interface{}{orgID, mcpServer}
	} else {
		query = `
			SELECT id, org_id, mcp_server, tool_name, classification,
				   requires_approval, description, created_at, updated_at, created_by
			FROM tool_classifications
			WHERE org_id = $1
			ORDER BY mcp_server, tool_name`
		args = []interface{}{orgID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tool classifications: %w", err)
	}
	defer rows.Close()

	var classifications []domain.ToolClassification
	for rows.Next() {
		var c domain.ToolClassification
		err := rows.Scan(
			&c.ID, &c.OrgID, &c.MCPServer, &c.ToolName, &c.Classification,
			&c.RequiresApproval, &c.Description, &c.CreatedAt, &c.UpdatedAt, &c.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tool classification: %w", err)
		}
		classifications = append(classifications, c)
	}

	return classifications, nil
}

// DeleteClassification removes a tool classification.
func (r *ToolRepository) DeleteClassification(ctx context.Context, orgID uuid.UUID, mcpServer, toolName string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM tool_classifications WHERE org_id = $1 AND mcp_server = $2 AND tool_name = $3",
		orgID, mcpServer, toolName,
	)
	if err != nil {
		return fmt.Errorf("delete tool classification: %w", err)
	}

	return nil
}

// CreateApproval inserts a new tool approval request.
func (r *ToolRepository) CreateApproval(ctx context.Context, approval *domain.ToolApproval) error {
	arguments, _ := json.Marshal(approval.Arguments)

	query := `
		INSERT INTO tool_approvals (
			id, org_id, team_id, mcp_server, tool_name,
			requested_by, requested_at, reason, arguments,
			status, trace_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		approval.ID, approval.OrgID, approval.TeamID, approval.MCPServer, approval.ToolName,
		approval.RequestedBy, approval.RequestedAt, approval.Reason, arguments,
		approval.Status, approval.TraceID,
	)
	if err != nil {
		return fmt.Errorf("insert tool approval: %w", err)
	}

	return nil
}

// GetApproval retrieves a tool approval by ID.
func (r *ToolRepository) GetApproval(ctx context.Context, id uuid.UUID) (*domain.ToolApproval, error) {
	query := `
		SELECT id, org_id, team_id, mcp_server, tool_name,
			   requested_by, requested_at, reason, arguments,
			   status, reviewed_by, reviewed_at, review_note, expires_at, trace_id
		FROM tool_approvals
		WHERE id = $1`

	var approval domain.ToolApproval
	var teamID, reviewedBy sql.NullString
	var reviewedAt, expiresAt sql.NullTime
	var arguments []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&approval.ID, &approval.OrgID, &teamID, &approval.MCPServer, &approval.ToolName,
		&approval.RequestedBy, &approval.RequestedAt, &approval.Reason, &arguments,
		&approval.Status, &reviewedBy, &reviewedAt, &approval.ReviewNote, &expiresAt, &approval.TraceID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query tool approval: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		approval.TeamID = &tid
	}
	if reviewedBy.Valid {
		rid, _ := uuid.Parse(reviewedBy.String)
		approval.ReviewedBy = &rid
	}
	if reviewedAt.Valid {
		approval.ReviewedAt = &reviewedAt.Time
	}
	if expiresAt.Valid {
		approval.ExpiresAt = &expiresAt.Time
	}
	if len(arguments) > 0 {
		json.Unmarshal(arguments, &approval.Arguments)
	}

	return &approval, nil
}

// UpdateApproval updates a tool approval (for reviewing).
func (r *ToolRepository) UpdateApproval(ctx context.Context, approval *domain.ToolApproval) error {
	query := `
		UPDATE tool_approvals SET
			status = $2, reviewed_by = $3, reviewed_at = $4,
			review_note = $5, expires_at = $6
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		approval.ID, approval.Status, approval.ReviewedBy, approval.ReviewedAt,
		approval.ReviewNote, approval.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("update tool approval: %w", err)
	}

	return nil
}

// ListApprovals retrieves tool approvals with filtering.
func (r *ToolRepository) ListApprovals(ctx context.Context, filter domain.ToolApprovalFilter) (*domain.ToolApprovalPage, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argNum))
		args = append(args, *filter.TeamID)
		argNum++
	}

	if filter.MCPServer != "" {
		conditions = append(conditions, fmt.Sprintf("mcp_server = $%d", argNum))
		args = append(args, filter.MCPServer)
		argNum++
	}

	if filter.ToolName != "" {
		conditions = append(conditions, fmt.Sprintf("tool_name = $%d", argNum))
		args = append(args, filter.ToolName)
		argNum++
	}

	if filter.RequestedBy != nil {
		conditions = append(conditions, fmt.Sprintf("requested_by = $%d", argNum))
		args = append(args, *filter.RequestedBy)
		argNum++
	}

	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, status)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tool_approvals WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count tool approvals: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, team_id, mcp_server, tool_name,
			   requested_by, requested_at, reason, arguments,
			   status, reviewed_by, reviewed_at, review_note, expires_at, trace_id
		FROM tool_approvals
		WHERE %s
		ORDER BY requested_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tool approvals: %w", err)
	}
	defer rows.Close()

	var approvals []domain.ToolApproval
	for rows.Next() {
		var approval domain.ToolApproval
		var teamID, reviewedBy sql.NullString
		var reviewedAt, expiresAt sql.NullTime
		var arguments []byte

		err := rows.Scan(
			&approval.ID, &approval.OrgID, &teamID, &approval.MCPServer, &approval.ToolName,
			&approval.RequestedBy, &approval.RequestedAt, &approval.Reason, &arguments,
			&approval.Status, &reviewedBy, &reviewedAt, &approval.ReviewNote, &expiresAt, &approval.TraceID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tool approval: %w", err)
		}

		if teamID.Valid {
			tid, _ := uuid.Parse(teamID.String)
			approval.TeamID = &tid
		}
		if reviewedBy.Valid {
			rid, _ := uuid.Parse(reviewedBy.String)
			approval.ReviewedBy = &rid
		}
		if reviewedAt.Valid {
			approval.ReviewedAt = &reviewedAt.Time
		}
		if expiresAt.Valid {
			approval.ExpiresAt = &expiresAt.Time
		}
		if len(arguments) > 0 {
			json.Unmarshal(arguments, &approval.Arguments)
		}

		approvals = append(approvals, approval)
	}

	return &domain.ToolApprovalPage{
		Approvals: approvals,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		HasMore:   int64(offset+len(approvals)) < total,
	}, nil
}

// GetActiveApproval retrieves an active (non-expired) approval for a tool.
func (r *ToolRepository) GetActiveApproval(ctx context.Context, orgID uuid.UUID, mcpServer, toolName string, userID uuid.UUID) (*domain.ToolApproval, error) {
	query := `
		SELECT id, org_id, team_id, mcp_server, tool_name,
			   requested_by, requested_at, reason, arguments,
			   status, reviewed_by, reviewed_at, review_note, expires_at, trace_id
		FROM tool_approvals
		WHERE org_id = $1 AND mcp_server = $2 AND tool_name = $3
			  AND requested_by = $4 AND status = 'approved'
			  AND (expires_at IS NULL OR expires_at > $5)
		ORDER BY requested_at DESC
		LIMIT 1`

	var approval domain.ToolApproval
	var teamID, reviewedBy sql.NullString
	var reviewedAt, expiresAt sql.NullTime
	var arguments []byte

	err := r.db.QueryRowContext(ctx, query, orgID, mcpServer, toolName, userID, time.Now()).Scan(
		&approval.ID, &approval.OrgID, &teamID, &approval.MCPServer, &approval.ToolName,
		&approval.RequestedBy, &approval.RequestedAt, &approval.Reason, &arguments,
		&approval.Status, &reviewedBy, &reviewedAt, &approval.ReviewNote, &expiresAt, &approval.TraceID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query active approval: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		approval.TeamID = &tid
	}
	if reviewedBy.Valid {
		rid, _ := uuid.Parse(reviewedBy.String)
		approval.ReviewedBy = &rid
	}
	if reviewedAt.Valid {
		approval.ReviewedAt = &reviewedAt.Time
	}
	if expiresAt.Valid {
		approval.ExpiresAt = &expiresAt.Time
	}
	if len(arguments) > 0 {
		json.Unmarshal(arguments, &approval.Arguments)
	}

	return &approval, nil
}

// ExpireApprovals marks expired approvals as expired.
func (r *ToolRepository) ExpireApprovals(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tool_approvals SET status = 'expired' WHERE status = 'approved' AND expires_at < $1",
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("expire approvals: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return count, nil
}
