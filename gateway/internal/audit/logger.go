// Package audit provides comprehensive audit logging.
package audit

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Logger implements audit logging functionality.
type Logger struct {
	logger   zerolog.Logger
	logs     []domain.AuditLog
	mu       sync.RWMutex
	maxLogs  int
}

// NewLogger creates a new audit logger.
func NewLogger(logger zerolog.Logger) *Logger {
	l := &Logger{
		logger:  logger,
		logs:    make([]domain.AuditLog, 0),
		maxLogs: 10000, // Keep last 10k logs in memory for demo
	}

	logger.Info().Msg("Audit logging initialized")
	return l
}

// LogEvent logs an audit event.
func (l *Logger) LogEvent(ctx context.Context, event Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log := domain.AuditLog{
		ID:         uuid.New(),
		OrgID:      event.OrgID,
		TeamID:     event.TeamID,
		UserID:     event.UserID,
		APIKeyID:   event.APIKeyID,
		TraceID:    event.TraceID,
		Action:     event.Action,
		Resource:   event.Resource,
		ResourceID: event.ResourceID,
		Outcome:    event.Outcome,
		Details:    event.Details,
		IPAddress:  event.IPAddress,
		UserAgent:  event.UserAgent,
		RequestID:  event.RequestID,
		DurationMS: event.DurationMS,
		CreatedAt:  time.Now(),
	}

	// Add to in-memory store
	if len(l.logs) >= l.maxLogs {
		l.logs = l.logs[1:]
	}
	l.logs = append(l.logs, log)

	// Also log to structured logger
	logEvent := l.logger.Info().
		Str("audit_id", log.ID.String()).
		Str("action", string(log.Action)).
		Str("resource", log.Resource).
		Str("outcome", string(log.Outcome))

	if log.UserID != nil {
		logEvent = logEvent.Str("user_id", log.UserID.String())
	}
	if log.APIKeyID != nil {
		logEvent = logEvent.Str("api_key_id", log.APIKeyID.String())
	}
	if log.ResourceID != "" {
		logEvent = logEvent.Str("resource_id", log.ResourceID)
	}
	if log.DurationMS > 0 {
		logEvent = logEvent.Int64("duration_ms", log.DurationMS)
	}

	logEvent.Msg("Audit event")
}

// GetLogs returns audit logs matching the filter.
func (l *Logger) GetLogs(filter domain.AuditLogFilter) domain.AuditLogPage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]domain.AuditLog, 0)

	for _, log := range l.logs {
		if !l.matchesFilter(log, filter) {
			continue
		}
		filtered = append(filtered, log)
	}

	// Sort by most recent first (reverse order)
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	total := int64(len(filtered))
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// Paginate
	start := offset
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return domain.AuditLogPage{
		Logs:    filtered[start:end],
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: end < len(filtered),
	}
}

// GetLog returns a single audit log by ID.
func (l *Logger) GetLog(id uuid.UUID) *domain.AuditLog {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i := range l.logs {
		if l.logs[i].ID == id {
			return &l.logs[i]
		}
	}
	return nil
}

// Search performs a text search across audit logs.
func (l *Logger) Search(query string, filter domain.AuditLogFilter) domain.AuditLogPage {
	l.mu.RLock()
	defer l.mu.RUnlock()

	query = strings.ToLower(query)
	filtered := make([]domain.AuditLog, 0)

	for _, log := range l.logs {
		if !l.matchesFilter(log, filter) {
			continue
		}
		if !l.matchesSearch(log, query) {
			continue
		}
		filtered = append(filtered, log)
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

	return domain.AuditLogPage{
		Logs:    filtered[start:end],
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: end < len(filtered),
	}
}

// Export exports audit logs in the specified format.
func (l *Logger) Export(filter domain.AuditLogFilter, format domain.AuditExportFormat) ([]byte, error) {
	page := l.GetLogs(filter)

	switch format {
	case domain.AuditExportJSON:
		return json.MarshalIndent(page.Logs, "", "  ")
	case domain.AuditExportCSV:
		return l.exportCSV(page.Logs)
	default:
		return json.MarshalIndent(page.Logs, "", "  ")
	}
}

// GetStats returns audit log statistics.
func (l *Logger) GetStats() AuditStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := AuditStats{
		TotalLogs:  int64(len(l.logs)),
		ByAction:   make(map[string]int64),
		ByOutcome:  make(map[string]int64),
		ByResource: make(map[string]int64),
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, log := range l.logs {
		stats.ByAction[string(log.Action)]++
		stats.ByOutcome[string(log.Outcome)]++
		stats.ByResource[log.Resource]++

		if log.CreatedAt.After(today) {
			stats.TodayLogs++
		}
	}

	return stats
}

// matchesFilter checks if a log matches the given filter.
func (l *Logger) matchesFilter(log domain.AuditLog, filter domain.AuditLogFilter) bool {
	// Filter by actions
	if len(filter.Actions) > 0 {
		found := false
		for _, a := range filter.Actions {
			if log.Action == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by outcomes
	if len(filter.Outcomes) > 0 {
		found := false
		for _, o := range filter.Outcomes {
			if log.Outcome == o {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by resource
	if filter.Resource != "" && log.Resource != filter.Resource {
		return false
	}

	// Filter by user ID
	if filter.UserID != nil && (log.UserID == nil || *log.UserID != *filter.UserID) {
		return false
	}

	// Filter by API key ID
	if filter.APIKeyID != nil && (log.APIKeyID == nil || *log.APIKeyID != *filter.APIKeyID) {
		return false
	}

	// Filter by time range
	if filter.StartTime != nil && log.CreatedAt.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && log.CreatedAt.After(*filter.EndTime) {
		return false
	}

	return true
}

// matchesSearch checks if a log matches the search query.
func (l *Logger) matchesSearch(log domain.AuditLog, query string) bool {
	// Search in action
	if strings.Contains(strings.ToLower(string(log.Action)), query) {
		return true
	}

	// Search in resource
	if strings.Contains(strings.ToLower(log.Resource), query) {
		return true
	}

	// Search in resource ID
	if strings.Contains(strings.ToLower(log.ResourceID), query) {
		return true
	}

	// Search in IP address
	if strings.Contains(log.IPAddress, query) {
		return true
	}

	// Search in details
	if log.Details != nil {
		detailsJSON, _ := json.Marshal(log.Details)
		if strings.Contains(strings.ToLower(string(detailsJSON)), query) {
			return true
		}
	}

	return false
}

// exportCSV exports logs as CSV.
func (l *Logger) exportCSV(logs []domain.AuditLog) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"ID", "Timestamp", "Action", "Resource", "ResourceID", "Outcome", "UserID", "APIKeyID", "IPAddress", "DurationMS"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write rows
	for _, log := range logs {
		row := []string{
			log.ID.String(),
			log.CreatedAt.Format(time.RFC3339),
			string(log.Action),
			log.Resource,
			log.ResourceID,
			string(log.Outcome),
			uuidToString(log.UserID),
			uuidToString(log.APIKeyID),
			log.IPAddress,
			intToString(log.DurationMS),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buf.String()), writer.Error()
}

// Helper functions
func uuidToString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func intToString(n int64) string {
	if n == 0 {
		return ""
	}
	return strings.TrimLeft(strings.Replace(string(rune(n)), "\x00", "", -1), "\x00")
}

// Event represents an audit event to be logged.
type Event struct {
	OrgID      uuid.UUID
	TeamID     *uuid.UUID
	UserID     *uuid.UUID
	APIKeyID   *uuid.UUID
	TraceID    string
	Action     domain.AuditAction
	Resource   string
	ResourceID string
	Outcome    domain.AuditOutcome
	Details    map[string]interface{}
	IPAddress  string
	UserAgent  string
	RequestID  string
	DurationMS int64
}

// AuditStats represents audit log statistics.
type AuditStats struct {
	TotalLogs  int64            `json:"total_logs"`
	TodayLogs  int64            `json:"today_logs"`
	ByAction   map[string]int64 `json:"by_action"`
	ByOutcome  map[string]int64 `json:"by_outcome"`
	ByResource map[string]int64 `json:"by_resource"`
}
