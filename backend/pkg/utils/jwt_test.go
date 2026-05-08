package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret      = "this-is-a-32-byte-dev-test-secret-xx"
	testOtherSecret = "another-32-byte-dev-test-secret-xxxxx"
)

func newTestService(t *testing.T, secret string) *JWTService {
	t.Helper()
	svc, err := NewJWTService(secret)
	if err != nil {
		t.Fatalf("NewJWTService(%q) returned error: %v", secret, err)
	}
	if svc == nil {
		t.Fatalf("NewJWTService(%q) returned nil service", secret)
	}
	return svc
}

func TestJWT_Generate_ReturnsNonEmpty(t *testing.T) {
	svc := newTestService(t, testSecret)

	token, err := svc.GenerateToken("usr_123", "hh_456", 1)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty string")
	}
}

func TestJWT_Validate_RoundTrip(t *testing.T) {
	svc := newTestService(t, testSecret)

	const (
		userID       = "usr_round_trip"
		householdID  = "hh_round_trip"
		tokenVersion = 7
	)

	token, err := svc.GenerateToken(userID, householdID, tokenVersion)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims == nil {
		t.Fatal("ValidateToken returned nil claims")
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.HouseholdID != householdID {
		t.Errorf("HouseholdID = %q, want %q", claims.HouseholdID, householdID)
	}
	if claims.TokenVersion != tokenVersion {
		t.Errorf("TokenVersion = %d, want %d", claims.TokenVersion, tokenVersion)
	}
}

func TestJWT_Validate_Expired(t *testing.T) {
	expiredClaims := Claims{
		UserID:       "usr_expired",
		HouseholdID:  "hh_expired",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	svc := newTestService(t, testSecret)
	if _, err := svc.ValidateToken(signed); err == nil {
		t.Fatal("ValidateToken accepted expired token, expected error")
	}
}

func TestJWT_Validate_WrongSecret(t *testing.T) {
	svcA := newTestService(t, testSecret)
	svcB := newTestService(t, testOtherSecret)

	token, err := svcA.GenerateToken("usr_sig", "hh_sig", 1)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := svcB.ValidateToken(token); err == nil {
		t.Fatal("ValidateToken accepted token signed with different secret, expected error")
	}
}

func TestJWT_Validate_RejectsNoneAlg(t *testing.T) {
	noneClaims := Claims{
		UserID:       "usr_none",
		HouseholdID:  "hh_none",
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodNone, noneClaims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("failed to sign none-alg token: %v", err)
	}

	svc := newTestService(t, testSecret)
	if _, err := svc.ValidateToken(signed); err == nil {
		t.Fatal("ValidateToken accepted none-alg token, expected error")
	}
}
