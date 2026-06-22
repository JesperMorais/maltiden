package services

import (
	"maltiden/internal/domain"
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "too short",
			password: "Ab1!",
			wantErr:  domain.ErrWeakPassword,
		},
		{
			name:     "exactly 8 chars but only lowercase",
			password: "abcdefgh",
			wantErr:  domain.ErrWeakPassword,
		},
		{
			name:     "missing upper class — only lower+digit+special",
			password: "abcdef1!",
			wantErr:  nil,
		},
		{
			name:     "missing lower class — only upper+digit+special",
			password: "ABCDEF1!",
			wantErr:  nil,
		},
		{
			name:     "missing digit — only upper+lower (2 classes)",
			password: "Abcdefgh",
			wantErr:  domain.ErrWeakPassword,
		},
		{
			name:     "valid strong password — upper+lower+digit",
			password: "Password1",
			wantErr:  nil,
		},
		{
			name:     "valid with all four classes",
			password: "Password1!",
			wantErr:  nil,
		},
		{
			name:     "only digits — one class",
			password: "12345678",
			wantErr:  domain.ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidatePasswordStrength(%q) = %v, want %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
