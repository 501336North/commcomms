package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateAccessToken_ValidClaims tests that GenerateAccessToken generates a valid
// access token containing the user_id claim and expires in 15 minutes.
func TestGenerateAccessToken_ValidClaims(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"
	userID := "user-12345"

	tokenService := NewJWTService(jwtSecret)

	// Act
	token, err := tokenService.GenerateAccessToken(userID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token and check claims
	claims, err := tokenService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Verify expiry is approximately 15 minutes from now (security best practice)
	expectedExpiry := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt, 5*time.Second)
}

// TestGenerateRefreshToken_7DayExpiry tests that GenerateRefreshToken generates a valid
// refresh token with a 7 day expiry period.
func TestGenerateRefreshToken_7DayExpiry(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"
	userID := "user-12345"

	tokenService := NewJWTService(jwtSecret)

	// Act
	token, err := tokenService.GenerateRefreshToken(userID)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token and check claims
	claims, err := tokenService.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Verify expiry is approximately 7 days from now
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt, 5*time.Second)
}

// TestValidateToken_ValidSignature tests that ValidateToken correctly validates
// a token with a valid signature and returns the correct claims.
func TestValidateToken_ValidSignature(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"
	userID := "user-67890"

	tokenService := NewJWTService(jwtSecret)

	// Generate a token
	token, err := tokenService.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Act
	claims, err := tokenService.ValidateToken(token)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

// TestValidateToken_InvalidSignature tests that ValidateToken rejects tokens
// signed with a different secret key.
func TestValidateToken_InvalidSignature(t *testing.T) {
	// Arrange
	jwtSecret1 := "original-secret-key"
	jwtSecret2 := "different-secret-key"
	userID := "user-12345"

	tokenService1 := NewJWTService(jwtSecret1)
	tokenService2 := NewJWTService(jwtSecret2)

	// Generate a token with the first secret
	token, err := tokenService1.GenerateAccessToken(userID)
	require.NoError(t, err)

	// Act - validate with a different secret
	claims, err := tokenService2.ValidateToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid")
}

// TestValidateToken_Expired tests that ValidateToken rejects tokens that have expired.
func TestValidateToken_Expired(t *testing.T) {
	// Arrange
	jwtSecret := "test-secret-key-for-jwt-signing"
	userID := "user-12345"

	tokenService := NewJWTService(jwtSecret)

	// Generate an expired token (negative duration)
	token, err := tokenService.generateTokenWithExpiry(userID, -1*time.Hour)
	require.NoError(t, err)

	// Act
	claims, err := tokenService.ValidateToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}
