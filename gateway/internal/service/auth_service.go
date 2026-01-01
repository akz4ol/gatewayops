package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
)

// AuthService handles SSO/OIDC authentication.
type AuthService struct {
	userRepo      *repository.UserRepository
	auditService  *AuditService
	encryptionKey []byte
	baseURL       string
	logger        *slog.Logger
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	userRepo *repository.UserRepository,
	auditService *AuditService,
	encryptionKey string,
	baseURL string,
	logger *slog.Logger,
) *AuthService {
	key := []byte(encryptionKey)
	if len(key) < 32 {
		// Pad or hash to get 32 bytes for AES-256
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	}

	return &AuthService{
		userRepo:      userRepo,
		auditService:  auditService,
		encryptionKey: key[:32],
		baseURL:       baseURL,
		logger:        logger,
	}
}

// CreateSSOProvider creates a new SSO provider configuration.
func (s *AuthService) CreateSSOProvider(ctx context.Context, orgID uuid.UUID, input domain.SSOProviderInput) (*domain.SSOProvider, error) {
	// Encrypt client secret
	encryptedSecret, err := s.encryptSecret(input.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("encrypt client secret: %w", err)
	}

	// Set default scopes if not provided
	scopes := input.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "email", "profile"}
	}

	provider := &domain.SSOProvider{
		ID:                    uuid.New(),
		OrgID:                 orgID,
		Type:                  input.Type,
		Name:                  input.Name,
		IssuerURL:             input.IssuerURL,
		ClientID:              input.ClientID,
		ClientSecretEncrypted: encryptedSecret,
		Scopes:                scopes,
		ClaimMappings:         input.ClaimMappings,
		GroupMappings:         input.GroupMappings,
		Enabled:               input.Enabled,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Discover OIDC endpoints
	if err := s.discoverOIDCEndpoints(ctx, provider); err != nil {
		s.logger.Warn("failed to discover OIDC endpoints, using manual configuration",
			"issuer", input.IssuerURL,
			"error", err,
		)
	}

	if err := s.userRepo.CreateSSOProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("create SSO provider: %w", err)
	}

	s.logger.Info("SSO provider created",
		"id", provider.ID,
		"type", provider.Type,
		"name", provider.Name,
	)

	return provider, nil
}

// discoverOIDCEndpoints discovers OIDC endpoints from the issuer.
func (s *AuthService) discoverOIDCEndpoints(ctx context.Context, provider *domain.SSOProvider) error {
	oidcProvider, err := oidc.NewProvider(ctx, provider.IssuerURL)
	if err != nil {
		return err
	}

	// Get the endpoint from the provider
	endpoint := oidcProvider.Endpoint()
	provider.AuthorizationURL = endpoint.AuthURL
	provider.TokenURL = endpoint.TokenURL

	// Try to get userinfo URL from claims
	var claims struct {
		UserInfoURL string `json:"userinfo_endpoint"`
	}
	if err := oidcProvider.Claims(&claims); err == nil && claims.UserInfoURL != "" {
		provider.UserInfoURL = claims.UserInfoURL
	}

	return nil
}

// GetSSOProvider retrieves an SSO provider by ID.
func (s *AuthService) GetSSOProvider(ctx context.Context, id uuid.UUID) (*domain.SSOProvider, error) {
	return s.userRepo.GetSSOProvider(ctx, id)
}

// ListSSOProviders retrieves all SSO providers for an organization.
func (s *AuthService) ListSSOProviders(ctx context.Context, orgID uuid.UUID) ([]domain.SSOProvider, error) {
	return s.userRepo.ListSSOProviders(ctx, orgID)
}

// UpdateSSOProvider updates an SSO provider.
func (s *AuthService) UpdateSSOProvider(ctx context.Context, id uuid.UUID, input domain.SSOProviderInput) (*domain.SSOProvider, error) {
	provider, err := s.userRepo.GetSSOProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("SSO provider not found")
	}

	provider.Name = input.Name
	provider.IssuerURL = input.IssuerURL
	provider.ClientID = input.ClientID
	provider.Scopes = input.Scopes
	provider.ClaimMappings = input.ClaimMappings
	provider.GroupMappings = input.GroupMappings
	provider.Enabled = input.Enabled
	provider.UpdatedAt = time.Now()

	// Update secret if provided
	if input.ClientSecret != "" {
		encryptedSecret, err := s.encryptSecret(input.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("encrypt client secret: %w", err)
		}
		provider.ClientSecretEncrypted = encryptedSecret
	}

	// Re-discover endpoints
	if err := s.discoverOIDCEndpoints(ctx, provider); err != nil {
		s.logger.Warn("failed to discover OIDC endpoints", "error", err)
	}

	if err := s.userRepo.UpdateSSOProvider(ctx, provider); err != nil {
		return nil, fmt.Errorf("update SSO provider: %w", err)
	}

	return provider, nil
}

// DeleteSSOProvider deletes an SSO provider.
func (s *AuthService) DeleteSSOProvider(ctx context.Context, id uuid.UUID) error {
	return s.userRepo.DeleteSSOProvider(ctx, id)
}

// GetAuthorizationURL generates the OAuth authorization URL for SSO.
func (s *AuthService) GetAuthorizationURL(ctx context.Context, providerID uuid.UUID, redirectURL string) (string, string, error) {
	provider, err := s.userRepo.GetSSOProvider(ctx, providerID)
	if err != nil {
		return "", "", err
	}
	if provider == nil {
		return "", "", fmt.Errorf("SSO provider not found")
	}
	if !provider.Enabled {
		return "", "", fmt.Errorf("SSO provider is disabled")
	}

	// Decrypt client secret
	clientSecret, err := s.decryptSecret(provider.ClientSecretEncrypted)
	if err != nil {
		return "", "", fmt.Errorf("decrypt client secret: %w", err)
	}

	// Generate state for CSRF protection
	state := uuid.New().String()

	oauth2Config := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthorizationURL,
			TokenURL: provider.TokenURL,
		},
		RedirectURL: fmt.Sprintf("%s/v1/sso/callback/%s", s.baseURL, providerID.String()),
		Scopes:      provider.Scopes,
	}

	authURL := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, state, nil
}

// HandleCallback processes the OAuth callback.
func (s *AuthService) HandleCallback(ctx context.Context, providerID uuid.UUID, code, state string, ipAddress, userAgent string) (*domain.TokenPair, *domain.User, error) {
	provider, err := s.userRepo.GetSSOProvider(ctx, providerID)
	if err != nil {
		return nil, nil, err
	}
	if provider == nil {
		return nil, nil, fmt.Errorf("SSO provider not found")
	}

	// Decrypt client secret
	clientSecret, err := s.decryptSecret(provider.ClientSecretEncrypted)
	if err != nil {
		return nil, nil, fmt.Errorf("decrypt client secret: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  provider.AuthorizationURL,
			TokenURL: provider.TokenURL,
		},
		RedirectURL: fmt.Sprintf("%s/v1/sso/callback/%s", s.baseURL, providerID.String()),
		Scopes:      provider.Scopes,
	}

	// Exchange code for token
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("exchange code: %w", err)
	}

	// Create OIDC provider for token verification
	oidcProvider, err := oidc.NewProvider(ctx, provider.IssuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("create OIDC provider: %w", err)
	}

	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientID})

	// Verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, nil, fmt.Errorf("no id_token in response")
	}

	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, nil, fmt.Errorf("verify ID token: %w", err)
	}

	// Extract claims
	var claims domain.OIDCClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, nil, fmt.Errorf("parse claims: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateUser(ctx, provider, &claims)
	if err != nil {
		return nil, nil, fmt.Errorf("find or create user: %w", err)
	}

	// Create session
	session := &domain.UserSession{
		ID:             uuid.New(),
		UserID:         user.ID,
		OrgID:          user.OrgID,
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresAt:      token.Expiry,
		LastActivityAt: time.Now(),
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	// Update user last login
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	s.userRepo.UpdateUser(ctx, user)

	// Log auth event
	s.auditService.LogAuthEvent(ctx, user.OrgID, user.ID, domain.AuditActionUserLogin, true, ipAddress, userAgent, map[string]interface{}{
		"provider_id":   providerID.String(),
		"provider_type": provider.Type,
	})

	return &domain.TokenPair{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}, user, nil
}

// findOrCreateUser finds an existing user or creates a new one based on OIDC claims.
func (s *AuthService) findOrCreateUser(ctx context.Context, provider *domain.SSOProvider, claims *domain.OIDCClaims) (*domain.User, error) {
	// First, try to find by SSO external ID
	user, err := s.userRepo.GetUserBySSOExternalID(ctx, provider.ID, claims.Subject)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// Update user info if changed
		updated := false
		if user.Email != claims.Email {
			user.Email = claims.Email
			updated = true
		}
		if user.Name != claims.Name {
			user.Name = claims.Name
			updated = true
		}
		if claims.Picture != "" && user.AvatarURL != claims.Picture {
			user.AvatarURL = claims.Picture
			updated = true
		}
		if updated {
			user.UpdatedAt = time.Now()
			s.userRepo.UpdateUser(ctx, user)
		}
		return user, nil
	}

	// Try to find by email
	user, err = s.userRepo.GetUserByEmail(ctx, provider.OrgID, claims.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// Link existing user to SSO provider
		user.SSOProviderID = &provider.ID
		user.SSOExternalID = claims.Subject
		user.UpdatedAt = time.Now()
		if err := s.userRepo.UpdateUser(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	// Create new user
	user = &domain.User{
		ID:            uuid.New(),
		OrgID:         provider.OrgID,
		Email:         claims.Email,
		Name:          claims.Name,
		AvatarURL:     claims.Picture,
		Status:        domain.UserStatusActive,
		SSOProviderID: &provider.ID,
		SSOExternalID: claims.Subject,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	s.logger.Info("user created via SSO",
		"user_id", user.ID,
		"email", user.Email,
		"provider_id", provider.ID,
	)

	return user, nil
}

// Logout invalidates a user session.
func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID, ipAddress, userAgent string) error {
	session, err := s.userRepo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return nil // Already logged out
	}

	if err := s.userRepo.DeleteSession(ctx, sessionID); err != nil {
		return err
	}

	s.auditService.LogAuthEvent(ctx, session.OrgID, session.UserID, domain.AuditActionUserLogout, true, ipAddress, userAgent, nil)

	return nil
}

// GetSession retrieves a session by ID.
func (s *AuthService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	return s.userRepo.GetSession(ctx, sessionID)
}

// UpdateSessionActivity updates the last activity time of a session.
func (s *AuthService) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	return s.userRepo.UpdateSessionActivity(ctx, sessionID)
}

// CleanupExpiredSessions removes all expired sessions.
func (s *AuthService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.userRepo.DeleteExpiredSessions(ctx)
}

// encryptSecret encrypts a secret using AES-GCM.
func (s *AuthService) encryptSecret(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// decryptSecret decrypts a secret using AES-GCM.
func (s *AuthService) decryptSecret(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateAPIToken generates a random API token.
func GenerateAPIToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
