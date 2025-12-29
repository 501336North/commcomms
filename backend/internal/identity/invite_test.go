package identity

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateInvite_UniqueCode tests that CreateInvite generates a unique 32-character alphanumeric code.
func TestCreateInvite_UniqueCode(t *testing.T) {
	// Arrange
	service := NewInviteService()
	opts := InviteOptions{}

	// Act
	invite, err := service.CreateInvite("community-123", "creator-456", opts)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, invite)
	assert.Len(t, invite.Code, 32, "invite code should be 32 characters")
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]+$`), invite.Code, "invite code should be alphanumeric")
}

// TestCreateInvite_DefaultExpiry tests that CreateInvite sets a default expiry of 7 days.
func TestCreateInvite_DefaultExpiry(t *testing.T) {
	// Arrange
	service := NewInviteService()
	opts := InviteOptions{}
	now := time.Now()

	// Act
	invite, err := service.CreateInvite("community-123", "creator-456", opts)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, invite)

	expectedExpiry := now.Add(7 * 24 * time.Hour)
	// Allow 1 second tolerance for test execution time
	assert.WithinDuration(t, expectedExpiry, invite.ExpiresAt, time.Second, "expiry should be 7 days from now")
}

// TestCreateInvite_CustomMaxUses tests that CreateInvite respects custom max uses.
func TestCreateInvite_CustomMaxUses(t *testing.T) {
	// Arrange
	service := NewInviteService()
	opts := InviteOptions{
		MaxUses: 5,
	}

	// Act
	invite, err := service.CreateInvite("community-123", "creator-456", opts)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, invite)
	assert.Equal(t, 5, invite.MaxUses, "max uses should be 5")
}

// TestGenerateInviteCode tests that generateInviteCode produces a 32-character alphanumeric string.
func TestGenerateInviteCode(t *testing.T) {
	// Act
	code := generateInviteCode()

	// Assert
	assert.Len(t, code, 32, "generated code should be 32 characters")
	assert.Regexp(t, regexp.MustCompile(`^[a-zA-Z0-9]+$`), code, "generated code should be alphanumeric")
}

// TestGenerateInviteCode_Uniqueness tests that generateInviteCode produces unique codes.
func TestGenerateInviteCode_Uniqueness(t *testing.T) {
	// Arrange
	codes := make(map[string]bool)

	// Act - Generate 100 codes
	for i := 0; i < 100; i++ {
		code := generateInviteCode()
		codes[code] = true
	}

	// Assert - All codes should be unique
	assert.Len(t, codes, 100, "all 100 generated codes should be unique")
}

// MockCommunityRepository is a mock implementation of CommunityRepository for testing.
type MockCommunityRepository struct {
	communities map[string]*Community
}

func NewMockCommunityRepository() *MockCommunityRepository {
	return &MockCommunityRepository{
		communities: make(map[string]*Community),
	}
}

func (m *MockCommunityRepository) FindByID(id string) (*Community, error) {
	if community, ok := m.communities[id]; ok {
		return community, nil
	}
	return nil, ErrCommunityNotFound
}

func (m *MockCommunityRepository) Add(community *Community) {
	m.communities[community.ID] = community
}

// MockInviteValidationRepository is a mock implementation of invite repository for validation tests.
type MockInviteValidationRepository struct {
	invites map[string]*Invite
}

func NewMockInviteValidationRepository() *MockInviteValidationRepository {
	return &MockInviteValidationRepository{
		invites: make(map[string]*Invite),
	}
}

func (m *MockInviteValidationRepository) FindByCode(code string) (*Invite, error) {
	if invite, ok := m.invites[code]; ok {
		return invite, nil
	}
	return nil, ErrInviteNotFound
}

func (m *MockInviteValidationRepository) IncrementUsage(code string) error {
	if invite, ok := m.invites[code]; ok {
		invite.UsedCount++
		return nil
	}
	return ErrInviteNotFound
}

func (m *MockInviteValidationRepository) Add(invite *Invite) {
	m.invites[invite.Code] = invite
}

// Sentinel errors for invite validation.
// Note: ErrInviteNotFound is declared in service_test.go
var (
	ErrInviteExpired      = errors.New("Invite link has expired")
	ErrInviteExhausted    = errors.New("Invite link exhausted")
	ErrCommunityNotFound  = errors.New("community not found")
)

// TestValidateInvite_Valid tests that ValidateInvite accepts a valid invite code and returns the community.
func TestValidateInvite_Valid(t *testing.T) {
	// Arrange
	mockInviteRepo := NewMockInviteValidationRepository()
	mockCommunityRepo := NewMockCommunityRepository()
	service := NewInviteServiceWithRepos(mockInviteRepo, mockCommunityRepo)

	// Add a valid invite
	validInvite := &Invite{
		Code:        "VALID_INVITE_CODE_12345678901234",
		CommunityID: "community-123",
		CreatorID:   "creator-456",
		MaxUses:     10,
		UsedCount:   0,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.Add(validInvite)

	// Add the community
	community := &Community{
		ID:   "community-123",
		Name: "Test Community",
	}
	mockCommunityRepo.Add(community)

	// Act
	result, err := service.ValidateInvite("VALID_INVITE_CODE_12345678901234")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "community-123", result.ID)
}

// TestValidateInvite_Expired tests that ValidateInvite rejects an expired invite.
func TestValidateInvite_Expired(t *testing.T) {
	// Arrange
	mockInviteRepo := NewMockInviteValidationRepository()
	service := NewInviteServiceWithRepos(mockInviteRepo, nil)

	// Add an expired invite
	expiredInvite := &Invite{
		Code:        "EXPIRED_INVITE_CODE_123456789012",
		CommunityID: "community-123",
		CreatorID:   "creator-456",
		MaxUses:     10,
		UsedCount:   0,
		ExpiresAt:   time.Now().Add(-24 * time.Hour), // Expired 24 hours ago
	}
	mockInviteRepo.Add(expiredInvite)

	// Act
	result, err := service.ValidateInvite("EXPIRED_INVITE_CODE_123456789012")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInviteExpired, err)
}

// TestValidateInvite_Exhausted tests that ValidateInvite rejects an exhausted invite (UsedCount >= MaxUses).
func TestValidateInvite_Exhausted(t *testing.T) {
	// Arrange
	mockInviteRepo := NewMockInviteValidationRepository()
	service := NewInviteServiceWithRepos(mockInviteRepo, nil)

	// Add an exhausted invite (UsedCount == MaxUses)
	exhaustedInvite := &Invite{
		Code:        "EXHAUSTED_INVITE_CODE_1234567890",
		CommunityID: "community-123",
		CreatorID:   "creator-456",
		MaxUses:     5,
		UsedCount:   5, // All uses consumed
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.Add(exhaustedInvite)

	// Act
	result, err := service.ValidateInvite("EXHAUSTED_INVITE_CODE_1234567890")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInviteExhausted, err)
}

// TestUseInvite_IncrementsCount tests that UseInvite increments the UsedCount by 1.
func TestUseInvite_IncrementsCount(t *testing.T) {
	// Arrange
	mockInviteRepo := NewMockInviteValidationRepository()
	service := NewInviteServiceWithRepos(mockInviteRepo, nil)

	// Add a valid invite with some usage
	validInvite := &Invite{
		Code:        "USE_INVITE_CODE_12345678901234",
		CommunityID: "community-123",
		CreatorID:   "creator-456",
		MaxUses:     10,
		UsedCount:   3,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}
	mockInviteRepo.Add(validInvite)

	// Act
	err := service.UseInvite("USE_INVITE_CODE_12345678901234")

	// Assert
	require.NoError(t, err)

	// Verify the count was incremented
	updatedInvite, _ := mockInviteRepo.FindByCode("USE_INVITE_CODE_12345678901234")
	assert.Equal(t, 4, updatedInvite.UsedCount, "UsedCount should be incremented by 1")
}
