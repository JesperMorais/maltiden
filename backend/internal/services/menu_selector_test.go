package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"testing"
)

// rec builds a tag-only recipe (no ingredients/metadata) for filter/quota tests.
func rec(id string, tags ...string) domain.Recipe {
	return domain.Recipe{ID: id, Name: id, Servings: 4, Tags: tags}
}

// ing is a convenience constructor for an ingredient with a canonical name and
// optional pantry-staple flag.
func ing(name, canonical string, staple bool) domain.Ingredient {
	return domain.Ingredient{Name: name, CanonicalName: canonical, IsPantryStaple: staple, Amount: 1, Unit: "st"}
}

// recWith builds a recipe with ingredients and recipe-level metadata.
func recWith(id, protein, dietClass string, tags []string, ingredients ...domain.Ingredient) domain.Recipe {
	return domain.Recipe{
		ID:          id,
		Name:        id,
		Servings:    4,
		Tags:        tags,
		Ingredients: ingredients,
		MainProtein: protein,
		DietClass:   dietClass,
	}
}

func testRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

func defaultPrefs() domain.MenuPreferences { return domain.DefaultMenuPreferences("hh_1") }

func TestFilterByExcludedTags(t *testing.T) {
	all := []domain.Recipe{
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
		{"no exclusions keeps everything", nil, []string{"rec_a", "rec_b", "rec_c", "rec_d"}},
		{"single exclusion drops matching recipe", []string{"fisk"}, []string{"rec_a", "rec_c", "rec_d"}},
		{"multiple exclusions", []string{"fisk", "fläsk"}, []string{"rec_a", "rec_d"}},
		{"case-insensitive and trimmed match", []string{" Fisk ", "FLÄSK"}, []string{"rec_a", "rec_d"}},
		{"exclude everything possible", []string{"vegetariskt", "fisk", "fläsk"}, []string{"rec_d"}},
		{"non-matching exclusion keeps everything", []string{"glutenfritt"}, []string{"rec_a", "rec_b", "rec_c", "rec_d"}},
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
	dc := func(id, dietClass string, tags ...string) domain.Recipe {
		r := rec(id, tags...)
		r.DietClass = dietClass
		return r
	}
	tests := []struct {
		name string
		r    domain.Recipe
		want bool
	}{
		{"tagged vegetariskt", rec("a", "vegetariskt"), true},
		{"mixed case tag", rec("b", "Vegetariskt"), true},
		{"not vegetarian", rec("c", "fisk"), false},
		{"no tags", rec("d"), false},
		// DietClass should qualify even without the Swedish tag.
		{"vegetarian DietClass, no tag", dc("e", domain.DietClassVegetarian), true},
		{"vegan DietClass, no tag", dc("f", domain.DietClassVegan), true},
		{"omnivore DietClass, no tag", dc("g", domain.DietClassOmnivore), false},
		{"pescetarian DietClass, no tag", dc("h", domain.DietClassPescetarian), false},
		{"omnivore DietClass but tagged", dc("i", domain.DietClassOmnivore, "vegetariskt"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isVegetarian(tt.r); got != tt.want {
				t.Errorf("isVegetarian() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectWeek_VegQuotaHonorsDietClassWithoutTag(t *testing.T) {
	// Recipes carry vegetarian/vegan DietClass but NOT the "vegetariskt" tag.
	// The veg quota must still be satisfied (regression: tag-only isVegetarian
	// silently no-opped the quota for metadata-driven households).
	recipes := []domain.Recipe{
		recWith("veg_1", "", domain.DietClassVegetarian, nil, ing("Halloumi", "halloumi", false)),
		recWith("veg_2", "", domain.DietClassVegan, nil, ing("Kikärtor", "kikärtor", false)),
		recWith("meat_1", "nöt", domain.DietClassOmnivore, nil, ing("Köttfärs", "köttfärs", false)),
		recWith("meat_2", "fisk", domain.DietClassPescetarian, nil, ing("Lax", "lax", false)),
		recWith("meat_3", "fågel", domain.DietClassOmnivore, nil, ing("Kyckling", "kyckling", false)),
	}
	prefs := defaultPrefs()
	prefs.VegetarianDays = 2

	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := selectWeek(t, sel, 5)
	placedVeg := map[string]int{}
	for _, id := range week {
		if id == "veg_1" || id == "veg_2" {
			placedVeg[id]++
		}
	}
	if len(placedVeg) != 2 {
		t.Errorf("expected 2 distinct DietClass-vegetarian recipes placed, got %d (%v) in %v", len(placedVeg), placedVeg, week)
	}
}

func TestNewMenuSelector_AllFilteredOut(t *testing.T) {
	recipes := []domain.Recipe{rec("rec_a", "fisk")}
	prefs := defaultPrefs()
	prefs.ExcludedTags = []string{"fisk"}

	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if ok {
		t.Fatalf("expected ok=false when all recipes filtered out")
	}
	if sel != nil {
		t.Errorf("expected nil selector, got %+v", sel)
	}
}

// helper: build a week with no skips/locks.
func selectWeek(t *testing.T, sel *menuSelector, days int) []string {
	t.Helper()
	return sel.SelectWeek(days, make([]bool, days), make([]string, days))
}

func TestSelectWeek_CyclesWhenFewerRecipesThanDays(t *testing.T) {
	recipes := []domain.Recipe{rec("rec_a"), rec("rec_b")}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	week := selectWeek(t, sel, 7)
	picks := make(map[string]int)
	for _, id := range week {
		if id == "" {
			t.Fatalf("got empty slot in week %v", week)
		}
		picks[id]++
	}
	if len(picks) != 2 {
		t.Fatalf("expected to cycle through exactly 2 recipes, got %d distinct: %v", len(picks), week)
	}
	// The cycling fallback must spread duplicates EVENLY (round-robin), not pile
	// every duplicate onto one recipe. With 7 slots over 2 recipes the only even
	// split is 4/3 — neither recipe may appear fewer than 3 or more than 4 times.
	for id, n := range picks {
		if n < 3 || n > 4 {
			t.Errorf("uneven cycling: recipe %s appears %d times (expected 3 or 4), week %v", id, n, week)
		}
	}
}

func TestSelectWeek_CyclingEvenSpreadThreeRecipes(t *testing.T) {
	// 8 slots over 3 recipes → even round-robin is 3/3/2; no recipe should appear
	// fewer than 2 or more than 3 times. Catches a regression to "always pick
	// candidates[0]" which would clump all extras on a single recipe.
	recipes := []domain.Recipe{rec("rec_a"), rec("rec_b"), rec("rec_c")}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	week := selectWeek(t, sel, 8)
	picks := make(map[string]int)
	for i, id := range week {
		if id == "" {
			t.Fatalf("got empty slot at %d in week %v", i, week)
		}
		picks[id]++
	}
	if len(picks) != 3 {
		t.Fatalf("expected all 3 recipes used, got %d distinct: %v", len(picks), week)
	}
	for id, n := range picks {
		if n < 2 || n > 3 {
			t.Errorf("uneven cycling: recipe %s appears %d times (expected 2 or 3), week %v", id, n, week)
		}
	}
}

func TestSelectWeek_HonorsVegetarianQuota(t *testing.T) {
	recipes := []domain.Recipe{
		rec("veg_1", "vegetariskt"),
		rec("veg_2", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
		rec("meat_3"),
	}
	prefs := defaultPrefs()
	prefs.VegetarianDays = 2

	vegIDs := map[string]bool{"veg_1": true, "veg_2": true}

	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	week := selectWeek(t, sel, 5)
	// The two vegetarian recipes must be PLACED and be DISTINCT (no under-fill,
	// no double-placing the same veg recipe to fake the quota).
	placedVeg := map[string]int{}
	for _, id := range week {
		if vegIDs[id] {
			placedVeg[id]++
		}
	}
	if len(placedVeg) != 2 {
		t.Errorf("expected 2 distinct vegetarian recipes placed, got %d (%v) in week %v", len(placedVeg), placedVeg, week)
	}
	for id, n := range placedVeg {
		if n != 1 {
			t.Errorf("vegetarian recipe %s placed %d times, expected exactly 1: week %v", id, n, week)
		}
	}
}

func TestSelectWeek_VegetarianQuotaCappedByAvailability(t *testing.T) {
	// Quota asks for 3 veg days but only 1 vegetarian recipe exists. Use exactly
	// 3 slots (= the catalog size) so the pool is not exhausted and there is no
	// cycling: the selector must PLACE the single veg recipe (quota capped at the
	// one available), place each of the three distinct recipes once, and leave no
	// empty slot.
	recipes := []domain.Recipe{
		rec("veg_1", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
	}
	prefs := defaultPrefs()
	prefs.VegetarianDays = 3

	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	week := selectWeek(t, sel, 3)
	placed := map[string]int{}
	for i, id := range week {
		if id == "" {
			t.Fatalf("got empty id at slot %d: %v", i, week)
		}
		placed[id]++
	}
	// All three distinct recipes placed exactly once (no over-fill of the lone
	// veg recipe trying to chase the unmet quota of 3).
	if len(placed) != 3 {
		t.Errorf("expected 3 distinct recipes placed, got %d (%v) in week %v", len(placed), placed, week)
	}
	if placed["veg_1"] != 1 {
		t.Errorf("expected the single vegetarian recipe placed exactly once, got %d in week %v", placed["veg_1"], week)
	}
}

func TestSelectWeek_Deterministic(t *testing.T) {
	recipes := []domain.Recipe{
		recWith("a", "nöt", "", nil, ing("Lök", "lök", false), ing("Köttfärs", "köttfärs", false)),
		recWith("b", "fågel", "", nil, ing("Lök", "lök", false), ing("Kyckling", "kyckling", false)),
		recWith("c", "fisk", "", nil, ing("Lax", "lax", false)),
		recWith("d", "fläsk", "", nil, ing("Fläsk", "fläsk", false)),
	}

	collect := func() []string {
		sel, _ := newMenuSelector(recipes, defaultPrefs(), nil, rand.New(rand.NewPCG(123, 456)))
		return selectWeek(t, sel, 4)
	}

	a, b := collect(), collect()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("selector not deterministic with seeded RNG: %v vs %v", a, b)
		}
	}
}

func TestSelectWeek_OverlapRewardPicksSharedCluster(t *testing.T) {
	// shared_1 and shared_2 share "lök" + "grädde" but each has two unique
	// ingredients too, so their Jaccard stays below the near-duplicate
	// threshold (overlap reward applies, duplicate penalty does not). The third
	// recipe shares nothing. With 2 open slots the selector should prefer the
	// overlapping pair.
	recipes := []domain.Recipe{
		recWith("shared_1", "nöt", "", nil,
			ing("Lök", "lök", false), ing("Grädde", "grädde", false),
			ing("Köttfärs", "köttfärs", false), ing("Potatis", "potatis", false)),
		recWith("shared_2", "fläsk", "", nil,
			ing("Lök", "lök", false), ing("Grädde", "grädde", false),
			ing("Fläskkött", "fläskkött", false), ing("Senap", "senap", false)),
		recWith("lonely", "fisk", "", nil,
			ing("Lax", "lax", false), ing("Citron", "citron", false),
			ing("Dill", "dill", false), ing("Ris", "ris", false)),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}

	week := sel.SelectWeek(2, []bool{false, false}, []string{"", ""})
	got := map[string]bool{week[0]: true, week[1]: true}
	if !got["shared_1"] || !got["shared_2"] {
		t.Errorf("expected the overlapping pair, got %v", week)
	}
}

func TestSelectWeek_StaplesExcludedFromOverlap_RealFlag(t *testing.T) {
	// A staple flagged IsPantryStaple (salt) must NOT count as a non-staple
	// canonical, so it never contributes overlap. Assert via nonStapleCanonicals
	// (the source of truth) — deterministic, independent of tie-breaking.
	r := recWith("r", "nöt", "", nil,
		ing("Salt", "salt", true),
		ing("Lök", "lök", false),
	)
	set := nonStapleCanonicals(r)
	if _, has := set["salt"]; has {
		t.Errorf("salt (IsPantryStaple) should be excluded, set=%v", set)
	}
	if _, has := set["lök"]; !has {
		t.Errorf("lök should be present, set=%v", set)
	}
}

func TestSelectWeek_StaplesExcludedFromOverlap_FallbackList(t *testing.T) {
	// No metadata at all (no canonical, no staple flags) → fallback staple list.
	// "salt" and "vatten" are in the fallback list; "lök" is not.
	r := domain.Recipe{
		ID:   "r",
		Name: "r",
		Ingredients: []domain.Ingredient{
			{Name: "Salt", Amount: 1, Unit: "tsk"},
			{Name: "Vatten", Amount: 1, Unit: "dl"},
			{Name: "Lök", Amount: 1, Unit: "st"},
		},
	}
	set := nonStapleCanonicals(r)
	if _, has := set["salt"]; has {
		t.Errorf("salt (fallback staple) should be excluded, set=%v", set)
	}
	if _, has := set["vatten"]; has {
		t.Errorf("vatten (fallback staple) should be excluded, set=%v", set)
	}
	if _, has := set["lök"]; !has {
		t.Errorf("lök should be present, set=%v", set)
	}
}

func TestSelectWeek_ProteinVarietyAvoidsRepeats(t *testing.T) {
	// Three recipes, two share the same protein. With 2 slots and a third
	// distinct-protein option, the selector should avoid the protein repeat —
	// unless overlap dominates, so give the same-protein pair no overlap.
	recipes := []domain.Recipe{
		recWith("beef_1", "nöt", "", nil, ing("Köttfärs", "köttfärs", false)),
		recWith("beef_2", "nöt", "", nil, ing("Biff", "biff", false)),
		recWith("fish", "fisk", "", nil, ing("Lax", "lax", false)),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(2, []bool{false, false}, []string{"", ""})
	proteins := map[string]int{}
	for _, id := range week {
		for _, r := range recipes {
			if r.ID == id {
				proteins[r.MainProtein]++
			}
		}
	}
	if proteins["nöt"] > 1 {
		t.Errorf("expected protein variety to avoid repeating nöt, got week %v", week)
	}
}

func TestSelectWeek_NearDuplicatesNotBothChosen(t *testing.T) {
	// dup_a and dup_b share 5 ingredients; dup_a adds one more, so their Jaccard
	// is 5/6 ≈ 0.83 ≥ 0.8 (near-duplicate). With an alternative available and 2
	// slots, the selector should not pick both.
	common := []domain.Ingredient{
		ing("Lök", "lök", false),
		ing("Vitlök", "vitlök2", false),
		ing("Tomat", "tomat", false),
		ing("Pasta", "pasta", false),
		ing("Basilika", "basilika", false),
	}
	dupA := append(append([]domain.Ingredient{}, common...), ing("Oregano", "oregano", false))
	dupB := append([]domain.Ingredient{}, common...)
	recipes := []domain.Recipe{
		recWith("dup_a", "nöt", "", nil, dupA...),
		recWith("dup_b", "fågel", "", nil, dupB...),
		recWith("other", "fisk", "", nil, ing("Lax", "lax", false), ing("Citron", "citron", false)),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(2, []bool{false, false}, []string{"", ""})
	got := map[string]bool{week[0]: true, week[1]: true}
	if got["dup_a"] && got["dup_b"] {
		t.Errorf("near-duplicates should not both be chosen, got %v", week)
	}
}

func TestSelectWeek_RecencyDeprioritizes(t *testing.T) {
	recipes := []domain.Recipe{
		recWith("fresh", "nöt", "", nil, ing("Köttfärs", "köttfärs", false)),
		recWith("recent", "fisk", "", nil, ing("Lax", "lax", false)),
	}
	recency := map[string]bool{"recent": true}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), recency, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	// Single slot: should prefer the non-recent recipe.
	week := sel.SelectWeek(1, []bool{false}, []string{""})
	if week[0] != "fresh" {
		t.Errorf("expected recency to deprioritize 'recent', got %v", week)
	}
}

func TestSelectWeek_LocksHonored(t *testing.T) {
	recipes := []domain.Recipe{
		recWith("locked", "nöt", "", nil, ing("Lök", "lök", false)),
		recWith("free_1", "fisk", "", nil, ing("Lax", "lax", false)),
		recWith("free_2", "fågel", "", nil, ing("Kyckling", "kyckling", false)),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(3, []bool{false, false, false}, []string{"", "locked", ""})
	if week[1] != "locked" {
		t.Errorf("expected slot 1 to stay locked, got %v", week)
	}
	// Locked recipe must not be re-picked in an open slot.
	if week[0] == "locked" || week[2] == "locked" {
		t.Errorf("locked recipe duplicated into an open slot: %v", week)
	}
}

func TestSelectWeek_UnknownLockTreatedAsOpen(t *testing.T) {
	recipes := []domain.Recipe{
		rec("a"), rec("b"),
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(2, []bool{false, false}, []string{"ghost_recipe", ""})
	for i, id := range week {
		if id == "" || id == "ghost_recipe" {
			t.Errorf("unknown lock should be filled with a real recipe, slot %d = %q", i, id)
		}
	}
}

func TestSelectWeek_SkipsStayEmpty(t *testing.T) {
	recipes := []domain.Recipe{rec("a"), rec("b"), rec("c")}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(3, []bool{false, true, false}, []string{"", "", ""})
	if week[1] != "" {
		t.Errorf("skip slot should be empty, got %q", week[1])
	}
	if week[0] == "" || week[2] == "" {
		t.Errorf("non-skip slots should be filled, got %v", week)
	}
}

func TestSelectWeek_VegQuotaReducedByLockedVeg(t *testing.T) {
	// VegetarianDays=2, one locked veg → only one more veg needed.
	recipes := []domain.Recipe{
		rec("veg_locked", "vegetariskt"),
		rec("veg_free", "vegetariskt"),
		rec("meat_1"),
		rec("meat_2"),
		rec("meat_3"),
	}
	prefs := defaultPrefs()
	prefs.VegetarianDays = 2
	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(4, make([]bool, 4), []string{"veg_locked", "", "", ""})
	vegCount := 0
	for _, id := range week {
		if id == "veg_locked" || id == "veg_free" {
			vegCount++
		}
	}
	// At least the quota of 2 (locked + one more) is satisfied.
	if vegCount < 2 {
		t.Errorf("expected veg quota of 2 satisfied (incl. lock), got %d in %v", vegCount, week)
	}
}

func TestSelectWeek_NoMetadataDoesNotPanic(t *testing.T) {
	// Recipes with no ingredients / no metadata must be handled gracefully.
	recipes := []domain.Recipe{
		{ID: "a", Name: "a"},
		{ID: "b", Name: "b"},
	}
	sel, ok := newMenuSelector(recipes, defaultPrefs(), nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := selectWeek(t, sel, 3)
	for i, id := range week {
		if id == "" {
			t.Errorf("slot %d empty in %v", i, week)
		}
	}
}

func TestDietCompatible(t *testing.T) {
	tests := []struct {
		profile string
		class   string
		want    bool
	}{
		{"", "omnivore", true},          // empty profile accepts all
		{"vegetarian", "", true},        // empty class graceful
		{"vegetarian", "vegetarian", true},
		{"vegetarian", "vegan", true},
		{"vegetarian", "omnivore", false},
		{"vegan", "vegan", true},
		{"vegan", "vegetarian", false},
		{"pescetarian", "pescetarian", true},
		{"pescetarian", "vegetarian", true},
		{"pescetarian", "vegan", true},
		{"pescetarian", "omnivore", false},
		{"omnivore", "omnivore", true},
		{"omnivore", "vegan", true},
	}
	for _, tt := range tests {
		if got := domain.DietCompatible(tt.profile, tt.class); got != tt.want {
			t.Errorf("DietCompatible(%q,%q)=%v want %v", tt.profile, tt.class, got, tt.want)
		}
	}
}

func TestNewMenuSelector_DietProfileFilter(t *testing.T) {
	// vegetarian profile excludes omnivore, keeps vegetarian and unknown-class.
	recipes := []domain.Recipe{
		recWith("veg", "", "vegetarian", nil, ing("Halloumi", "halloumi", false)),
		recWith("omni", "nöt", "omnivore", nil, ing("Köttfärs", "köttfärs", false)),
		recWith("unknown", "", "", nil, ing("Pasta", "pasta", false)),
	}
	prefs := defaultPrefs()
	prefs.DietProfile = "vegetarian"
	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	if _, has := sel.byID["omni"]; has {
		t.Errorf("omnivore recipe should be filtered out under vegetarian profile")
	}
	if _, has := sel.byID["veg"]; !has {
		t.Errorf("vegetarian recipe should survive")
	}
	if _, has := sel.byID["unknown"]; !has {
		t.Errorf("unknown-class recipe should survive (graceful)")
	}
}

func TestNewMenuSelector_DislikedIngredientsFilter(t *testing.T) {
	recipes := []domain.Recipe{
		recWith("has_coriander", "", "", nil, ing("Färsk koriander", "koriander", false)),
		recWith("has_coriander_raw", "", "", nil, domain.Ingredient{Name: "Koriander", Amount: 1, Unit: "knippe"}),
		recWith("clean", "", "", nil, ing("Persilja", "persilja", false)),
	}
	prefs := defaultPrefs()
	prefs.DislikedIngredients = []string{"Koriander"} // case-insensitive
	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	if _, has := sel.byID["has_coriander"]; has {
		t.Errorf("recipe with disliked canonical should be filtered")
	}
	if _, has := sel.byID["has_coriander_raw"]; has {
		t.Errorf("recipe with disliked raw name (fallback) should be filtered")
	}
	if _, has := sel.byID["clean"]; !has {
		t.Errorf("clean recipe should survive")
	}
}

// recWithProtein builds a recipe carrying per-serving macros (protein only set),
// for nutrition-target tests.
func recWithProtein(id, protein string, gramsProtein float64, ingredients ...domain.Ingredient) domain.Recipe {
	r := recWith(id, protein, "", nil, ingredients...)
	r.Macros = &domain.Macros{Protein: gramsProtein}
	return r
}

func TestSelectWeek_NutritionTargetPrefersHigherProtein(t *testing.T) {
	// Two single-protein recipes with no overlap; only protein differs. With a
	// positive per-day target the high-protein one wins a single open slot.
	recipes := []domain.Recipe{
		recWithProtein("low", "nöt", 5, ing("Sallad", "sallad", false)),
		recWithProtein("high", "fisk", 60, ing("Lax", "lax", false)),
	}
	prefs := defaultPrefs()
	prefs.ProteinTargetPerDay = 100
	sel, ok := newMenuSelector(recipes, prefs, nil, testRNG())
	if !ok {
		t.Fatalf("expected selector to build")
	}
	week := sel.SelectWeek(1, []bool{false}, []string{""})
	if week[0] != "high" {
		t.Errorf("expected high-protein recipe with a protein target, got %v", week)
	}
}

func TestSelectWeek_NutritionTargetZeroIsRegression(t *testing.T) {
	// With target 0 the nutrition term is inert: the selected week must be
	// byte-identical to one built from a selector whose recipes carry NO macros.
	withMacros := []domain.Recipe{
		recWithProtein("a", "nöt", 30, ing("Lök", "lök", false), ing("Köttfärs", "köttfärs", false)),
		recWithProtein("b", "fågel", 40, ing("Lök", "lök", false), ing("Kyckling", "kyckling", false)),
		recWithProtein("c", "fisk", 25, ing("Lax", "lax", false)),
		recWithProtein("d", "fläsk", 35, ing("Fläsk", "fläsk", false)),
	}
	noMacros := []domain.Recipe{
		recWith("a", "nöt", "", nil, ing("Lök", "lök", false), ing("Köttfärs", "köttfärs", false)),
		recWith("b", "fågel", "", nil, ing("Lök", "lök", false), ing("Kyckling", "kyckling", false)),
		recWith("c", "fisk", "", nil, ing("Lax", "lax", false)),
		recWith("d", "fläsk", "", nil, ing("Fläsk", "fläsk", false)),
	}
	build := func(recipes []domain.Recipe) []string {
		sel, _ := newMenuSelector(recipes, defaultPrefs(), nil, rand.New(rand.NewPCG(123, 456)))
		return selectWeek(t, sel, 4)
	}
	a := build(withMacros) // target 0 (default prefs)
	b := build(noMacros)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("target=0 changed selection: with-macros %v vs no-macros %v", a, b)
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
