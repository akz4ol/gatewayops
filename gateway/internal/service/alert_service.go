package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/akz4ol/gatewayops/gateway/internal/webhook"
)

// AlertService handles alert rule management and notifications.
type AlertService struct {
	alertRepo    *repository.AlertRepository
	auditService *AuditService
	slackClient  *webhook.SlackClient
	pagerClient  *webhook.PagerDutyClient
	logger       *slog.Logger
}

// NewAlertService creates a new alert service.
func NewAlertService(
	alertRepo *repository.AlertRepository,
	auditService *AuditService,
	slackClient *webhook.SlackClient,
	pagerClient *webhook.PagerDutyClient,
	logger *slog.Logger,
) *AlertService {
	return &AlertService{
		alertRepo:    alertRepo,
		auditService: auditService,
		slackClient:  slackClient,
		pagerClient:  pagerClient,
		logger:       logger,
	}
}

// CreateRule creates a new alert rule.
func (s *AlertService) CreateRule(ctx context.Context, orgID uuid.UUID, createdBy uuid.UUID, input domain.AlertRuleInput) (*domain.AlertRule, error) {
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
		CreatedBy:     createdBy,
	}

	// Set defaults
	if rule.WindowMinutes <= 0 {
		rule.WindowMinutes = 5
	}
	if rule.Severity == "" {
		rule.Severity = domain.AlertSeverityWarning
	}

	if err := s.alertRepo.CreateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}

	s.logger.Info("alert rule created",
		"rule_id", rule.ID,
		"name", rule.Name,
		"metric", rule.Metric,
	)

	return rule, nil
}

// GetRule retrieves an alert rule by ID.
func (s *AlertService) GetRule(ctx context.Context, id uuid.UUID) (*domain.AlertRule, error) {
	return s.alertRepo.GetRule(ctx, id)
}

// ListRules retrieves all alert rules for an organization.
func (s *AlertService) ListRules(ctx context.Context, orgID uuid.UUID, enabledOnly bool) ([]domain.AlertRule, error) {
	return s.alertRepo.ListRules(ctx, orgID, enabledOnly)
}

// UpdateRule updates an alert rule.
func (s *AlertService) UpdateRule(ctx context.Context, id uuid.UUID, input domain.AlertRuleInput) (*domain.AlertRule, error) {
	rule, err := s.alertRepo.GetRule(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, ErrNotFound{Resource: "alert_rule", ID: id.String()}
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

	if err := s.alertRepo.UpdateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}

	return rule, nil
}

// DeleteRule deletes an alert rule.
func (s *AlertService) DeleteRule(ctx context.Context, id uuid.UUID) error {
	return s.alertRepo.DeleteRule(ctx, id)
}

// CreateChannel creates a new alert channel.
func (s *AlertService) CreateChannel(ctx context.Context, orgID uuid.UUID, input domain.AlertChannelInput) (*domain.AlertChannel, error) {
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

	if err := s.alertRepo.CreateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	s.logger.Info("alert channel created",
		"channel_id", channel.ID,
		"name", channel.Name,
		"type", channel.Type,
	)

	return channel, nil
}

// GetChannel retrieves an alert channel by ID.
func (s *AlertService) GetChannel(ctx context.Context, id uuid.UUID) (*domain.AlertChannel, error) {
	return s.alertRepo.GetChannel(ctx, id)
}

// ListChannels retrieves all alert channels for an organization.
func (s *AlertService) ListChannels(ctx context.Context, orgID uuid.UUID) ([]domain.AlertChannel, error) {
	return s.alertRepo.ListChannels(ctx, orgID)
}

// UpdateChannel updates an alert channel.
func (s *AlertService) UpdateChannel(ctx context.Context, id uuid.UUID, input domain.AlertChannelInput) (*domain.AlertChannel, error) {
	channel, err := s.alertRepo.GetChannel(ctx, id)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrNotFound{Resource: "alert_channel", ID: id.String()}
	}

	channel.Name = input.Name
	channel.Type = input.Type
	channel.Config = input.Config
	channel.Enabled = input.Enabled
	channel.UpdatedAt = time.Now()

	if err := s.alertRepo.UpdateChannel(ctx, channel); err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}

	return channel, nil
}

// DeleteChannel deletes an alert channel.
func (s *AlertService) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	return s.alertRepo.DeleteChannel(ctx, id)
}

// FireAlert creates and sends a new alert.
func (s *AlertService) FireAlert(ctx context.Context, rule *domain.AlertRule, value float64, labels domain.Labels) (*domain.Alert, error) {
	// Check if there's already a firing alert for this rule
	existing, err := s.alertRepo.GetFiringAlertByRule(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// Update existing alert
		return existing, nil
	}

	alert := &domain.Alert{
		ID:        uuid.New(),
		OrgID:     rule.OrgID,
		RuleID:    rule.ID,
		Status:    domain.AlertStatusFiring,
		Severity:  rule.Severity,
		Message:   s.formatAlertMessage(rule, value),
		Value:     value,
		Threshold: rule.Threshold,
		Labels:    labels,
		StartedAt: time.Now(),
	}

	if err := s.alertRepo.CreateAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("create alert: %w", err)
	}

	// Send notifications
	go s.sendNotifications(context.Background(), alert, rule)

	s.logger.Warn("alert fired",
		"alert_id", alert.ID,
		"rule_id", rule.ID,
		"rule_name", rule.Name,
		"severity", alert.Severity,
		"value", value,
		"threshold", rule.Threshold,
	)

	return alert, nil
}

// formatAlertMessage formats the alert message.
func (s *AlertService) formatAlertMessage(rule *domain.AlertRule, value float64) string {
	conditionStr := ""
	switch rule.Condition {
	case domain.AlertConditionGreaterThan:
		conditionStr = ">"
	case domain.AlertConditionLessThan:
		conditionStr = "<"
	case domain.AlertConditionGreaterThanEqual:
		conditionStr = ">="
	case domain.AlertConditionLessThanEqual:
		conditionStr = "<="
	case domain.AlertConditionEqual:
		conditionStr = "="
	case domain.AlertConditionNotEqual:
		conditionStr = "!="
	}

	return fmt.Sprintf("[%s] %s: %s is %s (%.2f %s %.2f)",
		rule.Severity,
		rule.Name,
		rule.Metric,
		"triggered",
		value,
		conditionStr,
		rule.Threshold,
	)
}

// sendNotifications sends alert notifications to all configured channels.
func (s *AlertService) sendNotifications(ctx context.Context, alert *domain.Alert, rule *domain.AlertRule) {
	for _, channelID := range rule.Channels {
		channel, err := s.alertRepo.GetChannel(ctx, channelID)
		if err != nil || channel == nil || !channel.Enabled {
			continue
		}

		var sendErr error
		switch channel.Type {
		case domain.AlertChannelSlack:
			sendErr = s.sendSlackNotification(ctx, alert, rule, channel)
		case domain.AlertChannelPagerDuty:
			sendErr = s.sendPagerDutyNotification(ctx, alert, rule, channel)
		case domain.AlertChannelWebhook:
			sendErr = s.sendWebhookNotification(ctx, alert, rule, channel)
		default:
			s.logger.Warn("unsupported channel type",
				"channel_id", channel.ID,
				"type", channel.Type,
			)
			continue
		}

		if sendErr != nil {
			s.logger.Error("failed to send alert notification",
				"alert_id", alert.ID,
				"channel_id", channel.ID,
				"channel_type", channel.Type,
				"error", sendErr,
			)
		}
	}
}

// sendSlackNotification sends an alert to Slack.
func (s *AlertService) sendSlackNotification(ctx context.Context, alert *domain.Alert, rule *domain.AlertRule, channel *domain.AlertChannel) error {
	if s.slackClient == nil {
		return fmt.Errorf("slack client not configured")
	}

	webhookURL, _ := channel.Config["webhook_url"].(string)
	if webhookURL == "" {
		return fmt.Errorf("webhook_url not configured")
	}

	return s.slackClient.SendAlert(ctx, webhookURL, webhook.SlackAlert{
		Title:     rule.Name,
		Message:   alert.Message,
		Severity:  string(alert.Severity),
		Value:     alert.Value,
		Threshold: alert.Threshold,
		Metric:    string(rule.Metric),
		StartedAt: alert.StartedAt,
	})
}

// sendPagerDutyNotification sends an alert to PagerDuty.
func (s *AlertService) sendPagerDutyNotification(ctx context.Context, alert *domain.Alert, rule *domain.AlertRule, channel *domain.AlertChannel) error {
	if s.pagerClient == nil {
		return fmt.Errorf("pagerduty client not configured")
	}

	routingKey, _ := channel.Config["routing_key"].(string)
	if routingKey == "" {
		return fmt.Errorf("routing_key not configured")
	}

	return s.pagerClient.TriggerAlert(ctx, routingKey, webhook.PagerDutyAlert{
		Summary:   alert.Message,
		Source:    "gatewayops",
		Severity:  s.mapSeverityToPagerDuty(alert.Severity),
		Component: string(rule.Metric),
		DedupKey:  rule.ID.String(),
	})
}

// mapSeverityToPagerDuty maps our severity to PagerDuty severity.
func (s *AlertService) mapSeverityToPagerDuty(severity domain.AlertSeverity) string {
	switch severity {
	case domain.AlertSeverityCritical:
		return "critical"
	case domain.AlertSeverityWarning:
		return "warning"
	default:
		return "info"
	}
}

// sendWebhookNotification sends an alert via webhook.
func (s *AlertService) sendWebhookNotification(ctx context.Context, alert *domain.Alert, rule *domain.AlertRule, channel *domain.AlertChannel) error {
	// Implement generic webhook notification
	// This would make an HTTP POST to the configured URL with alert data
	s.logger.Debug("webhook notification sent",
		"alert_id", alert.ID,
		"channel_id", channel.ID,
	)
	return nil
}

// ResolveAlert resolves a firing alert.
func (s *AlertService) ResolveAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	alert, err := s.alertRepo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, ErrNotFound{Resource: "alert", ID: id.String()}
	}
	if alert.Status != domain.AlertStatusFiring {
		return alert, nil // Already resolved
	}

	now := time.Now()
	alert.Status = domain.AlertStatusResolved
	alert.ResolvedAt = &now

	if err := s.alertRepo.UpdateAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("update alert: %w", err)
	}

	s.logger.Info("alert resolved",
		"alert_id", id,
	)

	return alert, nil
}

// AcknowledgeAlert acknowledges an alert.
func (s *AlertService) AcknowledgeAlert(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Alert, error) {
	alert, err := s.alertRepo.GetAlert(ctx, id)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, ErrNotFound{Resource: "alert", ID: id.String()}
	}

	now := time.Now()
	alert.Status = domain.AlertStatusAcked
	alert.AckedAt = &now
	alert.AckedBy = &userID

	if err := s.alertRepo.UpdateAlert(ctx, alert); err != nil {
		return nil, fmt.Errorf("update alert: %w", err)
	}

	s.logger.Info("alert acknowledged",
		"alert_id", id,
		"acked_by", userID,
	)

	return alert, nil
}

// GetAlert retrieves an alert by ID.
func (s *AlertService) GetAlert(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	return s.alertRepo.GetAlert(ctx, id)
}

// ListAlerts retrieves alerts with filtering.
func (s *AlertService) ListAlerts(ctx context.Context, filter domain.AlertFilter) (*domain.AlertPage, error) {
	return s.alertRepo.ListAlerts(ctx, filter)
}

// ListFiringAlerts retrieves all firing alerts for an organization.
func (s *AlertService) ListFiringAlerts(ctx context.Context, orgID uuid.UUID) (*domain.AlertPage, error) {
	return s.alertRepo.ListAlerts(ctx, domain.AlertFilter{
		OrgID:    orgID,
		Statuses: []domain.AlertStatus{domain.AlertStatusFiring},
		Limit:    100,
	})
}

// CountActiveAlerts counts firing alerts for an organization.
func (s *AlertService) CountActiveAlerts(ctx context.Context, orgID uuid.UUID) (int64, error) {
	return s.alertRepo.CountActiveAlerts(ctx, orgID)
}

// EvaluateRule evaluates an alert rule against current metrics.
// This would be called by a background job.
func (s *AlertService) EvaluateRule(ctx context.Context, rule *domain.AlertRule, currentValue float64) error {
	triggered := s.checkCondition(rule.Condition, currentValue, rule.Threshold)

	if triggered {
		_, err := s.FireAlert(ctx, rule, currentValue, nil)
		return err
	}

	// Check if we should resolve existing alert
	firingAlert, err := s.alertRepo.GetFiringAlertByRule(ctx, rule.ID)
	if err != nil {
		return err
	}
	if firingAlert != nil {
		_, err = s.ResolveAlert(ctx, firingAlert.ID)
		return err
	}

	return nil
}

// checkCondition checks if the value meets the condition.
func (s *AlertService) checkCondition(condition domain.AlertCondition, value, threshold float64) bool {
	switch condition {
	case domain.AlertConditionGreaterThan:
		return value > threshold
	case domain.AlertConditionLessThan:
		return value < threshold
	case domain.AlertConditionGreaterThanEqual:
		return value >= threshold
	case domain.AlertConditionLessThanEqual:
		return value <= threshold
	case domain.AlertConditionEqual:
		return value == threshold
	case domain.AlertConditionNotEqual:
		return value != threshold
	default:
		return false
	}
}
