package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/akz4ol/gatewayops/gateway/internal/service"
)

// SSOHandler handles SSO/OIDC endpoints.
type SSOHandler struct {
	authService *service.AuthService
	states      map[string]stateInfo // In production, use Redis
}

type stateInfo struct {
	ProviderID  uuid.UUID
	RedirectURL string
	ExpiresAt   time.Time
}

// NewSSOHandler creates a new SSO handler.
func NewSSOHandler(authService *service.AuthService) *SSOHandler {
	return &SSOHandler{
		authService: authService,
		states:      make(map[string]stateInfo),
	}
}

// CreateProvider creates a new SSO provider.
func (h *SSOHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var input domain.SSOProviderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	// Validate required fields
	if input.Name == "" || input.IssuerURL == "" || input.ClientID == "" || input.ClientSecret == "" {
		WriteError(w, http.StatusBadRequest, "missing_fields", "Name, issuer_url, client_id, and client_secret are required")
		return
	}

	provider, err := h.authService.CreateSSOProvider(r.Context(), authInfo.OrgID, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to create SSO provider")
		return
	}

	WriteSuccess(w, provider)
}

// ListProviders lists all SSO providers for an organization.
func (h *SSOHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	authInfo := middleware.GetAuthInfo(r.Context())
	if authInfo == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	providers, err := h.authService.ListSSOProviders(r.Context(), authInfo.OrgID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to list SSO providers")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"providers": providers,
	})
}

// GetProvider retrieves an SSO provider by ID.
func (h *SSOHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	provider, err := h.authService.GetSSOProvider(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to get SSO provider")
		return
	}

	if provider == nil {
		WriteError(w, http.StatusNotFound, "not_found", "SSO provider not found")
		return
	}

	WriteSuccess(w, provider)
}

// UpdateProvider updates an SSO provider.
func (h *SSOHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	var input domain.SSOProviderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	provider, err := h.authService.UpdateSSOProvider(r.Context(), id, input)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to update SSO provider")
		return
	}

	WriteSuccess(w, provider)
}

// DeleteProvider deletes an SSO provider.
func (h *SSOHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	if err := h.authService.DeleteSSOProvider(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to delete SSO provider")
		return
	}

	WriteSuccess(w, map[string]string{"status": "deleted"})
}

// Authorize initiates the OAuth flow.
func (h *SSOHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	providerID, err := uuid.Parse(chi.URLParam(r, "provider"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_provider", "Invalid provider ID")
		return
	}

	// Get redirect URL from query param
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	authURL, state, err := h.authService.GetAuthorizationURL(r.Context(), providerID, redirectURL)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Store state for CSRF protection
	h.states[state] = stateInfo{
		ProviderID:  providerID,
		RedirectURL: redirectURL,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	// Redirect to authorization URL
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth callback.
func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
	providerID, err := uuid.Parse(chi.URLParam(r, "provider"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_provider", "Invalid provider ID")
		return
	}

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		errorCode := r.URL.Query().Get("error")
		errorDesc := r.URL.Query().Get("error_description")
		WriteError(w, http.StatusBadRequest, errorCode, errorDesc)
		return
	}

	// Validate state
	stateData, ok := h.states[state]
	if !ok || time.Now().After(stateData.ExpiresAt) {
		WriteError(w, http.StatusBadRequest, "invalid_state", "Invalid or expired state")
		return
	}
	delete(h.states, state)

	if stateData.ProviderID != providerID {
		WriteError(w, http.StatusBadRequest, "provider_mismatch", "Provider ID mismatch")
		return
	}

	// Get client info
	ipAddress := middleware.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Exchange code for token
	tokenPair, user, err := h.authService.HandleCallback(r.Context(), providerID, code, state, ipAddress, userAgent)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "callback_failed", err.Error())
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "gwo_session",
		Value:    tokenPair.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  tokenPair.ExpiresAt,
	})

	// Redirect to original URL or return JSON
	if stateData.RedirectURL != "" && stateData.RedirectURL != "/" {
		http.Redirect(w, r, stateData.RedirectURL, http.StatusTemporaryRedirect)
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"user":  user,
		"token": tokenPair,
	})
}

// Logout invalidates the current session.
func (h *SSOHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r.Context())
	if session == nil {
		WriteSuccess(w, map[string]string{"status": "logged_out"})
		return
	}

	ipAddress := middleware.GetClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	if err := h.authService.Logout(r.Context(), session.ID, ipAddress, userAgent); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to logout")
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "gwo_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})

	WriteSuccess(w, map[string]string{"status": "logged_out"})
}

// GetCurrentUser returns the current authenticated user.
func (h *SSOHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	session := middleware.GetSession(r.Context())
	if session == nil {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"user_id": session.UserID,
		"org_id":  session.OrgID,
	})
}
