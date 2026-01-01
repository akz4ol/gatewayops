package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/approval"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ApprovalHandler handles tool approval HTTP requests.
type ApprovalHandler struct {
	logger  zerolog.Logger
	service *approval.Service
}

// NewApprovalHandler creates a new approval handler.
func NewApprovalHandler(logger zerolog.Logger, service *approval.Service) *ApprovalHandler {
	return &ApprovalHandler{
		logger:  logger,
		service: service,
	}
}

// ListClassifications returns all tool classifications.
func (h *ApprovalHandler) ListClassifications(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	classifications := h.service.ListClassifications(server)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"classifications": classifications,
		"total":           len(classifications),
	})
}

// GetClassification returns a specific classification.
func (h *ApprovalHandler) GetClassification(w http.ResponseWriter, r *http.Request) {
	server := chi.URLParam(r, "server")
	tool := chi.URLParam(r, "tool")

	classification := h.service.GetClassification(server, tool)
	if classification == nil {
		// Return default classification
		defaultLevel := domain.GetDefaultClassification(tool)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"server":           server,
			"tool":             tool,
			"classification":   defaultLevel,
			"requires_approval": defaultLevel != domain.ToolRiskSafe,
			"is_default":       true,
		})
		return
	}

	WriteJSON(w, http.StatusOK, classification)
}

// SetClassification sets or updates a tool classification.
func (h *ApprovalHandler) SetClassification(w http.ResponseWriter, r *http.Request) {
	var input domain.ToolClassificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.MCPServer == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "MCP server is required")
		return
	}
	if input.ToolName == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Tool name is required")
		return
	}
	if input.Classification == "" {
		input.Classification = domain.ToolRiskSensitive
	}

	// Demo org and user
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	classification := h.service.SetClassification(input, orgID, userID)
	WriteJSON(w, http.StatusOK, classification)
}

// DeleteClassification removes a tool classification.
func (h *ApprovalHandler) DeleteClassification(w http.ResponseWriter, r *http.Request) {
	server := chi.URLParam(r, "server")
	tool := chi.URLParam(r, "tool")

	if !h.service.DeleteClassification(server, tool) {
		WriteError(w, http.StatusNotFound, "not_found", "Classification not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// CheckAccess checks if access is allowed to a tool.
func (h *ApprovalHandler) CheckAccess(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	tool := r.URL.Query().Get("tool")

	if server == "" || tool == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Server and tool are required")
		return
	}

	// Demo user
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	allowed, reason := h.service.CheckAccess(userID, nil, server, tool)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"allowed": allowed,
		"reason":  reason,
		"server":  server,
		"tool":    tool,
	})
}

// ListApprovals returns tool approval requests.
func (h *ApprovalHandler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	filter := h.parseApprovalFilter(r)
	page := h.service.ListApprovals(filter)
	WriteJSON(w, http.StatusOK, page)
}

// GetApproval returns a specific approval request.
func (h *ApprovalHandler) GetApproval(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "approvalID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	approval := h.service.GetApproval(id)
	if approval == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
		return
	}

	WriteJSON(w, http.StatusOK, approval)
}

// RequestApproval creates a new approval request.
func (h *ApprovalHandler) RequestApproval(w http.ResponseWriter, r *http.Request) {
	var input domain.ToolApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.MCPServer == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "MCP server is required")
		return
	}
	if input.ToolName == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Tool name is required")
		return
	}

	// Demo org and user
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	approval := h.service.RequestApproval(input, orgID, userID)
	WriteJSON(w, http.StatusCreated, approval)
}

// ApproveRequest approves an approval request.
func (h *ApprovalHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "approvalID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	var review domain.ToolApprovalReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		// Allow empty body for simple approval
		review = domain.ToolApprovalReview{}
	}
	review.Status = domain.ApprovalStatusApproved

	// Demo reviewer
	reviewerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	approval := h.service.ReviewApproval(id, review, reviewerID)
	if approval == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
		return
	}

	WriteJSON(w, http.StatusOK, approval)
}

// DenyRequest denies an approval request.
func (h *ApprovalHandler) DenyRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "approvalID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	var review domain.ToolApprovalReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		review = domain.ToolApprovalReview{}
	}
	review.Status = domain.ApprovalStatusDenied

	// Demo reviewer
	reviewerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	approval := h.service.ReviewApproval(id, review, reviewerID)
	if approval == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
		return
	}

	WriteJSON(w, http.StatusOK, approval)
}

// ListPermissions returns all tool permissions.
func (h *ApprovalHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Query().Get("server")
	permissions := h.service.ListPermissions(server)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"permissions": permissions,
		"total":       len(permissions),
	})
}

// GrantPermission grants a permission to use a tool.
func (h *ApprovalHandler) GrantPermission(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID     *uuid.UUID `json:"user_id,omitempty"`
		TeamID     *uuid.UUID `json:"team_id,omitempty"`
		MCPServer  string     `json:"mcp_server"`
		ToolName   string     `json:"tool_name"`
		ExpiresIn  *int       `json:"expires_in,omitempty"` // seconds
		MaxUsesDay *int       `json:"max_uses_day,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.MCPServer == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "MCP server is required")
		return
	}
	if input.ToolName == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Tool name is required")
		return
	}
	if input.UserID == nil && input.TeamID == nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Either user_id or team_id is required")
		return
	}

	// Demo org and granter
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	granterID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	permission := h.service.GrantPermission(
		orgID,
		input.UserID,
		input.TeamID,
		input.MCPServer,
		input.ToolName,
		granterID,
		input.ExpiresIn,
		input.MaxUsesDay,
	)

	if permission == nil {
		WriteError(w, http.StatusBadRequest, "grant_failed", "Failed to grant permission")
		return
	}

	WriteJSON(w, http.StatusCreated, permission)
}

// RevokePermission revokes a tool permission.
func (h *ApprovalHandler) RevokePermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "permissionID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid permission ID")
		return
	}

	if !h.service.RevokePermission(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Permission not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// GetPendingCount returns the count of pending approvals.
func (h *ApprovalHandler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
	count := h.service.GetPendingCount()
	WriteJSON(w, http.StatusOK, map[string]int{"pending_count": count})
}

func (h *ApprovalHandler) parseApprovalFilter(r *http.Request) domain.ToolApprovalFilter {
	query := r.URL.Query()

	filter := domain.ToolApprovalFilter{
		OrgID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	}

	if server := query.Get("server"); server != "" {
		filter.MCPServer = server
	}
	if tool := query.Get("tool"); tool != "" {
		filter.ToolName = tool
	}
	if statusesStr := query.Get("statuses"); statusesStr != "" {
		statuses := strings.Split(statusesStr, ",")
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.ApprovalStatus(strings.TrimSpace(s)))
		}
	}
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
