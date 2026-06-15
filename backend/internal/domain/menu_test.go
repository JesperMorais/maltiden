package domain

import (
	"errors"
	"strings"
	"testing"
)

func validPrefsReq() UpdateMenuPreferencesRequest {
	return UpdateMenuPreferencesRequest{
		ExcludedTags:        []string{},
		DislikedIngredients: []string{},
		DefaultDays:         7,
		DefaultServings:     4,
		VegetarianDays:      0,
	}
}

func TestUpdateMenuPreferencesRequest_ValidateDislikedIngredients(t *testing.T) {
	tests := []struct {
		name     string
		disliked []string
		wantErr  error
	}{
		{"empty is valid", []string{}, nil},
		{"normal entries valid", []string{"räkor", "koriander"}, nil},
		{"blank entry rejected", []string{""}, ErrInvalidDislikedIngredient},
		{"too-long entry rejected", []string{strings.Repeat("x", 81)}, ErrInvalidDislikedIngredient},
		{"too many entries rejected", make([]string, 51), ErrTooManyDislikedIngredients},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPrefsReq()
			// Fill the "too many" case with valid-length names so only the count trips.
			if tt.name == "too many entries rejected" {
				for i := range tt.disliked {
					tt.disliked[i] = "ingrediens"
				}
			}
			req.DislikedIngredients = tt.disliked
			err := req.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
