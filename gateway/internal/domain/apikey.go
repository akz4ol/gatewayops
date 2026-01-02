package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API key.
type APIKey struct {
	ID          uuid.UUID  `json:"id"`
	OrgID       uuid.UUID  `json:"org_id"`
	TeamID      *uuid.UUID `json:"team_id,omitempty"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"` // First 8 chars for identification
	KeyHash     string     `json:"-"`          // Hashed key, never exposed
	Environment string     `json:"environment"` // production, staging, development
	Permissions []string   `json:"permissions"`
	RateLimit   int        `json:"rate_limit"` // Requests per minute
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	Revoked     bool       `json:"revoked"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
}

// APIKeyCreate represents the request to create a new API key.
type APIKeyCreate struct {
	Name        string     `json:"name"`
	TeamID      *uuid.UUID `json:"team_id,omitempty"`
	Environment string     `json:"environment"`
	Permissions []string   `json:"permissions,omitempty"`
	RateLimit   int        `json:"rate_limit,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// APIKeyCreated is returned after creating an API key (includes raw key).
type APIKeyCreated struct {
	APIKey
	RawKey string `json:"key"` // Only returned once on creation
}

// APIKeyUsage represents usage stats for an API key.
type APIKeyUsage struct {
	KeyID         uuid.UUID `json:"key_id"`
	TotalRequests int64     `json:"total_requests"`
	TotalCost     float64   `json:"total_cost"`
	LastUsedAt    time.Time `json:"last_used_at"`
}

// APIKeyFilter represents filtering options for listing API keys.
type APIKeyFilter struct {
	OrgID          uuid.UUID
	TeamID         *uuid.UUID
	Environment    string
	IncludeRevoked bool
	Limit          int
	Offset         int
}
