package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/canary/commcomms/internal/auth"
	"github.com/canary/commcomms/internal/identity"
)

// MockInviteService mocks the invite service for handler tests.
type MockInviteService struct {
	mock.Mock
}

func (m *MockInviteService) CreateInvite(communityID, creatorID string, opts identity.InviteOptions) (*identity.Invite, error) {
	args := m.Called(communityID, creatorID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.Invite), args.Error(1)
}

// ============================================
// TestInviteHandler_CreateInvite
// ============================================

func TestInviteHandler_CreateInvite_Success(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invite := &identity.Invite{
		Code:        "ABC123XYZ",
		MaxUses:     10,
		ExpiresAt:   expiresAt,
		CommunityID: "test-community",
		CreatorID:   "user-123",
	}

	mockInviteService.On("CreateInvite", "test-community", "user-123", mock.MatchedBy(func(opts identity.InviteOptions) bool {
		return opts.MaxUses == 10
	})).Return(invite, nil)

	reqBody := `{"expiresInDays":7,"maxUses":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid_token")
	// Add user ID and community ID to context
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, CommunityIDKey, "test-community")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "ABC123XYZ", body["code"])
	assert.Contains(t, body["url"], "ABC123XYZ")
	assert.NotEmpty(t, body["expiresAt"])

	mockInviteService.AssertExpectations(t)
}

func TestInviteHandler_CreateInvite_DefaultExpiry(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	invite := &identity.Invite{
		Code:        "DEF456XYZ",
		MaxUses:     0, // unlimited
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CommunityID: "test-community",
		CreatorID:   "user-123",
	}

	mockInviteService.On("CreateInvite", "test-community", "user-123", mock.Anything).Return(invite, nil)

	// Request without expiresInDays - should use default
	reqBody := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid_token")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, CommunityIDKey, "test-community")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	mockInviteService.AssertExpectations(t)
}

func TestInviteHandler_CreateInvite_NoUserInContext(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	reqBody := `{"expiresInDays":7}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// No user ID in context
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Unauthorized")
}

func TestInviteHandler_CreateInvite_NoCommunityInContext(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	reqBody := `{"expiresInDays":7}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	// No community ID in context
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Community ID")
}

func TestInviteHandler_CreateInvite_InvalidJSON(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, CommunityIDKey, "test-community")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid request body")
}

func TestInviteHandler_CreateInvite_WithMaxUses(t *testing.T) {
	// Arrange
	mockInviteService := new(MockInviteService)
	handler := NewInviteHandler(mockInviteService, "https://example.com")

	invite := &identity.Invite{
		Code:        "LIMITED123",
		MaxUses:     5,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CommunityID: "test-community",
		CreatorID:   "user-123",
	}

	mockInviteService.On("CreateInvite", "test-community", "user-123", mock.MatchedBy(func(opts identity.InviteOptions) bool {
		return opts.MaxUses == 5
	})).Return(invite, nil)

	reqBody := `{"expiresInDays":7,"maxUses":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/communities/test-community/invites", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), auth.UserIDKey, "user-123")
	ctx = context.WithValue(ctx, CommunityIDKey, "test-community")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Act
	handler.CreateInvite(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "LIMITED123", body["code"])

	mockInviteService.AssertExpectations(t)
}
