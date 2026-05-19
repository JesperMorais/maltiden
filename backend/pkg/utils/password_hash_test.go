package utils

import (
	"strings"
	"testing"
)

const testPassword = "correct-horse-battery-staple"

func TestHashPassword_ReturnsNonEmpty(t *testing.T) {
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
}

func TestHashPassword_DiffersFromPlaintext(t *testing.T) {
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == testPassword {
		t.Fatal("HashPassword returned the plaintext password unchanged")
	}
}

func TestHashPassword_ProducesDifferentHashesEachCall(t *testing.T) {
	hash1, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("first HashPassword returned error: %v", err)
	}
	hash2, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("second HashPassword returned error: %v", err)
	}
	if hash1 == hash2 {
		t.Fatal("HashPassword produced identical hashes for two calls — bcrypt salt not applied")
	}
}

func TestHashPassword_CheckPasswordVerifies(t *testing.T) {
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !CheckPassword(testPassword, hash) {
		t.Fatal("CheckPassword returned false for a valid hash produced by HashPassword")
	}
}

func TestHashPassword_EmptyInput(t *testing.T) {
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword(\"\") returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword(\"\") returned empty hash")
	}
}

// TestHashPassword_LongInput uses a 70-byte password (safely under bcrypt's 72-byte cap)
// and asserts success.
func TestHashPassword_LongInput(t *testing.T) {
	pw := strings.Repeat("x", 70)
	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword(70-byte) returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword(70-byte) returned empty hash")
	}
	if hash == pw {
		t.Fatal("HashPassword(70-byte) returned plaintext unchanged")
	}
}
