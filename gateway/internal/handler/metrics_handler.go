package handler

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// MetricsHandler handles metrics and dashboard data requests.
type MetricsHandler struct {
	logger zerolog.Logger
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(logger zerolog.Logger) *MetricsHandler {
	return &MetricsHandler{logger: logger}
}

// Overview returns dashboard overview metrics.
func (h *MetricsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	// Generate sample overview data (in production, query from database)
	overview := map[string]interface{}{
		"total_requests": map[string]interface{}{
			"value":       1234567,
			"change":      12.5,
			"period":      "30d",
			"formatted":   "1.2M",
		},
		"total_cost": map[string]interface{}{
			"value":     4231.89,
			"change":    -3.2,
			"period":    "30d",
			"formatted": "$4,231.89",
		},
		"avg_latency": map[string]interface{}{
			"value":     127,
			"change":    -8.1,
			"period":    "30d",
			"formatted": "127ms",
			"percentile": "p95",
		},
		"error_rate": map[string]interface{}{
			"value":     0.12,
			"change":    0.02,
			"period":    "24h",
			"formatted": "0.12%",
		},
	}

	WriteJSON(w, http.StatusOK, overview)
}

// RequestsChart returns data for the requests volume chart.
func (h *MetricsHandler) RequestsChart(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	// Generate sample chart data (in production, query from database)
	data := make([]map[string]interface{}, 14)
	baseDate := time.Now().AddDate(0, 0, -13)
	baseRequests := 4000

	for i := 0; i < 14; i++ {
		date := baseDate.AddDate(0, 0, i)
		requests := baseRequests + (i * 500) + ((i % 3) * 1000)
		errors := requests / 100 * (1 + i%3)

		data[i] = map[string]interface{}{
			"date":     date.Format("Jan 2"),
			"requests": requests,
			"errors":   errors,
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"data": data,
	})
}

// TopServers returns the top MCP servers by usage.
func (h *MetricsHandler) TopServers(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	// Generate sample data (in production, query from database)
	servers := []map[string]interface{}{
		{"name": "filesystem", "requests": 45000, "cost": 1250.00},
		{"name": "database", "requests": 32000, "cost": 890.00},
		{"name": "github", "requests": 28000, "cost": 780.00},
		{"name": "slack", "requests": 15000, "cost": 420.00},
		{"name": "memory", "requests": 12000, "cost": 340.00},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"servers": servers,
	})
}

// RecentTraces returns recent traces for the dashboard.
func (h *MetricsHandler) RecentTraces(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	now := time.Now()

	// Generate sample recent traces (in production, query from database)
	traces := []map[string]interface{}{
		{
			"id":        "tr_1a2b3c4d",
			"server":    "filesystem",
			"operation": "tools/call",
			"tool":      "read_file",
			"status":    "success",
			"duration":  45,
			"time":      now.Add(-2 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":        "tr_2b3c4d5e",
			"server":    "database",
			"operation": "tools/call",
			"tool":      "query",
			"status":    "success",
			"duration":  128,
			"time":      now.Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":        "tr_3c4d5e6f",
			"server":    "github",
			"operation": "tools/call",
			"tool":      "create_issue",
			"status":    "error",
			"duration":  2340,
			"time":      now.Add(-8 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":        "tr_4d5e6f7g",
			"server":    "slack",
			"operation": "tools/call",
			"tool":      "send_message",
			"status":    "success",
			"duration":  89,
			"time":      now.Add(-12 * time.Minute).Format(time.RFC3339),
		},
		{
			"id":        "tr_5e6f7g8h",
			"server":    "filesystem",
			"operation": "resources/read",
			"tool":      "file://config.json",
			"status":    "success",
			"duration":  23,
			"time":      now.Add(-15 * time.Minute).Format(time.RFC3339),
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"traces": traces,
	})
}
