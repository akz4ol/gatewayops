// Package sso provides SSO/OIDC authentication functionality.
package sso

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service manages SSO providers, authentication, and sessions.
type Service struct {
	logger    zerolog.Logger
	providers map[uuid.UUID]*domain.SSOProvider
	states    map[string]*domain.AuthState // keyed by state value
	sessions  map[uuid.UUID]*domain.UserSession
	users     map[uuid.UUID]*domain.User
	mu        sync.RWMutex
}

// NewService creates a new SSO service.
func NewService(logger zerolog.Logger) *Service {
	s := &Service{
		logger:    logger,
		providers: make(map[uuid.UUID]*domain.SSOProvider),
		states:    make(map[string]*domain.AuthState),
		sessions:  make(map[uuid.UUID]*domain.UserSession),
		users:     make(map[uuid.UUID]*domain.User),
	}

	// Create demo provider and user
	s.createDemoData()

	logger.Info().Msg("SSO service initialized")
	return s
}

func (s *Service) createDemoData() {
	// Demo organization
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	demoUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Demo Okta provider
	oktaProvider := &domain.SSOProvider{
		ID:               uuid.New(),
		OrgID:            orgID,
		Type:             domain.SSOProviderOkta,
		Name:             "Demo Okta",
		IssuerURL:        "https://dev-demo.okta.com",
		ClientID:         "demo-client-id",
		AuthorizationURL: "https://dev-demo.okta.com/oauth2/v1/authorize",
		TokenURL:         "https://dev-demo.okta.com/oauth2/v1/token",
		UserInfoURL:      "https://dev-demo.okta.com/oauth2/v1/userinfo",
		Scopes:           []string{"openid", "profile", "email", "groups"},
		ClaimMappings: map[string]string{
			"email": "email",
			"name":  "name",
		},
		GroupMappings: map[string]string{
			"Admins":     "admin",
			"Developers": "developer",
			"Viewers":    "viewer",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.providers[oktaProvider.ID] = oktaProvider

	// Demo Azure AD provider
	azureProvider := &domain.SSOProvider{
		ID:               uuid.New(),
		OrgID:            orgID,
		Type:             domain.SSOProviderAzureAD,
		Name:             "Demo Azure AD",
		IssuerURL:        "https://login.microsoftonline.com/demo-tenant/v2.0",
		ClientID:         "demo-azure-client-id",
		AuthorizationURL: "https://login.microsoftonline.com/demo-tenant/oauth2/v2.0/authorize",
		TokenURL:         "https://login.microsoftonline.com/demo-tenant/oauth2/v2.0/token",
		UserInfoURL:      "https://graph.microsoft.com/oidc/userinfo",
		Scopes:           []string{"openid", "profile", "email"},
		Enabled:          false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	s.providers[azureProvider.ID] = azureProvider

	// Demo Google provider
	googleProvider := &domain.SSOProvider{
		ID:               uuid.New(),
		OrgID:            orgID,
		Type:             domain.SSOProviderGoogle,
		Name:             "Google Workspace",
		IssuerURL:        "https://accounts.google.com",
		ClientID:         "demo-google-client-id",
		AuthorizationURL: "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:         "https://oauth2.googleapis.com/token",
		UserInfoURL:      "https://openidconnect.googleapis.com/v1/userinfo",
		Scopes:           []string{"openid", "profile", "email"},
		Enabled:          true, // Enable for demo
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	s.providers[googleProvider.ID] = googleProvider

	// Demo Auth0 provider
	auth0Provider := &domain.SSOProvider{
		ID:               uuid.New(),
		OrgID:            orgID,
		Type:             domain.SSOProviderAuth0,
		Name:             "Auth0",
		IssuerURL:        "https://demo.auth0.com",
		ClientID:         "demo-auth0-client-id",
		AuthorizationURL: "https://demo.auth0.com/authorize",
		TokenURL:         "https://demo.auth0.com/oauth/token",
		UserInfoURL:      "https://demo.auth0.com/userinfo",
		Scopes:           []string{"openid", "profile", "email"},
		Enabled:          true, // Enable for demo
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	s.providers[auth0Provider.ID] = auth0Provider

	// Demo user
	now := time.Now()
	s.users[demoUserID] = &domain.User{
		ID:          demoUserID,
		OrgID:       orgID,
		Email:       "admin@demo.gatewayops.io",
		Name:        "Demo Admin",
		Status:      domain.UserStatusActive,
		LastLoginAt: &now,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ListProviders returns all SSO providers for an organization.
func (s *Service) ListProviders(orgID uuid.UUID, includeDisabled bool) []domain.SSOProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]domain.SSOProvider, 0)
	for _, p := range s.providers {
		if p.OrgID == orgID && (includeDisabled || p.Enabled) {
			providers = append(providers, *p)
		}
	}
	return providers
}

// GetProvider returns a specific SSO provider.
func (s *Service) GetProvider(id uuid.UUID) *domain.SSOProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.providers[id]
}

// GetProviderByType returns an SSO provider by type.
func (s *Service) GetProviderByType(orgID uuid.UUID, providerType domain.SSOProviderType) *domain.SSOProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.providers {
		if p.OrgID == orgID && p.Type == providerType && p.Enabled {
			return p
		}
	}
	return nil
}

// CreateProvider creates a new SSO provider.
func (s *Service) CreateProvider(input domain.SSOProviderInput, orgID uuid.UUID) *domain.SSOProvider {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set default scopes if not provided
	if len(input.Scopes) == 0 {
		input.Scopes = []string{"openid", "profile", "email"}
	}

	// Generate authorization/token URLs based on provider type
	authURL, tokenURL, userInfoURL := s.getProviderURLs(input.Type, input.IssuerURL)

	provider := &domain.SSOProvider{
		ID:                    uuid.New(),
		OrgID:                 orgID,
		Type:                  input.Type,
		Name:                  input.Name,
		IssuerURL:             input.IssuerURL,
		ClientID:              input.ClientID,
		ClientSecretEncrypted: []byte(input.ClientSecret), // In production, encrypt this
		AuthorizationURL:      authURL,
		TokenURL:              tokenURL,
		UserInfoURL:           userInfoURL,
		Scopes:                input.Scopes,
		ClaimMappings:         input.ClaimMappings,
		GroupMappings:         input.GroupMappings,
		Enabled:               input.Enabled,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	s.providers[provider.ID] = provider

	s.logger.Info().
		Str("provider_id", provider.ID.String()).
		Str("type", string(provider.Type)).
		Str("name", provider.Name).
		Msg("SSO provider created")

	return provider
}

func (s *Service) getProviderURLs(providerType domain.SSOProviderType, issuerURL string) (authURL, tokenURL, userInfoURL string) {
	switch providerType {
	case domain.SSOProviderOkta:
		authURL = issuerURL + "/oauth2/v1/authorize"
		tokenURL = issuerURL + "/oauth2/v1/token"
		userInfoURL = issuerURL + "/oauth2/v1/userinfo"
	case domain.SSOProviderAzureAD:
		authURL = issuerURL + "/oauth2/v2.0/authorize"
		tokenURL = issuerURL + "/oauth2/v2.0/token"
		userInfoURL = "https://graph.microsoft.com/oidc/userinfo"
	case domain.SSOProviderGoogle:
		authURL = "https://accounts.google.com/o/oauth2/v2/auth"
		tokenURL = "https://oauth2.googleapis.com/token"
		userInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
	case domain.SSOProviderOneLogin:
		authURL = issuerURL + "/oidc/2/auth"
		tokenURL = issuerURL + "/oidc/2/token"
		userInfoURL = issuerURL + "/oidc/2/me"
	case domain.SSOProviderAuth0:
		authURL = issuerURL + "/authorize"
		tokenURL = issuerURL + "/oauth/token"
		userInfoURL = issuerURL + "/userinfo"
	default: // Generic OIDC
		authURL = issuerURL + "/authorize"
		tokenURL = issuerURL + "/token"
		userInfoURL = issuerURL + "/userinfo"
	}
	return
}

// UpdateProvider updates an existing SSO provider.
func (s *Service) UpdateProvider(id uuid.UUID, input domain.SSOProviderInput) *domain.SSOProvider {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, exists := s.providers[id]
	if !exists {
		return nil
	}

	if input.Name != "" {
		provider.Name = input.Name
	}
	if input.IssuerURL != "" {
		provider.IssuerURL = input.IssuerURL
		// Update URLs based on new issuer
		provider.AuthorizationURL, provider.TokenURL, provider.UserInfoURL =
			s.getProviderURLs(provider.Type, input.IssuerURL)
	}
	if input.ClientID != "" {
		provider.ClientID = input.ClientID
	}
	if input.ClientSecret != "" {
		provider.ClientSecretEncrypted = []byte(input.ClientSecret)
	}
	if len(input.Scopes) > 0 {
		provider.Scopes = input.Scopes
	}
	if input.ClaimMappings != nil {
		provider.ClaimMappings = input.ClaimMappings
	}
	if input.GroupMappings != nil {
		provider.GroupMappings = input.GroupMappings
	}
	provider.Enabled = input.Enabled
	provider.UpdatedAt = time.Now()

	s.logger.Info().
		Str("provider_id", id.String()).
		Msg("SSO provider updated")

	return provider
}

// DeleteProvider deletes an SSO provider.
func (s *Service) DeleteProvider(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.providers[id]; !exists {
		return false
	}

	delete(s.providers, id)

	s.logger.Info().
		Str("provider_id", id.String()).
		Msg("SSO provider deleted")

	return true
}

// GenerateAuthState generates OAuth state for CSRF protection.
func (s *Service) GenerateAuthState(providerID uuid.UUID, redirectURL string) (*domain.AuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random state and nonce
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, err
	}
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}

	state := &domain.AuthState{
		State:       hex.EncodeToString(stateBytes),
		Nonce:       hex.EncodeToString(nonceBytes),
		RedirectURL: redirectURL,
		ProviderID:  providerID,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	s.states[state.State] = state

	return state, nil
}

// ValidateAuthState validates and consumes an OAuth state.
func (s *Service) ValidateAuthState(stateValue string) (*domain.AuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.states[stateValue]
	if !exists {
		return nil, fmt.Errorf("invalid state")
	}

	// Delete state (one-time use)
	delete(s.states, stateValue)

	if time.Now().After(state.ExpiresAt) {
		return nil, fmt.Errorf("state expired")
	}

	return state, nil
}

// GetAuthorizationURL returns the OAuth authorization URL for a provider.
func (s *Service) GetAuthorizationURL(providerID uuid.UUID, state *domain.AuthState, callbackURL string) (string, error) {
	s.mu.RLock()
	provider := s.providers[providerID]
	s.mu.RUnlock()

	if provider == nil {
		return "", fmt.Errorf("provider not found")
	}

	if !provider.Enabled {
		return "", fmt.Errorf("provider is disabled")
	}

	// Build authorization URL
	url := fmt.Sprintf("%s?client_id=%s&response_type=code&scope=%s&state=%s&nonce=%s&redirect_uri=%s",
		provider.AuthorizationURL,
		provider.ClientID,
		joinScopes(provider.Scopes),
		state.State,
		state.Nonce,
		callbackURL,
	)

	return url, nil
}

func joinScopes(scopes []string) string {
	result := ""
	for i, s := range scopes {
		if i > 0 {
			result += "%20"
		}
		result += s
	}
	return result
}

// ExchangeCode exchanges an authorization code for tokens.
// In demo mode, this simulates the exchange.
func (s *Service) ExchangeCode(providerID uuid.UUID, code string, redirectURI string) (*domain.TokenPair, *domain.OIDCClaims, error) {
	s.mu.RLock()
	provider := s.providers[providerID]
	s.mu.RUnlock()

	if provider == nil {
		return nil, nil, fmt.Errorf("provider not found")
	}

	// In demo mode, simulate token exchange
	// In production, this would make HTTP calls to the provider's token endpoint

	s.logger.Info().
		Str("provider_id", providerID.String()).
		Str("code", code[:min(8, len(code))]+"...").
		Msg("Demo: simulating token exchange")

	// Generate demo tokens
	accessToken := generateDemoToken("access")
	refreshToken := generateDemoToken("refresh")

	tokenPair := &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	// Simulate OIDC claims
	claims := &domain.OIDCClaims{
		Subject:       "demo-user-" + uuid.New().String()[:8],
		Email:         "user@demo.gatewayops.io",
		EmailVerified: true,
		Name:          "Demo User",
		Groups:        []string{"Developers"},
	}

	return tokenPair, claims, nil
}

func generateDemoToken(prefix string) string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, base64.URLEncoding.EncodeToString(b))
}

// CreateSession creates a new user session.
func (s *Service) CreateSession(user *domain.User, ipAddress, userAgent string) *domain.UserSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &domain.UserSession{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrgID:          user.OrgID,
		AccessToken:    generateDemoToken("session"),
		RefreshToken:   generateDemoToken("refresh"),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		LastActivityAt: time.Now(),
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
	}

	s.sessions[session.ID] = session

	s.logger.Info().
		Str("session_id", session.ID.String()).
		Str("user_id", user.ID.String()).
		Msg("Session created")

	return session
}

// GetSession returns a session by ID.
func (s *Service) GetSession(id uuid.UUID) *domain.UserSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[id]
}

// ValidateSession validates a session token.
func (s *Service) ValidateSession(token string) (*domain.UserSession, *domain.User) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.sessions {
		if session.AccessToken == token {
			if time.Now().After(session.ExpiresAt) {
				return nil, nil // Expired
			}
			user := s.users[session.UserID]
			return session, user
		}
	}
	return nil, nil
}

// RefreshSession refreshes an expired session.
func (s *Service) RefreshSession(refreshToken string) *domain.UserSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.sessions {
		if session.RefreshToken == refreshToken {
			// Generate new tokens
			session.AccessToken = generateDemoToken("session")
			session.ExpiresAt = time.Now().Add(24 * time.Hour)
			session.LastActivityAt = time.Now()
			return session
		}
	}
	return nil
}

// RevokeSession revokes a session.
func (s *Service) RevokeSession(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; !exists {
		return false
	}

	delete(s.sessions, id)

	s.logger.Info().
		Str("session_id", id.String()).
		Msg("Session revoked")

	return true
}

// ListUserSessions returns all active sessions for a user.
func (s *Service) ListUserSessions(userID uuid.UUID) []domain.UserSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]domain.UserSession, 0)
	for _, session := range s.sessions {
		if session.UserID == userID && time.Now().Before(session.ExpiresAt) {
			sessions = append(sessions, *session)
		}
	}
	return sessions
}

// RevokeAllUserSessions revokes all sessions for a user.
func (s *Service) RevokeAllUserSessions(userID uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for id, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, id)
			count++
		}
	}

	s.logger.Info().
		Str("user_id", userID.String()).
		Int("sessions_revoked", count).
		Msg("All user sessions revoked")

	return count
}

// GetOrCreateUser gets or creates a user from OIDC claims.
func (s *Service) GetOrCreateUser(orgID uuid.UUID, providerID uuid.UUID, claims *domain.OIDCClaims) *domain.User {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists by SSO external ID
	for _, user := range s.users {
		if user.SSOExternalID == claims.Subject && user.SSOProviderID != nil && *user.SSOProviderID == providerID {
			// Update last login
			now := time.Now()
			user.LastLoginAt = &now
			user.UpdatedAt = now
			return user
		}
	}

	// Check if user exists by email
	for _, user := range s.users {
		if user.Email == claims.Email && user.OrgID == orgID {
			// Link SSO
			user.SSOProviderID = &providerID
			user.SSOExternalID = claims.Subject
			now := time.Now()
			user.LastLoginAt = &now
			user.UpdatedAt = now
			return user
		}
	}

	// Create new user
	now := time.Now()
	user := &domain.User{
		ID:            uuid.New(),
		OrgID:         orgID,
		Email:         claims.Email,
		Name:          claims.Name,
		AvatarURL:     claims.Picture,
		Status:        domain.UserStatusActive,
		SSOProviderID: &providerID,
		SSOExternalID: claims.Subject,
		LastLoginAt:   &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.users[user.ID] = user

	s.logger.Info().
		Str("user_id", user.ID.String()).
		Str("email", user.Email).
		Msg("User created from SSO")

	return user
}

// GetUser returns a user by ID.
func (s *Service) GetUser(id uuid.UUID) *domain.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[id]
}

// ProviderStats returns statistics about SSO providers.
func (s *Service) ProviderStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enabledCount := 0
	byType := make(map[string]int)

	for _, p := range s.providers {
		byType[string(p.Type)]++
		if p.Enabled {
			enabledCount++
		}
	}

	return map[string]interface{}{
		"total_providers":   len(s.providers),
		"enabled_providers": enabledCount,
		"by_type":           byType,
		"active_sessions":   len(s.sessions),
		"total_users":       len(s.users),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
