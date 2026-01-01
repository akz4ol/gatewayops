package service

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
)

// InjectionService handles prompt injection detection.
type InjectionService struct {
	safetyRepo   *repository.SafetyRepository
	auditService *AuditService
	logger       *slog.Logger
}

// NewInjectionService creates a new injection detection service.
func NewInjectionService(
	safetyRepo *repository.SafetyRepository,
	auditService *AuditService,
	logger *slog.Logger,
) *InjectionService {
	return &InjectionService{
		safetyRepo:   safetyRepo,
		auditService: auditService,
		logger:       logger,
	}
}

// Detect checks input for prompt injection patterns.
func (s *InjectionService) Detect(ctx context.Context, orgID uuid.UUID, mcpServer string, input string) (*domain.DetectionResult, error) {
	// Get applicable policies
	policies, err := s.safetyRepo.GetPoliciesForServer(ctx, orgID, mcpServer)
	if err != nil {
		return nil, err
	}

	// If no policies, use default patterns
	if len(policies) == 0 {
		return s.detectWithDefaults(input), nil
	}

	// Check against each policy
	for _, policy := range policies {
		result := s.detectWithPolicy(input, &policy)
		if result.Detected {
			return result, nil
		}
	}

	return &domain.DetectionResult{
		Detected: false,
		Action:   domain.SafetyModeLog,
	}, nil
}

// detectWithDefaults uses default patterns for detection.
func (s *InjectionService) detectWithDefaults(input string) *domain.DetectionResult {
	normalizedInput := strings.ToLower(input)

	// Check allow patterns first
	for _, pattern := range domain.DefaultAllowPatterns {
		if strings.Contains(normalizedInput, strings.ToLower(pattern)) {
			return &domain.DetectionResult{
				Detected: false,
				Action:   domain.SafetyModeLog,
			}
		}
	}

	// Check block patterns
	for _, pattern := range domain.DefaultBlockPatterns {
		if strings.Contains(normalizedInput, strings.ToLower(pattern)) {
			return &domain.DetectionResult{
				Detected:       true,
				Type:           domain.DetectionTypePromptInjection,
				Severity:       domain.DetectionSeverityHigh,
				PatternMatched: pattern,
				Confidence:     0.8,
				Action:         domain.SafetyModeWarn, // Default to warn
				Message:        "Potential prompt injection detected",
			}
		}
	}

	return &domain.DetectionResult{
		Detected: false,
		Action:   domain.SafetyModeLog,
	}
}

// detectWithPolicy uses a specific policy for detection.
func (s *InjectionService) detectWithPolicy(input string, policy *domain.SafetyPolicy) *domain.DetectionResult {
	normalizedInput := strings.ToLower(input)

	// Check allow patterns first
	for _, pattern := range policy.Patterns.Allow {
		if s.matchPattern(normalizedInput, pattern, policy.Sensitivity) {
			return &domain.DetectionResult{
				Detected: false,
				Action:   domain.SafetyModeLog,
			}
		}
	}

	// Check block patterns
	for _, pattern := range policy.Patterns.Block {
		if s.matchPattern(normalizedInput, pattern, policy.Sensitivity) {
			severity := s.getSeverityForSensitivity(policy.Sensitivity)
			return &domain.DetectionResult{
				Detected:       true,
				Type:           domain.DetectionTypePromptInjection,
				Severity:       severity,
				PatternMatched: pattern,
				Confidence:     s.getConfidenceForSensitivity(policy.Sensitivity),
				Action:         policy.Mode,
				Message:        "Prompt injection pattern matched: " + pattern,
			}
		}
	}

	return &domain.DetectionResult{
		Detected: false,
		Action:   domain.SafetyModeLog,
	}
}

// matchPattern matches input against a pattern based on sensitivity.
func (s *InjectionService) matchPattern(input, pattern string, sensitivity domain.SafetySensitivity) bool {
	normalizedPattern := strings.ToLower(pattern)

	switch sensitivity {
	case domain.SafetySensitivityStrict:
		// Strict: match anywhere, case-insensitive
		return strings.Contains(input, normalizedPattern)

	case domain.SafetySensitivityModerate:
		// Moderate: require word boundaries
		re, err := regexp.Compile(`\b` + regexp.QuoteMeta(normalizedPattern) + `\b`)
		if err != nil {
			return strings.Contains(input, normalizedPattern)
		}
		return re.MatchString(input)

	case domain.SafetySensitivityPermissive:
		// Permissive: require exact phrase match with word boundaries
		re, err := regexp.Compile(`(?:^|\s)` + regexp.QuoteMeta(normalizedPattern) + `(?:\s|$)`)
		if err != nil {
			return strings.Contains(input, normalizedPattern)
		}
		return re.MatchString(input)

	default:
		return strings.Contains(input, normalizedPattern)
	}
}

// getSeverityForSensitivity determines severity based on sensitivity level.
func (s *InjectionService) getSeverityForSensitivity(sensitivity domain.SafetySensitivity) domain.DetectionSeverity {
	switch sensitivity {
	case domain.SafetySensitivityStrict:
		return domain.DetectionSeverityHigh
	case domain.SafetySensitivityModerate:
		return domain.DetectionSeverityMedium
	case domain.SafetySensitivityPermissive:
		return domain.DetectionSeverityLow
	default:
		return domain.DetectionSeverityMedium
	}
}

// getConfidenceForSensitivity determines confidence based on sensitivity level.
func (s *InjectionService) getConfidenceForSensitivity(sensitivity domain.SafetySensitivity) float64 {
	switch sensitivity {
	case domain.SafetySensitivityStrict:
		return 0.6 // More false positives expected
	case domain.SafetySensitivityModerate:
		return 0.8
	case domain.SafetySensitivityPermissive:
		return 0.95 // High confidence when matched
	default:
		return 0.8
	}
}

// RecordDetection saves a detection to the database.
func (s *InjectionService) RecordDetection(ctx context.Context, detection *domain.InjectionDetection) error {
	detection.ID = uuid.New()
	detection.CreatedAt = time.Now()

	if err := s.safetyRepo.CreateDetection(ctx, detection); err != nil {
		return err
	}

	s.logger.Warn("injection detection recorded",
		"id", detection.ID,
		"type", detection.Type,
		"severity", detection.Severity,
		"action", detection.ActionTaken,
		"mcp_server", detection.MCPServer,
		"tool_name", detection.ToolName,
	)

	return nil
}

// GetDetection retrieves a detection by ID.
func (s *InjectionService) GetDetection(ctx context.Context, id uuid.UUID) (*domain.InjectionDetection, error) {
	return s.safetyRepo.GetDetection(ctx, id)
}

// ListDetections retrieves detections with filtering.
func (s *InjectionService) ListDetections(ctx context.Context, filter domain.DetectionFilter) (*domain.DetectionPage, error) {
	return s.safetyRepo.ListDetections(ctx, filter)
}

// GetSummary retrieves a summary of detections.
func (s *InjectionService) GetSummary(ctx context.Context, orgID uuid.UUID, period string) (*domain.SafetySummary, error) {
	return s.safetyRepo.GetSummary(ctx, orgID, period)
}

// CreatePolicy creates a new safety policy.
func (s *InjectionService) CreatePolicy(ctx context.Context, orgID uuid.UUID, createdBy uuid.UUID, input domain.SafetyPolicyInput) (*domain.SafetyPolicy, error) {
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
		CreatedBy:   createdBy,
	}

	// Set defaults if not specified
	if policy.Sensitivity == "" {
		policy.Sensitivity = domain.SafetySensitivityModerate
	}
	if policy.Mode == "" {
		policy.Mode = domain.SafetyModeWarn
	}

	if err := s.safetyRepo.CreatePolicy(ctx, policy); err != nil {
		return nil, err
	}

	s.auditService.LogPolicyChange(ctx, orgID, createdBy, domain.AuditActionPolicyCreate, policy.ID, map[string]interface{}{
		"policy_name": policy.Name,
		"sensitivity": policy.Sensitivity,
		"mode":        policy.Mode,
	})

	s.logger.Info("safety policy created",
		"id", policy.ID,
		"name", policy.Name,
		"org_id", orgID,
	)

	return policy, nil
}

// GetPolicy retrieves a safety policy by ID.
func (s *InjectionService) GetPolicy(ctx context.Context, id uuid.UUID) (*domain.SafetyPolicy, error) {
	return s.safetyRepo.GetPolicy(ctx, id)
}

// ListPolicies retrieves all safety policies for an organization.
func (s *InjectionService) ListPolicies(ctx context.Context, orgID uuid.UUID, enabledOnly bool) ([]domain.SafetyPolicy, error) {
	return s.safetyRepo.ListPolicies(ctx, orgID, enabledOnly)
}

// UpdatePolicy updates a safety policy.
func (s *InjectionService) UpdatePolicy(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID, input domain.SafetyPolicyInput) (*domain.SafetyPolicy, error) {
	policy, err := s.safetyRepo.GetPolicy(ctx, id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, ErrNotFound{Resource: "safety_policy", ID: id.String()}
	}

	policy.Name = input.Name
	policy.Description = input.Description
	policy.Sensitivity = input.Sensitivity
	policy.Mode = input.Mode
	policy.Patterns = input.Patterns
	policy.MCPServers = input.MCPServers
	policy.Enabled = input.Enabled
	policy.UpdatedAt = time.Now()

	if err := s.safetyRepo.UpdatePolicy(ctx, policy); err != nil {
		return nil, err
	}

	s.auditService.LogPolicyChange(ctx, policy.OrgID, updatedBy, domain.AuditActionPolicyUpdate, policy.ID, map[string]interface{}{
		"policy_name": policy.Name,
		"sensitivity": policy.Sensitivity,
		"mode":        policy.Mode,
	})

	return policy, nil
}

// DeletePolicy deletes a safety policy.
func (s *InjectionService) DeletePolicy(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	policy, err := s.safetyRepo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}
	if policy == nil {
		return ErrNotFound{Resource: "safety_policy", ID: id.String()}
	}

	if err := s.safetyRepo.DeletePolicy(ctx, id); err != nil {
		return err
	}

	s.auditService.LogPolicyChange(ctx, policy.OrgID, deletedBy, domain.AuditActionPolicyDelete, policy.ID, map[string]interface{}{
		"policy_name": policy.Name,
	})

	s.logger.Info("safety policy deleted",
		"id", id,
		"name", policy.Name,
	)

	return nil
}

// TestInput tests input against safety policies.
func (s *InjectionService) TestInput(ctx context.Context, orgID uuid.UUID, input string, policyID *uuid.UUID) (*domain.SafetyTestResponse, error) {
	var result *domain.DetectionResult

	if policyID != nil {
		// Test against specific policy
		policy, err := s.safetyRepo.GetPolicy(ctx, *policyID)
		if err != nil {
			return nil, err
		}
		if policy == nil {
			return nil, ErrNotFound{Resource: "safety_policy", ID: policyID.String()}
		}
		result = s.detectWithPolicy(input, policy)
	} else {
		// Test against all policies
		var err error
		result, err = s.Detect(ctx, orgID, "", input)
		if err != nil {
			return nil, err
		}
	}

	return &domain.SafetyTestResponse{
		Result:   *result,
		PolicyID: policyID,
	}, nil
}

// ErrNotFound represents a resource not found error.
type ErrNotFound struct {
	Resource string
	ID       string
}

func (e ErrNotFound) Error() string {
	return e.Resource + " not found: " + e.ID
}
