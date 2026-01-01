// Package webhook provides webhook clients for alert notifications.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackClient handles Slack webhook notifications.
type SlackClient struct {
	httpClient *http.Client
}

// NewSlackClient creates a new Slack webhook client.
func NewSlackClient() *SlackClient {
	return &SlackClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SlackAlert represents an alert to send to Slack.
type SlackAlert struct {
	Title     string
	Message   string
	Severity  string
	Value     float64
	Threshold float64
	Metric    string
	StartedAt time.Time
}

// SlackMessage represents a Slack webhook message.
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment.
type SlackAttachment struct {
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitempty"`
	Text       string       `json:"text,omitempty"`
	Fields     []SlackField `json:"fields,omitempty"`
	Footer     string       `json:"footer,omitempty"`
	FooterIcon string       `json:"footer_icon,omitempty"`
	Ts         int64        `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment.
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

// SendAlert sends an alert notification to Slack.
func (c *SlackClient) SendAlert(ctx context.Context, webhookURL string, alert SlackAlert) error {
	color := c.getSeverityColor(alert.Severity)

	message := SlackMessage{
		Username:  "GatewayOps Alerts",
		IconEmoji: ":warning:",
		Attachments: []SlackAttachment{
			{
				Color: color,
				Title: fmt.Sprintf("[%s] %s", alert.Severity, alert.Title),
				Text:  alert.Message,
				Fields: []SlackField{
					{
						Title: "Metric",
						Value: alert.Metric,
						Short: true,
					},
					{
						Title: "Current Value",
						Value: fmt.Sprintf("%.2f", alert.Value),
						Short: true,
					},
					{
						Title: "Threshold",
						Value: fmt.Sprintf("%.2f", alert.Threshold),
						Short: true,
					},
					{
						Title: "Severity",
						Value: alert.Severity,
						Short: true,
					},
				},
				Footer:     "GatewayOps",
				FooterIcon: "https://gatewayops.com/favicon.ico",
				Ts:         alert.StartedAt.Unix(),
			},
		},
	}

	return c.sendWebhook(ctx, webhookURL, message)
}

// SendResolution sends a resolution notification to Slack.
func (c *SlackClient) SendResolution(ctx context.Context, webhookURL string, title, message string, resolvedAt time.Time) error {
	slackMessage := SlackMessage{
		Username:  "GatewayOps Alerts",
		IconEmoji: ":white_check_mark:",
		Attachments: []SlackAttachment{
			{
				Color: "#36a64f", // Green
				Title: fmt.Sprintf("[RESOLVED] %s", title),
				Text:  message,
				Footer:     "GatewayOps",
				FooterIcon: "https://gatewayops.com/favicon.ico",
				Ts:         resolvedAt.Unix(),
			},
		},
	}

	return c.sendWebhook(ctx, webhookURL, slackMessage)
}

// sendWebhook sends a message to a Slack webhook URL.
func (c *SlackClient) sendWebhook(ctx context.Context, webhookURL string, message SlackMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

// getSeverityColor returns the Slack attachment color for a severity level.
func (c *SlackClient) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#dc3545" // Red
	case "warning":
		return "#ffc107" // Yellow
	case "info":
		return "#17a2b8" // Blue
	default:
		return "#6c757d" // Gray
	}
}

// TestWebhook tests a Slack webhook URL.
func (c *SlackClient) TestWebhook(ctx context.Context, webhookURL string) error {
	message := SlackMessage{
		Text:      "🔔 GatewayOps alert webhook test - connection successful!",
		Username:  "GatewayOps Alerts",
		IconEmoji: ":bell:",
	}

	return c.sendWebhook(ctx, webhookURL, message)
}
