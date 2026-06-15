package services

import (
	"maltiden/internal/domain"
	"math/rand/v2"
	"testing"
)

// seedBatchableRecipes creates a mix of batchable and non-batchable recipes with
// distinct ingredients so the prep selector has a clear batch candidate.
func (env *menuTestEnv) seedBatchableRecipes(t *testing.T) {
	t.Helper()
	specs := []struct {
		name      string
		protein   string
		batchable bool
		ingr      string
	}{
		{"Chili con carne", "nöt", true, "köttfärs"},
		{"Lasagne", "nöt", true, "lasagneplattor"},
		{"Laxpasta", "fisk", false, "lax"},
		{"Kycklinggryta", "fågel", false, "kyckling"},
		{"Korv stroganoff", "fläsk", false, "falukorv"},
	}
	for _, s := range specs {
		req := domain.CreateRecipeRequest{
			Name:         s.name,
			Servings:     4,
			MainProtein:  s.protein,
			Batchable:    s.batchable,
			Ingredients:  []domain.Ingredient{{Name: s.ingr, CanonicalName: s.ingr, Amount: 1, Unit: "st"}},
			Instructions: []string{"Cook"},
		}
		if _, err := env.recipeService.Create(req, env.householdID); err != nil {
			t.Fatalf("seed batchable recipe %q: %v", s.name, err)
		}
	}
}

// findPrepPair locates the cook/leftover day pair in a generated response.
func findPrepPair(days []domain.MenuResponseDay) (cook, leftover *domain.MenuResponseDay) {
	for i := range days {
		if days[i].Leftover {
			leftover = &days[i]
		}
	}
	if leftover == nil {
		return nil, nil
	}
	for i := range days {
		if days[i].Date == leftover.CookDate {
			cook = &days[i]
		}
	}
	return cook, leftover
}

func TestMenuGenerate_PrepModePlacesBatchPair(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env.seedBatchableRecipes(t)

	prep := true
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		PrepMode: &prep,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	cook, leftover := findPrepPair(resp.Days)
	if cook == nil || leftover == nil {
		t.Fatalf("expected a cook+leftover pair in prep mode, got days %+v", resp.Days)
	}
	if cook.Servings != domain.BatchMultiplier*4 {
		t.Errorf("cook day servings = %d, want %d (2× base)", cook.Servings, domain.BatchMultiplier*4)
	}
	if leftover.Servings != 4 {
		t.Errorf("leftover day servings = %d, want 4 (base)", leftover.Servings)
	}
	if leftover.RecipeID != cook.RecipeID {
		t.Errorf("leftover recipe %q must match cook recipe %q", leftover.RecipeID, cook.RecipeID)
	}
	if leftover.CookDate != cook.Date {
		t.Errorf("leftover.CookDate = %q, want cook date %q", leftover.CookDate, cook.Date)
	}
	if cook.Leftover {
		t.Errorf("cook day must not be flagged leftover")
	}
}

func TestMenuGenerate_PrepModeOffNoPairs(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env.seedBatchableRecipes(t)

	prep := false
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		PrepMode: &prep,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for i, d := range resp.Days {
		if d.Leftover || d.CookDate != "" {
			t.Errorf("day %d unexpectedly flagged leftover with prep off: %+v", i, d)
		}
		if d.Servings != 4 {
			t.Errorf("day %d servings = %d, want 4 (no batching)", i, d.Servings)
		}
	}
}

func TestMenuGenerate_PrepModeFallsBackToPreference(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env.seedBatchableRecipes(t)

	// Default preference enables prep; request omits PrepMode (nil pointer).
	if _, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     5,
		DefaultServings: 4,
		PrepModeDefault: true,
	}); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		// PrepMode omitted → falls back to PrepModeDefault=true.
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	cook, leftover := findPrepPair(resp.Days)
	if cook == nil || leftover == nil {
		t.Fatalf("expected prep pair from PrepModeDefault, got %+v", resp.Days)
	}

	// Explicit false in the request must override the true default.
	prep := false
	resp2, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		PrepMode: &prep,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	c2, l2 := findPrepPair(resp2.Days)
	if c2 != nil || l2 != nil {
		t.Errorf("explicit prepMode=false must override the true default, got pair")
	}
}

func TestMenuGenerate_PrepModeRoundTripsThroughStorage(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env.seedBatchableRecipes(t)

	prep := true
	gen, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 4,
		PrepMode: &prep,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	genCook, genLeftover := findPrepPair(gen.Days)
	if genCook == nil || genLeftover == nil {
		t.Fatalf("expected a prep pair, got %+v", gen.Days)
	}

	got, err := env.menuService.GetCurrent(env.householdID)
	if err != nil {
		t.Fatalf("GetCurrent: %v", err)
	}
	cook, leftover := findPrepPair(got.Days)
	if cook == nil || leftover == nil {
		t.Fatalf("leftover/cookDate must survive storage round-trip, got %+v", got.Days)
	}
	if leftover.CookDate != cook.Date || leftover.RecipeID != cook.RecipeID {
		t.Errorf("round-trip mismatch: leftover %+v cook %+v", leftover, cook)
	}
	if cook.Servings != domain.BatchMultiplier*4 || leftover.Servings != 4 {
		t.Errorf("round-trip servings mismatch: cook=%d leftover=%d", cook.Servings, leftover.Servings)
	}
}

func TestMenuGenerate_PrepExtraPortionsFoldIntoCookDay(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env.seedBatchableRecipes(t)

	prep := true
	// First find which dates form the pair (deterministic seed), then re-generate
	// with extra portions on the leftover date and assert they fold into the cook.
	probe, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days: 5, Servings: 4, PrepMode: &prep,
	})
	if err != nil {
		t.Fatalf("probe Generate: %v", err)
	}
	pc, pl := findPrepPair(probe.Days)
	if pc == nil || pl == nil {
		t.Fatalf("probe expected a pair")
	}

	// New household, same seed, same recipes, so the pair lands on the same dates.
	env2 := newMenuTestEnv(t)
	env2.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })
	env2.seedBatchableRecipes(t)
	resp, err := env2.menuService.Generate(env2.householdID, domain.GenerateMenuRequest{
		Days:          5,
		Servings:      4,
		PrepMode:      &prep,
		ExtraPortions: map[string]int{pl.Date: 3},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	cook, leftover := findPrepPair(resp.Days)
	if cook == nil || leftover == nil {
		t.Fatalf("expected pair, got %+v", resp.Days)
	}
	// Extra portions on the leftover date fold into the cook day's batch.
	if cook.Servings != domain.BatchMultiplier*4+3 {
		t.Errorf("cook servings = %d, want %d (2×4 + 3 folded)", cook.Servings, domain.BatchMultiplier*4+3)
	}
	if leftover.Servings != 4 {
		t.Errorf("leftover servings = %d, want 4 (extras folded to cook)", leftover.Servings)
	}
}

func TestUpdateCurrent_InconsistentLeftoverDemoted(t *testing.T) {
	env := newMenuTestEnv(t)
	env.menuService.WithRNG(func() *rand.Rand { return rand.New(rand.NewPCG(7, 7)) })

	// Two distinct recipes with known ids.
	recA, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Köttfärssås",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Köttfärs", CanonicalName: "köttfärs", Amount: 500, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, env.householdID)
	if err != nil {
		t.Fatalf("create recA: %v", err)
	}
	recB, err := env.recipeService.Create(domain.CreateRecipeRequest{
		Name:         "Laxpasta",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "Lax", CanonicalName: "lax", Amount: 300, Unit: "g"}},
		Instructions: []string{"Koka"},
	}, env.householdID)
	if err != nil {
		t.Fatalf("create recB: %v", err)
	}

	// A menu must exist for UpdateCurrent to target.
	if _, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{Days: 4, Servings: 4}); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Craft four days:
	//   d0 cook recA (valid cook day)
	//   d1 VALID leftover of recA → cookDate=d0, recipe matches  → stays leftover
	//   d2 INVALID leftover: cookDate points to d0 but recipe is recB (mismatch)
	//   d3 INVALID leftover: cookDate=d99 dangles (no such day)
	d0, d1, d2, d3 := "2026-03-01", "2026-03-02", "2026-03-03", "2026-03-04"
	req := domain.UpdateMenuRequest{Days: []domain.MenuDay{
		{Date: d0, RecipeID: recA.ID, Servings: 8},
		{Date: d1, RecipeID: recA.ID, Servings: 4, Leftover: true, CookDate: d0},
		{Date: d2, RecipeID: recB.ID, Servings: 4, Leftover: true, CookDate: d0},
		{Date: d3, RecipeID: recB.ID, Servings: 4, Leftover: true, CookDate: "2026-12-99"},
	}}
	resp, err := env.menuService.UpdateCurrent(env.householdID, req)
	if err != nil {
		t.Fatalf("UpdateCurrent: %v", err)
	}

	byDate := make(map[string]domain.MenuResponseDay, len(resp.Days))
	for _, d := range resp.Days {
		byDate[d.Date] = d
	}

	// d1: consistent leftover survives.
	if !byDate[d1].Leftover || byDate[d1].CookDate != d0 {
		t.Errorf("d1 valid leftover should be preserved, got %+v", byDate[d1])
	}
	// d2: recipe mismatch → demoted.
	if byDate[d2].Leftover || byDate[d2].CookDate != "" {
		t.Errorf("d2 mismatched leftover should be demoted to normal day, got %+v", byDate[d2])
	}
	// d3: dangling cookDate → demoted.
	if byDate[d3].Leftover || byDate[d3].CookDate != "" {
		t.Errorf("d3 dangling leftover should be demoted to normal day, got %+v", byDate[d3])
	}
	// Demoted days keep their recipe so shopping still buys their ingredients.
	if byDate[d2].RecipeID != recB.ID || byDate[d3].RecipeID != recB.ID {
		t.Errorf("demoted days must retain their recipeId, got d2=%q d3=%q", byDate[d2].RecipeID, byDate[d3].RecipeID)
	}
}

func TestMenuPreferences_PrepModeDefaultRoundTrip(t *testing.T) {
	env := newMenuTestEnv(t)

	if _, err := env.menuService.UpdatePreferences(env.householdID, domain.UpdateMenuPreferencesRequest{
		DefaultDays:     7,
		DefaultServings: 4,
		PrepModeDefault: true,
	}); err != nil {
		t.Fatalf("UpdatePreferences: %v", err)
	}
	got, err := env.menuService.GetPreferences(env.householdID)
	if err != nil {
		t.Fatalf("GetPreferences: %v", err)
	}
	if !got.PrepModeDefault {
		t.Errorf("expected PrepModeDefault=true after round-trip, got %+v", got)
	}
}
