package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID    string
	ExpiresAt time.Time
	IssuedAt  time.Time
	TokenID   string
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret []byte
	issuer string
}

// NewJWTService creates a new JWTService with the given secret.
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: []byte(secret),
		issuer: "commcomms",
	}
}

// GenerateAccessToken generates a short-lived access token (15 minutes).
func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	return s.generateTokenWithExpiry(userID, 15*time.Minute)
}

// GenerateRefreshToken generates a longer-lived refresh token (7 days).
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	return s.generateTokenWithExpiry(userID, 7*24*time.Hour)
}

func (s *JWTService) generateTokenWithExpiry(userID string, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)
	tokenID := uuid.New().String()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"iss":     s.issuer,
		"aud":     "commcomms-api",
		"jti":     tokenID,
	})
	return token.SignedString(s.secret)
}

// ValidateToken validates a JWT token and returns its claims.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm to prevent algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Safe type assertion with error checking
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract user_id with type checking
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return nil, errors.New("invalid user_id claim")
	}

	// Extract expiration time
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil, errors.New("invalid expiration claim")
	}

	// Extract issued at time
	iat, err := claims.GetIssuedAt()
	if err != nil {
		return nil, errors.New("invalid issued at claim")
	}

	// Extract token ID (optional for backwards compatibility)
	tokenID, _ := claims["jti"].(string)

	return &Claims{
		UserID:    userID,
		ExpiresAt: exp.Time,
		IssuedAt:  iat.Time,
		TokenID:   tokenID,
	}, nil
}
