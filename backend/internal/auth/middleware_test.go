package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// TestGetClientIP_XForwardedFor_SingleIP tests that GetClientIP
// correctly extracts a single IP from X-Forwarded-For header.
func TestGetClientIP_XForwardedFor_SingleIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	// Act
	ip := GetClientIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip)
}

// TestGetClientIP_XForwardedFor_MultipleIPs tests that GetClientIP
// correctly extracts only the first IP from a comma-separated list.
func TestGetClientIP_XForwardedFor_MultipleIPs(t *testing.T) {
	// Arrange - multiple IPs in X-Forwarded-For (client, proxy1, proxy2)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1, 172.16.0.1")

	// Act
	ip := GetClientIP(req)

	// Assert - should only return the first IP (original client)
	assert.Equal(t, "192.168.1.100", ip)
}

// TestGetClientIP_XForwardedFor_WithSpaces tests that GetClientIP
// correctly trims whitespace from the extracted IP.
func TestGetClientIP_XForwardedFor_WithSpaces(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "  192.168.1.100  ,  10.0.0.1  ")

	// Act
	ip := GetClientIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip)
}

// TestGetClientIP_XRealIP tests that GetClientIP falls back to
// X-Real-IP when X-Forwarded-For is not present.
func TestGetClientIP_XRealIP(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "192.168.1.100")

	// Act
	ip := GetClientIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100", ip)
}

// TestGetClientIP_RemoteAddr tests that GetClientIP falls back to
// RemoteAddr when no proxy headers are present.
func TestGetClientIP_RemoteAddr(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// Act
	ip := GetClientIP(req)

	// Assert
	assert.Equal(t, "192.168.1.100:12345", ip)
}

// TestRateLimiter_AllowsWithinLimit tests that the rate limiter
// allows requests within the configured limit.
func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	// Arrange - 5 requests per minute
	limiter := NewRateLimiter(5, time.Minute)
	clientIP := "192.168.1.100"

	// Act & Assert - first 5 requests should be allowed
	for i := 0; i < 5; i++ {
		assert.True(t, limiter.Allow(clientIP), "request %d should be allowed", i+1)
	}
}

// TestRateLimiter_BlocksOverLimit tests that the rate limiter
// blocks requests that exceed the configured limit.
func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	// Arrange - 3 requests per minute
	limiter := NewRateLimiter(3, time.Minute)
	clientIP := "192.168.1.100"

	// Exhaust the burst capacity (2x rate = 6)
	for i := 0; i < 6; i++ {
		limiter.Allow(clientIP)
	}

	// Act & Assert - next request should be blocked
	assert.False(t, limiter.Allow(clientIP), "request over limit should be blocked")
}

// TestRateLimiter_SeparatesClients tests that the rate limiter
// tracks limits separately for different client IPs.
func TestRateLimiter_SeparatesClients(t *testing.T) {
	// Arrange - 2 requests per minute
	limiter := NewRateLimiter(2, time.Minute)
	client1 := "192.168.1.100"
	client2 := "192.168.1.200"

	// Exhaust client1's limit
	for i := 0; i < 4; i++ { // burst capacity = 2*2 = 4
		limiter.Allow(client1)
	}

	// Act & Assert - client2 should still be allowed
	assert.True(t, limiter.Allow(client2), "different client should have separate limit")
	assert.False(t, limiter.Allow(client1), "client1 should be blocked")
}

// TestRateLimitMiddleware_RejectsOverLimit tests that the rate limit
// middleware returns 429 Too Many Requests when limit is exceeded.
func TestRateLimitMiddleware_RejectsOverLimit(t *testing.T) {
	// Arrange - 1 request per minute with burst of 2
	limiter := NewRateLimiter(1, time.Minute)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RateLimitMiddleware(limiter, GetClientIP)
	handler := middleware(nextHandler)

	// Make 3 requests (burst capacity is 2)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Act - third request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Contains(t, rr.Body.String(), "Rate limit exceeded")
}
