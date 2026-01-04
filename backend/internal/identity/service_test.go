package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByHandle(ctx context.Context, handle string) (*User, error) {
	args := m.Called(ctx, handle)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// MockInviteRepository is a mock implementation of InviteRepository for testing.
type MockInviteRepository struct {
	mock.Mock
}

func (m *MockInviteRepository) FindByCode(ctx context.Context, code string) (*Invite, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Invite), args.Error(1)
}

func (m *MockInviteRepository) IncrementUsage(ctx context.Context, code string) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

// MockPasswordHasher is a mock implementation of PasswordHasher for testing.
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) Compare(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// TestRegister_ValidUser tests that a user can register with valid email, password, handle, and invite code.
// The user should be created with a hashed password and reputation set to 0.
func TestRegister_ValidUser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists with future expiry
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)
	mockInviteRepo.On("IncrementUsage", ctx, "VALID_CODE").Return(nil)

	// Email and handle don't exist
	mockUserRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, ErrUserNotFound)
	mockUserRepo.On("FindByHandle", ctx, "newuser").Return(nil, ErrUserNotFound)

	// Password will be hashed
	mockHasher.On("Hash", "SecurePass123").Return("hashed_password", nil)

	// User will be created
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*identity.User")).Return(nil)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "SecurePass123", "newuser", "VALID_CODE")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "newuser", user.Handle)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, 0, user.Reputation)

	mockUserRepo.AssertExpectations(t)
	mockInviteRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

// TestRegister_InvalidInvite tests that registration fails with an invalid invite code.
// The service should return an "Invalid invite code" error.
func TestRegister_InvalidInvite(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Invite does not exist
	mockInviteRepo.On("FindByCode", ctx, "INVALID_CODE").Return(nil, ErrInviteNotFound)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "SecurePass123", "newuser", "INVALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidInviteCode, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_ExpiredInvite tests that registration fails with an expired invite.
func TestRegister_ExpiredInvite(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Expired invite
	expiredInvite := &Invite{
		Code:      "EXPIRED_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Expired yesterday
	}
	mockInviteRepo.On("FindByCode", ctx, "EXPIRED_CODE").Return(expiredInvite, nil)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "SecurePass123", "newuser", "EXPIRED_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInviteExpired, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_ExhaustedInvite tests that registration fails when invite has reached max uses.
func TestRegister_ExhaustedInvite(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Exhausted invite
	exhaustedInvite := &Invite{
		Code:      "EXHAUSTED_CODE",
		MaxUses:   5,
		UsedCount: 5, // Already used max times
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "EXHAUSTED_CODE").Return(exhaustedInvite, nil)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "SecurePass123", "newuser", "EXHAUSTED_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInviteExhausted, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_DuplicateEmail tests that registration fails when the email is already registered.
// The service should return an "Email already registered" error.
func TestRegister_DuplicateEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Email already exists
	existingUser := &User{
		ID:    "existing-id",
		Email: "existing@example.com",
	}
	mockUserRepo.On("FindByEmail", ctx, "existing@example.com").Return(existingUser, nil)

	// Act
	user, err := service.Register(ctx, "existing@example.com", "SecurePass123", "newhandle", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrEmailAlreadyRegistered, err)

	mockUserRepo.AssertExpectations(t)
	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_WeakPassword tests that registration fails when password is less than 8 characters.
// The service should return a "Password must be at least 8 characters" error.
func TestRegister_WeakPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "short", "newuser", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrPasswordTooShort, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_PasswordNoNumbers tests that registration fails when password has no numbers.
func TestRegister_PasswordNoNumbers(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Act - password has 8+ chars but no numbers
	user, err := service.Register(ctx, "newuser@example.com", "OnlyLetters", "newuser", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrPasswordTooWeak, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_PasswordNoLetters tests that registration fails when password has no letters.
func TestRegister_PasswordNoLetters(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Act - password has 8+ chars but no letters
	user, err := service.Register(ctx, "newuser@example.com", "12345678", "newuser", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrPasswordTooWeak, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_InvalidEmail tests that registration fails with invalid email format.
func TestRegister_InvalidEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Act
	user, err := service.Register(ctx, "notanemail", "SecurePass123", "newuser", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrInvalidEmailFormat, err)

	mockInviteRepo.AssertExpectations(t)
}

// TestRegister_DuplicateHandle tests that registration fails when handle is already taken.
func TestRegister_DuplicateHandle(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Valid invite exists
	validInvite := &Invite{
		Code:      "VALID_CODE",
		MaxUses:   10,
		UsedCount: 0,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.On("FindByCode", ctx, "VALID_CODE").Return(validInvite, nil)

	// Email doesn't exist
	mockUserRepo.On("FindByEmail", ctx, "newuser@example.com").Return(nil, ErrUserNotFound)

	// Handle already exists
	existingUser := &User{
		ID:     "existing-id",
		Handle: "takenhandle",
	}
	mockUserRepo.On("FindByHandle", ctx, "takenhandle").Return(existingUser, nil)

	// Act
	user, err := service.Register(ctx, "newuser@example.com", "SecurePass123", "takenhandle", "VALID_CODE")

	// Assert
	require.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrHandleAlreadyTaken, err)

	mockUserRepo.AssertExpectations(t)
	mockInviteRepo.AssertExpectations(t)
}

// TestValidateHandle_Valid tests that a valid handle with letters, numbers, and underscores is accepted.
func TestValidateHandle_Valid(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Act
	err := service.validateHandle("john_doe_123")

	// Assert
	assert.NoError(t, err)
}

// TestValidateHandle_Spaces tests that a handle with spaces is rejected.
func TestValidateHandle_Spaces(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Act
	err := service.validateHandle("john doe")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrHandleInvalidChars, err)
}

// TestValidateHandle_TooLong tests that a handle longer than 20 characters is rejected.
func TestValidateHandle_TooLong(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Act
	err := service.validateHandle("this_handle_is_way_too_long_for_our_system")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrHandleTooLong, err)
}

// TestValidateHandle_TooShort tests that a handle shorter than 3 characters is rejected.
func TestValidateHandle_TooShort(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Act
	err := service.validateHandle("ab")

	// Assert
	require.Error(t, err)
	assert.Equal(t, ErrHandleTooShort, err)
}

// TestValidateHandle_Duplicate tests that a handle already taken by another user is rejected.
func TestValidateHandle_Duplicate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	// Handle already exists
	existingUser := &User{
		ID:     "existing-id",
		Handle: "taken_handle",
	}
	mockUserRepo.On("FindByHandle", ctx, "taken_handle").Return(existingUser, nil)

	// Act
	available, err := service.isHandleAvailable(ctx, "taken_handle")

	// Assert
	require.NoError(t, err)
	assert.False(t, available)

	mockUserRepo.AssertExpectations(t)
}

// MockTokenGenerator is a mock implementation of TokenGenerator for testing.
type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateAccessToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenGenerator) GenerateRefreshToken(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// TestLogin_ValidCredentials tests that a user can login with valid email and password.
// The service should return access and refresh tokens.
func TestLogin_ValidCredentials(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)

	service := NewServiceWithTokenGenerator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen)

	// User exists
	existingUser := &User{
		ID:           "user-123",
		Email:        "user@example.com",
		Handle:       "testuser",
		PasswordHash: "hashed_password",
		Reputation:   0,
	}
	mockUserRepo.On("FindByEmail", ctx, "user@example.com").Return(existingUser, nil)

	// Password matches
	mockHasher.On("Compare", "hashed_password", "correct_password").Return(nil)

	// Tokens will be generated
	mockTokenGen.On("GenerateAccessToken", "user-123").Return("access_token_abc", nil)
	mockTokenGen.On("GenerateRefreshToken", "user-123").Return("refresh_token_xyz", nil)

	// Act
	authResponse, err := service.Login(ctx, "user@example.com", "correct_password")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, authResponse)
	assert.Equal(t, "access_token_abc", authResponse.AccessToken)
	assert.Equal(t, "refresh_token_xyz", authResponse.RefreshToken)

	mockUserRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
}

// TestLogin_InvalidPassword tests that login fails with an invalid password.
// The service should return an "Invalid credentials" error.
func TestLogin_InvalidPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)

	service := NewServiceWithTokenGenerator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen)

	// User exists
	existingUser := &User{
		ID:           "user-123",
		Email:        "user@example.com",
		Handle:       "testuser",
		PasswordHash: "hashed_password",
		Reputation:   0,
	}
	mockUserRepo.On("FindByEmail", ctx, "user@example.com").Return(existingUser, nil)

	// Password does NOT match
	mockHasher.On("Compare", "hashed_password", "wrong_password").Return(errors.New("password mismatch"))

	// Act
	authResponse, err := service.Login(ctx, "user@example.com", "wrong_password")

	// Assert
	require.Error(t, err)
	assert.Nil(t, authResponse)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockUserRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

// TestLogin_NonExistentEmail tests that login fails with a non-existent email.
// The service should return an "Invalid credentials" error (same as invalid password for security).
// Importantly, it should still perform a password comparison to prevent timing attacks.
func TestLogin_NonExistentEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)

	service := NewServiceWithTokenGenerator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen)

	// User does NOT exist
	mockUserRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, ErrUserNotFound)

	// Timing attack prevention: password compare is called even for non-existent users
	// The dummy hash is used to consume similar CPU time as real hash comparison
	mockHasher.On("Compare", "$2a$10$XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "any_password").Return(ErrInvalidCredentials)

	// Act
	authResponse, err := service.Login(ctx, "nonexistent@example.com", "any_password")

	// Assert
	require.Error(t, err)
	assert.Nil(t, authResponse)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockUserRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
}

// TestLogin_TokenGenerationFailure tests that login fails if token generation fails.
func TestLogin_TokenGenerationFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)

	service := NewServiceWithTokenGenerator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen)

	// User exists
	existingUser := &User{
		ID:           "user-123",
		Email:        "user@example.com",
		Handle:       "testuser",
		PasswordHash: "hashed_password",
		Reputation:   0,
	}
	mockUserRepo.On("FindByEmail", ctx, "user@example.com").Return(existingUser, nil)

	// Password matches
	mockHasher.On("Compare", "hashed_password", "correct_password").Return(nil)

	// Token generation fails
	mockTokenGen.On("GenerateAccessToken", "user-123").Return("", errors.New("token generation failed"))

	// Act
	authResponse, err := service.Login(ctx, "user@example.com", "correct_password")

	// Assert
	require.Error(t, err)
	assert.Nil(t, authResponse)
	assert.Contains(t, err.Error(), "failed to generate access token")

	mockUserRepo.AssertExpectations(t)
	mockHasher.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
}

// MockTokenValidator is a mock implementation of TokenValidator for testing.
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateRefreshToken(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository for testing.
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) IsRevoked(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// TestRefreshTokens_Valid tests that new tokens are issued with a valid refresh token.
// The service should return new access and refresh tokens.
func TestRefreshTokens_Valid(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)
	mockTokenValidator := new(MockTokenValidator)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	service := NewServiceWithTokenValidator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen, mockTokenValidator, mockRefreshTokenRepo)

	// Refresh token is valid and returns user ID
	mockTokenValidator.On("ValidateRefreshToken", "valid_refresh_token").Return("user-123", nil)

	// Token is NOT revoked
	mockRefreshTokenRepo.On("IsRevoked", ctx, "valid_refresh_token").Return(false, nil)

	// Revoke old token
	mockRefreshTokenRepo.On("Revoke", ctx, "valid_refresh_token").Return(nil)

	// New tokens will be generated
	mockTokenGen.On("GenerateAccessToken", "user-123").Return("new_access_token", nil)
	mockTokenGen.On("GenerateRefreshToken", "user-123").Return("new_refresh_token", nil)

	// Act
	authResponse, err := service.RefreshTokens(ctx, "valid_refresh_token")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, authResponse)
	assert.Equal(t, "new_access_token", authResponse.AccessToken)
	assert.Equal(t, "new_refresh_token", authResponse.RefreshToken)

	mockTokenValidator.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockTokenGen.AssertExpectations(t)
}

// TestRefreshTokens_Revoked tests that a revoked refresh token is rejected.
// The service should return a "Token revoked" error.
func TestRefreshTokens_Revoked(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)
	mockTokenValidator := new(MockTokenValidator)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	service := NewServiceWithTokenValidator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen, mockTokenValidator, mockRefreshTokenRepo)

	// Refresh token is valid (not expired)
	mockTokenValidator.On("ValidateRefreshToken", "revoked_refresh_token").Return("user-123", nil)

	// Token IS revoked
	mockRefreshTokenRepo.On("IsRevoked", ctx, "revoked_refresh_token").Return(true, nil)

	// Act
	authResponse, err := service.RefreshTokens(ctx, "revoked_refresh_token")

	// Assert
	require.Error(t, err)
	assert.Nil(t, authResponse)
	assert.Equal(t, ErrTokenRevoked, err)

	mockTokenValidator.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestRefreshTokens_Expired tests that an expired refresh token is rejected.
// The service should return a "Token expired" error.
func TestRefreshTokens_Expired(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)
	mockTokenGen := new(MockTokenGenerator)
	mockTokenValidator := new(MockTokenValidator)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	service := NewServiceWithTokenValidator(mockUserRepo, mockInviteRepo, mockHasher, mockTokenGen, mockTokenValidator, mockRefreshTokenRepo)

	// Refresh token is expired
	mockTokenValidator.On("ValidateRefreshToken", "expired_refresh_token").Return("", ErrTokenExpired)

	// Act
	authResponse, err := service.RefreshTokens(ctx, "expired_refresh_token")

	// Assert
	require.Error(t, err)
	assert.Nil(t, authResponse)
	assert.Equal(t, ErrTokenExpired, err)

	mockTokenValidator.AssertExpectations(t)
}

// TestValidateEmail_Valid tests that valid email formats are accepted.
func TestValidateEmail_Valid(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user@sub.domain.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			err := service.validateEmail(email)
			assert.NoError(t, err)
		})
	}
}

// TestValidateEmail_Invalid tests that invalid email formats are rejected.
func TestValidateEmail_Invalid(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockInviteRepo := new(MockInviteRepository)
	mockHasher := new(MockPasswordHasher)

	service := NewService(mockUserRepo, mockInviteRepo, mockHasher)

	invalidEmails := []string{
		"notanemail",
		"@example.com",
		"user@",
		"user@.com",
		"user@example",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			err := service.validateEmail(email)
			assert.Equal(t, ErrInvalidEmailFormat, err)
		})
	}
}
