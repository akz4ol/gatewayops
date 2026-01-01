package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	auditService *service.AuditService
}

// NewAuditHandler creates a new audit handler.
func NewAuditHandler(auditService *service.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// List retrieves audit logs with filtering.
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	// Parse query parameters
	filter := domain.AuditLogFilter{
		OrgID: authInfo.OrgID,
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		id, err := uuid.Parse(teamID)
		if err == nil {
			filter.TeamID = &id
		}
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err == nil {
			filter.UserID = &id
		}
	}

	if apiKeyID := r.URL.Query().Get("api_key_id"); apiKeyID != "" {
		id, err := uuid.Parse(apiKeyID)
		if err == nil {
			filter.APIKeyID = &id
		}
	}

	if actions := r.URL.Query()["action"]; len(actions) > 0 {
		for _, a := range actions {
			filter.Actions = append(filter.Actions, domain.AuditAction(a))
		}
	}

	if outcomes := r.URL.Query()["outcome"]; len(outcomes) > 0 {
		for _, o := range outcomes {
			filter.Outcomes = append(filter.Outcomes, domain.AuditOutcome(o))
		}
	}

	if resource := r.URL.Query().Get("resource"); resource != "" {
		filter.Resource = resource
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err == nil {
			filter.EndTime = &t
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err == nil {
			filter.Limit = l
		}
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err == nil {
			filter.Offset = o
		}
	}

	// Get audit logs
	page, err := h.auditService.List(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list audit logs")
		return
	}

	WriteSuccess(w, page)
}

// Get retrieves a single audit log by ID.
func (h *AuditHandler) Get(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid audit log ID")
		return
	}

	log, err := h.auditService.Get(r.Context(), authInfo.OrgID, id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get audit log")
		return
	}

	if log == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Audit log not found")
		return
	}

	WriteSuccess(w, log)
}

// Search searches audit logs with advanced filtering.
func (h *AuditHandler) Search(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var filter domain.AuditLogFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Override org_id with authenticated org
	filter.OrgID = authInfo.OrgID

	page, err := h.auditService.Search(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to search audit logs")
		return
	}

	WriteSuccess(w, page)
}
