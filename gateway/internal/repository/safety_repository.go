package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
)

// SafetyRepository handles safety policy and detection persistence.
type SafetyRepository struct {
	db *sql.DB
}

// NewSafetyRepository creates a new safety repository.
func NewSafetyRepository(db *sql.DB) *SafetyRepository {
	return &SafetyRepository{db: db}
}

// CreatePolicy inserts a new safety policy.
func (r *SafetyRepository) CreatePolicy(ctx context.Context, policy *domain.SafetyPolicy) error {
	patterns, _ := json.Marshal(policy.Patterns)
	mcpServers, _ := json.Marshal(policy.MCPServers)

	query := `
		INSERT INTO safety_policies (
			id, org_id, name, description, sensitivity, mode,
			patterns, mcp_servers, enabled, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.OrgID, policy.Name, policy.Description, policy.Sensitivity,
		policy.Mode, patterns, mcpServers, policy.Enabled,
		policy.CreatedAt, policy.UpdatedAt, policy.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert safety policy: %w", err)
	}

	return nil
}

// GetPolicy retrieves a safety policy by ID.
func (r *SafetyRepository) GetPolicy(ctx context.Context, id uuid.UUID) (*domain.SafetyPolicy, error) {
	query := `
		SELECT id, org_id, name, description, sensitivity, mode,
			   patterns, mcp_servers, enabled, created_at, updated_at, created_by
		FROM safety_policies
		WHERE id = $1`

	var policy domain.SafetyPolicy
	var patterns, mcpServers []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&policy.ID, &policy.OrgID, &policy.Name, &policy.Description, &policy.Sensitivity,
		&policy.Mode, &patterns, &mcpServers, &policy.Enabled,
		&policy.CreatedAt, &policy.UpdatedAt, &policy.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query safety policy: %w", err)
	}

	json.Unmarshal(patterns, &policy.Patterns)
	json.Unmarshal(mcpServers, &policy.MCPServers)

	return &policy, nil
}

// ListPolicies retrieves all safety policies for an organization.
func (r *SafetyRepository) ListPolicies(ctx context.Context, orgID uuid.UUID, enabledOnly bool) ([]domain.SafetyPolicy, error) {
	var query string
	var args []interface{}

	if enabledOnly {
		query = `
			SELECT id, org_id, name, description, sensitivity, mode,
				   patterns, mcp_servers, enabled, created_at, updated_at, created_by
			FROM safety_policies
			WHERE org_id = $1 AND enabled = true
			ORDER BY created_at DESC`
		args = []interface{}{orgID}
	} else {
		query = `
			SELECT id, org_id, name, description, sensitivity, mode,
				   patterns, mcp_servers, enabled, created_at, updated_at, created_by
			FROM safety_policies
			WHERE org_id = $1
			ORDER BY created_at DESC`
		args = []interface{}{orgID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query safety policies: %w", err)
	}
	defer rows.Close()

	var policies []domain.SafetyPolicy
	for rows.Next() {
		var policy domain.SafetyPolicy
		var patterns, mcpServers []byte

		err := rows.Scan(
			&policy.ID, &policy.OrgID, &policy.Name, &policy.Description, &policy.Sensitivity,
			&policy.Mode, &patterns, &mcpServers, &policy.Enabled,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan safety policy: %w", err)
		}

		json.Unmarshal(patterns, &policy.Patterns)
		json.Unmarshal(mcpServers, &policy.MCPServers)

		policies = append(policies, policy)
	}

	return policies, nil
}

// GetPoliciesForServer retrieves enabled policies that apply to a specific MCP server.
func (r *SafetyRepository) GetPoliciesForServer(ctx context.Context, orgID uuid.UUID, mcpServer string) ([]domain.SafetyPolicy, error) {
	query := `
		SELECT id, org_id, name, description, sensitivity, mode,
			   patterns, mcp_servers, enabled, created_at, updated_at, created_by
		FROM safety_policies
		WHERE org_id = $1 AND enabled = true
			  AND (mcp_servers IS NULL OR mcp_servers = '[]' OR mcp_servers @> $2)
		ORDER BY created_at DESC`

	serverJSON, _ := json.Marshal([]string{mcpServer})

	rows, err := r.db.QueryContext(ctx, query, orgID, serverJSON)
	if err != nil {
		return nil, fmt.Errorf("query policies for server: %w", err)
	}
	defer rows.Close()

	var policies []domain.SafetyPolicy
	for rows.Next() {
		var policy domain.SafetyPolicy
		var patterns, mcpServers []byte

		err := rows.Scan(
			&policy.ID, &policy.OrgID, &policy.Name, &policy.Description, &policy.Sensitivity,
			&policy.Mode, &patterns, &mcpServers, &policy.Enabled,
			&policy.CreatedAt, &policy.UpdatedAt, &policy.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan safety policy: %w", err)
		}

		json.Unmarshal(patterns, &policy.Patterns)
		json.Unmarshal(mcpServers, &policy.MCPServers)

		policies = append(policies, policy)
	}

	return policies, nil
}

// UpdatePolicy updates a safety policy.
func (r *SafetyRepository) UpdatePolicy(ctx context.Context, policy *domain.SafetyPolicy) error {
	patterns, _ := json.Marshal(policy.Patterns)
	mcpServers, _ := json.Marshal(policy.MCPServers)

	query := `
		UPDATE safety_policies SET
			name = $2, description = $3, sensitivity = $4, mode = $5,
			patterns = $6, mcp_servers = $7, enabled = $8, updated_at = $9
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		policy.ID, policy.Name, policy.Description, policy.Sensitivity, policy.Mode,
		patterns, mcpServers, policy.Enabled, policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update safety policy: %w", err)
	}

	return nil
}

// DeletePolicy deletes a safety policy.
func (r *SafetyRepository) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM safety_policies WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete safety policy: %w", err)
	}

	return nil
}

// CreateDetection inserts a new injection detection.
func (r *SafetyRepository) CreateDetection(ctx context.Context, detection *domain.InjectionDetection) error {
	query := `
		INSERT INTO injection_detections (
			id, org_id, trace_id, span_id, policy_id, type, severity,
			pattern_matched, input, action_taken, mcp_server, tool_name,
			api_key_id, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	// Truncate input if too long
	input := detection.Input
	if len(input) > 1000 {
		input = input[:1000] + "..."
	}

	_, err := r.db.ExecContext(ctx, query,
		detection.ID, detection.OrgID, detection.TraceID, detection.SpanID,
		detection.PolicyID, detection.Type, detection.Severity,
		detection.PatternMatched, input, detection.ActionTaken,
		detection.MCPServer, detection.ToolName, detection.APIKeyID,
		detection.IPAddress, detection.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert injection detection: %w", err)
	}

	return nil
}

// GetDetection retrieves an injection detection by ID.
func (r *SafetyRepository) GetDetection(ctx context.Context, id uuid.UUID) (*domain.InjectionDetection, error) {
	query := `
		SELECT id, org_id, trace_id, span_id, policy_id, type, severity,
			   pattern_matched, input, action_taken, mcp_server, tool_name,
			   api_key_id, ip_address, created_at
		FROM injection_detections
		WHERE id = $1`

	var detection domain.InjectionDetection
	var policyID, apiKeyID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&detection.ID, &detection.OrgID, &detection.TraceID, &detection.SpanID,
		&policyID, &detection.Type, &detection.Severity,
		&detection.PatternMatched, &detection.Input, &detection.ActionTaken,
		&detection.MCPServer, &detection.ToolName, &apiKeyID,
		&detection.IPAddress, &detection.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query injection detection: %w", err)
	}

	if policyID.Valid {
		pid, _ := uuid.Parse(policyID.String)
		detection.PolicyID = &pid
	}
	if apiKeyID.Valid {
		kid, _ := uuid.Parse(apiKeyID.String)
		detection.APIKeyID = &kid
	}

	return &detection, nil
}

// ListDetections retrieves injection detections with filtering.
func (r *SafetyRepository) ListDetections(ctx context.Context, filter domain.DetectionFilter) (*domain.DetectionPage, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	if len(filter.Types) > 0 {
		placeholders := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, t)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("type IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Severities) > 0 {
		placeholders := make([]string, len(filter.Severities))
		for i, s := range filter.Severities {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, s)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("severity IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Actions) > 0 {
		placeholders := make([]string, len(filter.Actions))
		for i, a := range filter.Actions {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, a)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("action_taken IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.MCPServer != "" {
		conditions = append(conditions, fmt.Sprintf("mcp_server = $%d", argNum))
		args = append(args, filter.MCPServer)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM injection_detections WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count detections: %w", err)
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
		SELECT id, org_id, trace_id, span_id, policy_id, type, severity,
			   pattern_matched, input, action_taken, mcp_server, tool_name,
			   api_key_id, ip_address, created_at
		FROM injection_detections
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query detections: %w", err)
	}
	defer rows.Close()

	var detections []domain.InjectionDetection
	for rows.Next() {
		var detection domain.InjectionDetection
		var policyID, apiKeyID sql.NullString

		err := rows.Scan(
			&detection.ID, &detection.OrgID, &detection.TraceID, &detection.SpanID,
			&policyID, &detection.Type, &detection.Severity,
			&detection.PatternMatched, &detection.Input, &detection.ActionTaken,
			&detection.MCPServer, &detection.ToolName, &apiKeyID,
			&detection.IPAddress, &detection.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan detection: %w", err)
		}

		if policyID.Valid {
			pid, _ := uuid.Parse(policyID.String)
			detection.PolicyID = &pid
		}
		if apiKeyID.Valid {
			kid, _ := uuid.Parse(apiKeyID.String)
			detection.APIKeyID = &kid
		}

		detections = append(detections, detection)
	}

	return &domain.DetectionPage{
		Detections: detections,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    int64(offset+len(detections)) < total,
	}, nil
}

// GetSummary retrieves a summary of safety detections.
func (r *SafetyRepository) GetSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.SafetySummary, error) {
	var interval string
	switch period {
	case "day":
		interval = "1 day"
	case "week":
		interval = "7 days"
	case "month":
		interval = "30 days"
	default:
		interval = "24 hours"
		period = "day"
	}

	// Total count
	var total int64
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM injection_detections WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '%s'", interval),
		orgID,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count total detections: %w", err)
	}

	// By type
	byType := make(map[string]int64)
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT type, COUNT(*) FROM injection_detections WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '%s' GROUP BY type", interval),
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("query by type: %w", err)
	}
	for rows.Next() {
		var t string
		var count int64
		rows.Scan(&t, &count)
		byType[t] = count
	}
	rows.Close()

	// By severity
	bySeverity := make(map[string]int64)
	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT severity, COUNT(*) FROM injection_detections WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '%s' GROUP BY severity", interval),
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("query by severity: %w", err)
	}
	for rows.Next() {
		var s string
		var count int64
		rows.Scan(&s, &count)
		bySeverity[s] = count
	}
	rows.Close()

	// By action
	byAction := make(map[string]int64)
	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT action_taken, COUNT(*) FROM injection_detections WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '%s' GROUP BY action_taken", interval),
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("query by action: %w", err)
	}
	for rows.Next() {
		var a string
		var count int64
		rows.Scan(&a, &count)
		byAction[a] = count
	}
	rows.Close()

	// Top patterns
	var topPatterns []domain.PatternCount
	rows, err = r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT pattern_matched, COUNT(*) as cnt FROM injection_detections WHERE org_id = $1 AND created_at >= NOW() - INTERVAL '%s' AND pattern_matched IS NOT NULL GROUP BY pattern_matched ORDER BY cnt DESC LIMIT 10", interval),
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("query top patterns: %w", err)
	}
	for rows.Next() {
		var pc domain.PatternCount
		rows.Scan(&pc.Pattern, &pc.Count)
		topPatterns = append(topPatterns, pc)
	}
	rows.Close()

	return &domain.SafetySummary{
		TotalDetections: total,
		ByType:          byType,
		BySeverity:      bySeverity,
		ByAction:        byAction,
		TopPatterns:     topPatterns,
		Period:          period,
	}, nil
}
