// Package approval provides tool approval workflow management.
package approval

import (
	"context"
	"sync"
	"time"

	"github.com/akz4ol/gatewayops/gateway/internal/domain"
	"github.com/akz4ol/gatewayops/gateway/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Service manages tool classifications and approval workflows.
type Service struct {
	logger          zerolog.Logger
	repo            *repository.ToolRepository
	classifications map[string]*domain.ToolClassification // key: "server:tool"
	approvals       []domain.ToolApproval
	permissions     map[string]*domain.ToolPermission // key: "user_or_team:server:tool"
	mu              sync.RWMutex
}

// NewService creates a new approval service.
func NewService(logger zerolog.Logger, repo *repository.ToolRepository) *Service {
	s := &Service{
		logger:          logger,
		repo:            repo,
		classifications: make(map[string]*domain.ToolClassification),
		approvals:       make([]domain.ToolApproval, 0),
		permissions:     make(map[string]*domain.ToolPermission),
	}

	// Load from database if available
	if repo != nil {
		s.loadFromDatabase()
	} else {
		// Initialize demo classifications
		s.initDemoClassifications()
	}

	logger.Info().Msg("Tool approval service initialized")
	return s
}

// loadFromDatabase loads classifications and approvals from the database.
func (s *Service) loadFromDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	demoOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Load classifications
	classifications, err := s.repo.ListClassifications(ctx, demoOrgID, "")
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to load tool classifications from database")
	} else {
		for i := range classifications {
			key := classificationKey(classifications[i].MCPServer, classifications[i].ToolName)
			s.classifications[key] = &classifications[i]
		}
		s.logger.Info().Int("count", len(classifications)).Msg("Loaded tool classifications from database")
	}

	// Load pending approvals
	filter := domain.ToolApprovalFilter{
		OrgID:    demoOrgID,
		Statuses: []domain.ApprovalStatus{domain.ApprovalStatusPending},
		Limit:    100,
	}
	approvalPage, err := s.repo.ListApprovals(ctx, filter)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to load tool approvals from database")
	} else if approvalPage != nil {
		s.approvals = approvalPage.Approvals
		s.logger.Info().Int("count", len(s.approvals)).Msg("Loaded tool approvals from database")
	}

	// If no classifications, create defaults
	if len(s.classifications) == 0 {
		s.initDemoClassifications()
		// Persist defaults to database
		for _, c := range s.classifications {
			if err := s.repo.CreateClassification(ctx, c); err != nil {
				s.logger.Warn().Err(err).Msg("Failed to persist default classification")
			}
		}
	}
}

func (s *Service) initDemoClassifications() {
	demoOrg := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	demoUser := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Create sample classifications
	classifications := []domain.ToolClassification{
		{
			ID:               uuid.New(),
			OrgID:            demoOrg,
			MCPServer:        "filesystem",
			ToolName:         "read_file",
			Classification:   domain.ToolRiskSafe,
			RequiresApproval: false,
			Description:      "Read file contents - safe operation",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			CreatedBy:        demoUser,
		},
		{
			ID:               uuid.New(),
			OrgID:            demoOrg,
			MCPServer:        "filesystem",
			ToolName:         "write_file",
			Classification:   domain.ToolRiskSensitive,
			RequiresApproval: true,
			Description:      "Write file contents - requires approval",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			CreatedBy:        demoUser,
		},
		{
			ID:               uuid.New(),
			OrgID:            demoOrg,
			MCPServer:        "database",
			ToolName:         "execute_query",
			Classification:   domain.ToolRiskSensitive,
			RequiresApproval: true,
			Description:      "Execute database queries - requires approval",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			CreatedBy:        demoUser,
		},
		{
			ID:               uuid.New(),
			OrgID:            demoOrg,
			MCPServer:        "shell",
			ToolName:         "execute_command",
			Classification:   domain.ToolRiskDangerous,
			RequiresApproval: true,
			Description:      "Execute shell commands - dangerous, blocked by default",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			CreatedBy:        demoUser,
		},
	}

	for i := range classifications {
		key := classificationKey(classifications[i].MCPServer, classifications[i].ToolName)
		s.classifications[key] = &classifications[i]
	}
}

func classificationKey(server, tool string) string {
	return server + ":" + tool
}

func permissionKey(id uuid.UUID, server, tool string) string {
	return id.String() + ":" + server + ":" + tool
}

// GetClassification returns the classification for a tool.
func (s *Service) GetClassification(server, tool string) *domain.ToolClassification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := classificationKey(server, tool)
	if c, exists := s.classifications[key]; exists {
		return c
	}
	return nil
}

// ListClassifications returns all classifications.
func (s *Service) ListClassifications(server string) []domain.ToolClassification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.ToolClassification, 0)
	for _, c := range s.classifications {
		if server == "" || c.MCPServer == server {
			result = append(result, *c)
		}
	}
	return result
}

// SetClassification sets the classification for a tool.
func (s *Service) SetClassification(input domain.ToolClassificationInput, orgID, userID uuid.UUID) *domain.ToolClassification {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := classificationKey(input.MCPServer, input.ToolName)

	classification := &domain.ToolClassification{
		ID:               uuid.New(),
		OrgID:            orgID,
		MCPServer:        input.MCPServer,
		ToolName:         input.ToolName,
		Classification:   input.Classification,
		RequiresApproval: input.RequiresApproval,
		Description:      input.Description,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        userID,
	}

	// If exists, preserve the ID and created_at
	if existing, exists := s.classifications[key]; exists {
		classification.ID = existing.ID
		classification.CreatedAt = existing.CreatedAt
	}

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateClassification(ctx, classification); err != nil {
			s.logger.Error().Err(err).Msg("Failed to persist tool classification")
		}
	}

	s.classifications[key] = classification

	s.logger.Info().
		Str("server", input.MCPServer).
		Str("tool", input.ToolName).
		Str("classification", string(input.Classification)).
		Bool("requires_approval", input.RequiresApproval).
		Msg("Tool classification set")

	return classification
}

// DeleteClassification removes a classification.
func (s *Service) DeleteClassification(server, tool string, orgID uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := classificationKey(server, tool)
	if _, exists := s.classifications[key]; exists {
		// Delete from database
		if s.repo != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.repo.DeleteClassification(ctx, orgID, server, tool); err != nil {
				s.logger.Error().Err(err).Msg("Failed to delete tool classification from database")
			}
		}
		delete(s.classifications, key)
		return true
	}
	return false
}

// CheckAccess checks if a user/team has access to a tool.
func (s *Service) CheckAccess(userID uuid.UUID, teamID *uuid.UUID, server, tool string) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get classification
	key := classificationKey(server, tool)
	classification := s.classifications[key]

	// If no classification, use default
	if classification == nil {
		defaultLevel := domain.GetDefaultClassification(tool)
		if defaultLevel == domain.ToolRiskSafe {
			return true, ""
		}
		return false, "Tool requires approval - no classification found"
	}

	// Safe tools always allowed
	if classification.Classification == domain.ToolRiskSafe {
		return true, ""
	}

	// Dangerous tools blocked unless explicitly permitted
	if classification.Classification == domain.ToolRiskDangerous {
		// Check for explicit permission
		if s.hasPermission(userID, teamID, server, tool) {
			return true, ""
		}
		return false, "Tool is classified as dangerous and blocked"
	}

	// Sensitive tools - check if approval required
	if !classification.RequiresApproval {
		return true, ""
	}

	// Check for pre-approved permission
	if s.hasPermission(userID, teamID, server, tool) {
		return true, ""
	}

	// Check for pending/approved request
	hasApproval := s.hasApproval(userID, server, tool)
	if hasApproval {
		return true, ""
	}

	return false, "Tool requires approval"
}

func (s *Service) hasPermission(userID uuid.UUID, teamID *uuid.UUID, server, tool string) bool {
	// Check user-level permission
	userKey := permissionKey(userID, server, tool)
	if perm, exists := s.permissions[userKey]; exists {
		if perm.ExpiresAt == nil || perm.ExpiresAt.After(time.Now()) {
			return true
		}
	}

	// Check user wildcard permission
	userWildcardKey := permissionKey(userID, server, "*")
	if perm, exists := s.permissions[userWildcardKey]; exists {
		if perm.ExpiresAt == nil || perm.ExpiresAt.After(time.Now()) {
			return true
		}
	}

	// Check team-level permission
	if teamID != nil {
		teamKey := permissionKey(*teamID, server, tool)
		if perm, exists := s.permissions[teamKey]; exists {
			if perm.ExpiresAt == nil || perm.ExpiresAt.After(time.Now()) {
				return true
			}
		}

		// Check team wildcard permission
		teamWildcardKey := permissionKey(*teamID, server, "*")
		if perm, exists := s.permissions[teamWildcardKey]; exists {
			if perm.ExpiresAt == nil || perm.ExpiresAt.After(time.Now()) {
				return true
			}
		}
	}

	return false
}

func (s *Service) hasApproval(userID uuid.UUID, server, tool string) bool {
	for _, approval := range s.approvals {
		if approval.RequestedBy == userID &&
			approval.MCPServer == server &&
			approval.ToolName == tool &&
			approval.Status == domain.ApprovalStatusApproved {
			// Check if expired
			if approval.ExpiresAt == nil || approval.ExpiresAt.After(time.Now()) {
				return true
			}
		}
	}
	return false
}

// RequestApproval creates a new approval request.
func (s *Service) RequestApproval(input domain.ToolApprovalRequest, orgID, userID uuid.UUID) *domain.ToolApproval {
	s.mu.Lock()
	defer s.mu.Unlock()

	approval := domain.ToolApproval{
		ID:          uuid.New(),
		OrgID:       orgID,
		TeamID:      input.TeamID,
		MCPServer:   input.MCPServer,
		ToolName:    input.ToolName,
		RequestedBy: userID,
		RequestedAt: time.Now(),
		Reason:      input.Reason,
		Arguments:   input.Arguments,
		Status:      domain.ApprovalStatusPending,
		TraceID:     input.TraceID,
	}

	// Persist to database
	if s.repo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.CreateApproval(ctx, &approval); err != nil {
			s.logger.Error().Err(err).Msg("Failed to persist tool approval request")
		}
	}

	// Keep only last 1000 approvals
	if len(s.approvals) >= 1000 {
		s.approvals = s.approvals[1:]
	}
	s.approvals = append(s.approvals, approval)

	s.logger.Info().
		Str("approval_id", approval.ID.String()).
		Str("server", input.MCPServer).
		Str("tool", input.ToolName).
		Str("requested_by", userID.String()).
		Msg("Tool approval requested")

	return &approval
}

// GetApproval returns an approval by ID.
func (s *Service) GetApproval(id uuid.UUID) *domain.ToolApproval {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.approvals {
		if s.approvals[i].ID == id {
			return &s.approvals[i]
		}
	}
	return nil
}

// ListApprovals returns approvals matching the filter.
func (s *Service) ListApprovals(filter domain.ToolApprovalFilter) domain.ToolApprovalPage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]domain.ToolApproval, 0)
	for _, approval := range s.approvals {
		if !s.matchesFilter(approval, filter) {
			continue
		}
		filtered = append(filtered, approval)
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

	return domain.ToolApprovalPage{
		Approvals: filtered[start:end],
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		HasMore:   end < len(filtered),
	}
}

func (s *Service) matchesFilter(approval domain.ToolApproval, filter domain.ToolApprovalFilter) bool {
	if filter.MCPServer != "" && approval.MCPServer != filter.MCPServer {
		return false
	}
	if filter.ToolName != "" && approval.ToolName != filter.ToolName {
		return false
	}
	if filter.RequestedBy != nil && approval.RequestedBy != *filter.RequestedBy {
		return false
	}
	if len(filter.Statuses) > 0 {
		found := false
		for _, status := range filter.Statuses {
			if approval.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ReviewApproval approves or denies an approval request.
func (s *Service) ReviewApproval(id uuid.UUID, review domain.ToolApprovalReview, reviewerID uuid.UUID) *domain.ToolApproval {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.approvals {
		if s.approvals[i].ID == id {
			now := time.Now()
			s.approvals[i].Status = review.Status
			s.approvals[i].ReviewedBy = &reviewerID
			s.approvals[i].ReviewedAt = &now
			s.approvals[i].ReviewNote = review.ReviewNote

			if review.ExpiresIn != nil && *review.ExpiresIn > 0 {
				expiresAt := now.Add(time.Duration(*review.ExpiresIn) * time.Second)
				s.approvals[i].ExpiresAt = &expiresAt
			}

			// Persist to database
			if s.repo != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := s.repo.UpdateApproval(ctx, &s.approvals[i]); err != nil {
					s.logger.Error().Err(err).Msg("Failed to update tool approval in database")
				}
			}

			s.logger.Info().
				Str("approval_id", id.String()).
				Str("status", string(review.Status)).
				Str("reviewed_by", reviewerID.String()).
				Msg("Tool approval reviewed")

			return &s.approvals[i]
		}
	}
	return nil
}

// GrantPermission grants a permanent permission to use a tool.
func (s *Service) GrantPermission(
	orgID uuid.UUID,
	userID *uuid.UUID,
	teamID *uuid.UUID,
	server, tool string,
	grantedBy uuid.UUID,
	expiresIn *int,
	maxUsesDay *int,
) *domain.ToolPermission {
	s.mu.Lock()
	defer s.mu.Unlock()

	permission := &domain.ToolPermission{
		ID:         uuid.New(),
		OrgID:      orgID,
		TeamID:     teamID,
		UserID:     userID,
		MCPServer:  server,
		ToolName:   tool,
		GrantedBy:  grantedBy,
		GrantedAt:  time.Now(),
		MaxUsesDay: maxUsesDay,
	}

	if expiresIn != nil && *expiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(*expiresIn) * time.Second)
		permission.ExpiresAt = &expiresAt
	}

	// Determine key based on user or team
	var key string
	if userID != nil {
		key = permissionKey(*userID, server, tool)
	} else if teamID != nil {
		key = permissionKey(*teamID, server, tool)
	} else {
		return nil // Must specify user or team
	}

	s.permissions[key] = permission

	s.logger.Info().
		Str("permission_id", permission.ID.String()).
		Str("server", server).
		Str("tool", tool).
		Msg("Tool permission granted")

	return permission
}

// RevokePermission removes a permission.
func (s *Service) RevokePermission(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, perm := range s.permissions {
		if perm.ID == id {
			delete(s.permissions, key)
			return true
		}
	}
	return false
}

// ListPermissions returns all permissions.
func (s *Service) ListPermissions(server string) []domain.ToolPermission {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.ToolPermission, 0)
	for _, p := range s.permissions {
		if server == "" || p.MCPServer == server {
			result = append(result, *p)
		}
	}
	return result
}

// GetPendingCount returns the count of pending approvals.
func (s *Service) GetPendingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, approval := range s.approvals {
		if approval.Status == domain.ApprovalStatusPending {
			count++
		}
	}
	return count
}
