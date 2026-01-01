package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// AlertHandler handles alert endpoints.
type AlertHandler struct {
	alertService *service.AlertService
}

// NewAlertHandler creates a new alert handler.
func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// ListRules lists all alert rules.
func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	enabledOnly := r.URL.Query().Get("enabled_only") == "true"

	rules, err := h.alertService.ListRules(r.Context(), authInfo.OrgID, enabledOnly)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list rules")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"rules": rules,
	})
}

// GetRule retrieves an alert rule by ID.
func (h *AlertHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	rule, err := h.alertService.GetRule(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get rule")
		return
	}

	if rule == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Rule not found")
		return
	}

	WriteSuccess(w, rule)
}

// CreateRule creates a new alert rule.
func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	var input domain.AlertRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.Name == "" || input.Metric == "" {
		WriteError(w, http.StatusBadRequest, "missing_fields", "Name and metric are required")
		return
	}

	rule, err := h.alertService.CreateRule(r.Context(), authInfo.OrgID, *userID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create rule")
		return
	}

	WriteJSON(w, http.StatusCreated, rule)
}

// UpdateRule updates an alert rule.
func (h *AlertHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	var input domain.AlertRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	rule, err := h.alertService.UpdateRule(r.Context(), id, input)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Rule not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update rule")
		return
	}

	WriteSuccess(w, rule)
}

// DeleteRule deletes an alert rule.
func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	if err := h.alertService.DeleteRule(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete rule")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}

// ListChannels lists all alert channels.
func (h *AlertHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	channels, err := h.alertService.ListChannels(r.Context(), authInfo.OrgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list channels")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"channels": channels,
	})
}

// GetChannel retrieves an alert channel by ID.
func (h *AlertHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	channel, err := h.alertService.GetChannel(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get channel")
		return
	}

	if channel == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Channel not found")
		return
	}

	WriteSuccess(w, channel)
}

// CreateChannel creates a new alert channel.
func (h *AlertHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var input domain.AlertChannelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.Name == "" || input.Type == "" {
		WriteError(w, http.StatusBadRequest, "missing_fields", "Name and type are required")
		return
	}

	channel, err := h.alertService.CreateChannel(r.Context(), authInfo.OrgID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create channel")
		return
	}

	WriteJSON(w, http.StatusCreated, channel)
}

// UpdateChannel updates an alert channel.
func (h *AlertHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	var input domain.AlertChannelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	channel, err := h.alertService.UpdateChannel(r.Context(), id, input)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Channel not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update channel")
		return
	}

	WriteSuccess(w, channel)
}

// DeleteChannel deletes an alert channel.
func (h *AlertHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	if err := h.alertService.DeleteChannel(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete channel")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}

// ListAlerts lists alerts with filtering.
func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	filter := domain.AlertFilter{
		OrgID: authInfo.OrgID,
	}

	if ruleID := r.URL.Query().Get("rule_id"); ruleID != "" {
		id, err := uuid.Parse(ruleID)
		if err == nil {
			filter.RuleID = &id
		}
	}

	if statuses := r.URL.Query()["status"]; len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.AlertStatus(s))
		}
	}

	if severities := r.URL.Query()["severity"]; len(severities) > 0 {
		for _, s := range severities {
			filter.Severities = append(filter.Severities, domain.AlertSeverity(s))
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

	page, err := h.alertService.ListAlerts(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list alerts")
		return
	}

	WriteSuccess(w, page)
}

// GetAlert retrieves an alert by ID.
func (h *AlertHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid alert ID")
		return
	}

	alert, err := h.alertService.GetAlert(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get alert")
		return
	}

	if alert == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Alert not found")
		return
	}

	WriteSuccess(w, alert)
}

// ResolveAlert resolves an alert.
func (h *AlertHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid alert ID")
		return
	}

	alert, err := h.alertService.ResolveAlert(r.Context(), id)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Alert not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to resolve alert")
		return
	}

	WriteSuccess(w, alert)
}

// AcknowledgeAlert acknowledges an alert.
func (h *AlertHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid alert ID")
		return
	}

	alert, err := h.alertService.AcknowledgeAlert(r.Context(), id, *userID)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Alert not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to acknowledge alert")
		return
	}

	WriteSuccess(w, alert)
}

// ListFiring lists all firing alerts.
func (h *AlertHandler) ListFiring(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	page, err := h.alertService.ListFiringAlerts(r.Context(), authInfo.OrgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list firing alerts")
		return
	}

	WriteSuccess(w, page)
}

// CountActive counts active (firing) alerts.
func (h *AlertHandler) CountActive(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	count, err := h.alertService.CountActiveAlerts(r.Context(), authInfo.OrgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to count alerts")
		return
	}

	WriteSuccess(w, map[string]int64{
		"count": count,
	})
}
