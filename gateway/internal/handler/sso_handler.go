package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/sso"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SSOHandler handles SSO-related HTTP requests.
type SSOHandler struct {
	logger     zerolog.Logger
	service    *sso.Service
	baseURL    string
}

// NewSSOHandler creates a new SSO handler.
func NewSSOHandler(logger zerolog.Logger, service *sso.Service, baseURL string) *SSOHandler {
	return &SSOHandler{
		logger:  logger,
		service: service,
		baseURL: baseURL,
	}
}

// ListProviders returns all SSO providers for the organization.
func (h *SSOHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	includeDisabled := r.URL.Query().Get("include_disabled") == "true"

	// Demo organization
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	providers := h.service.ListProviders(orgID, includeDisabled)

	// Mask sensitive data
	safeProviders := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		safeProviders = append(safeProviders, h.sanitizeProvider(p))
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"providers": safeProviders,
		"total":     len(providers),
	})
}

// GetProvider returns a specific SSO provider.
func (h *SSOHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "providerID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	provider := h.service.GetProvider(id)
	if provider == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Provider not found")
		return
	}

	WriteJSON(w, http.StatusOK, h.sanitizeProvider(*provider))
}

// CreateProvider creates a new SSO provider.
func (h *SSOHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var input domain.SSOProviderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	if input.Type == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Provider type is required")
		return
	}
	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Name is required")
		return
	}
	if input.IssuerURL == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Issuer URL is required")
		return
	}
	if input.ClientID == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Client ID is required")
		return
	}
	if input.ClientSecret == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Client secret is required")
		return
	}

	// Demo organization
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	provider := h.service.CreateProvider(input, orgID)
	WriteJSON(w, http.StatusCreated, h.sanitizeProvider(*provider))
}

// UpdateProvider updates an existing SSO provider.
func (h *SSOHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "providerID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	var input domain.SSOProviderInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
		return
	}

	provider := h.service.UpdateProvider(id, input)
	if provider == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Provider not found")
		return
	}

	WriteJSON(w, http.StatusOK, h.sanitizeProvider(*provider))
}

// DeleteProvider deletes an SSO provider.
func (h *SSOHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "providerID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	if !h.service.DeleteProvider(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Provider not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Authorize initiates the OAuth authorization flow.
func (h *SSOHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	providerIDStr := chi.URLParam(r, "providerID")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	provider := h.service.GetProvider(providerID)
	if provider == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Provider not found")
		return
	}

	if !provider.Enabled {
		WriteError(w, http.StatusBadRequest, "provider_disabled", "This SSO provider is disabled")
		return
	}

	// Get redirect URL from query param or use default
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = h.baseURL + "/dashboard"
	}

	// Generate state for CSRF protection
	state, err := h.service.GenerateAuthState(providerID, redirectURL)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate auth state")
		WriteError(w, http.StatusInternalServerError, "state_error", "Failed to initiate login")
		return
	}

	// Build callback URL
	callbackURL := h.baseURL + "/v1/sso/callback/" + providerID.String()

	// Get authorization URL
	authURL, err := h.service.GetAuthorizationURL(providerID, state, callbackURL)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get authorization URL")
		WriteError(w, http.StatusInternalServerError, "auth_url_error", "Failed to initiate login")
		return
	}

	// For API calls, return the URL; for browser, redirect
	if r.Header.Get("Accept") == "application/json" {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"authorization_url": authURL,
			"state":             state.State,
			"expires_at":        state.ExpiresAt,
		})
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback handles the OAuth callback from the identity provider.
func (h *SSOHandler) Callback(w http.ResponseWriter, r *http.Request) {
	providerIDStr := chi.URLParam(r, "providerID")
	providerID, err := uuid.Parse(providerIDStr)
	if err != nil {
		h.renderError(w, r, "Invalid provider ID")
		return
	}

	// Check for error from provider
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		errDesc := r.URL.Query().Get("error_description")
		h.logger.Warn().
			Str("error", errCode).
			Str("description", errDesc).
			Msg("OAuth error from provider")
		h.renderError(w, r, "Authentication failed: "+errDesc)
		return
	}

	// Validate state
	stateValue := r.URL.Query().Get("state")
	state, err := h.service.ValidateAuthState(stateValue)
	if err != nil {
		h.logger.Warn().Err(err).Msg("Invalid OAuth state")
		h.renderError(w, r, "Invalid or expired login session")
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.renderError(w, r, "No authorization code received")
		return
	}

	// Exchange code for tokens
	callbackURL := h.baseURL + "/v1/sso/callback/" + providerID.String()
	tokenPair, claims, err := h.service.ExchangeCode(providerID, code, callbackURL)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to exchange code")
		h.renderError(w, r, "Failed to complete authentication")
		return
	}

	// Get or create user
	provider := h.service.GetProvider(providerID)
	user := h.service.GetOrCreateUser(provider.OrgID, providerID, claims)

	// Create session
	session := h.service.CreateSession(user, r.RemoteAddr, r.UserAgent())

	h.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Str("provider", string(provider.Type)).
		Msg("SSO login successful")

	// For API calls, return tokens; for browser, redirect with cookie
	if r.Header.Get("Accept") == "application/json" {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"user":         user,
			"session":      session,
			"access_token": tokenPair.AccessToken,
			"token_type":   tokenPair.TokenType,
			"expires_in":   tokenPair.ExpiresIn,
		})
		return
	}

	// Set session cookie and redirect
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.AccessToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, state.RedirectURL, http.StatusFound)
}

// Logout logs out the current user.
func (h *SSOHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session from cookie or header
	var token string
	if cookie, err := r.Cookie("session"); err == nil {
		token = cookie.Value
	} else if auth := r.Header.Get("Authorization"); len(auth) > 7 {
		token = auth[7:] // Remove "Bearer "
	}

	if token != "" {
		session, _ := h.service.ValidateSession(token)
		if session != nil {
			h.service.RevokeSession(session.ID)
		}
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	})

	if r.Header.Get("Accept") == "application/json" {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// ListSessions returns all active sessions for the current user.
func (h *SSOHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Demo user
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	sessions := h.service.ListUserSessions(userID)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// RevokeSession revokes a specific session.
func (h *SSOHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "sessionID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid session ID")
		return
	}

	if !h.service.RevokeSession(id) {
		WriteError(w, http.StatusNotFound, "not_found", "Session not found")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// RevokeAllSessions revokes all sessions for the current user.
func (h *SSOHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	// Demo user
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	count := h.service.RevokeAllUserSessions(userID)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":           "revoked",
		"sessions_revoked": count,
	})
}

// GetStats returns SSO statistics.
func (h *SSOHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.service.ProviderStats()
	WriteJSON(w, http.StatusOK, stats)
}

// GetSupportedProviders returns the list of supported SSO provider types.
func (h *SSOHandler) GetSupportedProviders(w http.ResponseWriter, r *http.Request) {
	providers := []map[string]interface{}{
		{
			"type":        "okta",
			"name":        "Okta",
			"description": "Enterprise identity management",
			"logo_url":    "https://www.okta.com/sites/default/files/Okta_Logo_BrightBlue_Medium.png",
		},
		{
			"type":        "azure_ad",
			"name":        "Microsoft Azure AD",
			"description": "Microsoft identity platform",
			"logo_url":    "https://azure.microsoft.com/svghandler/azure-active-directory/",
		},
		{
			"type":        "google",
			"name":        "Google Workspace",
			"description": "Google Cloud Identity",
			"logo_url":    "https://www.google.com/images/branding/googleg/1x/googleg_standard_color_128dp.png",
		},
		{
			"type":        "onelogin",
			"name":        "OneLogin",
			"description": "Cloud-based identity management",
			"logo_url":    "https://www.onelogin.com/assets/img/press/logo/onelogin-dark.svg",
		},
		{
			"type":        "auth0",
			"name":        "Auth0",
			"description": "Universal login platform",
			"logo_url":    "https://cdn.auth0.com/styleguide/latest/lib/logos/img/badge.png",
		},
		{
			"type":        "oidc",
			"name":        "Generic OIDC",
			"description": "Any OpenID Connect compatible provider",
			"logo_url":    "",
		},
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"supported_providers": providers,
		"total":               len(providers),
	})
}

// TestConnection tests the connection to an SSO provider.
func (h *SSOHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "providerID")
	id, err := uuid.Parse(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_id", "Invalid provider ID")
		return
	}

	provider := h.service.GetProvider(id)
	if provider == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Provider not found")
		return
	}

	// In demo mode, simulate connection test
	// In production, this would attempt to fetch the OIDC discovery document

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "success",
		"provider_id":   id,
		"provider_type": provider.Type,
		"issuer_url":    provider.IssuerURL,
		"message":       "Connection test successful (demo mode)",
		"endpoints": map[string]string{
			"authorization": provider.AuthorizationURL,
			"token":         provider.TokenURL,
			"userinfo":      provider.UserInfoURL,
		},
	})
}

func (h *SSOHandler) sanitizeProvider(p domain.SSOProvider) map[string]interface{} {
	return map[string]interface{}{
		"id":                p.ID,
		"org_id":            p.OrgID,
		"type":              p.Type,
		"name":              p.Name,
		"issuer_url":        p.IssuerURL,
		"client_id":         p.ClientID,
		"authorization_url": p.AuthorizationURL,
		"token_url":         p.TokenURL,
		"userinfo_url":      p.UserInfoURL,
		"scopes":            p.Scopes,
		"claim_mappings":    p.ClaimMappings,
		"group_mappings":    p.GroupMappings,
		"enabled":           p.Enabled,
		"created_at":        p.CreatedAt,
		"updated_at":        p.UpdatedAt,
	}
}

func (h *SSOHandler) renderError(w http.ResponseWriter, r *http.Request, message string) {
	if r.Header.Get("Accept") == "application/json" {
		WriteError(w, http.StatusBadRequest, "auth_error", message)
		return
	}

	// Redirect to error page with message
	errorURL := h.baseURL + "/login?error=" + url.QueryEscape(message)
	http.Redirect(w, r, errorURL, http.StatusFound)
}
