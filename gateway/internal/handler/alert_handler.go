package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/alerting"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// AlertHandler handles alert-related HTTP requests.
type AlertHandler struct {
	logger  zerolog.Logger
	service *alerting.Service
}

// NewAlertHandler creates a new alert handler.
func NewAlertHandler(logger zerolog.Logger, service *alerting.Service) *AlertHandler {
	return &AlertHandler{
		logger:  logger,
		service: service,
	}
}

// ListRules returns all alert rules.
func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules := h.service.ListRules()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"rules": rules,
		"total": len(rules),
	})
}

// GetRule returns a single rule by ID.
func (h *AlertHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "ruleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	rule := h.service.GetRule(id)
	if rule == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Rule not found")
		return
	}

	WriteJSON(w, http.StatusOK, rule)
}

// CreateRule creates a new alert rule.
func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var input domain.AlertRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if input.Metric == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Metric is required")
		return
	}
	if input.Condition == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Condition is required")
		return
	}
	if input.WindowMinutes <= 0 {
		input.WindowMinutes = 5 // default
	}
	if input.Severity == "" {
		input.Severity = domain.AlertSeverityWarning
	}

	// Demo org and user
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	rule := h.service.CreateRule(input, orgID, userID)
	WriteJSON(w, http.StatusCreated, rule)
}

// UpdateRule updates an existing rule.
func (h *AlertHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "ruleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	var input domain.AlertRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	rule := h.service.UpdateRule(id, input)
	if rule == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Rule not found")
		return
	}

	WriteJSON(w, http.StatusOK, rule)
}

// DeleteRule deletes a rule.
func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "ruleID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid rule ID")
		return
	}

	if !h.service.DeleteRule(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Rule not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ListChannels returns all alert channels.
func (h *AlertHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels := h.service.ListChannels()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channels": channels,
		"total":    len(channels),
	})
}

// GetChannel returns a single channel by ID.
func (h *AlertHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "channelID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	channel := h.service.GetChannel(id)
	if channel == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Channel not found")
		return
	}

	WriteJSON(w, http.StatusOK, channel)
}

// CreateChannel creates a new alert channel.
func (h *AlertHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var input domain.AlertChannelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if input.Type == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Type is required")
		return
	}
	if input.Config == nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Config is required")
		return
	}

	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	channel := h.service.CreateChannel(input, orgID)
	WriteJSON(w, http.StatusCreated, channel)
}

// UpdateChannel updates an existing channel.
func (h *AlertHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "channelID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	var input domain.AlertChannelInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	channel := h.service.UpdateChannel(id, input)
	if channel == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Channel not found")
		return
	}

	WriteJSON(w, http.StatusOK, channel)
}

// DeleteChannel deletes a channel.
func (h *AlertHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "channelID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	if !h.service.DeleteChannel(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Channel not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// TestChannel sends a test notification to a channel.
func (h *AlertHandler) TestChannel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "channelID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid channel ID")
		return
	}

	if err := h.service.TestChannel(id); err != nil {
		WriteError(w, http.StatusBadRequest, "test_failed", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Test notification sent",
	})
}

// ListAlerts returns alerts matching the filter.
func (h *AlertHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	filter := h.parseAlertFilter(r)
	page := h.service.GetAlerts(filter)
	WriteJSON(w, http.StatusOK, page)
}

// GetActiveAlerts returns all currently firing alerts.
func (h *AlertHandler) GetActiveAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := h.service.GetActiveAlerts()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// AcknowledgeAlert acknowledges an alert.
func (h *AlertHandler) AcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "alertID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid alert ID")
		return
	}

	// Demo user
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	alert := h.service.AcknowledgeAlert(id, userID)
	if alert == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Alert not found")
		return
	}

	WriteJSON(w, http.StatusOK, alert)
}

// ResolveAlert resolves an alert.
func (h *AlertHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "alertID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid alert ID")
		return
	}

	alert := h.service.ResolveAlert(id)
	if alert == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Alert not found")
		return
	}

	WriteJSON(w, http.StatusOK, alert)
}

// TriggerTestAlert triggers a test alert for demo purposes.
func (h *AlertHandler) TriggerTestAlert(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Metric string  `json:"metric"`
		Value  float64 `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Metric == "" {
		input.Metric = "error_rate"
	}
	if input.Value == 0 {
		input.Value = 10.0 // Default value above threshold
	}

	alert := h.service.TriggerTestAlert(input.Metric, input.Value)
	if alert == nil {
		WriteError(w, http.StatusInternalServerError, "trigger_failed", "Failed to trigger test alert")
		return
	}

	WriteJSON(w, http.StatusCreated, alert)
}

// parseAlertFilter parses filter parameters from the request.
func (h *AlertHandler) parseAlertFilter(r *http.Request) domain.AlertFilter {
	query := r.URL.Query()

	filter := domain.AlertFilter{
		OrgID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	}

	// Parse rule ID
	if ruleIDStr := query.Get("rule_id"); ruleIDStr != "" {
		if ruleID, err := uuid.Parse(ruleIDStr); err == nil {
			filter.RuleID = &ruleID
		}
	}

	// Parse statuses
	if statusesStr := query.Get("statuses"); statusesStr != "" {
		statuses := strings.Split(statusesStr, ",")
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.AlertStatus(strings.TrimSpace(s)))
		}
	}

	// Parse severities
	if severitiesStr := query.Get("severities"); severitiesStr != "" {
		severities := strings.Split(severitiesStr, ",")
		for _, s := range severities {
			filter.Severities = append(filter.Severities, domain.AlertSeverity(strings.TrimSpace(s)))
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
