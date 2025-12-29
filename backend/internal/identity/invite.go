package identity

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

type InviteOptions struct {
	ExpiresAt time.Time
	MaxUses   int
}

type Community struct {
	ID   string
	Name string
}

type CommunityRepository interface {
	FindByID(ctx context.Context, id string) (*Community, error)
}

type InviteValidationRepository interface {
	FindByCode(ctx context.Context, code string) (*Invite, error)
	IncrementUsage(ctx context.Context, code string) error
}

type InviteService struct {
	inviteRepo    InviteValidationRepository
	communityRepo CommunityRepository
}

func NewInviteService(inviteRepo InviteValidationRepository, communityRepo CommunityRepository) *InviteService {
	if inviteRepo == nil || communityRepo == nil {
		panic("InviteService requires non-nil repositories")
	}
	return &InviteService{
		inviteRepo:    inviteRepo,
		communityRepo: communityRepo,
	}
}

func (s *InviteService) CreateInvite(communityID, creatorID string, opts InviteOptions) (*Invite, error) {
	expiresAt := opts.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	}

	code, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite code: %w", err)
	}

	return &Invite{
		Code:        code,
		MaxUses:     opts.MaxUses,
		ExpiresAt:   expiresAt,
		CommunityID: communityID,
		CreatorID:   creatorID,
	}, nil
}

func generateInviteCode() (string, error) {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	for i := range b {
		b[i] = chars[b[i]%62]
	}
	return string(b), nil
}

func (s *InviteService) ValidateInvite(ctx context.Context, code string) (*Community, error) {
	invite, err := s.inviteRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, ErrInviteNotFound
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}
	// MaxUses of 0 means unlimited uses
	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return nil, ErrInviteExhausted
	}
	return s.communityRepo.FindByID(ctx, invite.CommunityID)
}

func (s *InviteService) UseInvite(ctx context.Context, code string) error {
	return s.inviteRepo.IncrementUsage(ctx, code)
}
