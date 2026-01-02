package handler

import (
	"net/http"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// CostHandler handles cost-related HTTP requests.
type CostHandler struct {
	logger zerolog.Logger
	repo   *repository.CostRepository
}

// NewCostHandler creates a new cost handler.
func NewCostHandler(logger zerolog.Logger, repo *repository.CostRepository) *CostHandler {
	return &CostHandler{logger: logger, repo: repo}
}

// Summary returns cost summary for the authenticated organization.
func (h *CostHandler) Summary(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	// Parse period from query params (default: month)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "month"
	}

	now := time.Now()
	var startDate time.Time
	switch period {
	case "day":
		startDate = now.AddDate(0, 0, -1)
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	filter := domain.CostFilter{
		OrgID:     orgID,
		StartDate: startDate,
		EndDate:   now,
	}

	// Query from database if repository is available
	if h.repo != nil {
		summary, err := h.repo.GetSummary(r.Context(), filter)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get cost summary")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get cost summary")
			return
		}

		WriteJSON(w, http.StatusOK, summary)
		return
	}

	// Fallback to sample summary
	summary := domain.CostSummary{
		TotalCost:     4231.89,
		TotalRequests: 1234567,
		AvgCostPerReq: 0.00343,
		Period:        period,
		StartDate:     startDate,
		EndDate:       now,
	}

	WriteJSON(w, http.StatusOK, summary)
}

// ByServer returns cost breakdown by MCP server.
func (h *CostHandler) ByServer(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	now := time.Now()
	filter := domain.CostFilter{
		OrgID:     orgID,
		StartDate: now.AddDate(0, -1, 0),
		EndDate:   now,
	}

	// Query from database if repository is available
	if h.repo != nil {
		data, err := h.repo.GetByServer(r.Context(), filter)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get cost by server")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get cost by server")
			return
		}

		var totalCost float64
		for _, s := range data {
			totalCost += s.TotalCost
		}

		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"total_cost": totalCost,
			"servers":    data,
		})
		return
	}

	// Fallback to sample data
	totalCost := 4231.89
	data := []domain.CostByServer{
		{MCPServer: "filesystem", TotalCost: 1250.00, TotalRequests: 450000, AvgCostPerReq: 0.00278, Percentage: 29.5},
		{MCPServer: "database", TotalCost: 890.00, TotalRequests: 320000, AvgCostPerReq: 0.00278, Percentage: 21.0},
		{MCPServer: "github", TotalCost: 780.00, TotalRequests: 280000, AvgCostPerReq: 0.00279, Percentage: 18.4},
		{MCPServer: "slack", TotalCost: 420.00, TotalRequests: 150000, AvgCostPerReq: 0.00280, Percentage: 9.9},
		{MCPServer: "memory", TotalCost: 340.00, TotalRequests: 120000, AvgCostPerReq: 0.00283, Percentage: 8.0},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"total_cost": totalCost,
		"servers":    data,
	})
}

// ByTeam returns cost breakdown by team.
func (h *CostHandler) ByTeam(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	now := time.Now()
	filter := domain.CostFilter{
		OrgID:     orgID,
		StartDate: now.AddDate(0, -1, 0),
		EndDate:   now,
	}

	// Query from database if repository is available
	if h.repo != nil {
		data, err := h.repo.GetByTeam(r.Context(), filter)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get cost by team")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get cost by team")
			return
		}

		var totalCost float64
		for _, t := range data {
			totalCost += t.TotalCost
		}

		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"total_cost": totalCost,
			"teams":      data,
		})
		return
	}

	// Fallback to sample data
	data := []domain.CostByTeam{
		{TeamID: uuid.New(), TeamName: "Engineering", TotalCost: 2500.00, TotalRequests: 750000, AvgCostPerReq: 0.00333, Percentage: 59.1},
		{TeamID: uuid.New(), TeamName: "Data Science", TotalCost: 1200.00, TotalRequests: 350000, AvgCostPerReq: 0.00343, Percentage: 28.4},
		{TeamID: uuid.New(), TeamName: "Product", TotalCost: 531.89, TotalRequests: 134567, AvgCostPerReq: 0.00395, Percentage: 12.5},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"total_cost": 4231.89,
		"teams":      data,
	})
}

// Daily returns daily cost data for charts.
func (h *CostHandler) Daily(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	now := time.Now()
	filter := domain.CostFilter{
		OrgID:     orgID,
		StartDate: now.AddDate(0, 0, -14),
		EndDate:   now,
	}

	// Query from database if repository is available
	if h.repo != nil {
		data, err := h.repo.GetByDay(r.Context(), filter)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get daily costs")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get daily costs")
			return
		}

		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"daily": data,
		})
		return
	}

	// Fallback to sample daily data
	data := make([]domain.CostByDay, 14)
	baseDate := time.Now().AddDate(0, 0, -13)
	baseCost := 280.0
	baseRequests := int64(80000)

	for i := 0; i < 14; i++ {
		date := baseDate.AddDate(0, 0, i)
		variation := float64(i%5) * 20
		data[i] = domain.CostByDay{
			Date:          date.Format("Jan 2"),
			TotalCost:     baseCost + variation + float64(i*10),
			TotalRequests: baseRequests + int64(i*5000) + int64(variation*100),
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"daily": data,
	})
}
