package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthMiddleware_ValidToken tests that the middleware allows requests
// with a valid Authorization: Bearer token header and sets the user ID in context.
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"
	userID := "user-12345"

	jwtService := NewJWTService(jwtSecret)
	token, err := jwtService.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Create a handler that checks the user ID in context
	var capturedUserID string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = GetUserFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(jwtService)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, userID, capturedUserID)
}

// TestAuthMiddleware_NoToken tests that the middleware rejects requests
// without an Authorization header with 401 Unauthorized.
func TestAuthMiddleware_NoToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"

	jwtService := NewJWTService(jwtSecret)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	middleware := AuthMiddleware(jwtService)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestAuthMiddleware_InvalidToken tests that the middleware rejects requests
// with a malformed or invalid token with 401 Unauthorized.
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"

	jwtService := NewJWTService(jwtSecret)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})

	middleware := AuthMiddleware(jwtService)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	rr := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestGetUserFromContext_ValidContext tests that GetUserFromContext
// returns the user ID when it exists in the context.
func TestGetUserFromContext_ValidContext(t *testing.T) {
	// Arrange
	expectedUserID := "user-12345"
	ctx := context.WithValue(context.Background(), userContextKey, expectedUserID)

	// Act
	userID, err := GetUserFromContext(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedUserID, userID)
}

// TestGetUserFromContext_NoUser tests that GetUserFromContext
// returns an error when no user ID exists in the context.
func TestGetUserFromContext_NoUser(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	userID, err := GetUserFromContext(ctx)

	// Assert
	require.Error(t, err)
	assert.Empty(t, userID)
}
