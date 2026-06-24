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

func TestUpdateMenuPreferencesRequest_Validate_HappyPath(t *testing.T) {
	req := validPrefsReq()
	if err := req.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestUpdateMenuPreferencesRequest_Validate_DefaultDays(t *testing.T) {
	tests := []struct {
		name    string
		days    int
		wantErr error
	}{
		{"zero rejected", 0, ErrInvalidDays},
		{"min valid", 1, nil},
		{"max valid", 31, nil},
		{"over max rejected", 32, ErrInvalidDays},
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

func TestUpdateMenuPreferencesRequest_Validate_DefaultServings(t *testing.T) {
	tests := []struct {
		name     string
		servings int
		wantErr  error
	}{
		{"zero rejected", 0, ErrInvalidServings},
		{"min valid", 1, nil},
		{"max valid", 100, nil},
		{"over max rejected", 101, ErrInvalidServings},
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

func TestUpdateMenuPreferencesRequest_Validate_VegetarianAndExcludedTags(t *testing.T) {
	t.Run("vegetarian days negative rejected", func(t *testing.T) {
		req := validPrefsReq()
		req.VegetarianDays = -1
		if err := req.Validate(); !errors.Is(err, ErrInvalidVegetarianDays) {
			t.Errorf("Validate() = %v, want ErrInvalidVegetarianDays", err)
		}
	})
	t.Run("vegetarian days exceeds default days rejected", func(t *testing.T) {
		req := validPrefsReq()
		req.VegetarianDays = req.DefaultDays + 1
		if err := req.Validate(); !errors.Is(err, ErrInvalidVegetarianDays) {
			t.Errorf("Validate() = %v, want ErrInvalidVegetarianDays", err)
		}
	})
	t.Run("vegetarian days equal to default days valid", func(t *testing.T) {
		req := validPrefsReq()
		req.VegetarianDays = req.DefaultDays
		if err := req.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})
	t.Run("too many excluded tags rejected", func(t *testing.T) {
		req := validPrefsReq()
		req.ExcludedTags = make([]string, 51)
		for i := range req.ExcludedTags {
			req.ExcludedTags[i] = "tag"
		}
		if err := req.Validate(); !errors.Is(err, ErrTooManyExcludedTags) {
			t.Errorf("Validate() = %v, want ErrTooManyExcludedTags", err)
		}
	})
	t.Run("blank excluded tag rejected", func(t *testing.T) {
		req := validPrefsReq()
		req.ExcludedTags = []string{""}
		if err := req.Validate(); !errors.Is(err, ErrInvalidExcludedTag) {
			t.Errorf("Validate() = %v, want ErrInvalidExcludedTag", err)
		}
	})
	t.Run("too-long excluded tag rejected", func(t *testing.T) {
		req := validPrefsReq()
		req.ExcludedTags = []string{strings.Repeat("x", 51)}
		if err := req.Validate(); !errors.Is(err, ErrInvalidExcludedTag) {
			t.Errorf("Validate() = %v, want ErrInvalidExcludedTag", err)
		}
	})
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
