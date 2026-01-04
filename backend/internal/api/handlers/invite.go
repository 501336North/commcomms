package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/canary/commcomms/internal/auth"
	"github.com/canary/commcomms/internal/identity"
)

type contextKey string

// CommunityIDKey is the context key for community ID.
const CommunityIDKey contextKey = "community_id"

// InviteService defines the interface for invite operations.
type InviteService interface {
	CreateInvite(communityID, creatorID string, opts identity.InviteOptions) (*identity.Invite, error)
}

// InviteHandler handles invite-related HTTP requests.
type InviteHandler struct {
	inviteService InviteService
	baseURL       string
}

// NewInviteHandler creates a new InviteHandler.
func NewInviteHandler(inviteService InviteService, baseURL string) *InviteHandler {
	return &InviteHandler{
		inviteService: inviteService,
		baseURL:       baseURL,
	}
}

// CreateInviteRequest represents the create invite request body.
type CreateInviteRequest struct {
	ExpiresInDays int `json:"expiresInDays"`
	MaxUses       int `json:"maxUses"`
}

// CreateInviteResponse represents the create invite response body.
type CreateInviteResponse struct {
	Code      string `json:"code"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

// CreateInvite handles POST /api/v1/communities/:id/invites
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	communityID, ok := r.Context().Value(CommunityIDKey).(string)
	if !ok || communityID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Community ID is required")
		return
	}

	var req CreateInviteRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	}

	// Default to 7 days if not specified
	expiresInDays := req.ExpiresInDays
	if expiresInDays <= 0 {
		expiresInDays = 7
	}

	opts := identity.InviteOptions{
		ExpiresAt: time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour),
		MaxUses:   req.MaxUses,
	}

	invite, err := h.inviteService.CreateInvite(communityID, userID, opts)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create invite")
		return
	}

	resp := CreateInviteResponse{
		Code:      invite.Code,
		URL:       fmt.Sprintf("%s/invite/%s", h.baseURL, invite.Code),
		ExpiresAt: invite.ExpiresAt.Format(time.RFC3339),
	}

	writeJSONResponse(w, http.StatusCreated, resp)
}

// GetCommunityIDFromContext retrieves the community ID from context.
func GetCommunityIDFromContext(r *http.Request) (string, bool) {
	communityID, ok := r.Context().Value(CommunityIDKey).(string)
	return communityID, ok
}
