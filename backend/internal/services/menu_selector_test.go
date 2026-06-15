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

func TestFilterByDislikedIngredients(t *testing.T) {
	all := []domain.RecipeSummary{
		rec("rec_a"),
		rec("rec_b"),
		rec("rec_c"),
		rec("rec_no_data"), // intentionally absent from the ingredient map
	}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("Kyckling", "Ris"),
		"rec_b": ing("Räkor", "Vitlök"),
		"rec_c": ing("Torsk", "Potatis"),
	}

	tests := []struct {
		name     string
		disliked []string
		wantIDs  []string
	}{
		{
			name:     "no dislikes keeps everything",
			disliked: nil,
			wantIDs:  []string{"rec_a", "rec_b", "rec_c", "rec_no_data"},
		},
		{
			name:     "single dislike drops matching recipe",
			disliked: []string{"räkor"},
			wantIDs:  []string{"rec_a", "rec_c", "rec_no_data"},
		},
		{
			name:     "case-insensitive and trimmed match",
			disliked: []string{" Räkor ", "TORSK"},
			wantIDs:  []string{"rec_a", "rec_no_data"},
		},
		{
			name:     "non-matching dislike keeps everything",
			disliked: []string{"quinoa"},
			wantIDs:  []string{"rec_a", "rec_b", "rec_c", "rec_no_data"},
		},
		{
			name:     "blank dislikes are ignored",
			disliked: []string{"  ", ""},
			wantIDs:  []string{"rec_a", "rec_b", "rec_c", "rec_no_data"},
		},
		{
			name:     "recipe with no ingredient data is never dropped",
			disliked: []string{"kyckling", "räkor", "torsk"},
			wantIDs:  []string{"rec_no_data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterByDislikedIngredients(all, tt.disliked, ings)
			gotIDs := make([]string, len(got))
			for i, r := range got {
				gotIDs[i] = r.ID
			}
			if !sameSet(gotIDs, tt.wantIDs) {
				t.Errorf("filterByDislikedIngredients() = %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestNewMenuSelector_DislikedIngredientsFilteredOut(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("rec_a"), rec("rec_b")}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("kyckling", "ris"),
		"rec_b": ing("räkor", "lök"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.DislikedIngredients = []string{"räkor"}

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(1, 2)), ings)
	if !ok {
		t.Fatalf("expected selector to build (rec_a survives)")
	}
	// rec_b must never be placed; a full week of picks should only ever yield rec_a.
	for i := 0; i < 7; i++ {
		if got := sel.Next(); got != "rec_a" {
			t.Fatalf("pick %d = %q, want rec_a (rec_b is disliked)", i, got)
		}
	}
}

func TestNewMenuSelector_AllDislikedReturnsNoSelector(t *testing.T) {
	recipes := []domain.RecipeSummary{rec("rec_a"), rec("rec_b")}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("räkor"),
		"rec_b": ing("Räkor", "ris"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.DislikedIngredients = []string{"räkor"}

	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(3, 4)), ings)
	if ok {
		t.Fatalf("expected ok=false when every recipe is disliked")
	}
	if sel != nil {
		t.Errorf("expected nil selector, got %+v", sel)
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

func TestVarietyTags_DropsScheduleAndVegTags(t *testing.T) {
	tests := []struct {
		name string
		r    domain.RecipeSummary
		want []string
	}{
		{"cuisine and protein kept, lowercased", rec("a", "Fisk", "Asiatiskt"), []string{"fisk", "asiatiskt"}},
		{"vegetarian tag dropped", rec("b", "vegetariskt", "indiskt"), []string{"indiskt"}},
		{"schedule tags dropped", rec("c", "vardag", "helg", "barn", "kyckling"), []string{"kyckling"}},
		{"only schedule/veg -> empty", rec("d", "vegetariskt", "vardag"), []string{}},
		{"no tags -> empty", rec("e"), []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := varietyTags(tt.r); !sameSet(got, tt.want) {
				t.Errorf("varietyTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// tagOf returns the variety tags of a picked recipe id, looked up against the
// recipe set keyed by ID.
func tagOf(id string, byID map[string]domain.RecipeSummary) []string {
	return varietyTags(byID[id])
}

func TestSelectorNext_VarietyPenaltySpacesSameCuisine(t *testing.T) {
	// Three fish dishes and three distinct others, no ingredient overlap. The
	// variety + recency penalties must keep fish from being front-loaded or
	// clustered: no two fish dishes back-to-back over a 6-day week.
	recipes := []domain.RecipeSummary{
		rec("fisk_1", "fisk"),
		rec("fisk_2", "fisk"),
		rec("fisk_3", "fisk"),
		rec("kott_1", "nötkött"),
		rec("kyck_1", "kyckling"),
		rec("flask_1", "fläsk"),
	}
	byID := map[string]domain.RecipeSummary{}
	for _, r := range recipes {
		byID[r.ID] = r
	}
	prefs := domain.DefaultMenuPreferences("hh_1")

	isFisk := func(id string) bool {
		ts := tagOf(id, byID)
		return len(ts) > 0 && ts[0] == "fisk"
	}

	for _, seed := range []uint64{1, 2, 3, 7, 42} {
		sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(seed, seed+1)), nil)
		if !ok {
			t.Fatalf("seed %d: selector failed to build", seed)
		}
		// Fill exactly the 6 distinct recipes (one full cycle): the penalties
		// must interleave the three fish dishes with the three others rather
		// than front-loading or clustering them. With three non-fish dishes
		// available, no two fish dishes should be adjacent while a non-fish
		// dish remains fresh — checked via the first four slots, where a
		// non-fish option always exists.
		picks := make([]string, 6)
		for i := range picks {
			picks[i] = sel.Next()
		}
		// (a) Not all three fish dishes crammed into the first three slots.
		fishInFirst3 := 0
		for _, p := range picks[:3] {
			if isFisk(p) {
				fishInFirst3++
			}
		}
		if fishInFirst3 == 3 {
			t.Errorf("seed %d: all fish front-loaded into first 3 slots (%v)", seed, picks)
		}
		// (b) No two adjacent fish dishes among the first four slots, where a
		// non-fish dish is always still available to break them up.
		for i := 1; i < 4; i++ {
			if isFisk(picks[i-1]) && isFisk(picks[i]) {
				t.Errorf("seed %d: fish dishes adjacent at %d while non-fish available (%v)", seed, i, picks)
			}
		}
	}
}

func TestSelectorNext_RecencyPenaltySpacesRepeats(t *testing.T) {
	// Only two cuisines available but more than two slots: cycling will repeat,
	// yet the recency penalty must alternate them rather than emit AABB.
	recipes := []domain.RecipeSummary{
		rec("ita_1", "italienskt"),
		rec("asi_1", "asiatiskt"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(5, 6)), nil)
	if !ok {
		t.Fatalf("selector failed to build")
	}
	picks := make([]string, 4)
	for i := range picks {
		picks[i] = sel.Next()
	}
	for i := 1; i < len(picks); i++ {
		if picks[i] == picks[i-1] {
			t.Errorf("recency penalty failed to space repeats at %d: %v", i, picks)
		}
	}
}

func TestSelectorNext_OverlapStillBeatsVarietyPenalty(t *testing.T) {
	// rec_a and rec_b share both a tag (fisk) AND an ingredient (torsk); rec_c
	// shares neither. After picking a fish dish, its torsk-sharing fish partner
	// carries a variety penalty but a +1 overlap reward that must still win it
	// the slot over the no-overlap rec_c — one overlap point outweighs one tag
	// repeat. Guards the #258 overlap feature against the new penalties.
	recipes := []domain.RecipeSummary{rec("rec_a", "fisk"), rec("rec_b", "fisk"), rec("rec_c", "kyckling")}
	ings := map[string][]domain.Ingredient{
		"rec_a": ing("torsk", "potatis"),
		"rec_b": ing("torsk", "dill"),
		"rec_c": ing("kyckling", "ris"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	for _, seed := range []uint64{1, 2, 3, 99} {
		sel := selWithIngredients(t, recipes, prefs, ings, seed, seed+1)
		first := sel.Next()
		second := sel.Next()
		var wantPartner string
		switch first {
		case "rec_a":
			wantPartner = "rec_b"
		case "rec_b":
			wantPartner = "rec_a"
		}
		if wantPartner != "" && second != wantPartner {
			t.Errorf("seed %d: after %q expected overlap to win %q despite variety penalty, got %q", seed, first, wantPartner, second)
		}
	}
}

func TestSelectorNext_VarietyDoesNotBreakVegetarianQuota(t *testing.T) {
	// Both veg dishes share a (non-veg) tag; the variety penalty between them
	// must not stop the quota from placing two vegetarian dishes first.
	recipes := []domain.RecipeSummary{
		rec("veg_1", "vegetariskt", "indiskt"),
		rec("veg_2", "vegetariskt", "indiskt"),
		rec("meat_1", "kyckling"),
		rec("meat_2", "fisk"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	prefs.VegetarianDays = 2
	sel, ok := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(13, 14)), nil)
	if !ok {
		t.Fatalf("selector failed to build")
	}
	first, second := sel.Next(), sel.Next()
	vegCount := 0
	for _, id := range []string{first, second} {
		if id == "veg_1" || id == "veg_2" {
			vegCount++
		}
	}
	if vegCount != 2 {
		t.Errorf("expected first 2 picks vegetarian despite shared-tag variety penalty, got %d (%s, %s)", vegCount, first, second)
	}
}

func TestSelectorNext_PenaltiesDeterministic(t *testing.T) {
	recipes := []domain.RecipeSummary{
		rec("a", "fisk"), rec("b", "fisk"), rec("c", "kyckling"), rec("d", "nötkött"),
	}
	prefs := domain.DefaultMenuPreferences("hh_1")
	collect := func() []string {
		sel, _ := newMenuSelector(recipes, prefs, rand.New(rand.NewPCG(321, 654)), nil)
		out := make([]string, 6)
		for i := range out {
			out[i] = sel.Next()
		}
		return out
	}
	a, b := collect(), collect()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("variety/recency selection not deterministic: %v vs %v", a, b)
		}
	}
}

func TestIsBatchable(t *testing.T) {
	tests := []struct {
		name string
		r    domain.RecipeSummary
		want bool
	}{
		{"tagged batchcook", rec("a", "batchcook"), true},
		{"mixed case tag", rec("b", "BatchCook"), true},
		{"trimmed tag", rec("c", " batchcook "), true},
		{"not batchable", rec("d", "fisk"), false},
		{"no tags", rec("e"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBatchable(tt.r); got != tt.want {
				t.Errorf("isBatchable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyBatchCooking(t *testing.T) {
	// day is a small builder: a non-skip cook-able day.
	day := func(date, recipeID string, servings int) domain.MenuDay {
		return domain.MenuDay{Date: date, RecipeID: recipeID, Servings: servings}
	}
	skipDay := func(date string) domain.MenuDay {
		return domain.MenuDay{Date: date, Servings: 4, Skip: true}
	}

	tests := []struct {
		name       string
		days       []domain.MenuDay
		batchable  map[string]bool
		wantPrep   map[string]string // date -> prepMode
		wantServ   map[string]int    // date -> servings
		wantLeftOf map[string]string // date -> leftoverOf
		wantRecipe map[string]string // date -> recipeID (post-pass)
	}{
		{
			name:       "no batchable recipes is a no-op",
			days:       []domain.MenuDay{day("d1", "r_a", 4), day("d2", "r_b", 4)},
			batchable:  map[string]bool{},
			wantPrep:   map[string]string{"d1": "", "d2": ""},
			wantServ:   map[string]int{"d1": 4, "d2": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": ""},
			wantRecipe: map[string]string{"d1": "r_a", "d2": "r_b"},
		},
		{
			name:       "batchable cook-day pairs with next day as leftovers",
			days:       []domain.MenuDay{day("d1", "r_a", 4), day("d2", "r_b", 4), day("d3", "r_c", 4)},
			batchable:  map[string]bool{"r_a": true},
			wantPrep:   map[string]string{"d1": domain.PrepModeBatch, "d2": "", "d3": ""},
			wantServ:   map[string]int{"d1": 8, "d2": 4, "d3": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": "d1", "d3": ""},
			wantRecipe: map[string]string{"d1": "r_a", "d2": "r_a", "d3": "r_c"},
		},
		{
			name:       "leftovers slot skips over a skipped day",
			days:       []domain.MenuDay{day("d1", "r_a", 4), skipDay("d2"), day("d3", "r_c", 4)},
			batchable:  map[string]bool{"r_a": true},
			wantPrep:   map[string]string{"d1": domain.PrepModeBatch, "d2": "", "d3": ""},
			wantServ:   map[string]int{"d1": 8, "d2": 4, "d3": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": "", "d3": "d1"},
			wantRecipe: map[string]string{"d1": "r_a", "d2": "", "d3": "r_a"},
		},
		{
			name:       "batchable on last day has no room for leftovers (no-op)",
			days:       []domain.MenuDay{day("d1", "r_b", 4), day("d2", "r_a", 4)},
			batchable:  map[string]bool{"r_a": true},
			wantPrep:   map[string]string{"d1": "", "d2": ""},
			wantServ:   map[string]int{"d1": 4, "d2": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": ""},
			wantRecipe: map[string]string{"d1": "r_b", "d2": "r_a"},
		},
		{
			name:      "each recipe batched at most once; no chaining",
			days:      []domain.MenuDay{day("d1", "r_a", 4), day("d2", "r_a", 4), day("d3", "r_a", 4), day("d4", "r_a", 4)},
			batchable: map[string]bool{"r_a": true},
			// d1 cooks (batch, 8), d2 is its leftovers (consumed). d3 is a fresh
			// cook of r_a but r_a is already batched, so it stays ordinary; d4 too.
			wantPrep:   map[string]string{"d1": domain.PrepModeBatch, "d2": "", "d3": "", "d4": ""},
			wantServ:   map[string]int{"d1": 8, "d2": 4, "d3": 4, "d4": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": "d1", "d3": "", "d4": ""},
			wantRecipe: map[string]string{"d1": "r_a", "d2": "r_a", "d3": "r_a", "d4": "r_a"},
		},
		{
			name:      "two distinct batchable recipes each get a pair",
			days:      []domain.MenuDay{day("d1", "r_a", 4), day("d2", "r_b", 4), day("d3", "r_x", 4), day("d4", "r_y", 4)},
			batchable: map[string]bool{"r_a": true, "r_b": true},
			// d1(r_a) cooks -> d2 becomes its leftovers. d2 is now consumed, so
			// r_b never cooks (its only slot got eaten by r_a's leftovers).
			wantPrep:   map[string]string{"d1": domain.PrepModeBatch, "d2": "", "d3": "", "d4": ""},
			wantServ:   map[string]int{"d1": 8, "d2": 4, "d3": 4, "d4": 4},
			wantLeftOf: map[string]string{"d1": "", "d2": "d1", "d3": "", "d4": ""},
			wantRecipe: map[string]string{"d1": "r_a", "d2": "r_a", "d3": "r_x", "d4": "r_y"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyBatchCooking(tt.days, tt.batchable)
			byDate := make(map[string]domain.MenuDay, len(got))
			for _, d := range got {
				byDate[d.Date] = d
			}
			for date, want := range tt.wantPrep {
				if byDate[date].PrepMode != want {
					t.Errorf("%s prepMode = %q, want %q", date, byDate[date].PrepMode, want)
				}
			}
			for date, want := range tt.wantServ {
				if byDate[date].Servings != want {
					t.Errorf("%s servings = %d, want %d", date, byDate[date].Servings, want)
				}
			}
			for date, want := range tt.wantLeftOf {
				if byDate[date].LeftoverOf != want {
					t.Errorf("%s leftoverOf = %q, want %q", date, byDate[date].LeftoverOf, want)
				}
			}
			for date, want := range tt.wantRecipe {
				if byDate[date].RecipeID != want {
					t.Errorf("%s recipeID = %q, want %q", date, byDate[date].RecipeID, want)
				}
			}
		})
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

func TestComputeSharedIngredients(t *testing.T) {
	ingredients := map[string][]domain.Ingredient{
		"rec_a": ing("Kyckling", "Ris", "Salt"),         // Salt = pantry staple
		"rec_b": ing("kyckling", "Paprika", "Olivolja"), // shares kyckling (diff case); olja = staple
		"rec_c": ing("Ris", "Lök", "Lök"),               // dup Lök within recipe must not self-share
		"rec_d": ing("Pasta"),                           // shares nothing
	}

	got := computeSharedIngredients([]string{"rec_a", "rec_b", "rec_c", "rec_d"}, ingredients)

	// Kyckling shared by a+b (case-insensitive) = 2; Ris shared by a+c = 2.
	// Lök appears in one recipe only (dup ignored) -> not shared.
	// Salt/Olivolja are pantry staples -> excluded.
	if len(got) != 2 {
		t.Fatalf("want 2 shared ingredients, got %d: %+v", len(got), got)
	}
	// Sorted by count desc then name asc; both count 2 so alphabetical: Kyckling, Ris.
	if got[0].Name != "Kyckling" || got[0].RecipeCount != 2 {
		t.Errorf("got[0] = %+v, want {Kyckling 2}", got[0])
	}
	if got[1].Name != "Ris" || got[1].RecipeCount != 2 {
		t.Errorf("got[1] = %+v, want {Ris 2}", got[1])
	}
}

func TestComputeSharedIngredientsDedupesCycledRecipes(t *testing.T) {
	ingredients := map[string][]domain.Ingredient{
		"rec_a": ing("Kyckling", "Ris"),
	}
	// Same recipe twice (cycling) must not make its own ingredients look shared.
	got := computeSharedIngredients([]string{"rec_a", "rec_a"}, ingredients)
	if len(got) != 0 {
		t.Fatalf("a recipe shared with itself must yield nothing, got %+v", got)
	}
}

func TestComputeSharedIngredientsEmptyWhenNoOverlap(t *testing.T) {
	ingredients := map[string][]domain.Ingredient{
		"rec_a": ing("Kyckling"),
		"rec_b": ing("Lax"),
	}
	if got := computeSharedIngredients([]string{"rec_a", "rec_b"}, ingredients); len(got) != 0 {
		t.Fatalf("want empty, got %+v", got)
	}
}
