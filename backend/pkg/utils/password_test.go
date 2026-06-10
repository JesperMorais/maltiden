package utils

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	const pw = "correct-horse-battery"

	hash, err := HashPassword(pw)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password verifies", pw, hash, true},
		{"wrong password fails", "wrong-password", hash, false},
		{"malformed hash returns false no panic", pw, "not-a-bcrypt-hash", false},
		{"empty hash returns false no panic", pw, "", false},
		{"empty password returns false no panic", "", hash, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPassword(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPassword(%q, %q) = %v, want %v", tt.password, tt.hash, got, tt.want)
			}
		})
	}
}
