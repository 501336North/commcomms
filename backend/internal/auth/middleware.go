package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const userContextKey contextKey = "user_id"

// UserIDKey is exported for external access to user context values.
var UserIDKey = userContextKey

func AuthMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userContextKey).(string)
	if !ok {
		return "", errors.New("user not found in context")
	}
	return userID, nil
}
