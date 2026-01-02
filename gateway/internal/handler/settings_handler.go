package handler

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SettingsHandler handles organization settings HTTP requests.
type SettingsHandler struct {
	logger   zerolog.Logger
	settings map[uuid.UUID]*OrgSettings
	mu       sync.RWMutex
}

// OrgSettings represents organization-level settings.
type OrgSettings struct {
	ID               uuid.UUID        `json:"id"`
	OrgID            uuid.UUID        `json:"org_id"`
	OrgName          string           `json:"org_name"`
	BillingEmail     string           `json:"billing_email"`
	RateLimits       RateLimitConfig  `json:"rate_limits"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// RateLimitConfig holds default rate limit settings.
type RateLimitConfig struct {
	ProductionRPM int `json:"production_rpm"`
	SandboxRPM    int `json:"sandbox_rpm"`
}

// UpdateSettingsInput represents input for updating settings.
type UpdateSettingsInput struct {
	OrgName      *string          `json:"org_name,omitempty"`
	BillingEmail *string          `json:"billing_email,omitempty"`
	RateLimits   *RateLimitConfig `json:"rate_limits,omitempty"`
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(logger zerolog.Logger) *SettingsHandler {
	h := &SettingsHandler{
		logger:   logger,
		settings: make(map[uuid.UUID]*OrgSettings),
	}

	// Initialize demo org settings
	demoOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	h.settings[demoOrgID] = &OrgSettings{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000100"),
		OrgID:        demoOrgID,
		OrgName:      "Acme Corp",
		BillingEmail: "billing@acme.com",
		RateLimits: RateLimitConfig{
			ProductionRPM: 1000,
			SandboxRPM:    100,
		},
		UpdatedAt: time.Now(),
	}

	return h
}

// GetSettings returns the organization settings.
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	h.mu.RLock()
	settings, ok := h.settings[orgID]
	h.mu.RUnlock()

	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Settings not found")
		return
	}

	WriteJSON(w, http.StatusOK, settings)
}

// UpdateSettings updates the organization settings.
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var input UpdateSettingsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	// Demo org
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	h.mu.Lock()
	defer h.mu.Unlock()

	settings, ok := h.settings[orgID]
	if !ok {
		WriteError(w, http.StatusNotFound, "not_found", "Settings not found")
		return
	}

	// Apply updates
	if input.OrgName != nil {
		settings.OrgName = *input.OrgName
	}
	if input.BillingEmail != nil {
		settings.BillingEmail = *input.BillingEmail
	}
	if input.RateLimits != nil {
		if input.RateLimits.ProductionRPM > 0 {
			settings.RateLimits.ProductionRPM = input.RateLimits.ProductionRPM
		}
		if input.RateLimits.SandboxRPM > 0 {
			settings.RateLimits.SandboxRPM = input.RateLimits.SandboxRPM
		}
	}
	settings.UpdatedAt = time.Now()

	h.logger.Info().
		Str("org_id", orgID.String()).
		Str("org_name", settings.OrgName).
		Msg("Organization settings updated")

	WriteJSON(w, http.StatusOK, settings)
}
