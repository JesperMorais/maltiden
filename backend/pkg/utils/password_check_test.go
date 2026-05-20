package utils

import "testing"

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !CheckPassword("correct horse battery staple", hash) {
		t.Error("CheckPassword returned false for correct password, want true")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if CheckPassword("wrong password", hash) {
		t.Error("CheckPassword returned true for wrong password, want false")
	}
}

func TestCheckPassword_MalformedHash(t *testing.T) {
	if CheckPassword("any password", "not-a-bcrypt-hash") {
		t.Error("CheckPassword returned true for malformed hash, want false")
	}
}
