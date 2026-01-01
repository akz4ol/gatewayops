// Package repository provides data access layer implementations.
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

// AuditRepository handles audit log persistence.
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository.
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create inserts a new audit log entry.
func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	details, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("marshal details: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			id, org_id, team_id, user_id, api_key_id, trace_id,
			action, resource, resource_id, outcome, details,
			ip_address, user_agent, request_id, duration_ms, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err = r.db.ExecContext(ctx, query,
		log.ID, log.OrgID, log.TeamID, log.UserID, log.APIKeyID, log.TraceID,
		log.Action, log.Resource, log.ResourceID, log.Outcome, details,
		log.IPAddress, log.UserAgent, log.RequestID, log.DurationMS, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}

// Get retrieves an audit log by ID.
func (r *AuditRepository) Get(ctx context.Context, orgID, id uuid.UUID) (*domain.AuditLog, error) {
	query := `
		SELECT id, org_id, team_id, user_id, api_key_id, trace_id,
			   action, resource, resource_id, outcome, details,
			   ip_address, user_agent, request_id, duration_ms, created_at
		FROM audit_logs
		WHERE id = $1 AND org_id = $2`

	var log domain.AuditLog
	var details []byte
	var teamID, userID, apiKeyID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&log.ID, &log.OrgID, &teamID, &userID, &apiKeyID, &log.TraceID,
		&log.Action, &log.Resource, &log.ResourceID, &log.Outcome, &details,
		&log.IPAddress, &log.UserAgent, &log.RequestID, &log.DurationMS, &log.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query audit log: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		log.TeamID = &tid
	}
	if userID.Valid {
		uid, _ := uuid.Parse(userID.String)
		log.UserID = &uid
	}
	if apiKeyID.Valid {
		kid, _ := uuid.Parse(apiKeyID.String)
		log.APIKeyID = &kid
	}

	if len(details) > 0 {
		if err := json.Unmarshal(details, &log.Details); err != nil {
			return nil, fmt.Errorf("unmarshal details: %w", err)
		}
	}

	return &log, nil
}

// List retrieves audit logs with filtering and pagination.
func (r *AuditRepository) List(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogPage, error) {
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

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *filter.UserID)
		argNum++
	}

	if filter.APIKeyID != nil {
		conditions = append(conditions, fmt.Sprintf("api_key_id = $%d", argNum))
		args = append(args, *filter.APIKeyID)
		argNum++
	}

	if len(filter.Actions) > 0 {
		placeholders := make([]string, len(filter.Actions))
		for i, action := range filter.Actions {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, action)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("action IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Outcomes) > 0 {
		placeholders := make([]string, len(filter.Outcomes))
		for i, outcome := range filter.Outcomes {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, outcome)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("outcome IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argNum))
		args = append(args, filter.Resource)
		argNum++
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argNum))
		args = append(args, *filter.StartTime)
		argNum++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argNum))
		args = append(args, *filter.EndTime)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count audit logs: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, team_id, user_id, api_key_id, trace_id,
			   action, resource, resource_id, outcome, details,
			   ip_address, user_agent, request_id, duration_ms, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var details []byte
		var teamID, userID, apiKeyID sql.NullString

		err := rows.Scan(
			&log.ID, &log.OrgID, &teamID, &userID, &apiKeyID, &log.TraceID,
			&log.Action, &log.Resource, &log.ResourceID, &log.Outcome, &details,
			&log.IPAddress, &log.UserAgent, &log.RequestID, &log.DurationMS, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}

		if teamID.Valid {
			tid, _ := uuid.Parse(teamID.String)
			log.TeamID = &tid
		}
		if userID.Valid {
			uid, _ := uuid.Parse(userID.String)
			log.UserID = &uid
		}
		if apiKeyID.Valid {
			kid, _ := uuid.Parse(apiKeyID.String)
			log.APIKeyID = &kid
		}

		if len(details) > 0 {
			json.Unmarshal(details, &log.Details)
		}

		logs = append(logs, log)
	}

	return &domain.AuditLogPage{
		Logs:    logs,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(logs)) < total,
	}, nil
}

// DeleteOlderThan deletes audit logs older than the specified time.
func (r *AuditRepository) DeleteOlderThan(ctx context.Context, orgID uuid.UUID, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM audit_logs WHERE org_id = $1 AND created_at < $2",
		orgID, before,
	)
	if err != nil {
		return 0, fmt.Errorf("delete audit logs: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return count, nil
}
