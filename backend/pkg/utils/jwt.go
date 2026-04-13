package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       string `json:"user_id"`
	HouseholdID  string `json:"household_id"`
	TokenVersion int    `json:"token_version"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
}

// NewJWTService creates a new JWT service with the provided secret.
// Requires a non-empty secret. In production (FLY_APP_NAME set), enforces 32+ characters.
func NewJWTService(secret string) (*JWTService, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if os.Getenv("FLY_APP_NAME") != "" && len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters in production")
	}
	return &JWTService{secret: []byte(secret)}, nil
}

func (s *JWTService) GenerateToken(userID, householdID string, tokenVersion int) (string, error) {
	claims := Claims{
		UserID:       userID,
		HouseholdID:  householdID,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{},
		error) {
		// Validate algorithm to prevent "none" attack
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
