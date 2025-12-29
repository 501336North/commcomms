package identity

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

// ReputationRepository is the interface for reputation data access.
type ReputationRepository interface {
	GetReputation(ctx context.Context, userID string) (int, error)
	RecordEvent(ctx context.Context, event *ReputationEvent) error
}

// MockReputationRepository is a mock implementation of ReputationRepository for testing.
type MockReputationRepository struct {
	mock.Mock
}

func (m *MockReputationRepository) GetReputation(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockReputationRepository) RecordEvent(ctx context.Context, event *ReputationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// TestGetReputation_InitialZero tests that a new user has reputation initialized to 0.
// This verifies that new users start with zero reputation.
func TestGetReputation_InitialZero(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// New user has no reputation events, should return 0
	mockReputationRepo.On("GetReputation", ctx, "new-user-123").Return(0, nil)

	// Act
	reputation, err := reputationService.GetReputation(ctx, "new-user-123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0, reputation, "new user should have reputation of 0")

	mockReputationRepo.AssertExpectations(t)
}

// TestRecordReputationEvent_CreatesEvent tests that recording a reputation event stores it correctly.
// This verifies that events are recorded with type, points, and reference.
func TestRecordReputationEvent_CreatesEvent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// Expect event to be recorded with correct values
	mockReputationRepo.On("RecordEvent", ctx, mock.MatchedBy(func(event *ReputationEvent) bool {
		return event.UserID == "user-123" &&
			event.EventType == "message_quality" &&
			event.Points == 10 &&
			event.RefID == "message-456"
	})).Return(nil)

	// Act
	err := reputationService.RecordReputationEvent(ctx, "user-123", "message_quality", 10, "message-456")

	// Assert
	require.NoError(t, err)

	mockReputationRepo.AssertExpectations(t)
}

// TestGetReputation_NoDecay tests that reputation does not decay over time.
// Even after 6 months, reputation should remain unchanged.
func TestGetReputation_NoDecay(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// User has accumulated reputation of 100 (6 months ago)
	// The repository returns the same value regardless of when events occurred
	mockReputationRepo.On("GetReputation", ctx, "old-user-789").Return(100, nil)

	// Act
	reputation, err := reputationService.GetReputation(ctx, "old-user-789")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 100, reputation, "reputation should not decay after 6 months")

	mockReputationRepo.AssertExpectations(t)
}
