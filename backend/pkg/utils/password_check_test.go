package utils

import "testing"

func TestCheckPassword(t *testing.T) {
	const plain = "correct horse battery staple"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "Correct",
			password: plain,
			hash:     hash,
			want:     true,
		},
		{
			name:     "Wrong",
			password: "wrong password",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Malformed",
			password: plain,
			hash:     "not-a-real-bcrypt-hash",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CheckPassword(tc.password, tc.hash)
			if got != tc.want {
				t.Errorf("CheckPassword(%q, %q) = %v, want %v", tc.password, tc.hash, got, tc.want)
			}
		})
	}
}
