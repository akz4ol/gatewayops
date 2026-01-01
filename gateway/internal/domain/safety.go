package domain

import (
	"time"

	"github.com/google/uuid"
)

// SafetyMode represents the action to take when a safety violation is detected.
type SafetyMode string

const (
	SafetyModeBlock SafetyMode = "block" // Block the request
	SafetyModeWarn  SafetyMode = "warn"  // Log and allow
	SafetyModeLog   SafetyMode = "log"   // Log only
)

// SafetySensitivity represents the detection sensitivity level.
type SafetySensitivity string

const (
	SafetySensitivityStrict   SafetySensitivity = "strict"   // Maximum detection, may have false positives
	SafetySensitivityModerate SafetySensitivity = "moderate" // Balanced detection
	SafetySensitivityPermissive SafetySensitivity = "permissive" // Minimal detection, low false positives
)

// SafetyPolicy represents a safety policy configuration.
type SafetyPolicy struct {
	ID               uuid.UUID              `json:"id"`
	OrgID            uuid.UUID              `json:"org_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Sensitivity      SafetySensitivity      `json:"sensitivity"`
	Mode             SafetyMode             `json:"mode"`
	Patterns         SafetyPatterns         `json:"patterns"`
	MCPServers       []string               `json:"mcp_servers,omitempty"` // Empty means all
	Enabled          bool                   `json:"enabled"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	CreatedBy        uuid.UUID              `json:"created_by"`
}

// SafetyPatterns defines block and allow patterns for detection.
type SafetyPatterns struct {
	Block []string `json:"block,omitempty"` // Patterns to block
	Allow []string `json:"allow,omitempty"` // Patterns to allow (override blocks)
}

// SafetyPolicyInput represents input for creating/updating a safety policy.
type SafetyPolicyInput struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Sensitivity SafetySensitivity `json:"sensitivity"`
	Mode        SafetyMode        `json:"mode"`
	Patterns    SafetyPatterns    `json:"patterns"`
	MCPServers  []string          `json:"mcp_servers,omitempty"`
	Enabled     bool              `json:"enabled"`
}

// DetectionSeverity represents the severity of a detected issue.
type DetectionSeverity string

const (
	DetectionSeverityLow      DetectionSeverity = "low"
	DetectionSeverityMedium   DetectionSeverity = "medium"
	DetectionSeverityHigh     DetectionSeverity = "high"
	DetectionSeverityCritical DetectionSeverity = "critical"
)

// DetectionType represents the type of detection.
type DetectionType string

const (
	DetectionTypePromptInjection DetectionType = "prompt_injection"
	DetectionTypePII             DetectionType = "pii"
	DetectionTypeSecret          DetectionType = "secret"
	DetectionTypeMalicious       DetectionType = "malicious"
)

// InjectionDetection represents a detected prompt injection attempt.
type InjectionDetection struct {
	ID             uuid.UUID         `json:"id"`
	OrgID          uuid.UUID         `json:"org_id"`
	TraceID        string            `json:"trace_id,omitempty"`
	SpanID         string            `json:"span_id,omitempty"`
	PolicyID       *uuid.UUID        `json:"policy_id,omitempty"`
	Type           DetectionType     `json:"type"`
	Severity       DetectionSeverity `json:"severity"`
	PatternMatched string            `json:"pattern_matched,omitempty"`
	Input          string            `json:"input"` // The input that triggered detection (may be truncated)
	ActionTaken    SafetyMode        `json:"action_taken"`
	MCPServer      string            `json:"mcp_server,omitempty"`
	ToolName       string            `json:"tool_name,omitempty"`
	APIKeyID       *uuid.UUID        `json:"api_key_id,omitempty"`
	IPAddress      string            `json:"ip_address,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// DetectionResult represents the result of safety detection.
type DetectionResult struct {
	Detected       bool              `json:"detected"`
	Type           DetectionType     `json:"type,omitempty"`
	Severity       DetectionSeverity `json:"severity,omitempty"`
	PatternMatched string            `json:"pattern_matched,omitempty"`
	Confidence     float64           `json:"confidence,omitempty"` // 0-1 for ML-based detection
	Action         SafetyMode        `json:"action"`
	Message        string            `json:"message,omitempty"`
}

// DetectionFilter defines filters for querying detections.
type DetectionFilter struct {
	OrgID      uuid.UUID           `json:"org_id"`
	Types      []DetectionType     `json:"types,omitempty"`
	Severities []DetectionSeverity `json:"severities,omitempty"`
	Actions    []SafetyMode        `json:"actions,omitempty"`
	MCPServer  string              `json:"mcp_server,omitempty"`
	StartTime  *time.Time          `json:"start_time,omitempty"`
	EndTime    *time.Time          `json:"end_time,omitempty"`
	Limit      int                 `json:"limit,omitempty"`
	Offset     int                 `json:"offset,omitempty"`
}

// DetectionPage represents a paginated list of detections.
type DetectionPage struct {
	Detections []InjectionDetection `json:"detections"`
	Total      int64                `json:"total"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	HasMore    bool                 `json:"has_more"`
}

// SafetyTestRequest represents a request to test safety detection.
type SafetyTestRequest struct {
	Input     string    `json:"input"`
	PolicyID  *uuid.UUID `json:"policy_id,omitempty"` // If nil, uses default policy
}

// SafetyTestResponse represents the response from a safety test.
type SafetyTestResponse struct {
	Result   DetectionResult `json:"result"`
	PolicyID *uuid.UUID      `json:"policy_id,omitempty"`
}

// DefaultBlockPatterns provides default patterns to block.
var DefaultBlockPatterns = []string{
	// Common prompt injection patterns
	"ignore previous instructions",
	"ignore all previous",
	"disregard all prior",
	"disregard previous instructions",
	"forget all previous",
	"forget your instructions",
	"you are now",
	"pretend you are",
	"act as if you",
	"new persona",
	"new role",
	"jailbreak",
	"DAN mode",
	"developer mode",
	"ignore your programming",
	"bypass your",
	"override your",
	"system prompt",
	"initial prompt",
	"reveal your prompt",
	"show your instructions",
	"what are your instructions",
}

// DefaultAllowPatterns provides default patterns to allow (override blocks).
var DefaultAllowPatterns = []string{
	"summarize the following",
	"please help me",
	"can you explain",
	"what is the",
	"how do I",
}

// SafetySummary represents a summary of safety events.
type SafetySummary struct {
	TotalDetections int64             `json:"total_detections"`
	ByType          map[string]int64  `json:"by_type"`
	BySeverity      map[string]int64  `json:"by_severity"`
	ByAction        map[string]int64  `json:"by_action"`
	TopPatterns     []PatternCount    `json:"top_patterns"`
	Period          string            `json:"period"`
}

// PatternCount represents a pattern and its occurrence count.
type PatternCount struct {
	Pattern string `json:"pattern"`
	Count   int64  `json:"count"`
}
