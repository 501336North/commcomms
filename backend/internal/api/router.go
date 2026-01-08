package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/canary/commcomms/internal/api/handlers"
	"github.com/canary/commcomms/internal/auth"
)

// Router handles HTTP routing for the API.
type Router struct {
	mux              *http.ServeMux
	authHandler      *handlers.AuthHandler
	userHandler      *handlers.UserHandler
	inviteHandler    *handlers.InviteHandler
	jwtService       *auth.JWTService
	membershipChecker MembershipChecker
}

// MembershipChecker verifies community membership.
type MembershipChecker interface {
	IsMember(ctx context.Context, communityID, userID string) (bool, error)
}

// RouterConfig contains configuration for creating a new router.
type RouterConfig struct {
	AuthHandler       *handlers.AuthHandler
	UserHandler       *handlers.UserHandler
	InviteHandler     *handlers.InviteHandler
	JWTService        *auth.JWTService
	MembershipChecker MembershipChecker
}

// NewRouter creates a new Router with the given configuration.
func NewRouter(config RouterConfig) *Router {
	r := &Router{
		mux:               http.NewServeMux(),
		authHandler:       config.AuthHandler,
		userHandler:       config.UserHandler,
		inviteHandler:     config.InviteHandler,
		jwtService:        config.JWTService,
		membershipChecker: config.MembershipChecker,
	}
	r.setupRoutes()
	return r
}

// ServeHTTP implements the http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Wrap with request ID middleware
	RequestIDMiddleware(r.mux).ServeHTTP(w, req)
}

// setupRoutes configures all routes.
func (r *Router) setupRoutes() {
	// Public routes (no auth required) - with specific rate limiters
	r.mux.HandleFunc("POST /api/v1/auth/register", r.withRateLimit(auth.RegisterRateLimiter, r.authHandler.Register))
	r.mux.HandleFunc("POST /api/v1/auth/login", r.withRateLimit(auth.LoginRateLimiter, r.authHandler.Login))
	r.mux.HandleFunc("POST /api/v1/auth/refresh", r.authHandler.Refresh)

	// Protected routes (auth required)
	r.mux.HandleFunc("POST /api/v1/auth/logout", r.withAuth(r.authHandler.Logout))
	r.mux.HandleFunc("GET /api/v1/users/me", r.withAuth(r.userHandler.GetProfile))
	r.mux.HandleFunc("GET /api/v1/users/me/reputation", r.withAuth(r.userHandler.GetReputation))

	// Community invite routes (auth required + community context + membership check)
	r.mux.HandleFunc("POST /api/v1/communities/{communityID}/invites", r.withAuth(r.withCommunity(r.withMembership(r.inviteHandler.CreateInvite))))
}

// withAuth wraps a handler with authentication middleware.
func (r *Router) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := r.jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), auth.UserIDKey, claims.UserID)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
}

// withCommunity extracts community ID from path and adds to context.
func (r *Router) withCommunity(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		communityID := req.PathValue("communityID")
		if communityID == "" {
			http.Error(w, `{"error":"Community ID is required"}`, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(req.Context(), handlers.CommunityIDKey, communityID)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
}

// withRateLimit wraps a handler with rate limiting middleware.
func (r *Router) withRateLimit(limiter *auth.RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		key := auth.GetClientIP(req)
		if !limiter.Allow(key) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, req)
	}
}

// withMembership verifies the user is a member of the community.
func (r *Router) withMembership(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Get user ID from context (set by withAuth)
		userID, ok := req.Context().Value(auth.UserIDKey).(string)
		if !ok || userID == "" {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Get community ID from context (set by withCommunity)
		communityID, ok := req.Context().Value(handlers.CommunityIDKey).(string)
		if !ok || communityID == "" {
			http.Error(w, `{"error":"Community ID is required"}`, http.StatusBadRequest)
			return
		}

		// Check membership if checker is available
		if r.membershipChecker != nil {
			isMember, err := r.membershipChecker.IsMember(req.Context(), communityID, userID)
			if err != nil {
				http.Error(w, `{"error":"Failed to verify membership"}`, http.StatusInternalServerError)
				return
			}
			if !isMember {
				http.Error(w, `{"error":"Not a member of this community"}`, http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, req)
	}
}
