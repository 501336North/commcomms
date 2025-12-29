package identity

import (
	"context"
	"time"
)

// ReputationService provides reputation management operations.
type ReputationService struct {
	repo ReputationRepository
}

// NewReputationService creates a new ReputationService.
func NewReputationService(repo ReputationRepository) *ReputationService {
	return &ReputationService{repo: repo}
}

// GetReputation returns the reputation score for a user.
func (s *ReputationService) GetReputation(ctx context.Context, userID string) (int, error) {
	return s.repo.GetReputation(ctx, userID)
}

// RecordReputationEvent records a reputation event for a user.
func (s *ReputationService) RecordReputationEvent(ctx context.Context, userID, eventType string, points int, refID string) error {
	event := &ReputationEvent{
		UserID:    userID,
		EventType: eventType,
		Points:    points,
		RefID:     refID,
		CreatedAt: time.Now(),
	}
	return s.repo.RecordEvent(ctx, event)
}
