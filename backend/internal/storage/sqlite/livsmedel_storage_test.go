package sqlite

import (
	"testing"
)

func setupLivsmedelTestStorage(t *testing.T) *LivsmedelStorage {
	t.Helper()
	db, err := Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Insert a few deterministic rows (the migration seed exists but we want
	// known values independent of it).
	rows := []struct {
		num                          int
		namn                         string
		kcal, protein, carbs, fat    float64
	}{
		{900001, "Zztestkyckling, filé, rå", 110, 23, 0, 1.5},
		{900002, "Zztestris, kokt", 130, 2.7, 28, 0.3},
		{900003, "Zztestkycklinggryta", 90, 8, 5, 4},
	}
	for _, r := range rows {
		if _, err := db.Exec(
			`INSERT INTO livsmedel (livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			r.num, r.namn, r.kcal, r.protein, r.carbs, r.fat); err != nil {
			t.Fatalf("seed livsmedel: %v", err)
		}
	}
	return NewLivsmedelStorage(db)
}

func TestLivsmedelStorage_GetByNumbers(t *testing.T) {
	s := setupLivsmedelTestStorage(t)

	got, err := s.GetByNumbers([]int{900001, 900002, 999999})
	if err != nil {
		t.Fatalf("GetByNumbers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rows (one number missing), got %d", len(got))
	}
	if got[900001].ProteinPer100g != 23 {
		t.Errorf("protein mismatch: %+v", got[900001])
	}
	if _, ok := got[999999]; ok {
		t.Errorf("unknown number should be absent")
	}
}

func TestLivsmedelStorage_GetByNumbers_EmptyInput(t *testing.T) {
	s := setupLivsmedelTestStorage(t)
	got, err := s.GetByNumbers(nil)
	if err != nil {
		t.Fatalf("GetByNumbers(nil): %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil empty map")
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}

func TestLivsmedelStorage_Search(t *testing.T) {
	s := setupLivsmedelTestStorage(t)

	got, err := s.Search("zztestkyckling", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Both "Zztestkyckling, filé, rå" and "Zztestkycklinggryta" match (case-insensitive).
	if len(got) != 2 {
		t.Fatalf("expected 2 matches for 'zztestkyckling', got %d (%+v)", len(got), got)
	}

	// Limit is honored.
	limited, err := s.Search("zztestkyckling", 1)
	if err != nil {
		t.Fatalf("Search limited: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected limit of 1, got %d", len(limited))
	}

	// No match returns empty.
	none, err := s.Search("zzzznotfound", 10)
	if err != nil {
		t.Fatalf("Search none: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("expected no matches, got %d", len(none))
	}
}

func TestLivsmedelStorage_GetAll(t *testing.T) {
	s := setupLivsmedelTestStorage(t)
	got, err := s.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	// The migration seed plus our 3 rows; assert our rows are present.
	found := map[int]bool{}
	for _, lv := range got {
		found[lv.Livsmedelsnummer] = true
	}
	for _, n := range []int{900001, 900002, 900003} {
		if !found[n] {
			t.Errorf("GetAll missing seeded row %d", n)
		}
	}
}
