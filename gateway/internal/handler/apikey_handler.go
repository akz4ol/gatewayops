package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/rs/zerolog"
)

// APIKeyHandler handles API key management HTTP requests.
type APIKeyHandler struct {
	logger zerolog.Logger
	repo   *repository.APIKeyRepository
}

// NewAPIKeyHandler creates a new API key handler.
func NewAPIKeyHandler(logger zerolog.Logger, repo *repository.APIKeyRepository) *APIKeyHandler {
	return &APIKeyHandler{logger: logger, repo: repo}
}

// List returns all API keys for the authenticated organization.
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	if authInfo != nil {
		orgID = authInfo.OrgID
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	environment := r.URL.Query().Get("environment")
	includeRevoked := r.URL.Query().Get("include_revoked") == "true"

	filter := domain.APIKeyFilter{
		OrgID:          orgID,
		Environment:    environment,
		IncludeRevoked: includeRevoked,
		Limit:          limit,
		Offset:         offset,
	}

	// Query from database if repository is available
	if h.repo != nil {
		keys, total, err := h.repo.List(r.Context(), filter)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to list API keys")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list API keys")
			return
		}

		// Return database results if any exist
		if len(keys) > 0 || total > 0 {
			WriteJSON(w, http.StatusOK, map[string]interface{}{
				"api_keys": keys,
				"total":    total,
				"limit":    limit,
				"offset":   offset,
			})
			return
		}
	}

	// Fallback to sample API keys
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
		"limit":    limit,
		"offset":   offset,
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

	// Generate a random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate key")
		return
	}
	envPrefix := req.Environment
	if len(envPrefix) > 4 {
		envPrefix = envPrefix[:4]
	}
	rawKey := "gwo_" + envPrefix + "_" + hex.EncodeToString(keyBytes)

	now := time.Now()
	key := domain.APIKeyCreated{
		APIKey: domain.APIKey{
			ID:          uuid.New(),
			OrgID:       orgID,
			TeamID:      req.TeamID,
			Name:        req.Name,
			KeyPrefix:   rawKey[:16],
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

	// Save to database if repository is available
	if h.repo != nil {
		if err := h.repo.Create(r.Context(), &key.APIKey, rawKey); err != nil {
			h.logger.Error().Err(err).Msg("Failed to create API key")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create API key")
			return
		}
	}

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

	// Query from database if repository is available
	if h.repo != nil {
		key, err := h.repo.Get(r.Context(), orgID, keyUUID)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get API key")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get API key")
			return
		}
		if key == nil {
			WriteError(w, http.StatusNotFound, "not_found", "API key not found")
			return
		}

		// Get usage stats
		usage, err := h.repo.GetUsage(r.Context(), keyUUID)
		if err != nil {
			h.logger.Warn().Err(err).Msg("Failed to get API key usage")
			usage = &domain.APIKeyUsage{KeyID: keyUUID}
		}

		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"api_key": key,
			"usage":   usage,
		})
		return
	}

	// Fallback to sample key
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

	// Revoke in database if repository is available
	if h.repo != nil {
		if err := h.repo.Revoke(r.Context(), orgID, keyUUID); err != nil {
			h.logger.Error().Err(err).Msg("Failed to revoke API key")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke API key")
			return
		}
	}

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

	keyUUID, err := uuid.Parse(keyID)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid key ID format")
		return
	}

	// Get the old key to preserve settings
	var oldKey *domain.APIKey
	environment := "production"
	permissions := []string{"mcp:*"}
	rateLimit := 1000
	name := "Rotated Key"

	if h.repo != nil {
		oldKey, err = h.repo.Get(r.Context(), orgID, keyUUID)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to get old API key")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get API key")
			return
		}
		if oldKey != nil {
			environment = oldKey.Environment
			permissions = oldKey.Permissions
			rateLimit = oldKey.RateLimit
			name = oldKey.Name + " (rotated)"
		}
	}

	// Generate a new key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate key")
		return
	}
	envPrefix := environment
	if len(envPrefix) > 4 {
		envPrefix = envPrefix[:4]
	}
	rawKey := "gwo_" + envPrefix + "_" + hex.EncodeToString(keyBytes)

	now := time.Now()
	key := domain.APIKeyCreated{
		APIKey: domain.APIKey{
			ID:          uuid.New(),
			OrgID:       orgID,
			Name:        name,
			KeyPrefix:   rawKey[:16],
			Environment: environment,
			Permissions: permissions,
			RateLimit:   rateLimit,
			CreatedAt:   now,
			CreatedBy:   userID,
			Revoked:     false,
		},
		RawKey: rawKey,
	}

	// Revoke old key and create new one in database
	if h.repo != nil {
		// Revoke old key
		if err := h.repo.Revoke(r.Context(), orgID, keyUUID); err != nil {
			h.logger.Warn().Err(err).Msg("Failed to revoke old API key during rotation")
		}

		// Create new key
		if err := h.repo.Create(r.Context(), &key.APIKey, rawKey); err != nil {
			h.logger.Error().Err(err).Msg("Failed to create rotated API key")
			WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create API key")
			return
		}
	}

	h.logger.Info().
		Str("old_key_id", keyID).
		Str("new_key_id", key.ID.String()).
		Msg("API key rotated")

	WriteJSON(w, http.StatusOK, key)
}
