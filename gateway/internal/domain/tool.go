package domain

import (
	"time"

	"github.com/google/uuid"
)

// ToolRiskLevel represents the risk classification of a tool.
type ToolRiskLevel string

const (
	ToolRiskSafe      ToolRiskLevel = "safe"      // Auto-approve (e.g., read_file, list_directory)
	ToolRiskSensitive ToolRiskLevel = "sensitive" // Require approval (e.g., write_file, execute_sql)
	ToolRiskDangerous ToolRiskLevel = "dangerous" // Blocked by default (e.g., execute_command, delete_*)
)

// ToolClassification represents the risk classification of a specific tool.
type ToolClassification struct {
	ID               uuid.UUID     `json:"id"`
	OrgID            uuid.UUID     `json:"org_id"`
	MCPServer        string        `json:"mcp_server"`
	ToolName         string        `json:"tool_name"`
	Classification   ToolRiskLevel `json:"classification"`
	RequiresApproval bool          `json:"requires_approval"`
	Description      string        `json:"description,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	CreatedBy        uuid.UUID     `json:"created_by"`
}

// ToolClassificationInput represents input for classifying a tool.
type ToolClassificationInput struct {
	MCPServer        string        `json:"mcp_server"`
	ToolName         string        `json:"tool_name"`
	Classification   ToolRiskLevel `json:"classification"`
	RequiresApproval bool          `json:"requires_approval"`
	Description      string        `json:"description,omitempty"`
}

// ApprovalStatus represents the status of a tool approval request.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusDenied   ApprovalStatus = "denied"
	ApprovalStatusExpired  ApprovalStatus = "expired"
)

// ToolApproval represents a request to use a classified tool.
type ToolApproval struct {
	ID           uuid.UUID              `json:"id"`
	OrgID        uuid.UUID              `json:"org_id"`
	TeamID       *uuid.UUID             `json:"team_id,omitempty"`
	MCPServer    string                 `json:"mcp_server"`
	ToolName     string                 `json:"tool_name"`
	RequestedBy  uuid.UUID              `json:"requested_by"`
	RequestedAt  time.Time              `json:"requested_at"`
	Reason       string                 `json:"reason,omitempty"`
	Arguments    map[string]interface{} `json:"arguments,omitempty"` // Tool arguments for context
	Status       ApprovalStatus         `json:"status"`
	ReviewedBy   *uuid.UUID             `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time             `json:"reviewed_at,omitempty"`
	ReviewNote   string                 `json:"review_note,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"` // For time-limited approvals
	TraceID      string                 `json:"trace_id,omitempty"`
}

// ToolApprovalRequest represents a request to approve a tool use.
type ToolApprovalRequest struct {
	MCPServer string                 `json:"mcp_server"`
	ToolName  string                 `json:"tool_name"`
	TeamID    *uuid.UUID             `json:"team_id,omitempty"`
	Reason    string                 `json:"reason,omitempty"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// ToolApprovalReview represents a review of a tool approval request.
type ToolApprovalReview struct {
	Status     ApprovalStatus `json:"status"`
	ReviewNote string         `json:"review_note,omitempty"`
	ExpiresIn  *int           `json:"expires_in,omitempty"` // Duration in seconds
}

// ToolApprovalFilter defines filters for querying tool approvals.
type ToolApprovalFilter struct {
	OrgID       uuid.UUID        `json:"org_id"`
	TeamID      *uuid.UUID       `json:"team_id,omitempty"`
	MCPServer   string           `json:"mcp_server,omitempty"`
	ToolName    string           `json:"tool_name,omitempty"`
	RequestedBy *uuid.UUID       `json:"requested_by,omitempty"`
	Statuses    []ApprovalStatus `json:"statuses,omitempty"`
	Limit       int              `json:"limit,omitempty"`
	Offset      int              `json:"offset,omitempty"`
}

// ToolApprovalPage represents a paginated list of tool approvals.
type ToolApprovalPage struct {
	Approvals []ToolApproval `json:"approvals"`
	Total     int64          `json:"total"`
	Limit     int            `json:"limit"`
	Offset    int            `json:"offset"`
	HasMore   bool           `json:"has_more"`
}

// ToolPermission represents a pre-approved permission for a user/team to use a tool.
type ToolPermission struct {
	ID         uuid.UUID  `json:"id"`
	OrgID      uuid.UUID  `json:"org_id"`
	TeamID     *uuid.UUID `json:"team_id,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"` // If nil, applies to whole team
	MCPServer  string     `json:"mcp_server"`
	ToolName   string     `json:"tool_name"` // Can be "*" for all tools on server
	GrantedBy  uuid.UUID  `json:"granted_by"`
	GrantedAt  time.Time  `json:"granted_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUsesDay *int       `json:"max_uses_day,omitempty"` // Rate limit per day
}

// DefaultToolClassifications provides default classifications for common tools.
var DefaultToolClassifications = map[string]ToolRiskLevel{
	// Safe tools
	"read_file":       ToolRiskSafe,
	"list_directory":  ToolRiskSafe,
	"get_file_info":   ToolRiskSafe,
	"search_files":    ToolRiskSafe,
	"read_resource":   ToolRiskSafe,
	"list_resources":  ToolRiskSafe,
	"list_prompts":    ToolRiskSafe,
	"get_prompt":      ToolRiskSafe,

	// Sensitive tools
	"write_file":      ToolRiskSensitive,
	"create_file":     ToolRiskSensitive,
	"update_file":     ToolRiskSensitive,
	"execute_sql":     ToolRiskSensitive,
	"query_database":  ToolRiskSensitive,
	"send_email":      ToolRiskSensitive,
	"create_webhook":  ToolRiskSensitive,

	// Dangerous tools
	"execute_command": ToolRiskDangerous,
	"run_shell":       ToolRiskDangerous,
	"delete_file":     ToolRiskDangerous,
	"delete_database": ToolRiskDangerous,
	"drop_table":      ToolRiskDangerous,
	"admin_action":    ToolRiskDangerous,
}

// GetDefaultClassification returns the default classification for a tool.
func GetDefaultClassification(toolName string) ToolRiskLevel {
	if level, ok := DefaultToolClassifications[toolName]; ok {
		return level
	}
	// Default to sensitive for unknown tools
	return ToolRiskSensitive
}
