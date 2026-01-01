package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/rs/zerolog"
)

// APIKeyHandler handles API key management HTTP requests.
type APIKeyHandler struct {
	logger zerolog.Logger
}

// NewAPIKeyHandler creates a new API key handler.
func NewAPIKeyHandler(logger zerolog.Logger) *APIKeyHandler {
	return &APIKeyHandler{logger: logger}
}

// List returns all API keys for the authenticated organization.
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	// Generate sample API keys (in production, query from database)
	now := time.Now()
	lastUsed := now.Add(-2 * time.Hour)
	keys := []domain.APIKey{
		{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        "Production API Key",
			KeyPrefix:   "gwo_prod",
			Environment: "production",
			Permissions: []string{"mcp:*", "traces:read"},
			RateLimit:   1000,
			LastUsedAt:  &lastUsed,
			CreatedAt:   now.AddDate(0, -3, 0),
			CreatedBy:   uuid.New(),
			Revoked:     false,
		},
		{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        "Staging API Key",
			KeyPrefix:   "gwo_stag",
			Environment: "staging",
			Permissions: []string{"mcp:*", "traces:read"},
			RateLimit:   500,
			LastUsedAt:  &now,
			CreatedAt:   now.AddDate(0, -2, 0),
			CreatedBy:   uuid.New(),
			Revoked:     false,
		},
		{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        "Development Key",
			KeyPrefix:   "gwo_dev_",
			Environment: "development",
			Permissions: []string{"*"},
			RateLimit:   100,
			LastUsedAt:  nil,
			CreatedAt:   now.AddDate(0, -1, 0),
			CreatedBy:   uuid.New(),
			Revoked:     false,
		},
		{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        "Old Key (Revoked)",
			KeyPrefix:   "gwo_old_",
			Environment: "production",
			Permissions: []string{"mcp:read"},
			RateLimit:   100,
			CreatedAt:   now.AddDate(0, -6, 0),
			CreatedBy:   uuid.New(),
			Revoked:     true,
			RevokedAt:   &now,
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"api_keys": keys,
		"total":    len(keys),
	})
}

// Create creates a new API key.
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	if authInfo != nil {
		orgID = authInfo.OrgID
		userID = authInfo.UserID
	}

	var req domain.APIKeyCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Name is required")
		return
	}

	// Generate a random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate key")
		return
	}
	rawKey := "gwo_" + req.Environment[:4] + "_" + hex.EncodeToString(keyBytes)

	// Set defaults
	if req.Environment == "" {
		req.Environment = "development"
	}
	if req.RateLimit <= 0 {
		req.RateLimit = 100
	}
	if len(req.Permissions) == 0 {
		req.Permissions = []string{"mcp:*"}
	}

	now := time.Now()
	key := domain.APIKeyCreated{
		APIKey: domain.APIKey{
			ID:          uuid.New(),
			OrgID:       orgID,
			TeamID:      req.TeamID,
			Name:        req.Name,
			KeyPrefix:   rawKey[:12],
			Environment: req.Environment,
			Permissions: req.Permissions,
			RateLimit:   req.RateLimit,
			ExpiresAt:   req.ExpiresAt,
			CreatedAt:   now,
			CreatedBy:   userID,
			Revoked:     false,
		},
		RawKey: rawKey,
	}

	// In production, save to database here

	h.logger.Info().
		Str("key_id", key.ID.String()).
		Str("name", key.Name).
		Str("environment", key.Environment).
		Msg("API key created")

	WriteJSON(w, http.StatusCreated, key)
}

// Get returns a single API key by ID.
func (h *APIKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	keyID := chi.URLParam(r, "keyID")
	if keyID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Key ID is required")
		return
	}

	keyUUID, err := uuid.Parse(keyID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid key ID format")
		return
	}

	// Generate sample key (in production, query from database)
	now := time.Now()
	lastUsed := now.Add(-2 * time.Hour)
	key := domain.APIKey{
		ID:          keyUUID,
		OrgID:       orgID,
		Name:        "Production API Key",
		KeyPrefix:   "gwo_prod",
		Environment: "production",
		Permissions: []string{"mcp:*", "traces:read"},
		RateLimit:   1000,
		LastUsedAt:  &lastUsed,
		CreatedAt:   now.AddDate(0, -3, 0),
		CreatedBy:   uuid.New(),
		Revoked:     false,
	}

	// Include usage stats
	usage := domain.APIKeyUsage{
		KeyID:         keyUUID,
		TotalRequests: 45230,
		TotalCost:     123.45,
		LastUsedAt:    lastUsed,
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"api_key": key,
		"usage":   usage,
	})
}

// Delete revokes an API key.
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Auth not required for demo

	keyID := chi.URLParam(r, "keyID")
	if keyID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Key ID is required")
		return
	}

	// In production, mark key as revoked in database

	h.logger.Info().
		Str("key_id", keyID).
		Msg("API key revoked")

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "API key revoked successfully",
	})
}

// Rotate generates a new key while revoking the old one.
func (h *APIKeyHandler) Rotate(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	if authInfo != nil {
		orgID = authInfo.OrgID
		userID = authInfo.UserID
	}

	keyID := chi.URLParam(r, "keyID")
	if keyID == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Key ID is required")
		return
	}

	// Generate a new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate key")
		return
	}
	rawKey := "gwo_prod_" + hex.EncodeToString(keyBytes)

	now := time.Now()
	key := domain.APIKeyCreated{
		APIKey: domain.APIKey{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        "Rotated Key",
			KeyPrefix:   rawKey[:12],
			Environment: "production",
			Permissions: []string{"mcp:*"},
			RateLimit:   1000,
			CreatedAt:   now,
			CreatedBy:   userID,
			Revoked:     false,
		},
		RawKey: rawKey,
	}

	// In production: revoke old key, save new key to database

	h.logger.Info().
		Str("old_key_id", keyID).
		Str("new_key_id", key.ID.String()).
		Msg("API key rotated")

	WriteJSON(w, http.StatusOK, key)
}
