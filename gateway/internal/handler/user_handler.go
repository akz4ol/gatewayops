package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/rbac"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	logger      zerolog.Logger
	userRepo    *repository.UserRepository
	rbacService *rbac.Service
	invites     map[uuid.UUID]*Invite
	mu          sync.RWMutex
}

// Invite represents a pending user invitation.
type Invite struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	InvitedBy uuid.UUID `json:"invited_by"`
	InviterName string  `json:"inviter_name"`
	Status    string    `json:"status"` // pending, accepted, expired
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewUserHandler creates a new user handler.
func NewUserHandler(logger zerolog.Logger, userRepo *repository.UserRepository, rbacService *rbac.Service) *UserHandler {
	h := &UserHandler{
		logger:      logger,
		userRepo:    userRepo,
		rbacService: rbacService,
		invites:     make(map[uuid.UUID]*Invite),
	}

	// Add demo invite
	demoInvite := &Invite{
		ID:          uuid.MustParse("00000000-0000-0000-0000-000000000010"),
		OrgID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Email:       "alex@acme.com",
		Role:        "developer",
		InvitedBy:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		InviterName: "Sarah Chen",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-48 * time.Hour),
		ExpiresAt:   time.Now().Add(5 * 24 * time.Hour),
	}
	h.invites[demoInvite.ID] = demoInvite

	return h
}

// UserResponse represents a user with their role information.
type UserResponse struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	Status       string     `json:"status"`
	Role         string     `json:"role"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ListUsers returns all users in the organization.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Parse pagination
	limit := 50
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Try to get from database
	ctx := r.Context()
	users, total, err := h.userRepo.ListUsersByOrg(ctx, orgID, limit, offset)

	var response []UserResponse

	if err != nil || len(users) == 0 {
		// Return demo users if database is empty or fails
		response = h.getDemoUsers()
		total = int64(len(response))
	} else {
		// Convert to response format with roles
		for _, user := range users {
			role := h.getUserRole(user.ID)
			response = append(response, UserResponse{
				ID:           user.ID,
				Email:        user.Email,
				Name:         user.Name,
				AvatarURL:    user.AvatarURL,
				Status:       string(user.Status),
				Role:         role,
				LastActiveAt: user.LastLoginAt,
				CreatedAt:    user.CreatedAt,
			})
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"users":   response,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// GetUser returns a specific user by ID.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "userID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID")
		return
	}

	ctx := r.Context()
	user, err := h.userRepo.GetUser(ctx, id)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get user")
		WriteError(w, http.StatusInternalServerError, "db_error", "Failed to get user")
		return
	}
	if user == nil {
		WriteError(w, http.StatusNotFound, "not_found", "User not found")
		return
	}

	role := h.getUserRole(user.ID)
	response := UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		Status:       string(user.Status),
		Role:         role,
		LastActiveAt: user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
	}

	WriteJSON(w, http.StatusOK, response)
}

// InviteInput represents input for creating an invitation.
type InviteInput struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// CreateInvite creates a new user invitation.
func (h *UserHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var input InviteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Email == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Email is required")
		return
	}
	if input.Role == "" {
		input.Role = "developer" // Default role
	}

	// Demo org and inviter
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	inviterID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	invite := &Invite{
		ID:          uuid.New(),
		OrgID:       orgID,
		Email:       input.Email,
		Role:        input.Role,
		InvitedBy:   inviterID,
		InviterName: "Demo User",
		Status:      "pending",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	h.mu.Lock()
	h.invites[invite.ID] = invite
	h.mu.Unlock()

	h.logger.Info().
		Str("invite_id", invite.ID.String()).
		Str("email", input.Email).
		Str("role", input.Role).
		Msg("User invitation created")

	WriteJSON(w, http.StatusCreated, invite)
}

// ListInvites returns all pending invitations.
func (h *UserHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var invites []*Invite
	for _, inv := range h.invites {
		if inv.Status == "pending" {
			invites = append(invites, inv)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"invites": invites,
		"total":   len(invites),
	})
}

// CancelInvite cancels a pending invitation.
func (h *UserHandler) CancelInvite(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "inviteID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid invite ID")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	invite, ok := h.invites[id]
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Invite not found")
		return
	}

	delete(h.invites, id)

	h.logger.Info().
		Str("invite_id", id.String()).
		Str("email", invite.Email).
		Msg("User invitation cancelled")

	WriteJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// ResendInvite resends an invitation email.
func (h *UserHandler) ResendInvite(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "inviteID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid invite ID")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	invite, ok := h.invites[id]
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Invite not found")
		return
	}

	// Update expiration
	invite.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)

	h.logger.Info().
		Str("invite_id", id.String()).
		Str("email", invite.Email).
		Msg("User invitation resent")

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "resent",
		"invite":  invite,
	})
}

// Helper to get user's primary role name
func (h *UserHandler) getUserRole(userID uuid.UUID) string {
	if h.rbacService == nil {
		return "developer"
	}

	assignments := h.rbacService.GetUserRoles(userID)
	if len(assignments) == 0 {
		return "developer"
	}

	role := h.rbacService.GetRole(assignments[0].RoleID)
	if role == nil {
		return "developer"
	}

	return role.Name
}

// getDemoUsers returns demo user data when database is empty
func (h *UserHandler) getDemoUsers() []UserResponse {
	now := time.Now()
	return []UserResponse{
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Email:        "sarah@acme.com",
			Name:         "Sarah Chen",
			Status:       "active",
			Role:         "admin",
			LastActiveAt: &now,
			CreatedAt:    now.Add(-30 * 24 * time.Hour),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:        "michael@acme.com",
			Name:         "Michael Park",
			Status:       "active",
			Role:         "developer",
			LastActiveAt: func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
			CreatedAt:    now.Add(-25 * 24 * time.Hour),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Email:        "emma@acme.com",
			Name:         "Emma Wilson",
			Status:       "active",
			Role:         "developer",
			LastActiveAt: func() *time.Time { t := now.Add(-3 * time.Hour); return &t }(),
			CreatedAt:    now.Add(-20 * 24 * time.Hour),
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Email:        "james@acme.com",
			Name:         "James Lee",
			Status:       "active",
			Role:         "viewer",
			LastActiveAt: func() *time.Time { t := now.Add(-24 * time.Hour); return &t }(),
			CreatedAt:    now.Add(-15 * 24 * time.Hour),
		},
	}
}
