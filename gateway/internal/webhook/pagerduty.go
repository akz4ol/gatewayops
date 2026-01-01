package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PagerDutyClient handles PagerDuty Events API v2 notifications.
type PagerDutyClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewPagerDutyClient creates a new PagerDuty client.
func NewPagerDutyClient() *PagerDutyClient {
	return &PagerDutyClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://events.pagerduty.com/v2/enqueue",
	}
}

// PagerDutyAlert represents an alert to send to PagerDuty.
type PagerDutyAlert struct {
	Summary   string
	Source    string
	Severity  string // critical, warning, error, info
	Component string
	Group     string
	Class     string
	DedupKey  string
	Details   map[string]interface{}
}

// PagerDutyEvent represents a PagerDuty Events API v2 event.
type PagerDutyEvent struct {
	RoutingKey  string              `json:"routing_key"`
	EventAction string              `json:"event_action"` // trigger, acknowledge, resolve
	DedupKey    string              `json:"dedup_key,omitempty"`
	Payload     PagerDutyPayload    `json:"payload"`
	Images      []PagerDutyImage    `json:"images,omitempty"`
	Links       []PagerDutyLink     `json:"links,omitempty"`
}

// PagerDutyPayload represents the payload of a PagerDuty event.
type PagerDutyPayload struct {
	Summary       string                 `json:"summary"`
	Source        string                 `json:"source"`
	Severity      string                 `json:"severity"`
	Timestamp     string                 `json:"timestamp,omitempty"`
	Component     string                 `json:"component,omitempty"`
	Group         string                 `json:"group,omitempty"`
	Class         string                 `json:"class,omitempty"`
	CustomDetails map[string]interface{} `json:"custom_details,omitempty"`
}

// PagerDutyImage represents an image in a PagerDuty event.
type PagerDutyImage struct {
	Src  string `json:"src"`
	Href string `json:"href,omitempty"`
	Alt  string `json:"alt,omitempty"`
}

// PagerDutyLink represents a link in a PagerDuty event.
type PagerDutyLink struct {
	Href string `json:"href"`
	Text string `json:"text,omitempty"`
}

// PagerDutyResponse represents the response from PagerDuty Events API.
type PagerDutyResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	DedupKey string `json:"dedup_key"`
}

// TriggerAlert sends a trigger event to PagerDuty.
func (c *PagerDutyClient) TriggerAlert(ctx context.Context, routingKey string, alert PagerDutyAlert) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "trigger",
		DedupKey:    alert.DedupKey,
		Payload: PagerDutyPayload{
			Summary:       alert.Summary,
			Source:        alert.Source,
			Severity:      c.normalizeSeverity(alert.Severity),
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			Component:     alert.Component,
			Group:         alert.Group,
			Class:         alert.Class,
			CustomDetails: alert.Details,
		},
		Links: []PagerDutyLink{
			{
				Href: "https://app.gatewayops.com/alerts",
				Text: "View in GatewayOps",
			},
		},
	}

	return c.sendEvent(ctx, event)
}

// AcknowledgeAlert sends an acknowledge event to PagerDuty.
func (c *PagerDutyClient) AcknowledgeAlert(ctx context.Context, routingKey, dedupKey string) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "acknowledge",
		DedupKey:    dedupKey,
	}

	return c.sendEvent(ctx, event)
}

// ResolveAlert sends a resolve event to PagerDuty.
func (c *PagerDutyClient) ResolveAlert(ctx context.Context, routingKey, dedupKey string) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "resolve",
		DedupKey:    dedupKey,
	}

	return c.sendEvent(ctx, event)
}

// sendEvent sends an event to the PagerDuty Events API.
func (c *PagerDutyClient) sendEvent(ctx context.Context, event PagerDutyEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
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
		var pdResp PagerDutyResponse
		json.NewDecoder(resp.Body).Decode(&pdResp)
		return fmt.Errorf("pagerduty error: %s (status %d)", pdResp.Message, resp.StatusCode)
	}

	return nil
}

// normalizeSeverity normalizes severity to PagerDuty-supported values.
func (c *PagerDutyClient) normalizeSeverity(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "warning":
		return "warning"
	case "error":
		return "error"
	case "info":
		return "info"
	default:
		return "warning"
	}
}

// TestConnection tests the PagerDuty routing key.
func (c *PagerDutyClient) TestConnection(ctx context.Context, routingKey string) error {
	event := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "trigger",
		DedupKey:    "gatewayops-test-" + time.Now().Format("20060102150405"),
		Payload: PagerDutyPayload{
			Summary:   "GatewayOps connection test - this is a test alert",
			Source:    "gatewayops",
			Severity:  "info",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Component: "test",
		},
	}

	if err := c.sendEvent(ctx, event); err != nil {
		return err
	}

	// Immediately resolve the test alert
	resolveEvent := PagerDutyEvent{
		RoutingKey:  routingKey,
		EventAction: "resolve",
		DedupKey:    event.DedupKey,
	}

	return c.sendEvent(ctx, resolveEvent)
}
