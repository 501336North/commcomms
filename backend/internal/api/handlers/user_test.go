package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/canary/commcomms/internal/auth"
	"github.com/canary/commcomms/internal/identity"
)

// MockUserService mocks the user service for handler tests.
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string) (*identity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

// MockReputationService mocks the reputation service for handler tests.
type MockReputationService struct {
	mock.Mock
}

func (m *MockReputationService) GetReputation(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockReputationService) GetReputationBreakdown(ctx context.Context, userID string) ([]ReputationBreakdownItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]ReputationBreakdownItem), args.Error(1)
}

// ============================================
// TestUserHandler_GetProfile
// ============================================

func TestUserHandler_GetProfile_Success(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	user := &identity.User{
		ID:         "user-123",
		Email:      "user@example.com",
		Handle:     "testuser",
		Reputation: 150,
	}
	mockUserService.On("GetUserByID", mock.Anything, "user-123").Return(user, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	// Add user ID to context (simulating auth middleware)
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.GetProfile(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "user-123", body["id"])
	assert.Equal(t, "testuser", body["handle"])
	assert.Equal(t, "user@example.com", body["email"])
	assert.Equal(t, float64(150), body["reputation"])

	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_UserNotFound(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	mockUserService.On("GetUserByID", mock.Anything, "nonexistent-user").Return(nil, identity.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "nonexistent-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.GetProfile(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "not found")

	mockUserService.AssertExpectations(t)
}

func TestUserHandler_GetProfile_NoUserInContext(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	// No user ID in context
	w := httptest.NewRecorder()

	// Act
	handler.GetProfile(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Unauthorized")
}

// ============================================
// TestUserHandler_GetReputation
// ============================================

func TestUserHandler_GetReputation_Success(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	mockReputationService.On("GetReputation", mock.Anything, "user-123").Return(150, nil)
	mockReputationService.On("GetReputationBreakdown", mock.Anything, "user-123").Return([]ReputationBreakdownItem{
		{EventType: "message_posted", Points: 50, Count: 10},
		{EventType: "message_upvoted", Points: 100, Count: 20},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/reputation", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.GetReputation(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, float64(150), body["total"])
	assert.NotNil(t, body["breakdown"])

	breakdown := body["breakdown"].([]interface{})
	assert.Len(t, breakdown, 2)

	mockReputationService.AssertExpectations(t)
}

func TestUserHandler_GetReputation_NoUserInContext(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/reputation", nil)
	// No user ID in context
	w := httptest.NewRecorder()

	// Act
	handler.GetReputation(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Unauthorized")
}

func TestUserHandler_GetReputation_NewUserWithNoEvents(t *testing.T) {
	// Arrange
	mockUserService := new(MockUserService)
	mockReputationService := new(MockReputationService)
	handler := NewUserHandler(mockUserService, mockReputationService)

	mockReputationService.On("GetReputation", mock.Anything, "new-user").Return(0, nil)
	mockReputationService.On("GetReputationBreakdown", mock.Anything, "new-user").Return([]ReputationBreakdownItem{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/reputation", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "new-user")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.GetReputation(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, float64(0), body["total"])
	assert.NotNil(t, body["breakdown"])

	mockReputationService.AssertExpectations(t)
}
