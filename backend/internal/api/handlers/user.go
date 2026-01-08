package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/canary/commcomms/internal/auth"
	"github.com/canary/commcomms/internal/identity"
)

// UserService defines the interface for user operations.
type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*identity.User, error)
}

// ReputationBreakdownItem represents a breakdown of reputation by event type.
type ReputationBreakdownItem struct {
	EventType string `json:"eventType"`
	Points    int    `json:"points"`
	Count     int    `json:"count"`
}

// ReputationService defines the interface for reputation operations.
type ReputationService interface {
	GetReputation(ctx context.Context, userID string) (int, error)
	GetReputationBreakdown(ctx context.Context, userID string) ([]ReputationBreakdownItem, error)
}

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService       UserService
	reputationService ReputationService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService UserService, reputationService ReputationService) *UserHandler {
	return &UserHandler{
		userService:       userService,
		reputationService: reputationService,
	}
}

// ProfileResponse represents the user profile response.
type ProfileResponse struct {
	ID         string `json:"id"`
	Handle     string `json:"handle"`
	Email      string `json:"email"`
	Reputation int    `json:"reputation"`
}

// ReputationResponse represents the reputation details response.
type ReputationResponse struct {
	Total     int                       `json:"total"`
	Breakdown []ReputationBreakdownItem `json:"breakdown"`
}

// GetProfile handles GET /api/v1/users/me
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			writeErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user profile")
		return
	}

	resp := ProfileResponse{
		ID:         user.ID,
		Handle:     user.Handle,
		Email:      user.Email,
		Reputation: user.Reputation,
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// GetReputation handles GET /api/v1/users/me/reputation
func (h *UserHandler) GetReputation(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	total, err := h.reputationService.GetReputation(r.Context(), userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reputation")
		return
	}

	breakdown, err := h.reputationService.GetReputationBreakdown(r.Context(), userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get reputation breakdown")
		return
	}

	resp := ReputationResponse{
		Total:     total,
		Breakdown: breakdown,
	}

	writeJSONResponse(w, http.StatusOK, resp)
}
