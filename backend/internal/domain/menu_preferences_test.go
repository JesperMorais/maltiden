package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func validPrefsRequest() UpdateMenuPreferencesRequest {
	return UpdateMenuPreferencesRequest{
		ExcludedTags:    []string{},
		DefaultDays:     7,
		DefaultServings: 4,
		VegetarianDays:  0,
	}
}

func TestUpdateMenuPreferencesRequest_Validate_DietProfile(t *testing.T) {
	t.Run("empty accepted", func(t *testing.T) {
		r := validPrefsRequest()
		if err := r.Validate(); err != nil {
			t.Errorf("expected empty diet profile accepted, got %v", err)
		}
	})

	t.Run("valid accepted", func(t *testing.T) {
		r := validPrefsRequest()
		r.DietProfile = DietClassPescetarian
		if err := r.Validate(); err != nil {
			t.Errorf("expected pescetarian accepted, got %v", err)
		}
	})

	t.Run("garbage rejected", func(t *testing.T) {
		r := validPrefsRequest()
		r.DietProfile = "carnivore"
		if err := r.Validate(); err != ErrInvalidDietProfile {
			t.Errorf("expected ErrInvalidDietProfile, got %v", err)
		}
	})
}

func TestUpdateMenuPreferencesRequest_Validate_DislikedIngredients(t *testing.T) {
	t.Run("empty list accepted", func(t *testing.T) {
		r := validPrefsRequest()
		r.DislikedIngredients = []string{}
		if err := r.Validate(); err != nil {
			t.Errorf("expected empty list accepted, got %v", err)
		}
	})

	t.Run("valid list accepted", func(t *testing.T) {
		r := validPrefsRequest()
		r.DislikedIngredients = []string{"koriander", "oliver"}
		if err := r.Validate(); err != nil {
			t.Errorf("expected valid list accepted, got %v", err)
		}
	})

	t.Run("over 50 entries rejected", func(t *testing.T) {
		r := validPrefsRequest()
		r.DislikedIngredients = make([]string, 51)
		for i := range r.DislikedIngredients {
			r.DislikedIngredients[i] = "x"
		}
		if err := r.Validate(); err != ErrTooManyDislikedIngredients {
			t.Errorf("expected ErrTooManyDislikedIngredients, got %v", err)
		}
	})

	t.Run("empty entry rejected", func(t *testing.T) {
		r := validPrefsRequest()
		r.DislikedIngredients = []string{""}
		if err := r.Validate(); err != ErrInvalidDislikedIngredient {
			t.Errorf("expected ErrInvalidDislikedIngredient, got %v", err)
		}
	})

	t.Run("over-long entry rejected", func(t *testing.T) {
		r := validPrefsRequest()
		r.DislikedIngredients = []string{strings.Repeat("a", 51)}
		if err := r.Validate(); err != ErrInvalidDislikedIngredient {
			t.Errorf("expected ErrInvalidDislikedIngredient, got %v", err)
		}
	})
}

func TestDefaultMenuPreferences_DietFields(t *testing.T) {
	d := DefaultMenuPreferences("hh_1")
	if d.DietProfile != "" {
		t.Errorf("expected empty diet profile, got %q", d.DietProfile)
	}
	if d.DislikedIngredients == nil || len(d.DislikedIngredients) != 0 {
		t.Errorf("expected non-nil empty disliked slice, got %v", d.DislikedIngredients)
	}
}

func TestMenuPreferences_RoundTrip(t *testing.T) {
	in := MenuPreferences{
		HouseholdID:         "hh_1",
		ExcludedTags:        []string{"fisk"},
		DefaultDays:         5,
		DefaultServings:     3,
		VegetarianDays:      2,
		DietProfile:         DietClassVegetarian,
		DislikedIngredients: []string{"koriander"},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out MenuPreferences
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.DietProfile != DietClassVegetarian {
		t.Errorf("diet profile mismatch: %q", out.DietProfile)
	}
	if len(out.DislikedIngredients) != 1 || out.DislikedIngredients[0] != "koriander" {
		t.Errorf("disliked ingredients mismatch: %v", out.DislikedIngredients)
	}
}
