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

// ApprovalHandler handles tool approval endpoints.
type ApprovalHandler struct {
	approvalService *service.ApprovalService
}

// NewApprovalHandler creates a new approval handler.
func NewApprovalHandler(approvalService *service.ApprovalService) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: approvalService,
	}
}

// List retrieves tool approvals with filtering.
func (h *ApprovalHandler) List(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	filter := domain.ToolApprovalFilter{
		OrgID: authInfo.OrgID,
	}

	if teamID := r.URL.Query().Get("team_id"); teamID != "" {
		id, err := uuid.Parse(teamID)
		if err == nil {
			filter.TeamID = &id
		}
	}

	if mcpServer := r.URL.Query().Get("mcp_server"); mcpServer != "" {
		filter.MCPServer = mcpServer
	}

	if toolName := r.URL.Query().Get("tool_name"); toolName != "" {
		filter.ToolName = toolName
	}

	if statuses := r.URL.Query()["status"]; len(statuses) > 0 {
		for _, s := range statuses {
			filter.Statuses = append(filter.Statuses, domain.ApprovalStatus(s))
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

	page, err := h.approvalService.ListApprovals(r.Context(), filter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list approvals")
		return
	}

	WriteSuccess(w, page)
}

// Get retrieves a tool approval by ID.
func (h *ApprovalHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	approval, err := h.approvalService.GetApproval(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get approval")
		return
	}

	if approval == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
		return
	}

	WriteSuccess(w, approval)
}

// Request creates a new tool approval request.
func (h *ApprovalHandler) Request(w http.ResponseWriter, r *http.Request) {
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

	var request domain.ToolApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if request.MCPServer == "" || request.ToolName == "" {
		WriteError(w, http.StatusBadRequest, "missing_fields", "mcp_server and tool_name are required")
		return
	}

	approval, err := h.approvalService.RequestApproval(r.Context(), authInfo.OrgID, *userID, request)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create approval request")
		return
	}

	WriteJSON(w, http.StatusCreated, approval)
}

// Approve approves a tool approval request.
func (h *ApprovalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	var review domain.ToolApprovalReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		// Allow empty body for simple approvals
		review = domain.ToolApprovalReview{
			Status: domain.ApprovalStatusApproved,
		}
	}
	review.Status = domain.ApprovalStatusApproved

	approval, err := h.approvalService.Approve(r.Context(), id, *userID, review)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to approve request")
		return
	}

	WriteSuccess(w, approval)
}

// Deny denies a tool approval request.
func (h *ApprovalHandler) Deny(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "User authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid approval ID")
		return
	}

	var review domain.ToolApprovalReview
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		review = domain.ToolApprovalReview{}
	}
	review.Status = domain.ApprovalStatusDenied

	approval, err := h.approvalService.Deny(r.Context(), id, *userID, review)
	if err != nil {
		if _, ok := err.(service.ErrNotFound); ok {
			WriteError(w, http.StatusNotFound, "not_found", "Approval not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to deny request")
		return
	}

	WriteSuccess(w, approval)
}

// ListPending lists pending approvals.
func (h *ApprovalHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	page, err := h.approvalService.ListPendingApprovals(r.Context(), authInfo.OrgID, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list pending approvals")
		return
	}

	WriteSuccess(w, page)
}

// ListClassifications lists tool classifications.
func (h *ApprovalHandler) ListClassifications(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	mcpServer := r.URL.Query().Get("mcp_server")

	classifications, err := h.approvalService.ListClassifications(r.Context(), authInfo.OrgID, mcpServer)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list classifications")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"classifications": classifications,
	})
}

// SetClassification sets or updates a tool classification.
func (h *ApprovalHandler) SetClassification(w http.ResponseWriter, r *http.Request) {
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

	var input domain.ToolClassificationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	if input.MCPServer == "" || input.ToolName == "" {
		WriteError(w, http.StatusBadRequest, "missing_fields", "mcp_server and tool_name are required")
		return
	}

	classification, err := h.approvalService.SetClassification(r.Context(), authInfo.OrgID, *userID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to set classification")
		return
	}

	WriteSuccess(w, classification)
}

// GetClassification retrieves a tool classification.
func (h *ApprovalHandler) GetClassification(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	mcpServer := chi.URLParam(r, "server")
	toolName := chi.URLParam(r, "tool")

	if mcpServer == "" || toolName == "" {
		WriteError(w, http.StatusBadRequest, "missing_params", "server and tool are required")
		return
	}

	classification, err := h.approvalService.GetClassification(r.Context(), authInfo.OrgID, mcpServer, toolName)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get classification")
		return
	}

	if classification == nil {
		// Return default classification
		defaultLevel := domain.GetDefaultClassification(toolName)
		WriteSuccess(w, map[string]interface{}{
			"mcp_server":        mcpServer,
			"tool_name":         toolName,
			"classification":    defaultLevel,
			"requires_approval": defaultLevel != domain.ToolRiskSafe,
			"is_default":        true,
		})
		return
	}

	WriteSuccess(w, classification)
}

// DeleteClassification removes a tool classification.
func (h *ApprovalHandler) DeleteClassification(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	mcpServer := chi.URLParam(r, "server")
	toolName := chi.URLParam(r, "tool")

	if err := h.approvalService.DeleteClassification(r.Context(), authInfo.OrgID, mcpServer, toolName); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete classification")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}
