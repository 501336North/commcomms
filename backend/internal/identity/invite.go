package identity

import (
	"crypto/rand"
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
	FindByID(id string) (*Community, error)
}

type InviteValidationRepository interface {
	FindByCode(code string) (*Invite, error)
	IncrementUsage(code string) error
}

type InviteService struct {
	inviteRepo    InviteValidationRepository
	communityRepo CommunityRepository
}

func NewInviteService() *InviteService {
	return &InviteService{}
}

func NewInviteServiceWithRepos(inviteRepo InviteValidationRepository, communityRepo CommunityRepository) *InviteService {
	return &InviteService{inviteRepo: inviteRepo, communityRepo: communityRepo}
}

func (s *InviteService) CreateInvite(communityID, creatorID string, opts InviteOptions) (*Invite, error) {
	expiresAt := opts.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	return &Invite{
		Code:        generateInviteCode(),
		MaxUses:     opts.MaxUses,
		ExpiresAt:   expiresAt,
		CommunityID: communityID,
		CreatorID:   creatorID,
	}, nil
}

func generateInviteCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	rand.Read(b)
	for i := range b {
		b[i] = chars[b[i]%62]
	}
	return string(b)
}

func (s *InviteService) ValidateInvite(code string) (*Community, error) {
	invite, err := s.inviteRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}
	if invite.UsedCount >= invite.MaxUses {
		return nil, ErrInviteExhausted
	}
	return s.communityRepo.FindByID(invite.CommunityID)
}

func (s *InviteService) UseInvite(code string) error {
	return s.inviteRepo.IncrementUsage(code)
}
