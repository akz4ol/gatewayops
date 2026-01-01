// Package domain contains the core domain models for GatewayOps.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action being audited.
type AuditAction string

const (
	AuditActionMCPToolCall    AuditAction = "mcp.tool.call"
	AuditActionMCPToolList    AuditAction = "mcp.tool.list"
	AuditActionMCPResourceGet AuditAction = "mcp.resource.get"
	AuditActionAPIKeyCreate   AuditAction = "api_key.create"
	AuditActionAPIKeyRevoke   AuditAction = "api_key.revoke"
	AuditActionAPIKeyRotate   AuditAction = "api_key.rotate"
	AuditActionUserLogin      AuditAction = "user.login"
	AuditActionUserLogout     AuditAction = "user.logout"
	AuditActionRoleCreate     AuditAction = "role.create"
	AuditActionRoleUpdate     AuditAction = "role.update"
	AuditActionRoleDelete     AuditAction = "role.delete"
	AuditActionRoleAssign     AuditAction = "role.assign"
	AuditActionRoleRevoke     AuditAction = "role.revoke"
	AuditActionPolicyCreate   AuditAction = "policy.create"
	AuditActionPolicyUpdate   AuditAction = "policy.update"
	AuditActionPolicyDelete   AuditAction = "policy.delete"
	AuditActionApprovalCreate AuditAction = "approval.create"
	AuditActionApprovalGrant  AuditAction = "approval.grant"
	AuditActionApprovalDeny   AuditAction = "approval.deny"
	AuditActionConfigChange   AuditAction = "config.change"
)

// AuditOutcome represents the result of an audited action.
type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeFailure AuditOutcome = "failure"
	AuditOutcomeBlocked AuditOutcome = "blocked"
)

// AuditLog represents a single audit log entry.
type AuditLog struct {
	ID          uuid.UUID              `json:"id"`
	OrgID       uuid.UUID              `json:"org_id"`
	TeamID      *uuid.UUID             `json:"team_id,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	APIKeyID    *uuid.UUID             `json:"api_key_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Action      AuditAction            `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Outcome     AuditOutcome           `json:"outcome"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	DurationMS  int64                  `json:"duration_ms,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AuditLogFilter defines filters for querying audit logs.
type AuditLogFilter struct {
	OrgID      uuid.UUID     `json:"org_id"`
	TeamID     *uuid.UUID    `json:"team_id,omitempty"`
	UserID     *uuid.UUID    `json:"user_id,omitempty"`
	APIKeyID   *uuid.UUID    `json:"api_key_id,omitempty"`
	Actions    []AuditAction `json:"actions,omitempty"`
	Outcomes   []AuditOutcome `json:"outcomes,omitempty"`
	Resource   string        `json:"resource,omitempty"`
	StartTime  *time.Time    `json:"start_time,omitempty"`
	EndTime    *time.Time    `json:"end_time,omitempty"`
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
}

// AuditLogPage represents a paginated list of audit logs.
type AuditLogPage struct {
	Logs       []AuditLog `json:"logs"`
	Total      int64      `json:"total"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	HasMore    bool       `json:"has_more"`
}

// AuditExportFormat defines the format for audit log exports.
type AuditExportFormat string

const (
	AuditExportJSON AuditExportFormat = "json"
	AuditExportCSV  AuditExportFormat = "csv"
)

// AuditExportConfig defines configuration for audit log exports.
type AuditExportConfig struct {
	Format      AuditExportFormat `json:"format"`
	Destination string            `json:"destination"` // webhook URL, S3 bucket, etc.
	Filter      AuditLogFilter    `json:"filter"`
}
