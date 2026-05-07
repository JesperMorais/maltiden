package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func TestJWTGenerateAndValidate(t *testing.T) {
	svc, err := NewJWTService(testSecret)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	token, err := svc.GenerateToken("usr_123", "hh_456", 1)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != "usr_123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "usr_123")
	}
	if claims.HouseholdID != "hh_456" {
		t.Errorf("HouseholdID = %q, want %q", claims.HouseholdID, "hh_456")
	}
	if claims.TokenVersion != 1 {
		t.Errorf("TokenVersion = %d, want 1", claims.TokenVersion)
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	signer, err := NewJWTService(testSecret)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}
	verifier, err := NewJWTService("different-secret")
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	token, err := signer.GenerateToken("usr_1", "hh_1", 0)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	if _, err := verifier.ValidateToken(token); err == nil {
		t.Fatal("expected error validating token with wrong secret, got nil")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	svc, err := NewJWTService(testSecret)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	expiredClaims := Claims{
		UserID:       "usr_exp",
		HouseholdID:  "hh_exp",
		TokenVersion: 0,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	signed, err := tok.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	if _, err := svc.ValidateToken(signed); err == nil {
		t.Fatal("expected error validating expired token, got nil")
	}
}

func TestJWTGenerateNonEmpty(t *testing.T) {
	svc, err := NewJWTService(testSecret)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	token, err := svc.GenerateToken("usr_x", "hh_y", 42)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}
