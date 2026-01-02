// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
)

// APIKeyRepository handles API key persistence.
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create inserts a new API key.
func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey, rawKey string) error {
	if r.db == nil {
		return nil
	}

	// Hash the raw key for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	permissions, err := json.Marshal(key.Permissions)
	if err != nil {
		permissions = []byte(`["*"]`)
	}

	query := `
		INSERT INTO api_keys (
			id, org_id, team_id, name, key_prefix, key_hash,
			environment, permissions, rate_limit, expires_at,
			created_at, created_by, revoked
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err = r.db.ExecContext(ctx, query,
		key.ID, key.OrgID, key.TeamID, key.Name, key.KeyPrefix, keyHash,
		key.Environment, permissions, key.RateLimit, key.ExpiresAt,
		key.CreatedAt, key.CreatedBy, key.Revoked,
	)
	if err != nil {
		return fmt.Errorf("insert api key: %w", err)
	}

	return nil
}

// Get retrieves an API key by ID.
func (r *APIKeyRepository) Get(ctx context.Context, orgID, id uuid.UUID) (*domain.APIKey, error) {
	if r.db == nil {
		return nil, nil
	}

	query := `
		SELECT id, org_id, team_id, name, key_prefix, environment,
			   permissions, rate_limit, expires_at, last_used_at,
			   created_at, created_by, revoked, revoked_at
		FROM api_keys
		WHERE id = $1 AND org_id = $2`

	var key domain.APIKey
	var teamID sql.NullString
	var permissions []byte
	var expiresAt, lastUsedAt, revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id, orgID).Scan(
		&key.ID, &key.OrgID, &teamID, &key.Name, &key.KeyPrefix, &key.Environment,
		&permissions, &key.RateLimit, &expiresAt, &lastUsedAt,
		&key.CreatedAt, &key.CreatedBy, &key.Revoked, &revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query api key: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		key.TeamID = &tid
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	if len(permissions) > 0 {
		json.Unmarshal(permissions, &key.Permissions)
	}

	return &key, nil
}

// GetByHash retrieves an API key by its hash (for authentication).
func (r *APIKeyRepository) GetByHash(ctx context.Context, rawKey string) (*domain.APIKey, error) {
	if r.db == nil {
		return nil, nil
	}

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	query := `
		SELECT id, org_id, team_id, name, key_prefix, environment,
			   permissions, rate_limit, expires_at, last_used_at,
			   created_at, created_by, revoked, revoked_at
		FROM api_keys
		WHERE key_hash = $1 AND revoked = false`

	var key domain.APIKey
	var teamID sql.NullString
	var permissions []byte
	var expiresAt, lastUsedAt, revokedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&key.ID, &key.OrgID, &teamID, &key.Name, &key.KeyPrefix, &key.Environment,
		&permissions, &key.RateLimit, &expiresAt, &lastUsedAt,
		&key.CreatedAt, &key.CreatedBy, &key.Revoked, &revokedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query api key by hash: %w", err)
	}

	if teamID.Valid {
		tid, _ := uuid.Parse(teamID.String)
		key.TeamID = &tid
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	if len(permissions) > 0 {
		json.Unmarshal(permissions, &key.Permissions)
	}

	return &key, nil
}

// List retrieves API keys with filtering and pagination.
func (r *APIKeyRepository) List(ctx context.Context, filter domain.APIKeyFilter) ([]domain.APIKey, int64, error) {
	if r.db == nil {
		return nil, 0, nil
	}

	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id = $%d", argNum))
		args = append(args, *filter.TeamID)
		argNum++
	}

	if filter.Environment != "" {
		conditions = append(conditions, fmt.Sprintf("environment = $%d", argNum))
		args = append(args, filter.Environment)
		argNum++
	}

	if !filter.IncludeRevoked {
		conditions = append(conditions, "revoked = false")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM api_keys WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count api keys: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, team_id, name, key_prefix, environment,
			   permissions, rate_limit, expires_at, last_used_at,
			   created_at, created_by, revoked, revoked_at
		FROM api_keys
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query api keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var key domain.APIKey
		var teamID sql.NullString
		var permissions []byte
		var expiresAt, lastUsedAt, revokedAt sql.NullTime

		err := rows.Scan(
			&key.ID, &key.OrgID, &teamID, &key.Name, &key.KeyPrefix, &key.Environment,
			&permissions, &key.RateLimit, &expiresAt, &lastUsedAt,
			&key.CreatedAt, &key.CreatedBy, &key.Revoked, &revokedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan api key: %w", err)
		}

		if teamID.Valid {
			tid, _ := uuid.Parse(teamID.String)
			key.TeamID = &tid
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		if revokedAt.Valid {
			key.RevokedAt = &revokedAt.Time
		}
		if len(permissions) > 0 {
			json.Unmarshal(permissions, &key.Permissions)
		}

		keys = append(keys, key)
	}

	return keys, total, rows.Err()
}

// Revoke marks an API key as revoked.
func (r *APIKeyRepository) Revoke(ctx context.Context, orgID, id uuid.UUID) error {
	if r.db == nil {
		return nil
	}

	query := `
		UPDATE api_keys
		SET revoked = true, revoked_at = NOW()
		WHERE id = $1 AND org_id = $2 AND revoked = false`

	result, err := r.db.ExecContext(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check revoke result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("api key not found or already revoked")
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp.
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return nil
	}

	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update last used: %w", err)
	}

	return nil
}

// GetUsage retrieves usage statistics for an API key.
func (r *APIKeyRepository) GetUsage(ctx context.Context, keyID uuid.UUID) (*domain.APIKeyUsage, error) {
	if r.db == nil {
		return &domain.APIKeyUsage{KeyID: keyID}, nil
	}

	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(cost), 0) as total_cost,
			MAX(created_at) as last_used_at
		FROM traces
		WHERE api_key_id = $1`

	var usage domain.APIKeyUsage
	var lastUsed sql.NullTime

	err := r.db.QueryRowContext(ctx, query, keyID).Scan(
		&usage.TotalRequests,
		&usage.TotalCost,
		&lastUsed,
	)
	if err != nil {
		return nil, fmt.Errorf("query api key usage: %w", err)
	}

	usage.KeyID = keyID
	if lastUsed.Valid {
		usage.LastUsedAt = lastUsed.Time
	}

	return &usage, nil
}
