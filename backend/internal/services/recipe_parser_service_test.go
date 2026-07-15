package services

import (
	"strings"
	"testing"
)

func TestParseRecipe(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: "raw text is required",
		},
		{
			name:    "oversized input",
			input:   strings.Repeat("a", 10001),
			wantErr: "too long",
		},
	}

	svc := NewRecipeParserService(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ParseRecipe(tt.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
