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

func TestUpdateMenuPreferencesRequest_ValidateDefaultDays(t *testing.T) {
	tests := []struct {
		name    string
		days    int
		wantErr error
	}{
		{"min valid", 1, nil},
		{"max valid", 31, nil},
		{"zero rejected", 0, ErrInvalidDays},
		{"negative rejected", -1, ErrInvalidDays},
		{"too large rejected", 32, ErrInvalidDays},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPrefsReq()
			req.DefaultDays = tt.days
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

func TestUpdateMenuPreferencesRequest_ValidateDefaultServings(t *testing.T) {
	tests := []struct {
		name     string
		servings int
		wantErr  error
	}{
		{"min valid", 1, nil},
		{"max valid", 100, nil},
		{"zero rejected", 0, ErrInvalidServings},
		{"negative rejected", -1, ErrInvalidServings},
		{"too large rejected", 101, ErrInvalidServings},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPrefsReq()
			req.DefaultServings = tt.servings
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

func TestUpdateMenuPreferencesRequest_ValidateVegetarianDays(t *testing.T) {
	tests := []struct {
		name           string
		vegetarianDays int
		defaultDays    int
		wantErr        error
	}{
		{"zero valid", 0, 7, nil},
		{"equal to defaultDays valid", 7, 7, nil},
		{"less than defaultDays valid", 3, 7, nil},
		{"exceeds defaultDays rejected", 8, 7, ErrInvalidVegetarianDays},
		{"negative rejected", -1, 7, ErrInvalidVegetarianDays},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPrefsReq()
			req.DefaultDays = tt.defaultDays
			req.VegetarianDays = tt.vegetarianDays
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

func TestUpdateMenuPreferencesRequest_ValidateExcludedTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		wantErr error
	}{
		{"empty is valid", []string{}, nil},
		{"normal tags valid", []string{"vegetarisk", "snabb"}, nil},
		{"blank tag rejected", []string{""}, ErrInvalidExcludedTag},
		{"too-long tag rejected", []string{strings.Repeat("x", 51)}, ErrInvalidExcludedTag},
		{"too many tags rejected", makeTags(51, "tag"), ErrTooManyExcludedTags},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPrefsReq()
			req.ExcludedTags = tt.tags
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

func makeTags(n int, val string) []string {
	tags := make([]string, n)
	for i := range tags {
		tags[i] = val
	}
	return tags
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
