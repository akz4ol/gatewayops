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

// PolicyHandler handles safety policy endpoints.
type PolicyHandler struct {
	injectionService *service.InjectionService
}

// NewPolicyHandler creates a new policy handler.
func NewPolicyHandler(injectionService *service.InjectionService) *PolicyHandler {
	return &PolicyHandler{
		injectionService: injectionService,
	}
}

// ListPolicies lists all safety policies.
func (h *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	enabledOnly := r.URL.Query().Get("enabled_only") == "true"

	policies, err := h.injectionService.ListPolicies(r.Context(), authInfo.OrgID, enabledOnly)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list policies")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"policies": policies,
	})
}

// GetPolicy retrieves a safety policy by ID.
func (h *PolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	policy, err := h.injectionService.GetPolicy(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get policy")
		return
	}

	if policy == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
		return
	}

	WriteSuccess(w, policy)
}

// CreatePolicy creates a new safety policy.
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
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

	var input domain.SafetyPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "missing_name", "Policy name is required")
		return
	}

	policy, err := h.injectionService.CreatePolicy(r.Context(), authInfo.OrgID, *userID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create policy")
		return
	}

	WriteJSON(w, http.StatusCreated, policy)
}

// UpdatePolicy updates a safety policy.
func (h *PolicyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	var input domain.SafetyPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	policy, err := h.injectionService.UpdatePolicy(r.Context(), id, *userID, input)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update policy")
		return
	}

	WriteSuccess(w, policy)
}

// DeletePolicy deletes a safety policy.
func (h *PolicyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	if err := h.injectionService.DeletePolicy(r.Context(), id, *userID); err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete policy")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}

// ListDetections lists injection detections.
func (h *PolicyHandler) ListDetections(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	filter := domain.DetectionFilter{
		OrgID: authInfo.OrgID,
	}

	if types := r.URL.Query()["type"]; len(types) > 0 {
		for _, t := range types {
			filter.Types = append(filter.Types, domain.DetectionType(t))
		}
	}

	if severities := r.URL.Query()["severity"]; len(severities) > 0 {
		for _, s := range severities {
			filter.Severities = append(filter.Severities, domain.DetectionSeverity(s))
		}
	}

	if actions := r.URL.Query()["action"]; len(actions) > 0 {
		for _, a := range actions {
			filter.Actions = append(filter.Actions, domain.SafetyMode(a))
		}
	}

	if mcpServer := r.URL.Query().Get("mcp_server"); mcpServer != "" {
		filter.MCPServer = mcpServer
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

	page, err := h.injectionService.ListDetections(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list detections")
		return
	}

	WriteSuccess(w, page)
}

// GetDetection retrieves a detection by ID.
func (h *PolicyHandler) GetDetection(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid detection ID")
		return
	}

	detection, err := h.injectionService.GetDetection(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get detection")
		return
	}

	if detection == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Detection not found")
		return
	}

	WriteSuccess(w, detection)
}

// GetSummary retrieves a summary of safety detections.
func (h *PolicyHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "day"
	}

	summary, err := h.injectionService.GetSummary(r.Context(), authInfo.OrgID, period)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get summary")
		return
	}

	WriteSuccess(w, summary)
}

// TestInput tests input against safety policies.
func (h *PolicyHandler) TestInput(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var request domain.SafetyTestRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if request.Input == "" {
		WriteError(w, http.StatusBadRequest, "missing_input", "Input is required")
		return
	}

	response, err := h.injectionService.TestInput(r.Context(), authInfo.OrgID, request.Input, request.PolicyID)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to test input")
		return
	}

	WriteSuccess(w, response)
}
