package sqlite

import (
	"testing"

	"maltiden/internal/domain"
)

func setupPrefsTestStorage(t *testing.T) (*MenuPreferencesStorage, string) {
	t.Helper()
	tmpFile := t.TempDir() + "/test.db"
	db, err := Open(tmpFile)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	const hh = "hh_prefs_test"
	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, hh, "Prefs Test Household"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	return NewMenuPreferencesStorage(db), hh
}

func TestMenuPreferencesStorage_GetMissingReturnsNil(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for unsaved preferences, got %+v", got)
	}
}

func TestMenuPreferencesStorage_UpsertThenGet(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	in := &domain.MenuPreferences{
		HouseholdID:     hh,
		ExcludedTags:    []string{"fisk", "fläsk"},
		DefaultDays:     5,
		DefaultServings: 3,
		VegetarianDays:  2,
	}
	if err := s.Upsert(in); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected stored preferences, got nil")
	}
	if got.DefaultDays != 5 || got.DefaultServings != 3 || got.VegetarianDays != 2 {
		t.Errorf("scalar mismatch: %+v", got)
	}
	if len(got.ExcludedTags) != 2 || got.ExcludedTags[0] != "fisk" || got.ExcludedTags[1] != "fläsk" {
		t.Errorf("excluded tags mismatch: %v", got.ExcludedTags)
	}
}

func TestMenuPreferencesStorage_UpsertReplaces(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	if err := s.Upsert(&domain.MenuPreferences{
		HouseholdID:     hh,
		ExcludedTags:    []string{"fisk"},
		DefaultDays:     7,
		DefaultServings: 4,
	}); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	if err := s.Upsert(&domain.MenuPreferences{
		HouseholdID:     hh,
		ExcludedTags:    []string{},
		DefaultDays:     3,
		DefaultServings: 2,
		VegetarianDays:  1,
	}); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DefaultDays != 3 || got.DefaultServings != 2 || got.VegetarianDays != 1 {
		t.Errorf("upsert did not replace scalars: %+v", got)
	}
	if len(got.ExcludedTags) != 0 {
		t.Errorf("expected excluded tags cleared, got %v", got.ExcludedTags)
	}
}

func TestMenuPreferencesStorage_DietFieldsRoundTrip(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	in := &domain.MenuPreferences{
		HouseholdID:         hh,
		ExcludedTags:        []string{"fisk"},
		DefaultDays:         7,
		DefaultServings:     4,
		VegetarianDays:      0,
		DietProfile:         domain.DietClassVegetarian,
		DislikedIngredients: []string{"koriander", "oliver"},
	}
	if err := s.Upsert(in); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DietProfile != domain.DietClassVegetarian {
		t.Errorf("diet profile mismatch: %q", got.DietProfile)
	}
	if len(got.DislikedIngredients) != 2 || got.DislikedIngredients[0] != "koriander" {
		t.Errorf("disliked ingredients mismatch: %v", got.DislikedIngredients)
	}

	// Upsert again replacing the diet fields.
	in.DietProfile = ""
	in.DislikedIngredients = []string{}
	if err := s.Upsert(in); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	got, err = s.Get(hh)
	if err != nil {
		t.Fatalf("Get after replace: %v", err)
	}
	if got.DietProfile != "" || len(got.DislikedIngredients) != 0 {
		t.Errorf("expected diet fields cleared, got profile=%q disliked=%v", got.DietProfile, got.DislikedIngredients)
	}
}

func TestMenuPreferencesStorage_NutritionFieldsRoundTrip(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	in := &domain.MenuPreferences{
		HouseholdID:         hh,
		ExcludedTags:        []string{},
		DefaultDays:         7,
		DefaultServings:     4,
		NutritionProfile:    "high protein",
		ProteinTargetPerDay: 120,
	}
	if err := s.Upsert(in); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.NutritionProfile != "high protein" || got.ProteinTargetPerDay != 120 {
		t.Errorf("nutrition fields mismatch: profile=%q target=%v", got.NutritionProfile, got.ProteinTargetPerDay)
	}

	// Upsert again clearing the nutrition fields.
	in.NutritionProfile = ""
	in.ProteinTargetPerDay = 0
	if err := s.Upsert(in); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	got, err = s.Get(hh)
	if err != nil {
		t.Fatalf("Get after replace: %v", err)
	}
	if got.NutritionProfile != "" || got.ProteinTargetPerDay != 0 {
		t.Errorf("expected nutrition fields cleared, got profile=%q target=%v", got.NutritionProfile, got.ProteinTargetPerDay)
	}
}

func TestMenuPreferencesStorage_NilDislikedStoredAsEmpty(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	if err := s.Upsert(&domain.MenuPreferences{
		HouseholdID:         hh,
		ExcludedTags:        []string{},
		DefaultDays:         7,
		DefaultServings:     4,
		DislikedIngredients: nil,
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.DislikedIngredients == nil {
		t.Errorf("expected non-nil empty slice, got nil")
	}
	if len(got.DislikedIngredients) != 0 {
		t.Errorf("expected empty disliked ingredients, got %v", got.DislikedIngredients)
	}
}

func TestMenuPreferencesStorage_NilTagsStoredAsEmpty(t *testing.T) {
	s, hh := setupPrefsTestStorage(t)

	if err := s.Upsert(&domain.MenuPreferences{
		HouseholdID:     hh,
		ExcludedTags:    nil,
		DefaultDays:     7,
		DefaultServings: 4,
	}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	got, err := s.Get(hh)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ExcludedTags == nil {
		t.Errorf("expected non-nil empty slice, got nil")
	}
	if len(got.ExcludedTags) != 0 {
		t.Errorf("expected empty excluded tags, got %v", got.ExcludedTags)
	}
}
