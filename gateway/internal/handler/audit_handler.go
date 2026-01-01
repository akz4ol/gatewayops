package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/audit"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AuditHandler handles audit log HTTP requests.
type AuditHandler struct {
	logger      zerolog.Logger
	auditLogger *audit.Logger
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(logger zerolog.Logger, auditLogger *audit.Logger) *AuditHandler {
	return &AuditHandler{
		logger:      logger,
		auditLogger: auditLogger,
	}
}

// List returns paginated audit logs.
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)
	page := h.auditLogger.GetLogs(filter)
	WriteJSON(w, http.StatusOK, page)
}

// Get returns a single audit log by ID.
func (h *AuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "logID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid audit log ID")
		return
	}

	log := h.auditLogger.GetLog(id)
	if log == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Audit log not found")
		return
	}

	WriteJSON(w, http.StatusOK, log)
}

// Search performs a text search across audit logs.
func (h *AuditHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Search query 'q' is required")
		return
	}

	filter := h.parseFilter(r)
	page := h.auditLogger.Search(query, filter)
	WriteJSON(w, http.StatusOK, page)
}

// Export exports audit logs in the specified format.
func (h *AuditHandler) Export(w http.ResponseWriter, r *http.Request) {
	filter := h.parseFilter(r)

	// Get format from query param
	formatStr := r.URL.Query().Get("format")
	format := domain.AuditExportJSON
	if formatStr == "csv" {
		format = domain.AuditExportCSV
	}

	// Remove limit for export
	filter.Limit = 10000

	data, err := h.auditLogger.Export(filter, format)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "export_error", "Failed to export audit logs")
		return
	}

	// Set appropriate headers
	switch format {
	case domain.AuditExportCSV:
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=audit-logs.csv")
	default:
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=audit-logs.json")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Stats returns audit log statistics.
func (h *AuditHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := h.auditLogger.GetStats()
	WriteJSON(w, http.StatusOK, stats)
}

// parseFilter parses filter parameters from the request.
func (h *AuditHandler) parseFilter(r *http.Request) domain.AuditLogFilter {
	query := r.URL.Query()

	filter := domain.AuditLogFilter{
		OrgID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Demo org
	}

	// Parse actions
	if actionsStr := query.Get("actions"); actionsStr != "" {
		actions := strings.Split(actionsStr, ",")
		for _, a := range actions {
			filter.Actions = append(filter.Actions, domain.AuditAction(strings.TrimSpace(a)))
		}
	}

	// Parse outcomes
	if outcomesStr := query.Get("outcomes"); outcomesStr != "" {
		outcomes := strings.Split(outcomesStr, ",")
		for _, o := range outcomes {
			filter.Outcomes = append(filter.Outcomes, domain.AuditOutcome(strings.TrimSpace(o)))
		}
	}

	// Parse resource
	if resource := query.Get("resource"); resource != "" {
		filter.Resource = resource
	}

	// Parse user ID
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	// Parse API key ID
	if apiKeyIDStr := query.Get("api_key_id"); apiKeyIDStr != "" {
		if apiKeyID, err := uuid.Parse(apiKeyIDStr); err == nil {
			filter.APIKeyID = &apiKeyID
		}
	}

	// Parse time range
	if startStr := query.Get("start_time"); startStr != "" {
		if start, err := time.Parse(time.RFC3339, startStr); err == nil {
			filter.StartTime = &start
		}
	}
	if endStr := query.Get("end_time"); endStr != "" {
		if end, err := time.Parse(time.RFC3339, endStr); err == nil {
			filter.EndTime = &end
		}
	}

	// Parse pagination
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}
