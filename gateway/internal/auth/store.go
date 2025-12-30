// Package auth provides API key authentication.
package auth

import (
	"context"
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/middleware"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidKey = errors.New("invalid API key")
	ErrExpiredKey = errors.New("API key has expired")
	ErrRevokedKey = errors.New("API key has been revoked")
)

// Store implements middleware.AuthStore for API key validation.
type Store struct {
	db     *sql.DB
	logger zerolog.Logger
	cache  *keyCache
}

// keyCache provides in-memory caching of validated keys.
type keyCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	ttl   time.Duration
}

type cacheItem struct {
	info      *middleware.AuthInfo
	expiresAt time.Time
}

// NewStore creates a new auth store.
func NewStore(db *sql.DB, logger zerolog.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger,
		cache: &keyCache{
			items: make(map[string]*cacheItem),
			ttl:   5 * time.Minute,
		},
	}
}

// ValidateAPIKey validates an API key and returns auth info.
func (s *Store) ValidateAPIKey(ctx context.Context, apiKey string) (*middleware.AuthInfo, error) {
	// Hash the key for cache lookup (never store raw keys)
	keyHash := hashKey(apiKey)

	// Check cache first
	if info := s.cache.get(keyHash); info != nil {
		return info, nil
	}

	// Extract key prefix for database lookup
	// Format: gwo_{env}_{32chars}
	// We store a hash of the full key, but we can lookup by prefix
	keyPrefix := extractKeyPrefix(apiKey)
	if keyPrefix == "" {
		return nil, ErrInvalidKey
	}

	// Query database
	info, storedHash, err := s.lookupKey(ctx, keyPrefix)
	if err != nil {
		return nil, err
	}

	// Verify key matches stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(apiKey)); err != nil {
		return nil, ErrInvalidKey
	}

	// Cache the result
	s.cache.set(keyHash, info)

	return info, nil
}

// lookupKey queries the database for a key by prefix.
func (s *Store) lookupKey(ctx context.Context, keyPrefix string) (*middleware.AuthInfo, string, error) {
	query := `
		SELECT
			k.id,
			k.org_id,
			k.team_id,
			k.key_hash,
			k.environment,
			k.permissions,
			k.rate_limit_rpm,
			k.expires_at,
			k.revoked_at
		FROM api_keys k
		WHERE k.key_prefix = $1
		LIMIT 1
	`

	var (
		keyID       string
		orgID       string
		teamID      sql.NullString
		keyHash     string
		environment string
		permissions string
		rateLimit   int
		expiresAt   sql.NullTime
		revokedAt   sql.NullTime
	)

	err := s.db.QueryRowContext(ctx, query, keyPrefix).Scan(
		&keyID,
		&orgID,
		&teamID,
		&keyHash,
		&environment,
		&permissions,
		&rateLimit,
		&expiresAt,
		&revokedAt,
	)

	if err == sql.ErrNoRows {
		return nil, "", ErrInvalidKey
	}
	if err != nil {
		s.logger.Error().Err(err).Str("key_prefix", keyPrefix).Msg("Database error looking up API key")
		return nil, "", err
	}

	// Check if key is revoked
	if revokedAt.Valid {
		return nil, "", ErrRevokedKey
	}

	// Check if key is expired
	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		return nil, "", ErrExpiredKey
	}

	// Parse permissions
	perms := strings.Split(permissions, ",")

	info := &middleware.AuthInfo{
		KeyID:       keyID,
		OrgID:       orgID,
		TeamID:      teamID.String,
		Environment: environment,
		Permissions: perms,
		RateLimit:   rateLimit,
	}

	return info, keyHash, nil
}

// hashKey creates a SHA-256 hash of the API key for cache lookup.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// extractKeyPrefix extracts the prefix portion of an API key.
// Format: gwo_{env}_{32chars} -> gwo_{env}_{first8chars}
func extractKeyPrefix(key string) string {
	parts := strings.SplitN(key, "_", 3)
	if len(parts) != 3 {
		return ""
	}
	if len(parts[2]) < 8 {
		return ""
	}
	return parts[0] + "_" + parts[1] + "_" + parts[2][:8]
}

// Cache methods

func (c *keyCache) get(keyHash string) *middleware.AuthInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[keyHash]
	if !ok {
		return nil
	}
	if time.Now().After(item.expiresAt) {
		return nil
	}
	return item.info
}

func (c *keyCache) set(keyHash string, info *middleware.AuthInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[keyHash] = &cacheItem{
		info:      info,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// GenerateAPIKey generates a new API key.
// Format: gwo_{env}_{32_random_chars}
func GenerateAPIKey(env string) (string, error) {
	// Generate 16 random bytes (32 hex chars)
	randomBytes := make([]byte, 16)
	if _, err := cryptoRand.Read(randomBytes); err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomBytes)
	return "gwo_" + env + "_" + randomStr, nil
}

// HashAPIKey creates a bcrypt hash of an API key for storage.
func HashAPIKey(apiKey string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(apiKey), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
