package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/rs/zerolog"
)

// TraceHandler handles trace-related HTTP requests.
type TraceHandler struct {
	logger zerolog.Logger
}

// NewTraceHandler creates a new trace handler.
func NewTraceHandler(logger zerolog.Logger) *TraceHandler {
	return &TraceHandler{logger: logger}
}

// List returns a list of traces for the authenticated organization.
func (h *TraceHandler) List(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	// Use demo org ID if not authenticated
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	mcpServer := r.URL.Query().Get("server")
	status := r.URL.Query().Get("status")

	// Generate sample traces (in production, query from database)
	traces := generateSampleTraces(orgID, limit, mcpServer, status)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"traces": traces,
		"total":  100, // Would come from COUNT query
		"limit":  limit,
		"offset": offset,
	})
}

// Get returns a single trace by ID.
func (h *TraceHandler) Get(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	traceID := chi.URLParam(r, "traceID")
	if traceID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Trace ID is required")
		return
	}

	// Generate sample trace detail (in production, query from database)
	detail := generateSampleTraceDetail(traceID, orgID)

	WriteJSON(w, http.StatusOK, detail)
}

// Stats returns aggregated trace statistics.
func (h *TraceHandler) Stats(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	// Generate sample stats (in production, query from database)
	stats := domain.TraceStats{
		TotalRequests: 1234567,
		SuccessCount:  1220000,
		ErrorCount:    14567,
		AvgDurationMs: 127.5,
		P50DurationMs: 85.0,
		P95DurationMs: 350.0,
		P99DurationMs: 890.0,
		TotalCost:     4231.89,
		ErrorRate:     0.0118,
	}

	WriteJSON(w, http.StatusOK, stats)
}

// generateSampleTraces creates sample trace data.
func generateSampleTraces(orgID uuid.UUID, limit int, filterServer, filterStatus string) []domain.Trace {
	servers := []string{"filesystem", "database", "github", "slack", "memory"}
	operations := []string{"tools/call", "tools/list", "resources/read", "resources/list"}
	tools := []string{"read_file", "write_file", "query", "create_issue", "send_message", "list_files"}
	statuses := []string{"success", "success", "success", "success", "error"} // 80% success

	traces := make([]domain.Trace, 0, limit)
	baseTime := time.Now()

	for i := 0; i < limit; i++ {
		server := servers[i%len(servers)]
		status := statuses[i%len(statuses)]

		// Apply filters
		if filterServer != "" && server != filterServer {
			continue
		}
		if filterStatus != "" && status != filterStatus {
			continue
		}

		trace := domain.Trace{
			ID:           uuid.New(),
			TraceID:      "tr_" + uuid.New().String()[:8],
			SpanID:       "sp_" + uuid.New().String()[:8],
			OrgID:        orgID,
			MCPServer:    server,
			Operation:    operations[i%len(operations)],
			ToolName:     tools[i%len(tools)],
			Status:       status,
			StatusCode:   200,
			DurationMs:   int64(20 + (i*13)%500),
			RequestSize:  256 + i*10,
			ResponseSize: 512 + i*20,
			Cost:         0.0001 * float64(1+i%10),
			CreatedAt:    baseTime.Add(-time.Duration(i) * time.Minute),
		}

		if status == "error" {
			trace.StatusCode = 500
			trace.ErrorMsg = "Connection timeout"
		}

		traces = append(traces, trace)
	}

	return traces
}

// generateSampleTraceDetail creates sample trace detail with spans.
func generateSampleTraceDetail(traceID string, orgID uuid.UUID) domain.TraceDetail {
	baseTime := time.Now().Add(-5 * time.Minute)

	trace := domain.Trace{
		ID:           uuid.New(),
		TraceID:      traceID,
		SpanID:       "sp_main",
		OrgID:        orgID,
		MCPServer:    "filesystem",
		Operation:    "tools/call",
		ToolName:     "read_file",
		Status:       "success",
		StatusCode:   200,
		DurationMs:   145,
		RequestSize:  256,
		ResponseSize: 4096,
		Cost:         0.0003,
		CreatedAt:    baseTime,
	}

	spans := []domain.TraceSpan{
		{
			ID:         uuid.New(),
			TraceID:    traceID,
			SpanID:     "sp_auth",
			Name:       "authenticate",
			Kind:       "internal",
			Status:     "success",
			StartTime:  baseTime,
			EndTime:    baseTime.Add(5 * time.Millisecond),
			DurationMs: 5,
		},
		{
			ID:         uuid.New(),
			TraceID:    traceID,
			SpanID:     "sp_validate",
			ParentID:   "sp_auth",
			Name:       "validate_request",
			Kind:       "internal",
			Status:     "success",
			StartTime:  baseTime.Add(5 * time.Millisecond),
			EndTime:    baseTime.Add(10 * time.Millisecond),
			DurationMs: 5,
		},
		{
			ID:         uuid.New(),
			TraceID:    traceID,
			SpanID:     "sp_proxy",
			ParentID:   "sp_validate",
			Name:       "proxy_to_mcp",
			Kind:       "client",
			Status:     "success",
			StartTime:  baseTime.Add(10 * time.Millisecond),
			EndTime:    baseTime.Add(140 * time.Millisecond),
			DurationMs: 130,
		},
		{
			ID:         uuid.New(),
			TraceID:    traceID,
			SpanID:     "sp_response",
			ParentID:   "sp_proxy",
			Name:       "process_response",
			Kind:       "internal",
			Status:     "success",
			StartTime:  baseTime.Add(140 * time.Millisecond),
			EndTime:    baseTime.Add(145 * time.Millisecond),
			DurationMs: 5,
		},
	}

	return domain.TraceDetail{
		Trace: trace,
		Spans: spans,
	}
}
