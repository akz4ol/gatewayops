// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
)

// TraceRepository handles trace persistence.
type TraceRepository struct {
	db *sql.DB
}

// NewTraceRepository creates a new trace repository.
func NewTraceRepository(db *sql.DB) *TraceRepository {
	return &TraceRepository{db: db}
}

// Create inserts a new trace.
func (r *TraceRepository) Create(ctx context.Context, trace *domain.Trace) error {
	if r.db == nil {
		return nil // Silently skip if no DB
	}

	metadata, err := json.Marshal(trace.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		INSERT INTO traces (
			id, trace_id, span_id, parent_id, org_id, team_id, api_key_id,
			mcp_server, operation, tool_name, status, status_code,
			duration_ms, request_size, response_size, cost, error_msg,
			metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	_, err = r.db.ExecContext(ctx, query,
		trace.ID, trace.TraceID, trace.SpanID, trace.ParentID,
		trace.OrgID, trace.TeamID, trace.APIKeyID,
		trace.MCPServer, trace.Operation, trace.ToolName,
		trace.Status, trace.StatusCode,
		trace.DurationMs, trace.RequestSize, trace.ResponseSize,
		trace.Cost, trace.ErrorMsg,
		metadata, trace.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert trace: %w", err)
	}

	return nil
}

// CreateSpan inserts a new trace span.
func (r *TraceRepository) CreateSpan(ctx context.Context, span *domain.TraceSpan) error {
	if r.db == nil {
		return nil
	}

	attrs, err := json.Marshal(span.Attributes)
	if err != nil {
		attrs = []byte("{}")
	}

	query := `
		INSERT INTO trace_spans (
			id, trace_id, span_id, parent_id, name, kind, status,
			start_time, end_time, duration_ms, attributes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)`

	_, err = r.db.ExecContext(ctx, query,
		span.ID, span.TraceID, span.SpanID, span.ParentID,
		span.Name, span.Kind, span.Status,
		span.StartTime, span.EndTime, span.DurationMs, attrs,
	)
	if err != nil {
		return fmt.Errorf("insert trace span: %w", err)
	}

	return nil
}

// Get retrieves a trace by ID.
func (r *TraceRepository) Get(ctx context.Context, orgID, id uuid.UUID) (*domain.Trace, error) {
	if r.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, trace_id, span_id, parent_id, org_id, team_id, api_key_id,
			   mcp_server, operation, tool_name, status, status_code,
			   duration_ms, request_size, response_size, cost, error_msg,
			   metadata, created_at
		FROM traces
		WHERE id = $1 AND org_id = $2`

	var trace domain.Trace
	var teamID sql.NullString
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&trace.ID, &trace.TraceID, &trace.SpanID, &trace.ParentID,
		&trace.OrgID, &teamID, &trace.APIKeyID,
		&trace.MCPServer, &trace.Operation, &trace.ToolName,
		&trace.Status, &trace.StatusCode,
		&trace.DurationMs, &trace.RequestSize, &trace.ResponseSize,
		&trace.Cost, &trace.ErrorMsg,
		&metadata, &trace.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query trace: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		trace.TeamID = &tid
	}

	if len(metadata) > 0 {
		json.Unmarshal(metadata, &trace.Metadata)
	}

	return &trace, nil
}

// GetByTraceID retrieves a trace by trace ID string.
func (r *TraceRepository) GetByTraceID(ctx context.Context, orgID uuid.UUID, traceID string) (*domain.TraceDetail, error) {
	if r.db == nil {
		return nil, nil
	}

	// Get the main trace
	query := `
		SELECT id, trace_id, span_id, parent_id, org_id, team_id, api_key_id,
			   mcp_server, operation, tool_name, status, status_code,
			   duration_ms, request_size, response_size, cost, error_msg,
			   metadata, created_at
		FROM traces
		WHERE trace_id = $1 AND org_id = $2
		LIMIT 1`

	var trace domain.Trace
	var teamID sql.NullString
	var metadata []byte

	err := r.db.QueryRowContext(ctx, query, traceID, orgID).Scan(
		&trace.ID, &trace.TraceID, &trace.SpanID, &trace.ParentID,
		&trace.OrgID, &teamID, &trace.APIKeyID,
		&trace.MCPServer, &trace.Operation, &trace.ToolName,
		&trace.Status, &trace.StatusCode,
		&trace.DurationMs, &trace.RequestSize, &trace.ResponseSize,
		&trace.Cost, &trace.ErrorMsg,
		&metadata, &trace.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query trace: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		trace.TeamID = &tid
	}
	if len(metadata) > 0 {
		json.Unmarshal(metadata, &trace.Metadata)
	}

	// Get spans for this trace
	spans, err := r.GetSpans(ctx, traceID)
	if err != nil {
		return nil, err
	}

	return &domain.TraceDetail{
		Trace: trace,
		Spans: spans,
	}, nil
}

// GetSpans retrieves all spans for a trace.
func (r *TraceRepository) GetSpans(ctx context.Context, traceID string) ([]domain.TraceSpan, error) {
	if r.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, trace_id, span_id, parent_id, name, kind, status,
			   start_time, end_time, duration_ms, attributes
		FROM trace_spans
		WHERE trace_id = $1
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, traceID)
	if err != nil {
		return nil, fmt.Errorf("query trace spans: %w", err)
	}
	defer rows.Close()

	var spans []domain.TraceSpan
	for rows.Next() {
		var span domain.TraceSpan
		var attrs []byte

		err := rows.Scan(
			&span.ID, &span.TraceID, &span.SpanID, &span.ParentID,
			&span.Name, &span.Kind, &span.Status,
			&span.StartTime, &span.EndTime, &span.DurationMs, &attrs,
		)
		if err != nil {
			return nil, fmt.Errorf("scan trace span: %w", err)
		}

		if len(attrs) > 0 {
			json.Unmarshal(attrs, &span.Attributes)
		}

		spans = append(spans, span)
	}

	return spans, rows.Err()
}

// List retrieves traces with filtering and pagination.
func (r *TraceRepository) List(ctx context.Context, filter domain.TraceFilter) ([]domain.Trace, int64, error) {
	if r.db == nil {
		return nil, 0, nil
	}

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

	if filter.Operation != "" {
		conditions = append(conditions, fmt.Sprintf("operation = $%d", argNum))
		args = append(args, filter.Operation)
		argNum++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, filter.Status)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traces WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count traces: %w", err)
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
		SELECT id, trace_id, span_id, parent_id, org_id, team_id, api_key_id,
			   mcp_server, operation, tool_name, status, status_code,
			   duration_ms, request_size, response_size, cost, error_msg,
			   metadata, created_at
		FROM traces
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query traces: %w", err)
	}
	defer rows.Close()

	var traces []domain.Trace
	for rows.Next() {
		var trace domain.Trace
		var teamID sql.NullString
		var metadata []byte

		err := rows.Scan(
			&trace.ID, &trace.TraceID, &trace.SpanID, &trace.ParentID,
			&trace.OrgID, &teamID, &trace.APIKeyID,
			&trace.MCPServer, &trace.Operation, &trace.ToolName,
			&trace.Status, &trace.StatusCode,
			&trace.DurationMs, &trace.RequestSize, &trace.ResponseSize,
			&trace.Cost, &trace.ErrorMsg,
			&metadata, &trace.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan trace: %w", err)
		}

		if teamID.Valid {
			tid, _ := uuid.Parse(teamID.String)
			trace.TeamID = &tid
		}
		if len(metadata) > 0 {
			json.Unmarshal(metadata, &trace.Metadata)
		}

		traces = append(traces, trace)
	}

	return traces, total, rows.Err()
}

// Stats returns aggregated trace statistics.
func (r *TraceRepository) Stats(ctx context.Context, filter domain.TraceFilter) (*domain.TraceStats, error) {
	if r.db == nil {
		return &domain.TraceStats{}, nil
	}

	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

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

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE status = 'success') as success_count,
			COUNT(*) FILTER (WHERE status = 'error') as error_count,
			COALESCE(AVG(duration_ms), 0) as avg_duration_ms,
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration_ms), 0) as p50_duration_ms,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 0) as p95_duration_ms,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms), 0) as p99_duration_ms,
			COALESCE(SUM(cost), 0) as total_cost
		FROM traces
		WHERE %s`, whereClause)

	var stats domain.TraceStats
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalRequests,
		&stats.SuccessCount,
		&stats.ErrorCount,
		&stats.AvgDurationMs,
		&stats.P50DurationMs,
		&stats.P95DurationMs,
		&stats.P99DurationMs,
		&stats.TotalCost,
	)
	if err != nil {
		return nil, fmt.Errorf("query trace stats: %w", err)
	}

	if stats.TotalRequests > 0 {
		stats.ErrorRate = float64(stats.ErrorCount) / float64(stats.TotalRequests) * 100
	}

	return &stats, nil
}
