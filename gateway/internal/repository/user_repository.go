package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
)

// UserRepository handles user and SSO provider persistence.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user.
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, org_id, email, name, avatar_url, status,
			sso_provider_id, sso_external_id, last_login_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.OrgID, user.Email, user.Name, user.AvatarURL, user.Status,
		user.SSOProviderID, user.SSOExternalID, user.LastLoginAt, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID.
func (r *UserRepository) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, org_id, email, name, avatar_url, status,
			   sso_provider_id, sso_external_id, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1`

	var user domain.User
	var ssoProviderID sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Name, &user.AvatarURL, &user.Status,
		&ssoProviderID, &user.SSOExternalID, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	if ssoProviderID.Valid {
		pid, _ := uuid.Parse(ssoProviderID.String)
		user.SSOProviderID = &pid
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email within an organization.
func (r *UserRepository) GetUserByEmail(ctx context.Context, orgID uuid.UUID, email string) (*domain.User, error) {
	query := `
		SELECT id, org_id, email, name, avatar_url, status,
			   sso_provider_id, sso_external_id, last_login_at, created_at, updated_at
		FROM users
		WHERE org_id = $1 AND email = $2`

	var user domain.User
	var ssoProviderID sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orgID, email).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Name, &user.AvatarURL, &user.Status,
		&ssoProviderID, &user.SSOExternalID, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	if ssoProviderID.Valid {
		pid, _ := uuid.Parse(ssoProviderID.String)
		user.SSOProviderID = &pid
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// GetUserBySSOExternalID retrieves a user by SSO external ID.
func (r *UserRepository) GetUserBySSOExternalID(ctx context.Context, providerID uuid.UUID, externalID string) (*domain.User, error) {
	query := `
		SELECT id, org_id, email, name, avatar_url, status,
			   sso_provider_id, sso_external_id, last_login_at, created_at, updated_at
		FROM users
		WHERE sso_provider_id = $1 AND sso_external_id = $2`

	var user domain.User
	var ssoProviderID sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, providerID, externalID).Scan(
		&user.ID, &user.OrgID, &user.Email, &user.Name, &user.AvatarURL, &user.Status,
		&ssoProviderID, &user.SSOExternalID, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by SSO external ID: %w", err)
	}

	if ssoProviderID.Valid {
		pid, _ := uuid.Parse(ssoProviderID.String)
		user.SSOProviderID = &pid
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// UpdateUser updates an existing user.
func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			email = $2, name = $3, avatar_url = $4, status = $5,
			last_login_at = $6, updated_at = $7
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.AvatarURL, user.Status,
		user.LastLoginAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// ListUsersByOrg retrieves all users in an organization.
func (r *UserRepository) ListUsersByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.User, int64, error) {
	// Count total
	var total int64
	if err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE org_id = $1", orgID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, org_id, email, name, avatar_url, status,
			   sso_provider_id, sso_external_id, last_login_at, created_at, updated_at
		FROM users
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		var ssoProviderID sql.NullString
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.OrgID, &user.Email, &user.Name, &user.AvatarURL, &user.Status,
			&ssoProviderID, &user.SSOExternalID, &lastLoginAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		if ssoProviderID.Valid {
			pid, _ := uuid.Parse(ssoProviderID.String)
			user.SSOProviderID = &pid
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	return users, total, nil
}

// CreateSession creates a new user session.
func (r *UserRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			id, user_id, org_id, access_token, refresh_token,
			expires_at, last_activity_at, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.OrgID, session.AccessToken, session.RefreshToken,
		session.ExpiresAt, session.LastActivityAt, session.IPAddress, session.UserAgent, session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID.
func (r *UserRepository) GetSession(ctx context.Context, id uuid.UUID) (*domain.UserSession, error) {
	query := `
		SELECT id, user_id, org_id, access_token, refresh_token,
			   expires_at, last_activity_at, ip_address, user_agent, created_at
		FROM user_sessions
		WHERE id = $1 AND expires_at > NOW()`

	var session domain.UserSession
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.OrgID, &session.AccessToken, &session.RefreshToken,
		&session.ExpiresAt, &session.LastActivityAt, &session.IPAddress, &session.UserAgent, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query session: %w", err)
	}

	return &session, nil
}

// UpdateSessionActivity updates the last activity time of a session.
func (r *UserRepository) UpdateSessionActivity(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE user_sessions SET last_activity_at = $2 WHERE id = $1",
		id, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update session activity: %w", err)
	}

	return nil
}

// DeleteSession deletes a session (logout).
func (r *UserRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM user_sessions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions removes all expired sessions.
func (r *UserRepository) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM user_sessions WHERE expires_at < NOW()")
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return count, nil
}

// CreateSSOProvider creates a new SSO provider.
func (r *UserRepository) CreateSSOProvider(ctx context.Context, provider *domain.SSOProvider) error {
	scopes, _ := json.Marshal(provider.Scopes)
	claimMappings, _ := json.Marshal(provider.ClaimMappings)
	groupMappings, _ := json.Marshal(provider.GroupMappings)

	query := `
		INSERT INTO sso_providers (
			id, org_id, type, name, issuer_url, client_id, client_secret_encrypted,
			authorization_url, token_url, userinfo_url, scopes, claim_mappings,
			group_mappings, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := r.db.ExecContext(ctx, query,
		provider.ID, provider.OrgID, provider.Type, provider.Name, provider.IssuerURL,
		provider.ClientID, provider.ClientSecretEncrypted, provider.AuthorizationURL,
		provider.TokenURL, provider.UserInfoURL, scopes, claimMappings,
		groupMappings, provider.Enabled, provider.CreatedAt, provider.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert SSO provider: %w", err)
	}

	return nil
}

// GetSSOProvider retrieves an SSO provider by ID.
func (r *UserRepository) GetSSOProvider(ctx context.Context, id uuid.UUID) (*domain.SSOProvider, error) {
	query := `
		SELECT id, org_id, type, name, issuer_url, client_id, client_secret_encrypted,
			   authorization_url, token_url, userinfo_url, scopes, claim_mappings,
			   group_mappings, enabled, created_at, updated_at
		FROM sso_providers
		WHERE id = $1`

	var provider domain.SSOProvider
	var scopes, claimMappings, groupMappings []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&provider.ID, &provider.OrgID, &provider.Type, &provider.Name, &provider.IssuerURL,
		&provider.ClientID, &provider.ClientSecretEncrypted, &provider.AuthorizationURL,
		&provider.TokenURL, &provider.UserInfoURL, &scopes, &claimMappings,
		&groupMappings, &provider.Enabled, &provider.CreatedAt, &provider.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query SSO provider: %w", err)
	}

	json.Unmarshal(scopes, &provider.Scopes)
	json.Unmarshal(claimMappings, &provider.ClaimMappings)
	json.Unmarshal(groupMappings, &provider.GroupMappings)

	return &provider, nil
}

// ListSSOProviders retrieves all SSO providers for an organization.
func (r *UserRepository) ListSSOProviders(ctx context.Context, orgID uuid.UUID) ([]domain.SSOProvider, error) {
	query := `
		SELECT id, org_id, type, name, issuer_url, client_id,
			   authorization_url, token_url, userinfo_url, scopes, claim_mappings,
			   group_mappings, enabled, created_at, updated_at
		FROM sso_providers
		WHERE org_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query SSO providers: %w", err)
	}
	defer rows.Close()

	var providers []domain.SSOProvider
	for rows.Next() {
		var provider domain.SSOProvider
		var scopes, claimMappings, groupMappings []byte

		err := rows.Scan(
			&provider.ID, &provider.OrgID, &provider.Type, &provider.Name, &provider.IssuerURL,
			&provider.ClientID, &provider.AuthorizationURL, &provider.TokenURL, &provider.UserInfoURL,
			&scopes, &claimMappings, &groupMappings, &provider.Enabled, &provider.CreatedAt, &provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan SSO provider: %w", err)
		}

		json.Unmarshal(scopes, &provider.Scopes)
		json.Unmarshal(claimMappings, &provider.ClaimMappings)
		json.Unmarshal(groupMappings, &provider.GroupMappings)

		providers = append(providers, provider)
	}

	return providers, nil
}

// UpdateSSOProvider updates an SSO provider.
func (r *UserRepository) UpdateSSOProvider(ctx context.Context, provider *domain.SSOProvider) error {
	scopes, _ := json.Marshal(provider.Scopes)
	claimMappings, _ := json.Marshal(provider.ClaimMappings)
	groupMappings, _ := json.Marshal(provider.GroupMappings)

	query := `
		UPDATE sso_providers SET
			name = $2, issuer_url = $3, client_id = $4, client_secret_encrypted = $5,
			authorization_url = $6, token_url = $7, userinfo_url = $8, scopes = $9,
			claim_mappings = $10, group_mappings = $11, enabled = $12, updated_at = $13
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		provider.ID, provider.Name, provider.IssuerURL, provider.ClientID,
		provider.ClientSecretEncrypted, provider.AuthorizationURL, provider.TokenURL,
		provider.UserInfoURL, scopes, claimMappings, groupMappings,
		provider.Enabled, provider.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update SSO provider: %w", err)
	}

	return nil
}

// DeleteSSOProvider deletes an SSO provider.
func (r *UserRepository) DeleteSSOProvider(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sso_providers WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete SSO provider: %w", err)
	}

	return nil
}
