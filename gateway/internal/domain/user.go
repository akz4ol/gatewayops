package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus represents the status of a user account.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// User represents a user in the system.
type User struct {
	ID            uuid.UUID  `json:"id"`
	OrgID         uuid.UUID  `json:"org_id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	AvatarURL     string     `json:"avatar_url,omitempty"`
	Status        UserStatus `json:"status"`
	SSOProviderID *uuid.UUID `json:"sso_provider_id,omitempty"`
	SSOExternalID string     `json:"sso_external_id,omitempty"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UserSession represents an active user session.
type UserSession struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	OrgID          uuid.UUID `json:"org_id"`
	AccessToken    string    `json:"-"` // Never serialize
	RefreshToken   string    `json:"-"` // Never serialize
	ExpiresAt      time.Time `json:"expires_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// SSOProviderType represents the type of SSO provider.
type SSOProviderType string

const (
	SSOProviderOkta            SSOProviderType = "okta"
	SSOProviderAzureAD         SSOProviderType = "azure_ad"
	SSOProviderGoogle          SSOProviderType = "google"
	SSOProviderOneLogin        SSOProviderType = "onelogin"
	SSOProviderAuth0           SSOProviderType = "auth0"
	SSOProviderGenericOIDC     SSOProviderType = "oidc"
)

// SSOProvider represents an SSO/OIDC provider configuration.
type SSOProvider struct {
	ID                    uuid.UUID              `json:"id"`
	OrgID                 uuid.UUID              `json:"org_id"`
	Type                  SSOProviderType        `json:"type"`
	Name                  string                 `json:"name"`
	IssuerURL             string                 `json:"issuer_url"`
	ClientID              string                 `json:"client_id"`
	ClientSecretEncrypted []byte                 `json:"-"` // Never serialize
	AuthorizationURL      string                 `json:"authorization_url,omitempty"`
	TokenURL              string                 `json:"token_url,omitempty"`
	UserInfoURL           string                 `json:"userinfo_url,omitempty"`
	Scopes                []string               `json:"scopes"`
	ClaimMappings         map[string]string      `json:"claim_mappings,omitempty"`
	GroupMappings         map[string]string      `json:"group_mappings,omitempty"` // SSO group -> Role name
	Enabled               bool                   `json:"enabled"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// SSOProviderInput represents input for creating/updating an SSO provider.
type SSOProviderInput struct {
	Type          SSOProviderType   `json:"type"`
	Name          string            `json:"name"`
	IssuerURL     string            `json:"issuer_url"`
	ClientID      string            `json:"client_id"`
	ClientSecret  string            `json:"client_secret"`
	Scopes        []string          `json:"scopes,omitempty"`
	ClaimMappings map[string]string `json:"claim_mappings,omitempty"`
	GroupMappings map[string]string `json:"group_mappings,omitempty"`
	Enabled       bool              `json:"enabled"`
}

// OIDCClaims represents claims from an OIDC token.
type OIDCClaims struct {
	Subject       string   `json:"sub"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture,omitempty"`
	Groups        []string `json:"groups,omitempty"`
}

// AuthState represents OAuth state for CSRF protection.
type AuthState struct {
	State       string    `json:"state"`
	Nonce       string    `json:"nonce"`
	RedirectURL string    `json:"redirect_url"`
	ProviderID  uuid.UUID `json:"provider_id"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// TokenPair represents access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}
