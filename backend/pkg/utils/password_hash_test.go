package utils

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_NonEmptyAndDiffersFromPlaintext(t *testing.T) {
	const pw = "correct-horse-battery"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}
	if hash == pw {
		t.Errorf("HashPassword returned plaintext: %q", hash)
	}
}

func TestHashPassword_BcryptShaped(t *testing.T) {
	const pw = "correct-horse-battery"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	validPrefix := strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") || strings.HasPrefix(hash, "$2y$")
	if !validPrefix {
		t.Errorf("HashPassword returned non-bcrypt hash: %q", hash)
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("bcrypt.Cost returned error: %v", err)
	}
	if cost != 12 {
		t.Errorf("bcrypt cost = %d, want 12", cost)
	}

	if !CheckPassword(pw, hash) {
		t.Error("CheckPassword returned false for correct password")
	}
}

func TestHashPassword_RandomSalt(t *testing.T) {
	const pw = "correct-horse-battery"

	hash1, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword (first call) returned error: %v", err)
	}

	hash2, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword (second call) returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("HashPassword produced identical hashes for same password — salt not random")
	}

	if !CheckPassword(pw, hash1) {
		t.Error("CheckPassword returned false for hash1")
	}
	if !CheckPassword(pw, hash2) {
		t.Error("CheckPassword returned false for hash2")
	}
}
