// Package safety provides prompt injection detection and safety features.
package safety

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Detector implements prompt injection detection.
type Detector struct {
	logger      zerolog.Logger
	repo        *repository.SafetyRepository
	policies    map[uuid.UUID]*domain.SafetyPolicy
	mu          sync.RWMutex
	detections  []domain.InjectionDetection
	detectionMu sync.RWMutex
}

// NewDetector creates a new injection detector.
func NewDetector(logger zerolog.Logger, repo *repository.SafetyRepository) *Detector {
	d := &Detector{
		logger:     logger,
		repo:       repo,
		policies:   make(map[uuid.UUID]*domain.SafetyPolicy),
		detections: make([]domain.InjectionDetection, 0),
	}

	// Load from database if available
	if repo != nil {
		d.loadFromDatabase()
	} else {
		// Create default policy
		defaultPolicy := d.createDefaultPolicy()
		d.policies[defaultPolicy.ID] = defaultPolicy
	}

	logger.Info().
		Int("default_block_patterns", len(domain.DefaultBlockPatterns)).
		Msg("Prompt injection detector initialized")

	return d
}

// loadFromDatabase loads policies from the database.
func (d *Detector) loadFromDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	demoOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	policies, err := d.repo.ListPolicies(ctx, demoOrgID, false)
	if err != nil {
		d.logger.Warn().Err(err).Msg("Failed to load safety policies from database")
	} else {
		for i := range policies {
			d.policies[policies[i].ID] = &policies[i]
		}
		d.logger.Info().Int("count", len(policies)).Msg("Loaded safety policies from database")
	}

	// If no policies, create default
	if len(d.policies) == 0 {
		defaultPolicy := d.createDefaultPolicy()
		d.policies[defaultPolicy.ID] = defaultPolicy
		// Persist to database
		if err := d.repo.CreatePolicy(ctx, defaultPolicy); err != nil {
			d.logger.Warn().Err(err).Msg("Failed to persist default safety policy")
		}
	}
}

// createDefaultPolicy creates the default safety policy.
func (d *Detector) createDefaultPolicy() *domain.SafetyPolicy {
	return &domain.SafetyPolicy{
		ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OrgID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:        "Default Policy",
		Description: "Default prompt injection detection policy",
		Sensitivity: domain.SafetySensitivityModerate,
		Mode:        domain.SafetyModeBlock,
		Patterns: domain.SafetyPatterns{
			Block: domain.DefaultBlockPatterns,
			Allow: domain.DefaultAllowPatterns,
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Detect checks input for prompt injection attempts.
func (d *Detector) Detect(input string, opts DetectOptions) domain.DetectionResult {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Get policy
	var policy *domain.SafetyPolicy
	if opts.PolicyID != nil {
		policy = d.policies[*opts.PolicyID]
	}
	if policy == nil {
		// Use default policy
		policy = d.policies[uuid.MustParse("00000000-0000-0000-0000-000000000001")]
	}

	// Skip if policy is disabled
	if policy == nil || !policy.Enabled {
		return domain.DetectionResult{
			Detected: false,
			Action:   domain.SafetyModeLog,
		}
	}

	// Normalize input for comparison
	normalizedInput := strings.ToLower(input)

	// Check allow patterns first (these override blocks)
	for _, pattern := range policy.Patterns.Allow {
		if strings.Contains(normalizedInput, strings.ToLower(pattern)) {
			return domain.DetectionResult{
				Detected: false,
				Action:   domain.SafetyModeLog,
				Message:  "Input matched allow pattern",
			}
		}
	}

	// Check block patterns
	for _, pattern := range policy.Patterns.Block {
		lowerPattern := strings.ToLower(pattern)
		if strings.Contains(normalizedInput, lowerPattern) {
			severity := d.determineSeverity(pattern, policy.Sensitivity)
			result := domain.DetectionResult{
				Detected:       true,
				Type:           domain.DetectionTypePromptInjection,
				Severity:       severity,
				PatternMatched: pattern,
				Confidence:     0.85, // Pattern-based detection confidence
				Action:         policy.Mode,
				Message:        "Potential prompt injection detected",
			}

			// Record detection
			d.recordDetection(opts, result)

			return result
		}
	}

	// Additional heuristic checks for moderate/strict sensitivity
	if policy.Sensitivity != domain.SafetySensitivityPermissive {
		if result := d.heuristicCheck(normalizedInput, policy); result.Detected {
			d.recordDetection(opts, result)
			return result
		}
	}

	return domain.DetectionResult{
		Detected: false,
		Action:   domain.SafetyModeLog,
	}
}

// heuristicCheck performs additional heuristic-based detection.
func (d *Detector) heuristicCheck(input string, policy *domain.SafetyPolicy) domain.DetectionResult {
	// Check for common injection patterns using regex
	injectionPatterns := []struct {
		pattern  string
		severity domain.DetectionSeverity
		message  string
	}{
		{`(?i)ignore\s+(all\s+)?(your|the|previous)\s+(instructions|rules|guidelines)`, domain.DetectionSeverityHigh, "Instruction override attempt"},
		{`(?i)(you\s+are|you're)\s+(now|going\s+to\s+be)\s+a`, domain.DetectionSeverityMedium, "Role manipulation attempt"},
		{`(?i)pretend\s+(to\s+be|that\s+you)`, domain.DetectionSeverityMedium, "Persona injection attempt"},
		{`(?i)from\s+now\s+on`, domain.DetectionSeverityLow, "Behavioral modification attempt"},
		{`(?i)\[\s*system\s*\]`, domain.DetectionSeverityHigh, "System prompt injection"},
		{`(?i)<\s*system\s*>`, domain.DetectionSeverityHigh, "System tag injection"},
		{`(?i)assistant:\s*\n`, domain.DetectionSeverityMedium, "Role tag injection"},
		{`(?i)human:\s*\n`, domain.DetectionSeverityMedium, "Role tag injection"},
	}

	for _, p := range injectionPatterns {
		matched, _ := regexp.MatchString(p.pattern, input)
		if matched {
			return domain.DetectionResult{
				Detected:       true,
				Type:           domain.DetectionTypePromptInjection,
				Severity:       p.severity,
				PatternMatched: p.pattern,
				Confidence:     0.75,
				Action:         policy.Mode,
				Message:        p.message,
			}
		}
	}

	return domain.DetectionResult{Detected: false}
}

// determineSeverity determines the severity based on pattern and sensitivity.
func (d *Detector) determineSeverity(pattern string, sensitivity domain.SafetySensitivity) domain.DetectionSeverity {
	lowerPattern := strings.ToLower(pattern)

	// High severity patterns
	highSeverity := []string{"jailbreak", "dan mode", "developer mode", "system prompt", "bypass"}
	for _, hs := range highSeverity {
		if strings.Contains(lowerPattern, hs) {
			return domain.DetectionSeverityHigh
		}
	}

	// Critical patterns
	criticalPatterns := []string{"ignore your programming", "override your"}
	for _, cp := range criticalPatterns {
		if strings.Contains(lowerPattern, cp) {
			return domain.DetectionSeverityCritical
		}
	}

	// Adjust based on sensitivity
	switch sensitivity {
	case domain.SafetySensitivityStrict:
		return domain.DetectionSeverityMedium
	case domain.SafetySensitivityPermissive:
		return domain.DetectionSeverityLow
	default:
		return domain.DetectionSeverityMedium
	}
}

// recordDetection records a detection event.
func (d *Detector) recordDetection(opts DetectOptions, result domain.DetectionResult) {
	d.detectionMu.Lock()
	defer d.detectionMu.Unlock()

	// Truncate input for storage
	inputTrunc := opts.Input
	if len(inputTrunc) > 500 {
		inputTrunc = inputTrunc[:500] + "..."
	}

	detection := domain.InjectionDetection{
		ID:             uuid.New(),
		OrgID:          opts.OrgID,
		TraceID:        opts.TraceID,
		SpanID:         opts.SpanID,
		PolicyID:       opts.PolicyID,
		Type:           result.Type,
		Severity:       result.Severity,
		PatternMatched: result.PatternMatched,
		Input:          inputTrunc,
		ActionTaken:    result.Action,
		MCPServer:      opts.MCPServer,
		ToolName:       opts.ToolName,
		APIKeyID:       opts.APIKeyID,
		IPAddress:      opts.IPAddress,
		CreatedAt:      time.Now(),
	}

	// Persist to database asynchronously
	if d.repo != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := d.repo.CreateDetection(ctx, &detection); err != nil {
				d.logger.Error().Err(err).Msg("Failed to persist injection detection")
			}
		}()
	}

	// Keep only last 1000 detections in memory (demo mode)
	if len(d.detections) >= 1000 {
		d.detections = d.detections[1:]
	}
	d.detections = append(d.detections, detection)

	d.logger.Warn().
		Str("type", string(result.Type)).
		Str("severity", string(result.Severity)).
		Str("pattern", result.PatternMatched).
		Str("action", string(result.Action)).
		Str("mcp_server", opts.MCPServer).
		Str("tool", opts.ToolName).
		Msg("Prompt injection detected")
}

// GetPolicies returns all policies.
func (d *Detector) GetPolicies() []domain.SafetyPolicy {
	d.mu.RLock()
	defer d.mu.RUnlock()

	policies := make([]domain.SafetyPolicy, 0, len(d.policies))
	for _, p := range d.policies {
		policies = append(policies, *p)
	}
	return policies
}

// GetPolicy returns a policy by ID.
func (d *Detector) GetPolicy(id uuid.UUID) *domain.SafetyPolicy {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.policies[id]
}

// CreatePolicy creates a new policy.
func (d *Detector) CreatePolicy(input domain.SafetyPolicyInput, orgID, userID uuid.UUID) *domain.SafetyPolicy {
	d.mu.Lock()
	defer d.mu.Unlock()

	policy := &domain.SafetyPolicy{
		ID:          uuid.New(),
		OrgID:       orgID,
		Name:        input.Name,
		Description: input.Description,
		Sensitivity: input.Sensitivity,
		Mode:        input.Mode,
		Patterns:    input.Patterns,
		MCPServers:  input.MCPServers,
		Enabled:     input.Enabled,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CreatedBy:   userID,
	}

	// Persist to database
	if d.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := d.repo.CreatePolicy(ctx, policy); err != nil {
			d.logger.Error().Err(err).Msg("Failed to persist safety policy")
		}
	}

	d.policies[policy.ID] = policy
	return policy
}

// UpdatePolicy updates an existing policy.
func (d *Detector) UpdatePolicy(id uuid.UUID, input domain.SafetyPolicyInput) *domain.SafetyPolicy {
	d.mu.Lock()
	defer d.mu.Unlock()

	policy, exists := d.policies[id]
	if !exists {
		return nil
	}

	policy.Name = input.Name
	policy.Description = input.Description
	policy.Sensitivity = input.Sensitivity
	policy.Mode = input.Mode
	policy.Patterns = input.Patterns
	policy.MCPServers = input.MCPServers
	policy.Enabled = input.Enabled
	policy.UpdatedAt = time.Now()

	// Persist to database
	if d.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := d.repo.UpdatePolicy(ctx, policy); err != nil {
			d.logger.Error().Err(err).Msg("Failed to update safety policy in database")
		}
	}

	return policy
}

// DeletePolicy deletes a policy.
func (d *Detector) DeletePolicy(id uuid.UUID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Don't allow deleting default policy
	if id == uuid.MustParse("00000000-0000-0000-0000-000000000001") {
		return false
	}

	if _, exists := d.policies[id]; exists {
		// Delete from database
		if d.repo != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := d.repo.DeletePolicy(ctx, id); err != nil {
				d.logger.Error().Err(err).Msg("Failed to delete safety policy from database")
			}
		}
		delete(d.policies, id)
		return true
	}
	return false
}

// GetDetections returns recent detections.
func (d *Detector) GetDetections(filter domain.DetectionFilter) domain.DetectionPage {
	d.detectionMu.RLock()
	defer d.detectionMu.RUnlock()

	// Filter detections
	filtered := make([]domain.InjectionDetection, 0)
	for _, det := range d.detections {
		// Apply filters
		if len(filter.Types) > 0 && !containsType(filter.Types, det.Type) {
			continue
		}
		if len(filter.Severities) > 0 && !containsSeverity(filter.Severities, det.Severity) {
			continue
		}
		if len(filter.Actions) > 0 && !containsAction(filter.Actions, det.ActionTaken) {
			continue
		}
		if filter.MCPServer != "" && det.MCPServer != filter.MCPServer {
			continue
		}
		if filter.StartTime != nil && det.CreatedAt.Before(*filter.StartTime) {
			continue
		}
		if filter.EndTime != nil && det.CreatedAt.After(*filter.EndTime) {
			continue
		}
		filtered = append(filtered, det)
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

	return domain.DetectionPage{
		Detections: filtered[start:end],
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		HasMore:    end < len(filtered),
	}
}

// GetSummary returns a summary of detections.
func (d *Detector) GetSummary() domain.SafetySummary {
	d.detectionMu.RLock()
	defer d.detectionMu.RUnlock()

	summary := domain.SafetySummary{
		TotalDetections: int64(len(d.detections)),
		ByType:          make(map[string]int64),
		BySeverity:      make(map[string]int64),
		ByAction:        make(map[string]int64),
		TopPatterns:     make([]domain.PatternCount, 0),
		Period:          "all",
	}

	patternCounts := make(map[string]int64)

	for _, det := range d.detections {
		summary.ByType[string(det.Type)]++
		summary.BySeverity[string(det.Severity)]++
		summary.ByAction[string(det.ActionTaken)]++
		if det.PatternMatched != "" {
			patternCounts[det.PatternMatched]++
		}
	}

	// Get top patterns
	for pattern, count := range patternCounts {
		summary.TopPatterns = append(summary.TopPatterns, domain.PatternCount{
			Pattern: pattern,
			Count:   count,
		})
	}

	// Sort by count (simple bubble sort for small lists)
	for i := 0; i < len(summary.TopPatterns)-1; i++ {
		for j := i + 1; j < len(summary.TopPatterns); j++ {
			if summary.TopPatterns[j].Count > summary.TopPatterns[i].Count {
				summary.TopPatterns[i], summary.TopPatterns[j] = summary.TopPatterns[j], summary.TopPatterns[i]
			}
		}
	}

	// Limit to top 10
	if len(summary.TopPatterns) > 10 {
		summary.TopPatterns = summary.TopPatterns[:10]
	}

	return summary
}

// DetectOptions contains options for detection.
type DetectOptions struct {
	Input     string
	PolicyID  *uuid.UUID
	OrgID     uuid.UUID
	TraceID   string
	SpanID    string
	MCPServer string
	ToolName  string
	APIKeyID  *uuid.UUID
	IPAddress string
}

// Helper functions
func containsType(types []domain.DetectionType, t domain.DetectionType) bool {
	for _, dt := range types {
		if dt == t {
			return true
		}
	}
	return false
}

func containsSeverity(severities []domain.DetectionSeverity, s domain.DetectionSeverity) bool {
	for _, ds := range severities {
		if ds == s {
			return true
		}
	}
	return false
}

func containsAction(actions []domain.SafetyMode, a domain.SafetyMode) bool {
	for _, da := range actions {
		if da == a {
			return true
		}
	}
	return false
}
