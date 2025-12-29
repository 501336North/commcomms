package identity

import (
	"context"
	"regexp"
	"time"

	"github.com/google/uuid"
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
	_, err := s.inviteRepo.FindByCode(ctx, inviteCode)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	_, _ = s.userRepo.FindByHandle(ctx, handle)

	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           uuid.New().String(),
		Email:        email,
		Handle:       handle,
		PasswordHash: hashedPassword,
		Reputation:   0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	_ = s.inviteRepo.IncrementUsage(ctx, inviteCode)

	return user, nil
}

func (s *Service) validateHandle(handle string) error {
	if len(handle) > 20 {
		return ErrHandleTooLong
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(handle) {
		return ErrHandleInvalidChars
	}
	return nil
}

func (s *Service) isHandleAvailable(ctx context.Context, handle string) (bool, error) {
	_, err := s.userRepo.FindByHandle(ctx, handle)
	if err != nil {
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
	accessToken, _ := s.tokenGen.GenerateAccessToken(user.ID)
	refreshToken, _ := s.tokenGen.GenerateRefreshToken(user.ID)
	return &AuthResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	userID, err := s.tokenValidator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	revoked, _ := s.refreshTokenRepo.IsRevoked(ctx, refreshToken)
	if revoked {
		return nil, ErrTokenRevoked
	}
	_ = s.refreshTokenRepo.Revoke(ctx, refreshToken)
	accessToken, _ := s.tokenGen.GenerateAccessToken(userID)
	newRefreshToken, _ := s.tokenGen.GenerateRefreshToken(userID)
	return &AuthResponse{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}
