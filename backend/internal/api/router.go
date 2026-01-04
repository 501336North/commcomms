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
	mux           *http.ServeMux
	authHandler   *handlers.AuthHandler
	userHandler   *handlers.UserHandler
	inviteHandler *handlers.InviteHandler
	jwtService    *auth.JWTService
}

// RouterConfig contains configuration for creating a new router.
type RouterConfig struct {
	AuthHandler   *handlers.AuthHandler
	UserHandler   *handlers.UserHandler
	InviteHandler *handlers.InviteHandler
	JWTService    *auth.JWTService
}

// NewRouter creates a new Router with the given configuration.
func NewRouter(config RouterConfig) *Router {
	r := &Router{
		mux:           http.NewServeMux(),
		authHandler:   config.AuthHandler,
		userHandler:   config.UserHandler,
		inviteHandler: config.InviteHandler,
		jwtService:    config.JWTService,
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
	// Public routes (no auth required)
	r.mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)
	r.mux.HandleFunc("POST /api/v1/auth/refresh", r.authHandler.Refresh)

	// Protected routes (auth required)
	r.mux.HandleFunc("POST /api/v1/auth/logout", r.withAuth(r.authHandler.Logout))
	r.mux.HandleFunc("GET /api/v1/users/me", r.withAuth(r.userHandler.GetProfile))
	r.mux.HandleFunc("GET /api/v1/users/me/reputation", r.withAuth(r.userHandler.GetReputation))

	// Community invite routes (auth required + community context)
	r.mux.HandleFunc("POST /api/v1/communities/{communityID}/invites", r.withAuth(r.withCommunity(r.inviteHandler.CreateInvite)))
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
