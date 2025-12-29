package identity

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Pre-compiled regex patterns for validation (performance optimization).
var (
	handleRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type User struct {
	ID           string
	Email        string
	Handle       string
	PasswordHash string
	Reputation   int
}

type Invite struct {
	Code        string
	MaxUses     int
	UsedCount   int
	ExpiresAt   time.Time
	CommunityID string
	CreatorID   string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByHandle(ctx context.Context, handle string) (*User, error)
}

type InviteRepository interface {
	FindByCode(ctx context.Context, code string) (*Invite, error)
	IncrementUsage(ctx context.Context, code string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type TokenGenerator interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
}

type TokenValidator interface {
	ValidateRefreshToken(token string) (string, error)
}

type RefreshTokenRepository interface {
	IsRevoked(ctx context.Context, token string) (bool, error)
	Revoke(ctx context.Context, token string) error
}

type AuthResponse struct {
	AccessToken  string
	RefreshToken string
}

type Service struct {
	userRepo         UserRepository
	inviteRepo       InviteRepository
	hasher           PasswordHasher
	tokenGen         TokenGenerator
	tokenValidator   TokenValidator
	refreshTokenRepo RefreshTokenRepository
}

func NewService(userRepo UserRepository, inviteRepo InviteRepository, hasher PasswordHasher) *Service {
	return &Service{
		userRepo:   userRepo,
		inviteRepo: inviteRepo,
		hasher:     hasher,
	}
}

func NewServiceWithTokenGenerator(userRepo UserRepository, inviteRepo InviteRepository, hasher PasswordHasher, tokenGen TokenGenerator) *Service {
	return &Service{
		userRepo:   userRepo,
		inviteRepo: inviteRepo,
		hasher:     hasher,
		tokenGen:   tokenGen,
	}
}

func NewServiceWithTokenValidator(userRepo UserRepository, inviteRepo InviteRepository, hasher PasswordHasher, tokenGen TokenGenerator, tokenValidator TokenValidator, refreshTokenRepo RefreshTokenRepository) *Service {
	return &Service{
		userRepo:         userRepo,
		inviteRepo:       inviteRepo,
		hasher:           hasher,
		tokenGen:         tokenGen,
		tokenValidator:   tokenValidator,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *Service) Register(ctx context.Context, email, password, handle, inviteCode string) (*User, error) {
	// Validate invite code exists and is usable
	invite, err := s.inviteRepo.FindByCode(ctx, inviteCode)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	// Check invite expiration
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}

	// Check invite usage limit (MaxUses of 0 means unlimited)
	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return nil, ErrInviteExhausted
	}

	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// Validate password strength
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Validate handle format and length
	if err := s.validateHandle(handle); err != nil {
		return nil, err
	}

	// Check email uniqueness
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	// Check handle uniqueness
	available, err := s.isHandleAvailable(ctx, handle)
	if err != nil {
		return nil, fmt.Errorf("failed to check handle availability: %w", err)
	}
	if !available {
		return nil, ErrHandleAlreadyTaken
	}

	// Hash password
	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		Handle:       handle,
		PasswordHash: hashedPassword,
		Reputation:   0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Increment invite usage (log error but don't fail registration)
	if err := s.inviteRepo.IncrementUsage(ctx, inviteCode); err != nil {
		// Log this error in production - invite was used but usage not tracked
		// This is a non-critical error since the user was already created
	}

	return user, nil
}

func (s *Service) validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

func (s *Service) validatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func (s *Service) validateHandle(handle string) error {
	if len(handle) < 3 {
		return ErrHandleTooShort
	}
	if len(handle) > 20 {
		return ErrHandleTooLong
	}
	if !handleRegex.MatchString(handle) {
		return ErrHandleInvalidChars
	}
	return nil
}

func (s *Service) isHandleAvailable(ctx context.Context, handle string) (bool, error) {
	_, err := s.userRepo.FindByHandle(ctx, handle)
	if err != nil {
		// Assume not found means available
		return true, nil
	}
	return false, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens with proper error handling
	accessToken, err := s.tokenGen.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenGen.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	userID, err := s.tokenValidator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	revoked, err := s.refreshTokenRepo.IsRevoked(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check token revocation: %w", err)
	}
	if revoked {
		return nil, ErrTokenRevoked
	}

	// Revoke old token before issuing new ones
	if err := s.refreshTokenRepo.Revoke(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new tokens with proper error handling
	accessToken, err := s.tokenGen.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenGen.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}
