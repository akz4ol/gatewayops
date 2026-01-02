// Package alerting provides alert rule management and notification.
package alerting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service manages alert rules, channels, and notifications.
type Service struct {
	logger   zerolog.Logger
	repo     *repository.AlertRepository
	rules    map[uuid.UUID]*domain.AlertRule
	channels map[uuid.UUID]*domain.AlertChannel
	alerts   []domain.Alert
	mu       sync.RWMutex
	client   *http.Client

	// Simulated metrics for demo
	metrics map[string]float64
}

// NewService creates a new alerting service.
func NewService(logger zerolog.Logger, repo *repository.AlertRepository) *Service {
	s := &Service{
		logger:   logger,
		repo:     repo,
		rules:    make(map[uuid.UUID]*domain.AlertRule),
		channels: make(map[uuid.UUID]*domain.AlertChannel),
		alerts:   make([]domain.Alert, 0),
		client:   &http.Client{Timeout: 10 * time.Second},
		metrics:  make(map[string]float64),
	}

	// Load from database if available
	if repo != nil {
		s.loadFromDatabase()
	} else {
		// Create default demo data if no database
		s.createDemoChannel()
		s.createDemoRule()
	}

	// Initialize demo metrics
	s.initDemoMetrics()

	logger.Info().Msg("Alerting service initialized")
	return s
}

// loadFromDatabase loads rules and channels from the database.
func (s *Service) loadFromDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Load all rules (for demo org)
	demoOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	rules, err := s.repo.ListRules(ctx, demoOrgID, false)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to load alert rules from database")
	} else {
		for i := range rules {
			s.rules[rules[i].ID] = &rules[i]
		}
		s.logger.Info().Int("count", len(rules)).Msg("Loaded alert rules from database")
	}

	// Load all channels
	channels, err := s.repo.ListChannels(ctx, demoOrgID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to load alert channels from database")
	} else {
		for i := range channels {
			s.channels[channels[i].ID] = &channels[i]
		}
		s.logger.Info().Int("count", len(channels)).Msg("Loaded alert channels from database")
	}

	// If no data, create defaults
	if len(s.rules) == 0 && len(s.channels) == 0 {
		s.createDemoChannel()
		s.createDemoRule()
		// Persist defaults to database
		s.persistDefaults(ctx)
	}
}

// persistDefaults saves the default demo data to the database.
func (s *Service) persistDefaults(ctx context.Context) {
	if s.repo == nil {
		return
	}

	for _, channel := range s.channels {
		if err := s.repo.CreateChannel(ctx, channel); err != nil {
			s.logger.Warn().Err(err).Msg("Failed to persist default channel")
		}
	}

	for _, rule := range s.rules {
		if err := s.repo.CreateRule(ctx, rule); err != nil {
			s.logger.Warn().Err(err).Msg("Failed to persist default rule")
		}
	}
}

func (s *Service) createDemoChannel() {
	channel := &domain.AlertChannel{
		ID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OrgID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:  "Demo Slack Channel",
		Type:  domain.AlertChannelSlack,
		Config: map[string]interface{}{
			"webhook_url": "https://hooks.slack.com/services/DEMO/WEBHOOK/URL",
			"channel":     "#alerts",
			"username":    "GatewayOps Bot",
			"icon_emoji":  ":warning:",
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.channels[channel.ID] = channel
}

func (s *Service) createDemoRule() {
	rule := &domain.AlertRule{
		ID:            uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		OrgID:         uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:          "High Error Rate",
		Description:   "Alert when error rate exceeds 5%",
		Metric:        domain.AlertMetricErrorRate,
		Condition:     domain.AlertConditionGreaterThan,
		Threshold:     5.0,
		WindowMinutes: 5,
		Severity:      domain.AlertSeverityWarning,
		Channels:      []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-000000000001")},
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	s.rules[rule.ID] = rule
}

func (s *Service) initDemoMetrics() {
	s.metrics["error_rate"] = 0.12
	s.metrics["latency_p50"] = 45.0
	s.metrics["latency_p95"] = 127.0
	s.metrics["latency_p99"] = 250.0
	s.metrics["request_rate"] = 1500.0
	s.metrics["cost_per_hour"] = 175.0
	s.metrics["cost_per_day"] = 4200.0
}

// CreateRule creates a new alert rule.
func (s *Service) CreateRule(input domain.AlertRuleInput, orgID, userID uuid.UUID) *domain.AlertRule {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule := &domain.AlertRule{
		ID:            uuid.New(),
		OrgID:         orgID,
		Name:          input.Name,
		Description:   input.Description,
		Metric:        input.Metric,
		Condition:     input.Condition,
		Threshold:     input.Threshold,
		WindowMinutes: input.WindowMinutes,
		Severity:      input.Severity,
		Channels:      input.Channels,
		Filters:       input.Filters,
		Enabled:       input.Enabled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedBy:     userID,
	}

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateRule(ctx, rule); err != nil {
			s.logger.Error().Err(err).Msg("Failed to persist alert rule")
		}
	}

	s.rules[rule.ID] = rule

	s.logger.Info().
		Str("rule_id", rule.ID.String()).
		Str("name", rule.Name).
		Str("metric", string(rule.Metric)).
		Msg("Alert rule created")

	return rule
}

// GetRule returns a rule by ID.
func (s *Service) GetRule(id uuid.UUID) *domain.AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rules[id]
}

// ListRules returns all rules.
func (s *Service) ListRules() []domain.AlertRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]domain.AlertRule, 0, len(s.rules))
	for _, r := range s.rules {
		rules = append(rules, *r)
	}
	return rules
}

// UpdateRule updates an existing rule.
func (s *Service) UpdateRule(id uuid.UUID, input domain.AlertRuleInput) *domain.AlertRule {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil
	}

	rule.Name = input.Name
	rule.Description = input.Description
	rule.Metric = input.Metric
	rule.Condition = input.Condition
	rule.Threshold = input.Threshold
	rule.WindowMinutes = input.WindowMinutes
	rule.Severity = input.Severity
	rule.Channels = input.Channels
	rule.Filters = input.Filters
	rule.Enabled = input.Enabled
	rule.UpdatedAt = time.Now()

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.UpdateRule(ctx, rule); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update alert rule in database")
		}
	}

	return rule
}

// DeleteRule deletes a rule.
func (s *Service) DeleteRule(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rules[id]; exists {
		// Delete from database
		if s.repo != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.repo.DeleteRule(ctx, id); err != nil {
				s.logger.Error().Err(err).Msg("Failed to delete alert rule from database")
			}
		}
		delete(s.rules, id)
		return true
	}
	return false
}

// CreateChannel creates a new alert channel.
func (s *Service) CreateChannel(input domain.AlertChannelInput, orgID uuid.UUID) *domain.AlertChannel {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel := &domain.AlertChannel{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      input.Name,
		Type:      input.Type,
		Config:    input.Config,
		Enabled:   input.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateChannel(ctx, channel); err != nil {
			s.logger.Error().Err(err).Msg("Failed to persist alert channel")
		}
	}

	s.channels[channel.ID] = channel

	s.logger.Info().
		Str("channel_id", channel.ID.String()).
		Str("name", channel.Name).
		Str("type", string(channel.Type)).
		Msg("Alert channel created")

	return channel
}

// GetChannel returns a channel by ID.
func (s *Service) GetChannel(id uuid.UUID) *domain.AlertChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.channels[id]
}

// ListChannels returns all channels.
func (s *Service) ListChannels() []domain.AlertChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels := make([]domain.AlertChannel, 0, len(s.channels))
	for _, c := range s.channels {
		channels = append(channels, *c)
	}
	return channels
}

// UpdateChannel updates an existing channel.
func (s *Service) UpdateChannel(id uuid.UUID, input domain.AlertChannelInput) *domain.AlertChannel {
	s.mu.Lock()
	defer s.mu.Unlock()

	channel, exists := s.channels[id]
	if !exists {
		return nil
	}

	channel.Name = input.Name
	channel.Type = input.Type
	channel.Config = input.Config
	channel.Enabled = input.Enabled
	channel.UpdatedAt = time.Now()

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.UpdateChannel(ctx, channel); err != nil {
			s.logger.Error().Err(err).Msg("Failed to update alert channel in database")
		}
	}

	return channel
}

// DeleteChannel deletes a channel.
func (s *Service) DeleteChannel(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.channels[id]; exists {
		// Delete from database
		if s.repo != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.repo.DeleteChannel(ctx, id); err != nil {
				s.logger.Error().Err(err).Msg("Failed to delete alert channel from database")
			}
		}
		delete(s.channels, id)
		return true
	}
	return false
}

// TestChannel tests a channel by sending a test notification.
func (s *Service) TestChannel(id uuid.UUID) error {
	s.mu.RLock()
	channel, exists := s.channels[id]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel not found")
	}

	testAlert := domain.Alert{
		ID:        uuid.New(),
		Severity:  domain.AlertSeverityInfo,
		Status:    domain.AlertStatusFiring,
		Message:   "This is a test alert from GatewayOps",
		Value:     0,
		Threshold: 0,
		StartedAt: time.Now(),
	}

	return s.sendNotification(*channel, testAlert, "Test Alert Rule")
}

// CreateAlert creates a new alert and sends notifications.
func (s *Service) CreateAlert(ruleID uuid.UUID, value float64, message string) *domain.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[ruleID]
	if !exists {
		return nil
	}

	alert := domain.Alert{
		ID:        uuid.New(),
		OrgID:     rule.OrgID,
		RuleID:    ruleID,
		Status:    domain.AlertStatusFiring,
		Severity:  rule.Severity,
		Message:   message,
		Value:     value,
		Threshold: rule.Threshold,
		Labels: domain.Labels{
			"metric":    string(rule.Metric),
			"rule_name": rule.Name,
		},
		StartedAt: time.Now(),
	}

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateAlert(ctx, &alert); err != nil {
			s.logger.Error().Err(err).Msg("Failed to persist alert")
		}
	}

	// Keep only last 1000 alerts in memory
	if len(s.alerts) >= 1000 {
		s.alerts = s.alerts[1:]
	}
	s.alerts = append(s.alerts, alert)

	// Send notifications
	go s.notifyChannels(alert, *rule)

	s.logger.Warn().
		Str("alert_id", alert.ID.String()).
		Str("rule_id", ruleID.String()).
		Str("severity", string(alert.Severity)).
		Float64("value", value).
		Float64("threshold", rule.Threshold).
		Msg("Alert created")

	return &alert
}

// ResolveAlert resolves an existing alert.
func (s *Service) ResolveAlert(id uuid.UUID) *domain.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == id {
			now := time.Now()
			s.alerts[i].Status = domain.AlertStatusResolved
			s.alerts[i].ResolvedAt = &now

			// Persist to database
			if s.repo != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := s.repo.UpdateAlert(ctx, &s.alerts[i]); err != nil {
					s.logger.Error().Err(err).Msg("Failed to update resolved alert in database")
				}
			}

			return &s.alerts[i]
		}
	}
	return nil
}

// AcknowledgeAlert acknowledges an alert.
func (s *Service) AcknowledgeAlert(id uuid.UUID, userID uuid.UUID) *domain.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.alerts {
		if s.alerts[i].ID == id {
			now := time.Now()
			s.alerts[i].Status = domain.AlertStatusAcked
			s.alerts[i].AckedAt = &now
			s.alerts[i].AckedBy = &userID

			// Persist to database
			if s.repo != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := s.repo.UpdateAlert(ctx, &s.alerts[i]); err != nil {
					s.logger.Error().Err(err).Msg("Failed to update acknowledged alert in database")
				}
			}

			return &s.alerts[i]
		}
	}
	return nil
}

// GetAlerts returns alerts matching the filter.
func (s *Service) GetAlerts(filter domain.AlertFilter) domain.AlertPage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]domain.Alert, 0)
	for _, alert := range s.alerts {
		if !s.matchesAlertFilter(alert, filter) {
			continue
		}
		filtered = append(filtered, alert)
	}

	// Sort by most recent first
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	total := int64(len(filtered))
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset

	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return domain.AlertPage{
		Alerts:  filtered[start:end],
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: end < len(filtered),
	}
}

// GetActiveAlerts returns all currently firing alerts.
func (s *Service) GetActiveAlerts() []domain.Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]domain.Alert, 0)
	for _, alert := range s.alerts {
		if alert.Status == domain.AlertStatusFiring {
			active = append(active, alert)
		}
	}
	return active
}

// TriggerTestAlert creates a test alert for demo purposes.
func (s *Service) TriggerTestAlert(metric string, value float64) *domain.Alert {
	s.mu.RLock()
	// Find a rule that matches this metric
	var matchingRule *domain.AlertRule
	for _, rule := range s.rules {
		if string(rule.Metric) == metric && rule.Enabled {
			matchingRule = rule
			break
		}
	}
	s.mu.RUnlock()

	if matchingRule == nil {
		// Create alert without a rule
		s.mu.Lock()
		alert := domain.Alert{
			ID:       uuid.New(),
			OrgID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Status:   domain.AlertStatusFiring,
			Severity: domain.AlertSeverityWarning,
			Message:  fmt.Sprintf("Test alert: %s = %.2f", metric, value),
			Value:    value,
			Labels: domain.Labels{
				"metric": metric,
				"type":   "test",
			},
			StartedAt: time.Now(),
		}
		s.alerts = append(s.alerts, alert)
		s.mu.Unlock()
		return &alert
	}

	return s.CreateAlert(matchingRule.ID, value, fmt.Sprintf("%s exceeded threshold: %.2f > %.2f", matchingRule.Name, value, matchingRule.Threshold))
}

func (s *Service) matchesAlertFilter(alert domain.Alert, filter domain.AlertFilter) bool {
	if filter.RuleID != nil && alert.RuleID != *filter.RuleID {
		return false
	}

	if len(filter.Statuses) > 0 {
		found := false
		for _, status := range filter.Statuses {
			if alert.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.Severities) > 0 {
		found := false
		for _, sev := range filter.Severities {
			if alert.Severity == sev {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if filter.StartTime != nil && alert.StartedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && alert.StartedAt.After(*filter.EndTime) {
		return false
	}

	return true
}

func (s *Service) notifyChannels(alert domain.Alert, rule domain.AlertRule) {
	for _, channelID := range rule.Channels {
		s.mu.RLock()
		channel, exists := s.channels[channelID]
		s.mu.RUnlock()

		if !exists || !channel.Enabled {
			continue
		}

		if err := s.sendNotification(*channel, alert, rule.Name); err != nil {
			s.logger.Error().
				Err(err).
				Str("channel_id", channelID.String()).
				Str("channel_type", string(channel.Type)).
				Msg("Failed to send notification")
		}
	}
}

func (s *Service) sendNotification(channel domain.AlertChannel, alert domain.Alert, ruleName string) error {
	switch channel.Type {
	case domain.AlertChannelSlack:
		return s.sendSlackNotification(channel, alert, ruleName)
	case domain.AlertChannelPagerDuty:
		return s.sendPagerDutyNotification(channel, alert, ruleName)
	case domain.AlertChannelWebhook:
		return s.sendWebhookNotification(channel, alert, ruleName)
	default:
		s.logger.Debug().
			Str("channel_type", string(channel.Type)).
			Msg("Notification type not implemented - skipping (demo mode)")
		return nil
	}
}

func (s *Service) sendSlackNotification(channel domain.AlertChannel, alert domain.Alert, ruleName string) error {
	webhookURL, ok := channel.Config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("slack webhook_url not configured")
	}

	// In demo mode, just log the notification
	if webhookURL == "https://hooks.slack.com/services/DEMO/WEBHOOK/URL" {
		s.logger.Info().
			Str("channel", channel.Name).
			Str("alert_id", alert.ID.String()).
			Str("severity", string(alert.Severity)).
			Msg("Demo mode: Would send Slack notification")
		return nil
	}

	color := "#36a64f" // green
	switch alert.Severity {
	case domain.AlertSeverityWarning:
		color = "#ffcc00"
	case domain.AlertSeverityCritical:
		color = "#ff0000"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  fmt.Sprintf("[%s] %s", alert.Severity, ruleName),
				"text":   alert.Message,
				"fields": []map[string]interface{}{
					{"title": "Value", "value": fmt.Sprintf("%.2f", alert.Value), "short": true},
					{"title": "Threshold", "value": fmt.Sprintf("%.2f", alert.Threshold), "short": true},
					{"title": "Status", "value": string(alert.Status), "short": true},
				},
				"footer": "GatewayOps",
				"ts":     alert.StartedAt.Unix(),
			},
		},
	}

	return s.postJSON(webhookURL, payload)
}

func (s *Service) sendPagerDutyNotification(channel domain.AlertChannel, alert domain.Alert, ruleName string) error {
	routingKey, ok := channel.Config["routing_key"].(string)
	if !ok || routingKey == "" {
		return fmt.Errorf("pagerduty routing_key not configured")
	}

	severity := "warning"
	switch alert.Severity {
	case domain.AlertSeverityCritical:
		severity = "critical"
	case domain.AlertSeverityInfo:
		severity = "info"
	}

	payload := map[string]interface{}{
		"routing_key":  routingKey,
		"event_action": "trigger",
		"dedup_key":    alert.ID.String(),
		"payload": map[string]interface{}{
			"summary":   fmt.Sprintf("[GatewayOps] %s: %s", ruleName, alert.Message),
			"severity":  severity,
			"source":    "gatewayops",
			"timestamp": alert.StartedAt.Format(time.RFC3339),
			"custom_details": map[string]interface{}{
				"value":     alert.Value,
				"threshold": alert.Threshold,
			},
		},
	}

	return s.postJSON("https://events.pagerduty.com/v2/enqueue", payload)
}

func (s *Service) sendWebhookNotification(channel domain.AlertChannel, alert domain.Alert, ruleName string) error {
	webhookURL, ok := channel.Config["url"].(string)
	if !ok || webhookURL == "" {
		return fmt.Errorf("webhook url not configured")
	}

	payload := map[string]interface{}{
		"alert_id":   alert.ID.String(),
		"rule_name":  ruleName,
		"severity":   alert.Severity,
		"status":     alert.Status,
		"message":    alert.Message,
		"value":      alert.Value,
		"threshold":  alert.Threshold,
		"started_at": alert.StartedAt.Format(time.RFC3339),
	}

	return s.postJSON(webhookURL, payload)
}

func (s *Service) postJSON(url string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
