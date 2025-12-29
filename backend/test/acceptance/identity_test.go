// Package acceptance contains acceptance tests for the CommComms API.
// These tests operate at the HTTP boundary (black-box testing).
// They test WHAT the system does, not HOW it does it.
package acceptance

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test HTTP server for acceptance tests.
// This will be initialized in TestMain once the actual server is implemented.
var TestServer *httptest.Server

// ============================================
// US-ID-001: User Registration
// ============================================

// TestUserRegistration_Acceptance tests the user registration flow at the API boundary.
//
// User Story: As a guest, I want to register with my email so that I can join communities.
//
// Acceptance Criteria:
// - AC-ID-001.1: Valid registration returns user ID and adds to community
// - AC-ID-001.2: No phone number required
// - AC-ID-001.3: Password must be >= 8 characters
// - AC-ID-001.4: Email must be unique
func TestUserRegistration_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-ID-001.1: should register user with valid email and invite", func(t *testing.T) {
		// GIVEN - A valid invite code exists
		inviteCode := createTestInvite(t)

		// WHEN - I register with valid credentials
		reqBody := map[string]string{
			"email":      "newuser@example.com",
			"password":   "SecurePass123!",
			"handle":     "newuser",
			"inviteCode": inviteCode,
		}
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should receive confirmation with user ID
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.NotEmpty(t, body["accessToken"])
		assert.NotEmpty(t, body["refreshToken"])
		assert.NotEmpty(t, body["user"])

		user := body["user"].(map[string]interface{})
		assert.NotEmpty(t, user["id"])
		assert.Equal(t, "newuser", user["handle"])
		assert.Equal(t, float64(0), user["reputation"])
	})

	t.Run("AC-ID-001.2: should not require phone number", func(t *testing.T) {
		// GIVEN - Registration data without phone number
		inviteCode := createTestInvite(t)
		reqBody := map[string]string{
			"email":      "nophone@example.com",
			"password":   "SecurePass123!",
			"handle":     "nophone",
			"inviteCode": inviteCode,
		}

		// WHEN - I register without phone
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - Registration should succeed
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("AC-ID-001.3: should reject password less than 8 characters", func(t *testing.T) {
		// GIVEN - A password that's too short
		inviteCode := createTestInvite(t)
		reqBody := map[string]string{
			"email":      "shortpass@example.com",
			"password":   "short",
			"handle":     "shortpass",
			"inviteCode": inviteCode,
		}

		// WHEN - I try to register
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should see a password validation error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "8 characters")
	})

	t.Run("AC-ID-001.4: should reject duplicate email", func(t *testing.T) {
		// GIVEN - A user already exists with this email
		inviteCode := createTestInvite(t)
		firstReq := map[string]string{
			"email":      "duplicate@example.com",
			"password":   "FirstPass123!",
			"handle":     "firstuser",
			"inviteCode": inviteCode,
		}
		resp1 := postJSON(t, "/api/v1/auth/register", firstReq)
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// WHEN - I try to register with the same email
		secondInvite := createTestInvite(t)
		secondReq := map[string]string{
			"email":      "duplicate@example.com",
			"password":   "SecondPass123!",
			"handle":     "seconduser",
			"inviteCode": secondInvite,
		}
		resp2 := postJSON(t, "/api/v1/auth/register", secondReq)

		// THEN - I should see an error about duplicate email
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&body)
		assert.Contains(t, body["error"], "already registered")
	})

	t.Run("should reject registration without valid invite", func(t *testing.T) {
		// GIVEN - An invalid invite code
		reqBody := map[string]string{
			"email":      "noinvite@example.com",
			"password":   "SecurePass123!",
			"handle":     "noinvite",
			"inviteCode": "INVALID_CODE",
		}

		// WHEN - I try to register
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should see an invite error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "invite")
	})
}

// ============================================
// US-ID-002: Pseudonymous Profile
// ============================================

// TestHandleValidation_Acceptance tests handle creation and validation.
//
// User Story: As a new member, I want to create a pseudonymous handle
// so that I control my identity.
func TestHandleValidation_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-ID-002.1: should accept valid handle with letters and underscores", func(t *testing.T) {
		// GIVEN - A valid handle format
		inviteCode := createTestInvite(t)
		reqBody := map[string]string{
			"email":      "validhandle@example.com",
			"password":   "SecurePass123!",
			"handle":     "valid_user_123",
			"inviteCode": inviteCode,
		}

		// WHEN - I register
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - Registration should succeed
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("AC-ID-002.2: should reject duplicate handle", func(t *testing.T) {
		// GIVEN - A handle that already exists
		inviteCode := createTestInvite(t)
		firstReq := map[string]string{
			"email":      "first@example.com",
			"password":   "SecurePass123!",
			"handle":     "taken_handle",
			"inviteCode": inviteCode,
		}
		resp1 := postJSON(t, "/api/v1/auth/register", firstReq)
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// WHEN - I try to use the same handle
		secondInvite := createTestInvite(t)
		secondReq := map[string]string{
			"email":      "second@example.com",
			"password":   "SecurePass123!",
			"handle":     "taken_handle",
			"inviteCode": secondInvite,
		}
		resp2 := postJSON(t, "/api/v1/auth/register", secondReq)

		// THEN - I should see an error
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&body)
		assert.Contains(t, body["error"], "Handle already taken")
	})

	t.Run("AC-ID-002.3: should reject handle with spaces", func(t *testing.T) {
		// GIVEN - A handle with spaces
		inviteCode := createTestInvite(t)
		reqBody := map[string]string{
			"email":      "spaces@example.com",
			"password":   "SecurePass123!",
			"handle":     "invalid handle",
			"inviteCode": inviteCode,
		}

		// WHEN - I try to register
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should see a validation error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "letters, numbers, underscores")
	})

	t.Run("should reject handle over 20 characters", func(t *testing.T) {
		// GIVEN - A handle that's too long
		inviteCode := createTestInvite(t)
		reqBody := map[string]string{
			"email":      "longhandle@example.com",
			"password":   "SecurePass123!",
			"handle":     "this_handle_is_way_too_long",
			"inviteCode": inviteCode,
		}

		// WHEN - I try to register
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should see a validation error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "20 characters")
	})
}

// ============================================
// User Login
// ============================================

// TestUserLogin_Acceptance tests user authentication.
//
// User Story: As a registered user, I want to login with my credentials
// so that I can access the platform.
func TestUserLogin_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should login with valid credentials", func(t *testing.T) {
		// GIVEN - A registered user
		user := createTestUser(t)

		// WHEN - I login with correct credentials
		reqBody := map[string]string{
			"email":    user.Email,
			"password": "TestPass123!",
		}
		resp := postJSON(t, "/api/v1/auth/login", reqBody)

		// THEN - I should receive tokens
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["accessToken"])
		assert.NotEmpty(t, body["refreshToken"])
		assert.NotEmpty(t, body["expiresIn"])
	})

	t.Run("should reject invalid password", func(t *testing.T) {
		// GIVEN - A registered user
		user := createTestUser(t)

		// WHEN - I login with wrong password
		reqBody := map[string]string{
			"email":    user.Email,
			"password": "WrongPassword!",
		}
		resp := postJSON(t, "/api/v1/auth/login", reqBody)

		// THEN - I should see an error (generic for security)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "Invalid credentials")
	})

	t.Run("should reject non-existent email", func(t *testing.T) {
		// GIVEN - An email that doesn't exist

		// WHEN - I try to login
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "AnyPassword!",
		}
		resp := postJSON(t, "/api/v1/auth/login", reqBody)

		// THEN - I should see same generic error (for security)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "Invalid credentials")
	})
}

// ============================================
// Token Refresh
// ============================================

// TestTokenRefresh_Acceptance tests token refresh flow.
func TestTokenRefresh_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should issue new tokens with valid refresh token", func(t *testing.T) {
		// GIVEN - A logged in user with refresh token
		user := createTestUser(t)
		loginResp := loginUser(t, user.Email, "TestPass123!")
		refreshToken := loginResp.RefreshToken

		// WHEN - I refresh the tokens
		reqBody := map[string]string{
			"refreshToken": refreshToken,
		}
		resp := postJSON(t, "/api/v1/auth/refresh", reqBody)

		// THEN - I should receive new tokens
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["accessToken"])
		assert.NotEmpty(t, body["refreshToken"])
		// New refresh token should be different
		assert.NotEqual(t, refreshToken, body["refreshToken"])
	})

	t.Run("should reject revoked refresh token", func(t *testing.T) {
		// GIVEN - A user who logged out
		user := createTestUser(t)
		loginResp := loginUser(t, user.Email, "TestPass123!")
		refreshToken := loginResp.RefreshToken

		// Logout to revoke the token
		logoutUser(t, loginResp.AccessToken)

		// WHEN - I try to use the old refresh token
		reqBody := map[string]string{
			"refreshToken": refreshToken,
		}
		resp := postJSON(t, "/api/v1/auth/refresh", reqBody)

		// THEN - I should see an error
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "revoked")
	})
}

// ============================================
// Protected Routes
// ============================================

// TestProtectedRoutes_Acceptance tests authentication middleware.
func TestProtectedRoutes_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("should reject request without token", func(t *testing.T) {
		// WHEN - I request a protected resource without auth
		resp := getJSON(t, "/api/v1/users/me", "")

		// THEN - I should get 401
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should reject request with invalid token", func(t *testing.T) {
		// WHEN - I request with an invalid token
		resp := getJSON(t, "/api/v1/users/me", "invalid.jwt.token")

		// THEN - I should get 401
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("should accept request with valid token", func(t *testing.T) {
		// GIVEN - A logged in user
		user := createTestUser(t)
		loginResp := loginUser(t, user.Email, "TestPass123!")

		// WHEN - I request with valid token
		resp := getJSON(t, "/api/v1/users/me", loginResp.AccessToken)

		// THEN - I should get my profile
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, user.Handle, body["handle"])
	})
}

// ============================================
// US-ID-004: Invite-Only Access
// ============================================

// TestInviteManagement_Acceptance tests invite link functionality.
//
// User Story: As an admin, I want to control who joins my community
// so that I maintain quality.
func TestInviteManagement_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-ID-004.1: should generate unique invite code", func(t *testing.T) {
		// GIVEN - An admin user
		admin := createAdminUser(t)
		token := loginUser(t, admin.Email, "TestPass123!").AccessToken

		// WHEN - I generate an invite
		reqBody := map[string]interface{}{
			"expiresInDays": 7,
		}
		resp := postJSONAuth(t, "/api/v1/communities/test-community/invites", reqBody, token)

		// THEN - I should receive an invite with code and URL
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.NotEmpty(t, body["code"])
		assert.NotEmpty(t, body["url"])
		assert.NotEmpty(t, body["expiresAt"])
	})

	t.Run("AC-ID-004.2: should reject expired invite", func(t *testing.T) {
		// GIVEN - An expired invite code
		expiredCode := createExpiredInvite(t)

		// WHEN - I try to register with it
		reqBody := map[string]string{
			"email":      "expired@example.com",
			"password":   "SecurePass123!",
			"handle":     "expired",
			"inviteCode": expiredCode,
		}
		resp := postJSON(t, "/api/v1/auth/register", reqBody)

		// THEN - I should see an expiration error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body["error"], "expired")
	})

	t.Run("should reject exhausted invite", func(t *testing.T) {
		// GIVEN - An invite with maxUses = 1 that's been used
		inviteCode := createLimitedInvite(t, 1)

		// First use should succeed
		firstReq := map[string]string{
			"email":      "first@example.com",
			"password":   "SecurePass123!",
			"handle":     "firstuse",
			"inviteCode": inviteCode,
		}
		resp1 := postJSON(t, "/api/v1/auth/register", firstReq)
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// WHEN - Second user tries to use same invite
		secondReq := map[string]string{
			"email":      "second@example.com",
			"password":   "SecurePass123!",
			"handle":     "seconduse",
			"inviteCode": inviteCode,
		}
		resp2 := postJSON(t, "/api/v1/auth/register", secondReq)

		// THEN - Should be rejected
		assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&body)
		assert.Contains(t, body["error"], "exhausted")
	})
}

// ============================================
// US-ID-003: Reputation Tracking
// ============================================

// TestReputation_Acceptance tests reputation initialization and display.
func TestReputation_Acceptance(t *testing.T) {
	t.Skip("Skipping until server implementation exists")

	t.Run("AC-ID-003.1: should initialize reputation to 0", func(t *testing.T) {
		// GIVEN - I just registered
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken

		// WHEN - I view my profile
		resp := getJSON(t, "/api/v1/users/me", token)

		// THEN - My reputation should be 0
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, float64(0), body["reputation"])
	})

	t.Run("AC-ID-003.2: should display reputation details", func(t *testing.T) {
		// GIVEN - A user with some reputation
		user := createTestUser(t)
		token := loginUser(t, user.Email, "TestPass123!").AccessToken

		// WHEN - I view reputation details
		resp := getJSON(t, "/api/v1/users/me/reputation", token)

		// THEN - I should see breakdown
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Contains(t, body, "total")
		assert.Contains(t, body, "breakdown")
	})
}

// ============================================
// Helper Functions
// ============================================

// postJSON sends a POST request with JSON body.
func postJSON(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, TestServer.URL+path, bytes.NewReader(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// postJSONAuth sends an authenticated POST request.
func postJSONAuth(t *testing.T, path string, body interface{}, token string) *http.Response {
	t.Helper()

	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, TestServer.URL+path, bytes.NewReader(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// getJSON sends a GET request with optional auth token.
func getJSON(t *testing.T, path string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, TestServer.URL+path, nil)
	require.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

// TestUser represents a test user.
type TestUser struct {
	ID     string
	Email  string
	Handle string
}

// LoginResponse represents login response.
type LoginResponse struct {
	AccessToken  string
	RefreshToken string
}

// createTestInvite creates a test invite and returns the code.
func createTestInvite(t *testing.T) string {
	t.Helper()
	// TODO: Implement when server exists
	return "TEST_INVITE_CODE"
}

// createExpiredInvite creates an expired invite for testing.
func createExpiredInvite(t *testing.T) string {
	t.Helper()
	// TODO: Implement when server exists
	return "EXPIRED_INVITE_CODE"
}

// createLimitedInvite creates an invite with max uses.
func createLimitedInvite(t *testing.T, maxUses int) string {
	t.Helper()
	// TODO: Implement when server exists
	return "LIMITED_INVITE_CODE"
}

// createTestUser creates a test user and returns it.
func createTestUser(t *testing.T) TestUser {
	t.Helper()
	// TODO: Implement when server exists
	return TestUser{
		ID:     "test-user-id",
		Email:  "testuser@example.com",
		Handle: "testuser",
	}
}

// createAdminUser creates an admin test user.
func createAdminUser(t *testing.T) TestUser {
	t.Helper()
	// TODO: Implement when server exists
	return TestUser{
		ID:     "admin-user-id",
		Email:  "admin@example.com",
		Handle: "admin",
	}
}

// loginUser logs in a user and returns tokens.
func loginUser(t *testing.T, email, password string) LoginResponse {
	t.Helper()
	// TODO: Implement when server exists
	return LoginResponse{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
	}
}

// logoutUser logs out a user.
func logoutUser(t *testing.T, token string) {
	t.Helper()
	// TODO: Implement when server exists
}
