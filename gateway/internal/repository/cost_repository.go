// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
)

// CostRepository handles cost aggregation queries.
type CostRepository struct {
	db *sql.DB
}

// NewCostRepository creates a new cost repository.
func NewCostRepository(db *sql.DB) *CostRepository {
	return &CostRepository{db: db}
}

// GetSummary returns aggregated cost summary for a period.
func (r *CostRepository) GetSummary(ctx context.Context, filter domain.CostFilter) (*domain.CostSummary, error) {
	if r.db == nil {
		return &domain.CostSummary{}, nil
	}

	query := `
		SELECT
			COALESCE(SUM(cost), 0) as total_cost,
			COUNT(*) as total_requests
		FROM traces
		WHERE org_id = $1
			AND created_at >= $2
			AND created_at <= $3`

	args := []interface{}{filter.OrgID, filter.StartDate, filter.EndDate}
	argNum := 4

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d", argNum)
		args = append(args, *filter.TeamID)
		argNum++
	}

	if filter.MCPServer != "" {
		query += fmt.Sprintf(" AND mcp_server = $%d", argNum)
		args = append(args, filter.MCPServer)
	}

	var summary domain.CostSummary
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalCost,
		&summary.TotalRequests,
	)
	if err != nil {
		return nil, fmt.Errorf("query cost summary: %w", err)
	}

	if summary.TotalRequests > 0 {
		summary.AvgCostPerReq = summary.TotalCost / float64(summary.TotalRequests)
	}

	summary.StartDate = filter.StartDate
	summary.EndDate = filter.EndDate

	// Determine period based on date range
	days := filter.EndDate.Sub(filter.StartDate).Hours() / 24
	switch {
	case days <= 1:
		summary.Period = "day"
	case days <= 7:
		summary.Period = "week"
	default:
		summary.Period = "month"
	}

	return &summary, nil
}

// GetByServer returns cost breakdown by MCP server.
func (r *CostRepository) GetByServer(ctx context.Context, filter domain.CostFilter) ([]domain.CostByServer, error) {
	if r.db == nil {
		return nil, nil
	}

	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argNum))
	args = append(args, filter.StartDate)
	argNum++

	conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argNum))
	args = append(args, filter.EndDate)
	argNum++

	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argNum))
		args = append(args, *filter.TeamID)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		WITH totals AS (
			SELECT COALESCE(SUM(cost), 0) as grand_total
			FROM traces
			WHERE %s
		)
		SELECT
			mcp_server,
			COALESCE(SUM(cost), 0) as total_cost,
			COUNT(*) as total_requests,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(cost), 0) / COUNT(*) ELSE 0 END as avg_cost,
			CASE WHEN t.grand_total > 0 THEN COALESCE(SUM(cost), 0) / t.grand_total * 100 ELSE 0 END as percentage
		FROM traces, totals t
		WHERE %s
		GROUP BY mcp_server, t.grand_total
		ORDER BY total_cost DESC`,
		whereClause, whereClause)

	// Duplicate args for the second WHERE clause
	args = append(args, args...)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cost by server: %w", err)
	}
	defer rows.Close()

	var results []domain.CostByServer
	for rows.Next() {
		var c domain.CostByServer
		err := rows.Scan(
			&c.MCPServer,
			&c.TotalCost,
			&c.TotalRequests,
			&c.AvgCostPerReq,
			&c.Percentage,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cost by server: %w", err)
		}
		results = append(results, c)
	}

	return results, rows.Err()
}

// GetByTeam returns cost breakdown by team.
func (r *CostRepository) GetByTeam(ctx context.Context, filter domain.CostFilter) ([]domain.CostByTeam, error) {
	if r.db == nil {
		return nil, nil
	}

	query := `
		WITH totals AS (
			SELECT COALESCE(SUM(cost), 0) as grand_total
			FROM traces
			WHERE org_id = $1
				AND created_at >= $2
				AND created_at <= $3
		)
		SELECT
			t.team_id,
			COALESCE(tm.name, 'Unknown') as team_name,
			COALESCE(SUM(t.cost), 0) as total_cost,
			COUNT(*) as total_requests,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(t.cost), 0) / COUNT(*) ELSE 0 END as avg_cost,
			CASE WHEN totals.grand_total > 0 THEN COALESCE(SUM(t.cost), 0) / totals.grand_total * 100 ELSE 0 END as percentage
		FROM traces t
		CROSS JOIN totals
		LEFT JOIN teams tm ON t.team_id = tm.id
		WHERE t.org_id = $1
			AND t.created_at >= $2
			AND t.created_at <= $3
			AND t.team_id IS NOT NULL
		GROUP BY t.team_id, tm.name, totals.grand_total
		ORDER BY total_cost DESC`

	rows, err := r.db.QueryContext(ctx, query, filter.OrgID, filter.StartDate, filter.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query cost by team: %w", err)
	}
	defer rows.Close()

	var results []domain.CostByTeam
	for rows.Next() {
		var c domain.CostByTeam
		var teamID sql.NullString
		err := rows.Scan(
			&teamID,
			&c.TeamName,
			&c.TotalCost,
			&c.TotalRequests,
			&c.AvgCostPerReq,
			&c.Percentage,
		)
		if err != nil {
			return nil, fmt.Errorf("scan cost by team: %w", err)
		}
		if teamID.Valid {
			c.TeamID, _ = uuid.Parse(teamID.String)
		}
		results = append(results, c)
	}

	return results, rows.Err()
}

// GetByDay returns daily cost breakdown for charts.
func (r *CostRepository) GetByDay(ctx context.Context, filter domain.CostFilter) ([]domain.CostByDay, error) {
	if r.db == nil {
		return nil, nil
	}

	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argNum))
	args = append(args, filter.StartDate)
	argNum++

	conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argNum))
	args = append(args, filter.EndDate)
	argNum++

	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argNum))
		args = append(args, *filter.TeamID)
		argNum++
	}

	if filter.MCPServer != "" {
		conditions = append(conditions, fmt.Sprintf("mcp_server = $%d", argNum))
		args = append(args, filter.MCPServer)
	}

	whereClause := strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT
			TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as date,
			COALESCE(SUM(cost), 0) as total_cost,
			COUNT(*) as total_requests
		FROM traces
		WHERE %s
		GROUP BY DATE(created_at)
		ORDER BY date ASC`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query cost by day: %w", err)
	}
	defer rows.Close()

	var results []domain.CostByDay
	for rows.Next() {
		var c domain.CostByDay
		err := rows.Scan(&c.Date, &c.TotalCost, &c.TotalRequests)
		if err != nil {
			return nil, fmt.Errorf("scan cost by day: %w", err)
		}
		results = append(results, c)
	}

	return results, rows.Err()
}
