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
	"github.com/google/uuid"
	"github.com/rs/zerolog"
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
// In demo mode, returns mock auth info for any key starting with "gwo_".
func (s *Store) ValidateAPIKey(ctx context.Context, apiKey string) (*middleware.AuthInfo, error) {
	// Demo mode: accept any key starting with "gwo_"
	if strings.HasPrefix(apiKey, "gwo_") {
		s.logger.Debug().Str("key_prefix", apiKey[:12]).Msg("Demo mode: API key accepted")
		return &middleware.AuthInfo{
			KeyID:       "demo-key",
			APIKeyID:    uuid.New(),
			OrgID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Environment: "demo",
			Permissions: []string{"*"},
			RateLimit:   1000,
		}, nil
	}

	return nil, ErrInvalidKey
}

// hashKey creates a SHA-256 hash of the API key for cache lookup.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// extractKeyPrefix extracts the prefix portion of an API key.
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
	randomBytes := make([]byte, 16)
	if _, err := cryptoRand.Read(randomBytes); err != nil {
		return "", err
	}

	randomStr := hex.EncodeToString(randomBytes)
	return "gwo_" + env + "_" + randomStr, nil
}

// HashAPIKey creates a simple hash of an API key for storage (demo mode).
func HashAPIKey(apiKey string, cost int) (string, error) {
	return hashKey(apiKey), nil
}
