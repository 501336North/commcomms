package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/canary/commcomms/internal/identity"
)

// IdentityService defines the interface for identity operations.
type IdentityService interface {
	Register(ctx context.Context, email, password, handle, inviteCode string) (*identity.User, error)
	Login(ctx context.Context, email, password string) (*identity.AuthResponse, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*identity.AuthResponse, error)
	GetUserByID(ctx context.Context, userID string) (*identity.User, error)
}

// TokenService defines the interface for token generation.
type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
}

// LogoutService defines the interface for token revocation.
type LogoutService interface {
	RevokeToken(ctx context.Context, token string) error
}

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	identityService IdentityService
	tokenService    TokenService
	logoutService   LogoutService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(identityService IdentityService, tokenService TokenService, logoutService LogoutService) *AuthHandler {
	return &AuthHandler{
		identityService: identityService,
		tokenService:    tokenService,
		logoutService:   logoutService,
	}
}

// RegisterRequest represents the registration request body.
type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Handle     string `json:"handle"`
	InviteCode string `json:"inviteCode"`
}

// RegisterResponse represents the registration response body.
type RegisterResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID         string `json:"id"`
	Handle     string `json:"handle"`
	Email      string `json:"email,omitempty"`
	Reputation int    `json:"reputation"`
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response body.
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// RefreshRequest represents the refresh token request body.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse represents the refresh token response body.
type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// LogoutRequest represents the logout request body.
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.identityService.Register(r.Context(), req.Email, req.Password, req.Handle, req.InviteCode)
	if err != nil {
		h.handleRegistrationError(w, err)
		return
	}

	// Generate tokens for the newly registered user
	accessToken, err := h.tokenService.GenerateAccessToken(user.ID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	resp := RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:         user.ID,
			Handle:     user.Handle,
			Reputation: user.Reputation,
		},
	}

	writeJSONResponse(w, http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	authResp, err := h.identityService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, identity.ErrInvalidCredentials) {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	resp := LoginResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	authResp, err := h.identityService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, identity.ErrTokenRevoked) {
			writeErrorResponse(w, http.StatusUnauthorized, "Token has been revoked")
			return
		}
		if errors.Is(err, identity.ErrTokenExpired) {
			writeErrorResponse(w, http.StatusUnauthorized, "Token has expired")
			return
		}
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	resp := RefreshResponse{
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Check for Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeErrorResponse(w, http.StatusUnauthorized, "Missing or invalid authorization header")
		return
	}

	var req LogoutRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	if req.RefreshToken != "" && h.logoutService != nil {
		if err := h.logoutService.RevokeToken(r.Context(), req.RefreshToken); err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to revoke token")
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// handleRegistrationError maps registration errors to HTTP responses.
func (h *AuthHandler) handleRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrEmailAlreadyRegistered):
		writeErrorResponse(w, http.StatusConflict, "Email already registered")
	case errors.Is(err, identity.ErrHandleAlreadyTaken):
		writeErrorResponse(w, http.StatusConflict, "Handle already taken")
	case errors.Is(err, identity.ErrPasswordTooShort):
		writeErrorResponse(w, http.StatusBadRequest, "Password must be at least 8 characters")
	case errors.Is(err, identity.ErrPasswordTooWeak):
		writeErrorResponse(w, http.StatusBadRequest, "Password must contain at least one letter and one number")
	case errors.Is(err, identity.ErrInvalidInviteCode):
		writeErrorResponse(w, http.StatusBadRequest, "Invalid invite code")
	case errors.Is(err, identity.ErrInviteExpired):
		writeErrorResponse(w, http.StatusBadRequest, "Invite has expired")
	case errors.Is(err, identity.ErrInviteExhausted):
		writeErrorResponse(w, http.StatusBadRequest, "Invite has been exhausted")
	case errors.Is(err, identity.ErrHandleInvalidChars):
		writeErrorResponse(w, http.StatusBadRequest, "Handle can only contain letters, numbers, and underscores")
	case errors.Is(err, identity.ErrHandleTooLong):
		writeErrorResponse(w, http.StatusBadRequest, "Handle must be 20 characters or less")
	case errors.Is(err, identity.ErrHandleTooShort):
		writeErrorResponse(w, http.StatusBadRequest, "Handle must be at least 3 characters")
	case errors.Is(err, identity.ErrInvalidEmailFormat):
		writeErrorResponse(w, http.StatusBadRequest, "Invalid email format")
	default:
		writeErrorResponse(w, http.StatusInternalServerError, "Registration failed")
	}
}

// writeJSONResponse writes a JSON response with the given status code.
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response with the given status code.
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
