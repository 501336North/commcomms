package identity

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

func (m *MockReputationRepository) HasRecordedEvent(ctx context.Context, userID, eventType, refID string) (bool, error) {
	args := m.Called(ctx, userID, eventType, refID)
	return args.Bool(0), args.Error(1)
}

func (m *MockReputationRepository) GetReputationBreakdown(ctx context.Context, userID string) ([]ReputationBreakdown, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ReputationBreakdown), args.Error(1)
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

	// No duplicate event exists
	mockReputationRepo.On("HasRecordedEvent", ctx, "target-user", "message_posted", "message-456").Return(false, nil)

	// Expect event to be recorded with correct values
	mockReputationRepo.On("RecordEvent", ctx, mock.MatchedBy(func(event *ReputationEvent) bool {
		return event.UserID == "target-user" &&
			event.EventType == "message_posted" &&
			event.Points == 5 &&
			event.RefID == "message-456"
	})).Return(nil)

	// Act - callerID is different from targetUserID
	err := reputationService.RecordReputationEvent(ctx, "caller-user", "target-user", "message_posted", 5, "message-456")

	// Assert
	require.NoError(t, err)

	mockReputationRepo.AssertExpectations(t)
}

// TestRecordReputationEvent_PreventsSelfReputation tests that users cannot modify their own reputation.
func TestRecordReputationEvent_PreventsSelfReputation(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// Act - callerID equals targetUserID (trying to modify own reputation)
	err := reputationService.RecordReputationEvent(ctx, "user-123", "user-123", "message_upvoted", 10, "message-456")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrSelfReputation, err)
}

// TestRecordReputationEvent_ValidatesEventType tests that invalid event types are rejected.
func TestRecordReputationEvent_ValidatesEventType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// Act - invalid event type
	err := reputationService.RecordReputationEvent(ctx, "caller-user", "target-user", "invalid_event_type", 10, "ref-123")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrInvalidEventType, err)
}

// TestRecordReputationEvent_ValidatesPointsRange tests that points outside valid range are rejected.
func TestRecordReputationEvent_ValidatesPointsRange(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// Act - points exceed max for event type (message_posted max is 5)
	err := reputationService.RecordReputationEvent(ctx, "caller-user", "target-user", "message_posted", 100, "ref-123")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrInvalidPointsValue, err)
}

// TestRecordReputationEvent_PreventsDuplicateEvents tests that duplicate events are rejected.
func TestRecordReputationEvent_PreventsDuplicateEvents(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockReputationRepo := new(MockReputationRepository)

	reputationService := NewReputationService(mockReputationRepo)

	// Duplicate event exists
	mockReputationRepo.On("HasRecordedEvent", ctx, "target-user", "message_upvoted", "message-456").Return(true, nil)

	// Act - try to record duplicate event
	err := reputationService.RecordReputationEvent(ctx, "caller-user", "target-user", "message_upvoted", 5, "message-456")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrDuplicateEvent, err)

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

// TestValidateReputationEvent tests the event type and points validation function.
func TestValidateReputationEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		points    int
		wantErr   error
	}{
		{"valid message_posted", "message_posted", 3, nil},
		{"valid message_upvoted", "message_upvoted", 5, nil},
		{"valid message_downvoted", "message_downvoted", -5, nil},
		{"invalid event type", "unknown_event", 10, ErrInvalidEventType},
		{"points too high", "message_posted", 100, ErrInvalidPointsValue},
		{"points too low", "message_upvoted", -5, ErrInvalidPointsValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReputationEvent(tt.eventType, tt.points)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
