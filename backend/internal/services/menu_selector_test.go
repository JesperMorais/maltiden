package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"testing"
)

func rec(id string, tags ...string) domain.RecipeSummary {
	return domain.RecipeSummary{ID: id, Name: id, Servings: 4, Tags: tags}
}

func TestFilterByExcludedTags(t *testing.T) {
	all := []domain.RecipeSummary{
		rec("rec_a", "vegetariskt", "vardag"),
		rec("rec_b", "fisk"),
		rec("rec_c", "fläsk", "helg"),
		rec("rec_d"), // no tags
	}

	tests := []struct {
		name     string
		excluded []string
		wantIDs  []string
	}{
		{
			name:     "no exclusions keeps everything",
			excluded: nil,
			wantIDs:  []string{"rec_a", "rec_b", "rec_c", "rec_d"},
		},
		{
			name:     "single exclusion drops matching recipe",
			excluded: []string{"fisk"},
			wantIDs:  []string{"rec_a", "rec_c", "rec_d"},
		},
		{
			name:     "multiple exclusions",
			excluded: []string{"fisk", "fläsk"},
			wantIDs:  []string{"rec_a", "rec_d"},
		},
		{
			name:     "case-insensitive and trimmed match",
			excluded: []string{" Fisk ", "FLÄSK"},
			wantIDs:  []string{"rec_a", "rec_d"},
		},
		{
			name:     "exclude everything possible",
			excluded: []string{"vegetariskt", "fisk", "fläsk"},
			wantIDs:  []string{"rec_d"},
		},
		{
			name:     "non-matching exclusion keeps everything",
			excluded: []string{"glutenfritt"},
			wantIDs:  []string{"rec_a", "rec_b", "rec_c", "rec_d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByExcludedTags(all, tt.excluded)
			gotIDs := make([]string, len(got))
			for i, r := range got {
				gotIDs[i] = r.ID
			}
			if !sameSet(gotIDs, tt.wantIDs) {
				t.Errorf("filterByExcludedTags() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestIsVegetarian(t *testing.T) {
	tests := []struct {
		name string
		r    domain.RecipeSummary
		want bool
	}{
		{"tagged vegetariskt", rec("a", "vegetariskt"), true},
		{"mixed case tag", rec("b", "Vegetariskt"), true},
		{"not vegetarian", rec("c", "fisk"), false},
		{"no tags", rec("d"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVegetarian(tt.r); got != tt.want {
				t.Errorf("isVegetarian() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMenuSelector_AllFilteredOut(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("rec_a", "fisk")}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.ExcludedTags = []string{"fisk"}

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(1, 2)))
	if ok {
		t.Fatalf("expected ok=false when all recipes filtered out")
	}
	if sel != nil {
		t.Errorf("expected nil selector, got %+v", sel)
	}
}

func TestSelectorNext_CyclesWhenFewerRecipesThanDays(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("rec_a"), rec("rec_b")}
	prefs := domain.DefaultMenuPreferences("hh_1")

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(42, 42)))
	if !ok {
		t.Fatalf("expected selector to build")
	}

	picks := make(map[string]int)
	for i := 0; i < 7; i++ {
		picks[sel.Next()]++
	}
	if len(picks) != 2 {
		t.Errorf("expected to cycle through exactly 2 recipes, got %d distinct", len(picks))
	}
	for id, n := range picks {
		if n == 0 {
			t.Errorf("recipe %s never selected", id)
		}
	}
}

func TestSelectorNext_HonorsVegetarianQuota(t *testing.T) {
	recipes := []domain.RecipeSummary{
		rec("veg_1", "vegetariskt"),
		rec("veg_2", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
		rec("meat_3"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.VegetarianDays = 2

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(7, 7)))
	if !ok {
		t.Fatalf("expected selector to build")
	}

	// First two picks should satisfy the vegetarian quota.
	vegFirstTwo := 0
	first := sel.Next()
	second := sel.Next()
	for _, id := range []string{first, second} {
		if id == "veg_1" || id == "veg_2" {
			vegFirstTwo++
		}
	}
	if vegFirstTwo != 2 {
		t.Errorf("expected first 2 picks to be vegetarian, got %d (%s, %s)", vegFirstTwo, first, second)
	}
	if first == second {
		t.Errorf("vegetarian picks should be distinct, both were %s", first)
	}
}

func TestSelectorNext_VegetarianQuotaCappedByAvailability(t *testing.T) {
	// Quota asks for 3 veg days but only 1 vegetarian recipe exists; selector
	// must not loop forever or panic — it falls back to the general pool.
	recipes := []domain.RecipeSummary{
		rec("veg_1", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.VegetarianDays = 3

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(9, 9)))
	if !ok {
		t.Fatalf("expected selector to build")
	}

	for i := 0; i < 5; i++ {
		if got := sel.Next(); got == "" {
			t.Fatalf("Next() returned empty id at iteration %d", i)
		}
	}
}

func TestSelectorNext_Deterministic(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("a"), rec("b"), rec("c"), rec("d")}
	prefs := domain.DefaultMenuPreferences("hh_1")

	collect := func() []string {
		sel, _ := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(123, 456)))
		out := make([]string, 4)
		for i := range out {
			out[i] = sel.Next()
		}
		return out
	}

	a, b := collect(), collect()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("selector not deterministic with seeded RNG: %v vs %v", a, b)
		}
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int)
	for _, x := range a {
		m[x]++
	}
	for _, x := range b {
		m[x]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
