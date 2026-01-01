package domain

import (
	"time"

	"github.com/google/uuid"
)

// AlertSeverity represents the severity level of an alert.
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus represents the status of an alert.
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusAcked    AlertStatus = "acknowledged"
)

// AlertMetric represents the metric type for alert rules.
type AlertMetric string

const (
	AlertMetricErrorRate    AlertMetric = "error_rate"
	AlertMetricLatencyP50   AlertMetric = "latency_p50"
	AlertMetricLatencyP95   AlertMetric = "latency_p95"
	AlertMetricLatencyP99   AlertMetric = "latency_p99"
	AlertMetricRequestRate  AlertMetric = "request_rate"
	AlertMetricCostPerHour  AlertMetric = "cost_per_hour"
	AlertMetricCostPerDay   AlertMetric = "cost_per_day"
	AlertMetricRateLimitHit AlertMetric = "rate_limit_hit"
	AlertMetricInjectionDetected AlertMetric = "injection_detected"
)

// AlertCondition represents the comparison condition.
type AlertCondition string

const (
	AlertConditionGreaterThan      AlertCondition = "gt"
	AlertConditionLessThan         AlertCondition = "lt"
	AlertConditionGreaterThanEqual AlertCondition = "gte"
	AlertConditionLessThanEqual    AlertCondition = "lte"
	AlertConditionEqual            AlertCondition = "eq"
	AlertConditionNotEqual         AlertCondition = "neq"
)

// AlertRule represents a rule for triggering alerts.
type AlertRule struct {
	ID            uuid.UUID      `json:"id"`
	OrgID         uuid.UUID      `json:"org_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Metric        AlertMetric    `json:"metric"`
	Condition     AlertCondition `json:"condition"`
	Threshold     float64        `json:"threshold"`
	WindowMinutes int            `json:"window_minutes"`
	Severity      AlertSeverity  `json:"severity"`
	Channels      []uuid.UUID    `json:"channels"` // Alert channel IDs
	Filters       AlertFilters   `json:"filters,omitempty"`
	Enabled       bool           `json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CreatedBy     uuid.UUID      `json:"created_by"`
}

// AlertFilters defines optional filters for alert rules.
type AlertFilters struct {
	MCPServers  []string   `json:"mcp_servers,omitempty"`
	Teams       []uuid.UUID `json:"teams,omitempty"`
	Environments []string   `json:"environments,omitempty"`
}

// AlertRuleInput represents input for creating/updating an alert rule.
type AlertRuleInput struct {
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Metric        AlertMetric    `json:"metric"`
	Condition     AlertCondition `json:"condition"`
	Threshold     float64        `json:"threshold"`
	WindowMinutes int            `json:"window_minutes"`
	Severity      AlertSeverity  `json:"severity"`
	Channels      []uuid.UUID    `json:"channels"`
	Filters       AlertFilters   `json:"filters,omitempty"`
	Enabled       bool           `json:"enabled"`
}

// AlertChannelType represents the type of alert channel.
type AlertChannelType string

const (
	AlertChannelSlack     AlertChannelType = "slack"
	AlertChannelPagerDuty AlertChannelType = "pagerduty"
	AlertChannelOpsgenie  AlertChannelType = "opsgenie"
	AlertChannelWebhook   AlertChannelType = "webhook"
	AlertChannelEmail     AlertChannelType = "email"
	AlertChannelTeams     AlertChannelType = "teams"
)

// AlertChannel represents a notification channel for alerts.
type AlertChannel struct {
	ID        uuid.UUID              `json:"id"`
	OrgID     uuid.UUID              `json:"org_id"`
	Name      string                 `json:"name"`
	Type      AlertChannelType       `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// AlertChannelInput represents input for creating/updating an alert channel.
type AlertChannelInput struct {
	Name    string                 `json:"name"`
	Type    AlertChannelType       `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Enabled bool                   `json:"enabled"`
}

// SlackChannelConfig represents Slack-specific channel configuration.
type SlackChannelConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel,omitempty"`
	Username   string `json:"username,omitempty"`
	IconEmoji  string `json:"icon_emoji,omitempty"`
}

// PagerDutyChannelConfig represents PagerDuty-specific channel configuration.
type PagerDutyChannelConfig struct {
	RoutingKey string `json:"routing_key"`
	ServiceID  string `json:"service_id,omitempty"`
}

// WebhookChannelConfig represents webhook-specific channel configuration.
type WebhookChannelConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"` // POST by default
	Headers map[string]string `json:"headers,omitempty"`
}

// Alert represents an active or historical alert.
type Alert struct {
	ID         uuid.UUID     `json:"id"`
	OrgID      uuid.UUID     `json:"org_id"`
	RuleID     uuid.UUID     `json:"rule_id"`
	Status     AlertStatus   `json:"status"`
	Severity   AlertSeverity `json:"severity"`
	Message    string        `json:"message"`
	Value      float64       `json:"value"` // The actual metric value that triggered the alert
	Threshold  float64       `json:"threshold"`
	Labels     Labels        `json:"labels,omitempty"`
	StartedAt  time.Time     `json:"started_at"`
	ResolvedAt *time.Time    `json:"resolved_at,omitempty"`
	AckedAt    *time.Time    `json:"acked_at,omitempty"`
	AckedBy    *uuid.UUID    `json:"acked_by,omitempty"`
}

// Labels represents key-value labels for an alert.
type Labels map[string]string

// AlertFilter defines filters for querying alerts.
type AlertFilter struct {
	OrgID      uuid.UUID       `json:"org_id"`
	RuleID     *uuid.UUID      `json:"rule_id,omitempty"`
	Statuses   []AlertStatus   `json:"statuses,omitempty"`
	Severities []AlertSeverity `json:"severities,omitempty"`
	StartTime  *time.Time      `json:"start_time,omitempty"`
	EndTime    *time.Time      `json:"end_time,omitempty"`
	Limit      int             `json:"limit,omitempty"`
	Offset     int             `json:"offset,omitempty"`
}

// AlertPage represents a paginated list of alerts.
type AlertPage struct {
	Alerts  []Alert `json:"alerts"`
	Total   int64   `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	HasMore bool    `json:"has_more"`
}

// AlertNotification represents a notification to be sent.
type AlertNotification struct {
	Alert   Alert        `json:"alert"`
	Rule    AlertRule    `json:"rule"`
	Channel AlertChannel `json:"channel"`
}
