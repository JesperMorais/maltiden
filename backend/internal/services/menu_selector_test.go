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

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(1, 2)), nil)
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

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(42, 42)), nil)
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

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(7, 7)), nil)
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

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(9, 9)), nil)
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
		sel, _ := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(123, 456)), nil)
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

// ing builds an ingredient with only the name set (amount/unit are irrelevant
// to overlap scoring).
func ing(names ...string) []domain.Ingredient {
	out := make([]domain.Ingredient, len(names))
	for i, n := range names {
		out[i] = domain.Ingredient{Name: n, Amount: 1, Unit: "st"}
	}
	return out
}

func TestOverlapKeys_DropsPantryStaples(t *testing.T) {
	tests := []struct {
		name string
		ings []domain.Ingredient
		want []string
	}{
		{"perishables kept, lowercased", ing("Kyckling", "Broccoli"), []string{"kyckling", "broccoli"}},
		{"spices dropped", ing("Kyckling", "salt", "peppar"), []string{"kyckling"}},
		{"oils and sauces dropped", ing("Pasta", "olivolja", "soja"), []string{"pasta"}},
		{"blank names ignored", ing("Ris", "  "), []string{"ris"}},
		{"all staples -> empty", ing("salt", "olivolja"), []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overlapKeys(tt.ings)
			if !sameSet(got, tt.want) {
				t.Errorf("overlapKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

// selWithIngredients is a small helper for the overlap tests.
func selWithIngredients(t *testing.T, recipes []domain.RecipeSummary, prefs domain.MenuPreferences, ings map[string][]domain.Ingredient, seed1, seed2 uint64) *menuSelector {
	t.Helper()
	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(seed1, seed2)), ings)
	if !ok {
		t.Fatalf("expected selector to build")
	}
	return sel
}

func TestSelectorNext_PrefersIngredientOverlap(t *testing.T) {
	// rec_a shares "kyckling" with rec_b but nothing with rec_c. After picking
	// rec_a, the next pick should be rec_b (overlap=1) over rec_c (overlap=0),
	// regardless of shuffle order.
	recipes := []domain.RecipeSummary{rec("rec_a"), rec("rec_b"), rec("rec_c")}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("kyckling", "ris"),
		"rec_b": ing("kyckling", "broccoli"),
		"rec_c": ing("torsk", "potatis"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")

	// Try several seeds so we are not just getting lucky with one shuffle.
	for _, seed := range []uint64{1, 2, 3, 99} {
		sel := selWithIngredients(t, recipes, prefs, ings, seed, seed+1)
		first := sel.Next()
		second := sel.Next()
		if second == first {
			t.Fatalf("seed %d: second pick repeated first %q while fresh recipes remained", seed, first)
		}
		// The greedy step should pick the recipe that shares an ingredient with
		// `first` when one exists.
		want := overlapPartner(first)
		if want != "" && second != want {
			t.Errorf("seed %d: after %q expected overlap pick %q, got %q", seed, first, want, second)
		}
	}
}

// overlapPartner returns the recipe that shares an ingredient with id in the
// PrefersIngredientOverlap fixture, or "" if id has no single clear partner.
func overlapPartner(id string) string {
	switch id {
	case "rec_a":
		return "rec_b"
	case "rec_b":
		return "rec_a"
	default:
		return "" // rec_c shares with neither; either pick is acceptable
	}
}

func TestSelectorNext_PantryStaplesDoNotDriveOverlap(t *testing.T) {
	// Every recipe shares only salt/olivolja (pantry staples). Because staples
	// are excluded, overlap is always 0 and selection must still fill the week
	// with distinct recipes (no crash, no degenerate single-recipe lock-in).
	recipes := []domain.RecipeSummary{rec("rec_a"), rec("rec_b"), rec("rec_c")}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("salt", "olivolja", "kyckling"),
		"rec_b": ing("salt", "olivolja", "torsk"),
		"rec_c": ing("salt", "olivolja", "fläsk"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	sel := selWithIngredients(t, recipes, prefs, ings, 5, 6)

	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		seen[sel.Next()] = true
	}
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct recipes when only staples overlap, got %d (%v)", len(seen), seen)
	}
}

func TestSelectorNext_OverlapDeterministic(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("a"), rec("b"), rec("c"), rec("d")}
	ings := map[string][]domain.Ingredient{
		"a": ing("kyckling", "ris"),
		"b": ing("kyckling", "lök"),
		"c": ing("torsk", "lök"),
		"d": ing("nötkött", "potatis"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")

	collect := func() []string {
		sel := selWithIngredients(t, recipes, prefs, ings, 321, 654)
		out := make([]string, 4)
		for i := range out {
			out[i] = sel.Next()
		}
		return out
	}
	a, b := collect(), collect()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("overlap selection not deterministic with seeded RNG: %v vs %v", a, b)
		}
	}
}

func TestSelectorNext_OverlapDoesNotBreakVegetarianQuota(t *testing.T) {
	// Even with ingredient data present, the vegetarian quota still wins: the
	// first two picks must be vegetarian.
	recipes := []domain.RecipeSummary{
		rec("veg_1", "vegetariskt"),
		rec("veg_2", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
	}
	ings := map[string][]domain.Ingredient{
		// Make the meat dishes share an ingredient with each other so a naive
		// overlap-only selector would prefer them; the quota must override.
		"veg_1":  ing("halloumi", "spenat"),
		"veg_2":  ing("tofu", "morot"),
		"meat_1": ing("kyckling", "ris"),
		"meat_2": ing("kyckling", "broccoli"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.VegetarianDays = 2
	sel := selWithIngredients(t, recipes, prefs, ings, 11, 22)

	first, second := sel.Next(), sel.Next()
	vegCount := 0
	for _, id := range []string{first, second} {
		if id == "veg_1" || id == "veg_2" {
			vegCount++
		}
	}
	if vegCount != 2 {
		t.Errorf("expected first 2 picks vegetarian despite overlap scoring, got %d (%s, %s)", vegCount, first, second)
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
