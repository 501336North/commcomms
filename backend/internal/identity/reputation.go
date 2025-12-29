package identity

import (
	"context"
	"fmt"
	"time"
)

// ReputationEvent represents a single reputation change event.
type ReputationEvent struct {
	ID        string
	UserID    string
	EventType string
	Points    int
	RefID     string
	CreatedAt time.Time
}

// ReputationRepository defines the interface for reputation data access.
type ReputationRepository interface {
	GetReputation(ctx context.Context, userID string) (int, error)
	RecordEvent(ctx context.Context, event *ReputationEvent) error
	HasRecordedEvent(ctx context.Context, userID, eventType, refID string) (bool, error)
}

// ReputationService provides reputation management operations.
type ReputationService struct {
	repo ReputationRepository
}

// NewReputationService creates a new ReputationService.
func NewReputationService(repo ReputationRepository) *ReputationService {
	if repo == nil {
		panic("ReputationService requires non-nil repository")
	}
	return &ReputationService{repo: repo}
}

// GetReputation returns the reputation score for a user.
func (s *ReputationService) GetReputation(ctx context.Context, userID string) (int, error) {
	return s.repo.GetReputation(ctx, userID)
}

// RecordReputationEvent records a reputation event for a user with proper validation.
// callerID is the user initiating the action (for authorization checks).
// targetUserID is the user whose reputation is being modified.
func (s *ReputationService) RecordReputationEvent(ctx context.Context, callerID, targetUserID, eventType string, points int, refID string) error {
	// Prevent self-reputation modification (except for system events)
	if callerID == targetUserID && eventType != string(EventModeratorAction) {
		return ErrSelfReputation
	}

	// Validate event type and points
	if err := ValidateReputationEvent(eventType, points); err != nil {
		return err
	}

	// Check for duplicate events (prevent gaming the system)
	if refID != "" {
		exists, err := s.repo.HasRecordedEvent(ctx, targetUserID, eventType, refID)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate event: %w", err)
		}
		if exists {
			return ErrDuplicateEvent
		}
	}

	event := &ReputationEvent{
		UserID:    targetUserID,
		EventType: eventType,
		Points:    points,
		RefID:     refID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.RecordEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to record reputation event: %w", err)
	}

	return nil
}
