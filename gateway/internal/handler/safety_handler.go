package handler

import (
	"encoding/json"
	"net/http"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/safety"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SafetyHandler handles safety-related HTTP requests.
type SafetyHandler struct {
	logger   zerolog.Logger
	detector *safety.Detector
}

// NewSafetyHandler creates a new safety handler.
func NewSafetyHandler(logger zerolog.Logger, detector *safety.Detector) *SafetyHandler {
	return &SafetyHandler{
		logger:   logger,
		detector: detector,
	}
}

// ListPolicies returns all safety policies.
func (h *SafetyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	policies := h.detector.GetPolicies()
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"policies": policies,
	})
}

// GetPolicy returns a specific policy by ID.
func (h *SafetyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "policyID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	policy := h.detector.GetPolicy(id)
	if policy == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
		return
	}

	WriteJSON(w, http.StatusOK, policy)
}

// CreatePolicy creates a new safety policy.
func (h *SafetyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var input domain.SafetyPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Validate input
	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Policy name is required")
		return
	}

	// Set defaults
	if input.Sensitivity == "" {
		input.Sensitivity = domain.SafetySensitivityModerate
	}
	if input.Mode == "" {
		input.Mode = domain.SafetyModeBlock
	}

	// Demo mode: use fixed org and user IDs
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	policy := h.detector.CreatePolicy(input, orgID, userID)

	h.logger.Info().
		Str("policy_id", policy.ID.String()).
		Str("name", policy.Name).
		Msg("Safety policy created")

	WriteJSON(w, http.StatusCreated, policy)
}

// UpdatePolicy updates an existing safety policy.
func (h *SafetyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "policyID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	var input domain.SafetyPolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	policy := h.detector.UpdatePolicy(id, input)
	if policy == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
		return
	}

	h.logger.Info().
		Str("policy_id", policy.ID.String()).
		Str("name", policy.Name).
		Msg("Safety policy updated")

	WriteJSON(w, http.StatusOK, policy)
}

// DeletePolicy deletes a safety policy.
func (h *SafetyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "policyID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid policy ID")
		return
	}

	// Check if it's the default policy
	if id == uuid.MustParse("00000000-0000-0000-0000-000000000001") {
		WriteError(w, http.StatusForbidden, "forbidden", "Cannot delete default policy")
		return
	}

	if !h.detector.DeletePolicy(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Policy not found")
		return
	}

	h.logger.Info().
		Str("policy_id", id.String()).
		Msg("Safety policy deleted")

	w.WriteHeader(http.StatusNoContent)
}

// TestInput tests input against safety detection.
func (h *SafetyHandler) TestInput(w http.ResponseWriter, r *http.Request) {
	var req domain.SafetyTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if req.Input == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Input is required")
		return
	}

	opts := safety.DetectOptions{
		Input:    req.Input,
		PolicyID: req.PolicyID,
		OrgID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Demo org
	}

	result := h.detector.Detect(req.Input, opts)

	WriteJSON(w, http.StatusOK, domain.SafetyTestResponse{
		Result:   result,
		PolicyID: req.PolicyID,
	})
}

// ListDetections returns recent injection detections.
func (h *SafetyHandler) ListDetections(w http.ResponseWriter, r *http.Request) {
	filter := domain.DetectionFilter{
		OrgID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), // Demo org
	}

	// Parse query params
	query := r.URL.Query()
	if mcpServer := query.Get("mcp_server"); mcpServer != "" {
		filter.MCPServer = mcpServer
	}
	if limit := query.Get("limit"); limit != "" {
		var l int
		if _, err := parseIntParam(limit, &l); err == nil {
			filter.Limit = l
		}
	}
	if offset := query.Get("offset"); offset != "" {
		var o int
		if _, err := parseIntParam(offset, &o); err == nil {
			filter.Offset = o
		}
	}

	page := h.detector.GetDetections(filter)
	WriteJSON(w, http.StatusOK, page)
}

// GetSummary returns a summary of safety detections.
func (h *SafetyHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary := h.detector.GetSummary()
	WriteJSON(w, http.StatusOK, summary)
}

// Helper to parse int query params
func parseIntParam(s string, target *int) (bool, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	*target = n
	return true, nil
}
