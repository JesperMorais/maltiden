package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

type Claims struct {
	UserID      string `json:"user_id"`
	HouseholdID string `json:"household_id"`
	jwt.RegisteredClaims
}

// InitJWTSecret validates and initializes the JWT secret.
// Must be called once at application startup before any JWT operations.
func InitJWTSecret(secret string) error {
	if len(secret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	jwtSecret = []byte(secret)
	return nil
}

func GenerateToken(userID, householdID string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret not initialized")
	}

	claims := Claims{
		UserID:      userID,
		HouseholdID: householdID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{},
		error) {
		// Validate algorithm to prevent "none" attack
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
