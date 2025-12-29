package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string
	ExpiresAt time.Time
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	return s.generateTokenWithExpiry(userID, 24*time.Hour)
}

func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	return s.generateTokenWithExpiry(userID, 7*24*time.Hour)
}

func (s *JWTService) generateTokenWithExpiry(userID string, duration time.Duration) (string, error) {
	expiresAt := time.Now().Add(duration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
	})
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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
	claims := token.Claims.(jwt.MapClaims)
	exp, _ := claims.GetExpirationTime()
	return &Claims{
		UserID:    claims["user_id"].(string),
		ExpiresAt: exp.Time,
	}, nil
}
