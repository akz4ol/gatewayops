package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
)

// AlertRepository handles alert rules, channels, and alerts persistence.
type AlertRepository struct {
	db *sql.DB
}

// NewAlertRepository creates a new alert repository.
func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// CreateRule inserts a new alert rule.
func (r *AlertRepository) CreateRule(ctx context.Context, rule *domain.AlertRule) error {
	channels, _ := json.Marshal(rule.Channels)
	filters, _ := json.Marshal(rule.Filters)

	query := `
		INSERT INTO alert_rules (
			id, org_id, name, description, metric, condition,
			threshold, window_minutes, severity, channels, filters,
			enabled, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.OrgID, rule.Name, rule.Description, rule.Metric, rule.Condition,
		rule.Threshold, rule.WindowMinutes, rule.Severity, channels, filters,
		rule.Enabled, rule.CreatedAt, rule.UpdatedAt, rule.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert alert rule: %w", err)
	}

	return nil
}

// GetRule retrieves an alert rule by ID.
func (r *AlertRepository) GetRule(ctx context.Context, id uuid.UUID) (*domain.AlertRule, error) {
	query := `
		SELECT id, org_id, name, description, metric, condition,
			   threshold, window_minutes, severity, channels, filters,
			   enabled, created_at, updated_at, created_by
		FROM alert_rules
		WHERE id = $1`

	var rule domain.AlertRule
	var channels, filters []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.OrgID, &rule.Name, &rule.Description, &rule.Metric, &rule.Condition,
		&rule.Threshold, &rule.WindowMinutes, &rule.Severity, &channels, &filters,
		&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query alert rule: %w", err)
	}

	json.Unmarshal(channels, &rule.Channels)
	json.Unmarshal(filters, &rule.Filters)

	return &rule, nil
}

// ListRules retrieves all alert rules for an organization.
func (r *AlertRepository) ListRules(ctx context.Context, orgID uuid.UUID, enabledOnly bool) ([]domain.AlertRule, error) {
	var query string
	var args []interface{}

	if enabledOnly {
		query = `
			SELECT id, org_id, name, description, metric, condition,
				   threshold, window_minutes, severity, channels, filters,
				   enabled, created_at, updated_at, created_by
			FROM alert_rules
			WHERE org_id = $1 AND enabled = true
			ORDER BY created_at DESC`
		args = []interface{}{orgID}
	} else {
		query = `
			SELECT id, org_id, name, description, metric, condition,
				   threshold, window_minutes, severity, channels, filters,
				   enabled, created_at, updated_at, created_by
			FROM alert_rules
			WHERE org_id = $1
			ORDER BY created_at DESC`
		args = []interface{}{orgID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.AlertRule
	for rows.Next() {
		var rule domain.AlertRule
		var channels, filters []byte

		err := rows.Scan(
			&rule.ID, &rule.OrgID, &rule.Name, &rule.Description, &rule.Metric, &rule.Condition,
			&rule.Threshold, &rule.WindowMinutes, &rule.Severity, &channels, &filters,
			&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt, &rule.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan alert rule: %w", err)
		}

		json.Unmarshal(channels, &rule.Channels)
		json.Unmarshal(filters, &rule.Filters)

		rules = append(rules, rule)
	}

	return rules, nil
}

// UpdateRule updates an alert rule.
func (r *AlertRepository) UpdateRule(ctx context.Context, rule *domain.AlertRule) error {
	channels, _ := json.Marshal(rule.Channels)
	filters, _ := json.Marshal(rule.Filters)

	query := `
		UPDATE alert_rules SET
			name = $2, description = $3, metric = $4, condition = $5,
			threshold = $6, window_minutes = $7, severity = $8, channels = $9,
			filters = $10, enabled = $11, updated_at = $12
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.Name, rule.Description, rule.Metric, rule.Condition,
		rule.Threshold, rule.WindowMinutes, rule.Severity, channels,
		filters, rule.Enabled, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update alert rule: %w", err)
	}

	return nil
}

// DeleteRule deletes an alert rule.
func (r *AlertRepository) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM alert_rules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}

	return nil
}

// CreateChannel inserts a new alert channel.
func (r *AlertRepository) CreateChannel(ctx context.Context, channel *domain.AlertChannel) error {
	config, _ := json.Marshal(channel.Config)

	query := `
		INSERT INTO alert_channels (
			id, org_id, name, type, config, enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		channel.ID, channel.OrgID, channel.Name, channel.Type,
		config, channel.Enabled, channel.CreatedAt, channel.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert alert channel: %w", err)
	}

	return nil
}

// GetChannel retrieves an alert channel by ID.
func (r *AlertRepository) GetChannel(ctx context.Context, id uuid.UUID) (*domain.AlertChannel, error) {
	query := `
		SELECT id, org_id, name, type, config, enabled, created_at, updated_at
		FROM alert_channels
		WHERE id = $1`

	var channel domain.AlertChannel
	var config []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID, &channel.OrgID, &channel.Name, &channel.Type,
		&config, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query alert channel: %w", err)
	}

	json.Unmarshal(config, &channel.Config)

	return &channel, nil
}

// ListChannels retrieves all alert channels for an organization.
func (r *AlertRepository) ListChannels(ctx context.Context, orgID uuid.UUID) ([]domain.AlertChannel, error) {
	query := `
		SELECT id, org_id, name, type, config, enabled, created_at, updated_at
		FROM alert_channels
		WHERE org_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query alert channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.AlertChannel
	for rows.Next() {
		var channel domain.AlertChannel
		var config []byte

		err := rows.Scan(
			&channel.ID, &channel.OrgID, &channel.Name, &channel.Type,
			&config, &channel.Enabled, &channel.CreatedAt, &channel.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan alert channel: %w", err)
		}

		json.Unmarshal(config, &channel.Config)
		channels = append(channels, channel)
	}

	return channels, nil
}

// UpdateChannel updates an alert channel.
func (r *AlertRepository) UpdateChannel(ctx context.Context, channel *domain.AlertChannel) error {
	config, _ := json.Marshal(channel.Config)

	query := `
		UPDATE alert_channels SET
			name = $2, type = $3, config = $4, enabled = $5, updated_at = $6
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		channel.ID, channel.Name, channel.Type, config, channel.Enabled, channel.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update alert channel: %w", err)
	}

	return nil
}

// DeleteChannel deletes an alert channel.
func (r *AlertRepository) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM alert_channels WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete alert channel: %w", err)
	}

	return nil
}

// CreateAlert inserts a new alert.
func (r *AlertRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	labels, _ := json.Marshal(alert.Labels)

	query := `
		INSERT INTO alerts (
			id, org_id, rule_id, status, severity, message,
			value, threshold, labels, started_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.OrgID, alert.RuleID, alert.Status, alert.Severity,
		alert.Message, alert.Value, alert.Threshold, labels, alert.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}

	return nil
}

// GetAlert retrieves an alert by ID.
func (r *AlertRepository) GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, org_id, rule_id, status, severity, message,
			   value, threshold, labels, started_at, resolved_at, acked_at, acked_by
		FROM alerts
		WHERE id = $1`

	var alert domain.Alert
	var labels []byte
	var resolvedAt, ackedAt sql.NullTime
	var ackedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&alert.ID, &alert.OrgID, &alert.RuleID, &alert.Status, &alert.Severity,
		&alert.Message, &alert.Value, &alert.Threshold, &labels,
		&alert.StartedAt, &resolvedAt, &ackedAt, &ackedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query alert: %w", err)
	}

	json.Unmarshal(labels, &alert.Labels)

	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}
	if ackedAt.Valid {
		alert.AckedAt = &ackedAt.Time
	}
	if ackedBy.Valid {
		aid, _ := uuid.Parse(ackedBy.String)
		alert.AckedBy = &aid
	}

	return &alert, nil
}

// UpdateAlert updates an alert (for resolving or acknowledging).
func (r *AlertRepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	query := `
		UPDATE alerts SET
			status = $2, resolved_at = $3, acked_at = $4, acked_by = $5
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		alert.ID, alert.Status, alert.ResolvedAt, alert.AckedAt, alert.AckedBy,
	)
	if err != nil {
		return fmt.Errorf("update alert: %w", err)
	}

	return nil
}

// ListAlerts retrieves alerts with filtering.
func (r *AlertRepository) ListAlerts(ctx context.Context, filter domain.AlertFilter) (*domain.AlertPage, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
	args = append(args, filter.OrgID)
	argNum++

	if filter.RuleID != nil {
		conditions = append(conditions, fmt.Sprintf("rule_id = $%d", argNum))
		args = append(args, *filter.RuleID)
		argNum++
	}

	if len(filter.Statuses) > 0 {
		placeholders := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, status)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Severities) > 0 {
		placeholders := make([]string, len(filter.Severities))
		for i, severity := range filter.Severities {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, severity)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("severity IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.StartTime != nil {
		conditions = append(conditions, fmt.Sprintf("started_at >= $%d", argNum))
		args = append(args, *filter.StartTime)
		argNum++
	}

	if filter.EndTime != nil {
		conditions = append(conditions, fmt.Sprintf("started_at <= $%d", argNum))
		args = append(args, *filter.EndTime)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM alerts WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count alerts: %w", err)
	}

	// Apply pagination
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, rule_id, status, severity, message,
			   value, threshold, labels, started_at, resolved_at, acked_at, acked_by
		FROM alerts
		WHERE %s
		ORDER BY started_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var labels []byte
		var resolvedAt, ackedAt sql.NullTime
		var ackedBy sql.NullString

		err := rows.Scan(
			&alert.ID, &alert.OrgID, &alert.RuleID, &alert.Status, &alert.Severity,
			&alert.Message, &alert.Value, &alert.Threshold, &labels,
			&alert.StartedAt, &resolvedAt, &ackedAt, &ackedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}

		json.Unmarshal(labels, &alert.Labels)

		if resolvedAt.Valid {
			alert.ResolvedAt = &resolvedAt.Time
		}
		if ackedAt.Valid {
			alert.AckedAt = &ackedAt.Time
		}
		if ackedBy.Valid {
			aid, _ := uuid.Parse(ackedBy.String)
			alert.AckedBy = &aid
		}

		alerts = append(alerts, alert)
	}

	return &domain.AlertPage{
		Alerts:  alerts,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(offset+len(alerts)) < total,
	}, nil
}

// GetFiringAlertByRule retrieves a firing alert for a specific rule.
func (r *AlertRepository) GetFiringAlertByRule(ctx context.Context, ruleID uuid.UUID) (*domain.Alert, error) {
	query := `
		SELECT id, org_id, rule_id, status, severity, message,
			   value, threshold, labels, started_at, resolved_at, acked_at, acked_by
		FROM alerts
		WHERE rule_id = $1 AND status = 'firing'
		ORDER BY started_at DESC
		LIMIT 1`

	var alert domain.Alert
	var labels []byte
	var resolvedAt, ackedAt sql.NullTime
	var ackedBy sql.NullString

	err := r.db.QueryRowContext(ctx, query, ruleID).Scan(
		&alert.ID, &alert.OrgID, &alert.RuleID, &alert.Status, &alert.Severity,
		&alert.Message, &alert.Value, &alert.Threshold, &labels,
		&alert.StartedAt, &resolvedAt, &ackedAt, &ackedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query firing alert: %w", err)
	}

	json.Unmarshal(labels, &alert.Labels)

	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}
	if ackedAt.Valid {
		alert.AckedAt = &ackedAt.Time
	}
	if ackedBy.Valid {
		aid, _ := uuid.Parse(ackedBy.String)
		alert.AckedBy = &aid
	}

	return &alert, nil
}

// CountActiveAlerts counts firing alerts for an organization.
func (r *AlertRepository) CountActiveAlerts(ctx context.Context, orgID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM alerts WHERE org_id = $1 AND status = 'firing'",
		orgID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active alerts: %w", err)
	}

	return count, nil
}
