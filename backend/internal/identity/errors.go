package identity

import "errors"

// Sentinel errors for identity operations.
var (
	// User errors
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email already registered")

	// Password errors
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooWeak  = errors.New("password must contain at least one letter and one number")

	// Handle errors
	ErrHandleInvalidChars = errors.New("handle can only contain letters, numbers, and underscores")
	ErrHandleAlreadyTaken = errors.New("handle already taken")
	ErrHandleTooLong      = errors.New("handle must be 20 characters or less")
	ErrHandleTooShort     = errors.New("handle must be at least 3 characters")

	// Email errors
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// Invite errors
	ErrInviteNotFound    = errors.New("invite not found")
	ErrInvalidInviteCode = errors.New("invalid invite code")
	ErrInviteExpired     = errors.New("invite has expired")
	ErrInviteExhausted   = errors.New("invite has reached maximum uses")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("invalid token")

	// Authorization errors
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInsufficientRep     = errors.New("insufficient reputation for this action")
	ErrNotCommunityMember  = errors.New("not a member of this community")
	ErrNotResourceOwner    = errors.New("not the owner of this resource")
	ErrAdminRequired       = errors.New("admin privileges required")

	// Reputation errors
	ErrInvalidEventType    = errors.New("invalid reputation event type")
	ErrDuplicateEvent      = errors.New("reputation event already recorded")
	ErrInvalidPointsValue  = errors.New("invalid points value for event type")
	ErrSelfReputation      = errors.New("cannot modify own reputation")
)

// ReputationEventType defines valid reputation event types.
type ReputationEventType string

const (
	EventMessagePosted    ReputationEventType = "message_posted"
	EventMessageUpvoted   ReputationEventType = "message_upvoted"
	EventMessageDownvoted ReputationEventType = "message_downvoted"
	EventInviteUsed       ReputationEventType = "invite_used"
	EventReportedAbuse    ReputationEventType = "reported_abuse"
	EventModeratorAction  ReputationEventType = "moderator_action"
)

// ReputationPointLimits defines min/max points per event type.
var ReputationPointLimits = map[ReputationEventType]struct{ Min, Max int }{
	EventMessagePosted:    {Min: 1, Max: 5},
	EventMessageUpvoted:   {Min: 1, Max: 10},
	EventMessageDownvoted: {Min: -10, Max: -1},
	EventInviteUsed:       {Min: 5, Max: 20},
	EventReportedAbuse:    {Min: -50, Max: -10},
	EventModeratorAction:  {Min: -100, Max: 100},
}

// ValidateReputationEvent validates that the event type and points are valid.
func ValidateReputationEvent(eventType string, points int) error {
	repType := ReputationEventType(eventType)
	limits, ok := ReputationPointLimits[repType]
	if !ok {
		return ErrInvalidEventType
	}
	if points < limits.Min || points > limits.Max {
		return ErrInvalidPointsValue
	}
	return nil
}
