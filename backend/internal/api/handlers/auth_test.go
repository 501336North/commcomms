package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/canary/commcomms/internal/identity"
)

// MockIdentityService mocks the identity service for handler tests.
type MockIdentityService struct {
	mock.Mock
}

func (m *MockIdentityService) Register(ctx context.Context, email, password, handle, inviteCode string) (*identity.User, error) {
	args := m.Called(ctx, email, password, handle, inviteCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

func (m *MockIdentityService) Login(ctx context.Context, email, password string) (*identity.AuthResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.AuthResponse), args.Error(1)
}

func (m *MockIdentityService) RefreshTokens(ctx context.Context, refreshToken string) (*identity.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.AuthResponse), args.Error(1)
}

func (m *MockIdentityService) GetUserByID(ctx context.Context, userID string) (*identity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.User), args.Error(1)
}

// MockTokenService mocks the token service for handler tests.
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// MockLogoutService mocks logout functionality.
type MockLogoutService struct {
	mock.Mock
}

func (m *MockLogoutService) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// ============================================
// TestAuthHandler_Register
// ============================================

func TestAuthHandler_Register_Success(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	user := &identity.User{
		ID:         "user-123",
		Email:      "newuser@example.com",
		Handle:     "newuser",
		Reputation: 0,
	}

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "newuser", "VALID_CODE").
		Return(user, nil)
	mockTokenService.On("GenerateAccessToken", "user-123").Return("access_token_abc", nil)
	mockTokenService.On("GenerateRefreshToken", "user-123").Return("refresh_token_xyz", nil)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"newuser","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "access_token_abc", body["accessToken"])
	assert.Equal(t, "refresh_token_xyz", body["refreshToken"])

	userResp := body["user"].(map[string]interface{})
	assert.Equal(t, "user-123", userResp["id"])
	assert.Equal(t, "newuser", userResp["handle"])
	assert.Equal(t, float64(0), userResp["reputation"])

	mockIdentityService.AssertExpectations(t)
	mockTokenService.AssertExpectations(t)
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "existing@example.com", "SecurePass123!", "newhandle", "VALID_CODE").
		Return(nil, identity.ErrEmailAlreadyRegistered)

	reqBody := `{"email":"existing@example.com","password":"SecurePass123!","handle":"newhandle","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "already registered")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_DuplicateHandle(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "takenhandle", "VALID_CODE").
		Return(nil, identity.ErrHandleAlreadyTaken)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"takenhandle","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Handle already taken")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_PasswordTooShort(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "short", "newuser", "VALID_CODE").
		Return(nil, identity.ErrPasswordTooShort)

	reqBody := `{"email":"newuser@example.com","password":"short","handle":"newuser","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "8 characters")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidInviteCode(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "newuser", "INVALID_CODE").
		Return(nil, identity.ErrInvalidInviteCode)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"newuser","inviteCode":"INVALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "invite")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_InviteExpired(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "newuser", "EXPIRED_CODE").
		Return(nil, identity.ErrInviteExpired)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"newuser","inviteCode":"EXPIRED_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "expired")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_InviteExhausted(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "newuser", "EXHAUSTED_CODE").
		Return(nil, identity.ErrInviteExhausted)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"newuser","inviteCode":"EXHAUSTED_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "exhausted")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid request body")
}

func TestAuthHandler_Register_HandleInvalidChars(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "invalid handle", "VALID_CODE").
		Return(nil, identity.ErrHandleInvalidChars)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"invalid handle","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "letters, numbers")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Register_HandleTooLong(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Register", mock.Anything, "newuser@example.com", "SecurePass123!", "this_handle_is_way_too_long", "VALID_CODE").
		Return(nil, identity.ErrHandleTooLong)

	reqBody := `{"email":"newuser@example.com","password":"SecurePass123!","handle":"this_handle_is_way_too_long","inviteCode":"VALID_CODE"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Register(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "20 characters")

	mockIdentityService.AssertExpectations(t)
}

// ============================================
// TestAuthHandler_Login
// ============================================

func TestAuthHandler_Login_Success(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	authResp := &identity.AuthResponse{
		AccessToken:  "access_token_abc",
		RefreshToken: "refresh_token_xyz",
	}
	mockIdentityService.On("Login", mock.Anything, "user@example.com", "TestPass123!").
		Return(authResp, nil)

	reqBody := `{"email":"user@example.com","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "access_token_abc", body["accessToken"])
	assert.Equal(t, "refresh_token_xyz", body["refreshToken"])
	assert.NotEmpty(t, body["expiresIn"])

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Login", mock.Anything, "user@example.com", "WrongPassword").
		Return(nil, identity.ErrInvalidCredentials)

	reqBody := `{"email":"user@example.com","password":"WrongPassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid credentials")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Login_NonExistentEmail(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("Login", mock.Anything, "nonexistent@example.com", "AnyPassword").
		Return(nil, identity.ErrInvalidCredentials)

	reqBody := `{"email":"nonexistent@example.com","password":"AnyPassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid credentials")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Login(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid request body")
}

// ============================================
// TestAuthHandler_Refresh
// ============================================

func TestAuthHandler_Refresh_Success(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	authResp := &identity.AuthResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
	}
	mockIdentityService.On("RefreshTokens", mock.Anything, "old_refresh_token").
		Return(authResp, nil)

	reqBody := `{"refreshToken":"old_refresh_token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Refresh(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "new_access_token", body["accessToken"])
	assert.Equal(t, "new_refresh_token", body["refreshToken"])

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_TokenRevoked(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("RefreshTokens", mock.Anything, "revoked_token").
		Return(nil, identity.ErrTokenRevoked)

	reqBody := `{"refreshToken":"revoked_token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Refresh(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "revoked")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_TokenExpired(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	mockIdentityService.On("RefreshTokens", mock.Anything, "expired_token").
		Return(nil, identity.ErrTokenExpired)

	reqBody := `{"refreshToken":"expired_token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Refresh(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "expired")

	mockIdentityService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_InvalidJSON(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, nil)

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.Refresh(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Contains(t, body["error"], "Invalid request body")
}

// ============================================
// TestAuthHandler_Logout
// ============================================

func TestAuthHandler_Logout_Success(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	mockLogoutService := new(MockLogoutService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, mockLogoutService)

	mockLogoutService.On("RevokeToken", mock.Anything, "valid_refresh_token").Return(nil)

	reqBody := `{"refreshToken":"valid_refresh_token"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid_access_token")
	w := httptest.NewRecorder()

	// Act
	handler.Logout(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockLogoutService.AssertExpectations(t)
}

func TestAuthHandler_Logout_MissingToken(t *testing.T) {
	// Arrange
	mockIdentityService := new(MockIdentityService)
	mockTokenService := new(MockTokenService)
	mockLogoutService := new(MockLogoutService)
	handler := NewAuthHandler(mockIdentityService, mockTokenService, mockLogoutService)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	// Act
	handler.Logout(w, req)

	// Assert
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
