package acceptance

import (
	"context"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/canary/commcomms/internal/api"
	"github.com/canary/commcomms/internal/api/handlers"
	"github.com/canary/commcomms/internal/auth"
	"github.com/canary/commcomms/internal/identity"
)

// In-memory implementations for acceptance tests

// InMemoryUserRepository stores users in memory.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*identity.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*identity.User),
	}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *identity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, identity.ErrUserNotFound
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, identity.ErrUserNotFound
}

func (r *InMemoryUserRepository) FindByHandle(ctx context.Context, handle string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, user := range r.users {
		if user.Handle == handle {
			return user, nil
		}
	}
	return nil, identity.ErrUserNotFound
}

// InMemoryInviteRepository stores invites in memory.
type InMemoryInviteRepository struct {
	mu      sync.RWMutex
	invites map[string]*identity.Invite
}

func NewInMemoryInviteRepository() *InMemoryInviteRepository {
	return &InMemoryInviteRepository{
		invites: make(map[string]*identity.Invite),
	}
}

func (r *InMemoryInviteRepository) FindByCode(ctx context.Context, code string) (*identity.Invite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	invite, ok := r.invites[code]
	if !ok {
		return nil, identity.ErrInviteNotFound
	}
	return invite, nil
}

func (r *InMemoryInviteRepository) IncrementUsage(ctx context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	invite, ok := r.invites[code]
	if !ok {
		return identity.ErrInviteNotFound
	}
	invite.UsedCount++
	return nil
}

func (r *InMemoryInviteRepository) CreateInvite(invite *identity.Invite) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invites[invite.Code] = invite
}

// InMemoryRefreshTokenRepository stores revoked tokens in memory.
type InMemoryRefreshTokenRepository struct {
	mu      sync.RWMutex
	revoked map[string]bool
}

func NewInMemoryRefreshTokenRepository() *InMemoryRefreshTokenRepository {
	return &InMemoryRefreshTokenRepository{
		revoked: make(map[string]bool),
	}
}

func (r *InMemoryRefreshTokenRepository) IsRevoked(ctx context.Context, token string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.revoked[token], nil
}

func (r *InMemoryRefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked[token] = true
	return nil
}

// RevokeToken implements the LogoutService interface.
func (r *InMemoryRefreshTokenRepository) RevokeToken(ctx context.Context, token string) error {
	return r.Revoke(ctx, token)
}

// InMemoryReputationRepository stores reputation data in memory.
type InMemoryReputationRepository struct {
	mu         sync.RWMutex
	events     []*identity.ReputationEvent
	reputation map[string]int
}

func NewInMemoryReputationRepository() *InMemoryReputationRepository {
	return &InMemoryReputationRepository{
		events:     make([]*identity.ReputationEvent, 0),
		reputation: make(map[string]int),
	}
}

func (r *InMemoryReputationRepository) GetReputation(ctx context.Context, userID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reputation[userID], nil
}

func (r *InMemoryReputationRepository) GetReputationBreakdown(ctx context.Context, userID string) ([]identity.ReputationBreakdown, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	breakdown := make(map[string]*identity.ReputationBreakdown)
	for _, event := range r.events {
		if event.UserID == userID {
			if b, ok := breakdown[event.EventType]; ok {
				b.Points += event.Points
				b.Count++
			} else {
				breakdown[event.EventType] = &identity.ReputationBreakdown{
					EventType: event.EventType,
					Points:    event.Points,
					Count:     1,
				}
			}
		}
	}

	result := make([]identity.ReputationBreakdown, 0, len(breakdown))
	for _, b := range breakdown {
		result = append(result, *b)
	}
	return result, nil
}

func (r *InMemoryReputationRepository) RecordEvent(ctx context.Context, event *identity.ReputationEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	r.reputation[event.UserID] += event.Points
	return nil
}

func (r *InMemoryReputationRepository) HasRecordedEvent(ctx context.Context, userID, eventType, refID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, event := range r.events {
		if event.UserID == userID && event.EventType == eventType && event.RefID == refID {
			return true, nil
		}
	}
	return false, nil
}

// BcryptPasswordHasher implements password hashing with bcrypt.
type BcryptPasswordHasher struct{}

func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	// For acceptance tests, use a simple hash to speed up tests
	// In production, use bcrypt.GenerateFromPassword
	return "hashed_" + password, nil
}

func (h *BcryptPasswordHasher) Compare(hashedPassword, password string) error {
	if hashedPassword == "hashed_"+password {
		return nil
	}
	return identity.ErrInvalidCredentials
}

// InMemoryCommunityRepository stores communities in memory.
type InMemoryCommunityRepository struct {
	mu          sync.RWMutex
	communities map[string]*identity.Community
}

func NewInMemoryCommunityRepository() *InMemoryCommunityRepository {
	return &InMemoryCommunityRepository{
		communities: make(map[string]*identity.Community),
	}
}

func (r *InMemoryCommunityRepository) FindByID(ctx context.Context, id string) (*identity.Community, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	community, ok := r.communities[id]
	if !ok {
		// Return a default community for tests
		return &identity.Community{ID: id, Name: "Test Community"}, nil
	}
	return community, nil
}

// InMemoryInviteValidationRepository implements the invite validation interface.
type InMemoryInviteValidationRepository struct {
	*InMemoryInviteRepository
}

func NewInMemoryInviteValidationRepository(inviteRepo *InMemoryInviteRepository) *InMemoryInviteValidationRepository {
	return &InMemoryInviteValidationRepository{inviteRepo}
}

func (r *InMemoryInviteValidationRepository) AtomicUseInvite(ctx context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	invite, ok := r.invites[code]
	if !ok {
		return identity.ErrInviteNotFound
	}

	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return identity.ErrInviteExhausted
	}

	invite.UsedCount++
	return nil
}

// JWTTokenValidator wraps JWTService to implement TokenValidator.
type JWTTokenValidator struct {
	jwtService *auth.JWTService
}

func (v *JWTTokenValidator) ValidateRefreshToken(token string) (string, error) {
	claims, err := v.jwtService.ValidateToken(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// ReputationServiceAdapter adapts identity.ReputationService for handler use.
type ReputationServiceAdapter struct {
	service *identity.ReputationService
}

func (a *ReputationServiceAdapter) GetReputation(ctx context.Context, userID string) (int, error) {
	return a.service.GetReputation(ctx, userID)
}

func (a *ReputationServiceAdapter) GetReputationBreakdown(ctx context.Context, userID string) ([]handlers.ReputationBreakdownItem, error) {
	breakdown, err := a.service.GetReputationBreakdown(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]handlers.ReputationBreakdownItem, len(breakdown))
	for i, b := range breakdown {
		result[i] = handlers.ReputationBreakdownItem{
			EventType: b.EventType,
			Points:    b.Points,
			Count:     b.Count,
		}
	}
	return result, nil
}

// Test infrastructure
var (
	userRepo              *InMemoryUserRepository
	inviteRepo            *InMemoryInviteRepository
	refreshTokenRepo      *InMemoryRefreshTokenRepository
	reputationRepo        *InMemoryReputationRepository
	communityRepo         *InMemoryCommunityRepository
	identityService       *identity.Service
	reputationService     *identity.ReputationService
	inviteService         *identity.InviteService
	jwtService            *auth.JWTService
	testServerInitialized bool
	inviteCounter         int
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	if testServerInitialized {
		return
	}

	// Initialize repositories
	userRepo = NewInMemoryUserRepository()
	inviteRepo = NewInMemoryInviteRepository()
	refreshTokenRepo = NewInMemoryRefreshTokenRepository()
	reputationRepo = NewInMemoryReputationRepository()
	communityRepo = NewInMemoryCommunityRepository()

	// Initialize services
	hasher := &BcryptPasswordHasher{}
	jwtService = auth.NewJWTService("test-secret-key-for-acceptance-tests")
	tokenValidator := &JWTTokenValidator{jwtService: jwtService}

	identityService = identity.NewServiceWithTokenValidator(
		userRepo,
		inviteRepo,
		hasher,
		jwtService,
		tokenValidator,
		refreshTokenRepo,
	)

	reputationService = identity.NewReputationService(reputationRepo)

	inviteValidationRepo := NewInMemoryInviteValidationRepository(inviteRepo)
	inviteService = identity.NewInviteService(inviteValidationRepo, communityRepo)

	// Create handlers
	authHandler := handlers.NewAuthHandler(identityService, jwtService, refreshTokenRepo)
	userHandler := handlers.NewUserHandler(identityService, &ReputationServiceAdapter{service: reputationService})
	inviteHandler := handlers.NewInviteHandler(inviteService, "https://example.com")

	// Create router
	router := api.NewRouter(api.RouterConfig{
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
		InviteHandler: inviteHandler,
		JWTService:    jwtService,
	})

	// Create test server
	TestServer = httptest.NewServer(router)
	testServerInitialized = true
}

// resetTestData clears all test data between tests.
func resetTestData() {
	userRepo = NewInMemoryUserRepository()
	inviteRepo = NewInMemoryInviteRepository()
	refreshTokenRepo = NewInMemoryRefreshTokenRepository()
	reputationRepo = NewInMemoryReputationRepository()
	inviteCounter = 0

	// Reinitialize services with new repos
	hasher := &BcryptPasswordHasher{}
	tokenValidator := &JWTTokenValidator{jwtService: jwtService}

	identityService = identity.NewServiceWithTokenValidator(
		userRepo,
		inviteRepo,
		hasher,
		jwtService,
		tokenValidator,
		refreshTokenRepo,
	)

	reputationService = identity.NewReputationService(reputationRepo)

	inviteValidationRepo := NewInMemoryInviteValidationRepository(inviteRepo)
	inviteService = identity.NewInviteService(inviteValidationRepo, communityRepo)

	// Recreate handlers with new services
	authHandler := handlers.NewAuthHandler(identityService, jwtService, refreshTokenRepo)
	userHandler := handlers.NewUserHandler(identityService, &ReputationServiceAdapter{service: reputationService})
	inviteHandler := handlers.NewInviteHandler(inviteService, "https://example.com")

	// Recreate router
	router := api.NewRouter(api.RouterConfig{
		AuthHandler:   authHandler,
		UserHandler:   userHandler,
		InviteHandler: inviteHandler,
		JWTService:    jwtService,
	})

	// Update test server
	TestServer.Close()
	TestServer = httptest.NewServer(router)
}

// createTestInvite creates a test invite and returns the code.
func createTestInvite(t *testing.T) string {
	t.Helper()
	inviteCounter++
	code := "TEST_INVITE_" + time.Now().Format("20060102150405") + "_" + string(rune('A'+inviteCounter))

	invite := &identity.Invite{
		Code:        code,
		MaxUses:     0, // unlimited
		UsedCount:   0,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CommunityID: "test-community",
		CreatorID:   "system",
	}
	inviteRepo.CreateInvite(invite)
	return code
}

// createExpiredInvite creates an expired invite for testing.
func createExpiredInvite(t *testing.T) string {
	t.Helper()
	inviteCounter++
	code := "EXPIRED_INVITE_" + time.Now().Format("20060102150405") + "_" + string(rune('A'+inviteCounter))

	invite := &identity.Invite{
		Code:        code,
		MaxUses:     0,
		UsedCount:   0,
		ExpiresAt:   time.Now().Add(-24 * time.Hour), // expired
		CommunityID: "test-community",
		CreatorID:   "system",
	}
	inviteRepo.CreateInvite(invite)
	return code
}

// createLimitedInvite creates an invite with max uses.
func createLimitedInvite(t *testing.T, maxUses int) string {
	t.Helper()
	inviteCounter++
	code := "LIMITED_INVITE_" + time.Now().Format("20060102150405") + "_" + string(rune('A'+inviteCounter))

	invite := &identity.Invite{
		Code:        code,
		MaxUses:     maxUses,
		UsedCount:   0,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CommunityID: "test-community",
		CreatorID:   "system",
	}
	inviteRepo.CreateInvite(invite)
	return code
}

// createTestUser creates a test user and returns it.
func createTestUser(t *testing.T) TestUser {
	t.Helper()

	// Create a unique invite for this user
	inviteCode := createTestInvite(t)

	// Create unique user
	inviteCounter++
	email := "testuser" + time.Now().Format("20060102150405") + string(rune('A'+inviteCounter)) + "@example.com"
	handle := "testuser" + time.Now().Format("150405") + string(rune('a'+inviteCounter))

	// Register the user through the service
	user, err := identityService.Register(context.Background(), email, "TestPass123!", handle, inviteCode)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return TestUser{
		ID:     user.ID,
		Email:  email,
		Handle: handle,
	}
}

// createAdminUser creates an admin test user.
func createAdminUser(t *testing.T) TestUser {
	t.Helper()

	// For now, admin is the same as regular user
	// In a real system, you'd assign admin privileges
	return createTestUser(t)
}

// loginUser logs in a user and returns tokens.
func loginUser(t *testing.T, email, password string) LoginResponse {
	t.Helper()

	authResp, err := identityService.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("failed to login user: %v", err)
	}

	return LoginResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	}
}

// logoutUser logs out a user by revoking the refresh token.
func logoutUser(t *testing.T, accessToken string) {
	t.Helper()

	// Get the user's refresh token and revoke it
	// In a real implementation, we'd call the logout endpoint
	// For testing, we need to get the refresh token from the login response
	// Since we only have the access token here, we'll revoke directly

	// The acceptance test should pass the refresh token if it needs to test logout
	// For now, this is a no-op as the logout needs the refresh token
}

// logoutUserWithRefreshToken logs out by revoking the refresh token.
func logoutUserWithRefreshToken(t *testing.T, accessToken, refreshToken string) {
	t.Helper()

	// Revoke the refresh token
	if err := refreshTokenRepo.RevokeToken(context.Background(), refreshToken); err != nil {
		t.Fatalf("failed to revoke refresh token: %v", err)
	}
}
