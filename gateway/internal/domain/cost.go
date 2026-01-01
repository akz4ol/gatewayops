package domain

import (
	"time"

	"github.com/google/uuid"
)

// CostSummary represents aggregated cost data.
type CostSummary struct {
	TotalCost     float64   `json:"total_cost"`
	TotalRequests int64     `json:"total_requests"`
	AvgCostPerReq float64   `json:"avg_cost_per_request"`
	Period        string    `json:"period"` // day, week, month
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
}

// CostByServer represents cost breakdown by MCP server.
type CostByServer struct {
	MCPServer     string  `json:"mcp_server"`
	TotalCost     float64 `json:"total_cost"`
	TotalRequests int64   `json:"total_requests"`
	AvgCostPerReq float64 `json:"avg_cost_per_request"`
	Percentage    float64 `json:"percentage"`
}

// CostByTeam represents cost breakdown by team.
type CostByTeam struct {
	TeamID        uuid.UUID `json:"team_id"`
	TeamName      string    `json:"team_name"`
	TotalCost     float64   `json:"total_cost"`
	TotalRequests int64     `json:"total_requests"`
	AvgCostPerReq float64   `json:"avg_cost_per_request"`
	Percentage    float64   `json:"percentage"`
}

// CostByDay represents daily cost data for charts.
type CostByDay struct {
	Date          string  `json:"date"`
	TotalCost     float64 `json:"total_cost"`
	TotalRequests int64   `json:"total_requests"`
}

// CostFilter represents filters for cost queries.
type CostFilter struct {
	OrgID     uuid.UUID  `json:"org_id"`
	TeamID    *uuid.UUID `json:"team_id,omitempty"`
	MCPServer string     `json:"mcp_server,omitempty"`
	StartDate time.Time  `json:"start_date"`
	EndDate   time.Time  `json:"end_date"`
}
