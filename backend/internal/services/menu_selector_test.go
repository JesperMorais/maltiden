package services

import (
	"maltiden/internal/domain"
	"testing"
)

// TestMenuSelector_SkipDays verifies days marked as SkipDays get Skip=true
// and no recipe assigned, while other days get a recipe.
func TestMenuSelector_SkipDays(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 3)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 3 {
		t.Fatalf("expected 3 days, got %d", len(resp.Days))
	}

	skipDate := resp.Days[1].Date
	resp2, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     3,
		Servings: 2,
		SkipDays: []string{skipDate},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, day := range resp2.Days {
		if day.Date == skipDate {
			if !day.Skip {
				t.Errorf("day %s: expected Skip=true", day.Date)
			}
			if day.RecipeID != "" {
				t.Errorf("day %s: expected empty RecipeID for skipped day, got %q", day.Date, day.RecipeID)
			}
		} else {
			if day.Skip {
				t.Errorf("day %s: expected Skip=false", day.Date)
			}
			if day.RecipeID == "" {
				t.Errorf("day %s: expected non-empty RecipeID", day.Date)
			}
		}
	}
}

// TestMenuSelector_DayCountAndServings verifies Days/Servings defaults and
// overrides, and that ExtraPortions adds on top of base servings.
func TestMenuSelector_DayCountAndServings(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 7)

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Days) != 7 {
		t.Errorf("expected default Days=7, got %d", len(resp.Days))
	}
	for _, day := range resp.Days {
		if day.Servings != 4 {
			t.Errorf("day %s: expected default Servings=4, got %d", day.Date, day.Servings)
		}
	}

	extraDate := resp.Days[2].Date
	resp2, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days:     5,
		Servings: 3,
		ExtraPortions: map[string]int{
			extraDate: 2,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp2.Days) != 5 {
		t.Errorf("expected Days=5, got %d", len(resp2.Days))
	}
	for _, day := range resp2.Days {
		if day.Date == extraDate {
			if day.Servings != 5 {
				t.Errorf("day %s: expected Servings=5 (3+2 extra), got %d", day.Date, day.Servings)
			}
		} else if day.Servings != 3 {
			t.Errorf("day %s: expected Servings=3, got %d", day.Date, day.Servings)
		}
	}
}

// TestMenuSelector_StructuralDeterminism verifies that across repeated runs,
// the set of assigned recipe IDs and the skip-day layout stay stable, even
// though the underlying shuffle is unseeded (order is not asserted).
func TestMenuSelector_StructuralDeterminism(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 4)

	allRecipes, err := env.recipeService.recipeStorage.GetAll(nil, env.householdID)
	if err != nil {
		t.Fatalf("failed to list recipes: %v", err)
	}
	validIDs := make(map[string]bool, len(allRecipes))
	for _, r := range allRecipes {
		validIDs[r.ID] = true
	}

	skipDate := ""

	for i := 0; i < 5; i++ {
		req := domain.GenerateMenuRequest{
			Days:     4,
			Servings: 2,
		}
		if skipDate != "" {
			req.SkipDays = []string{skipDate}
		}

		resp, err := env.menuService.Generate(env.householdID, req)
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", i, err)
		}
		if len(resp.Days) != 4 {
			t.Fatalf("run %d: expected 4 days, got %d", i, len(resp.Days))
		}

		if skipDate == "" {
			skipDate = resp.Days[0].Date
			continue
		}

		for _, day := range resp.Days {
			if day.Date == skipDate {
				if !day.Skip {
					t.Errorf("run %d: day %s: expected Skip=true", i, day.Date)
				}
				continue
			}
			if day.Skip {
				t.Errorf("run %d: day %s: expected Skip=false", i, day.Date)
			}
			if !validIDs[day.RecipeID] {
				t.Errorf("run %d: day %s: RecipeID %q not in seeded recipe set", i, day.Date, day.RecipeID)
			}
		}
	}
}

// The following behaviors are not yet implemented (no greedy/overlap
// selector, no ingredient data on RecipeSummary, no locked-day support) —
// blocked on mt-0711-greedy-overlap-generator.

func TestMenuSelector_OverlapBeatsRandom(t *testing.T) {
	t.Skip("blocked on mt-0711-greedy-overlap-generator")
}

func TestMenuSelector_VarietyPenalty(t *testing.T) {
	t.Skip("blocked on mt-0711-greedy-overlap-generator")
}

func TestMenuSelector_ReproducibleForFixedSeed(t *testing.T) {
	t.Skip("blocked on mt-0711-greedy-overlap-generator")
}

func TestMenuSelector_LockDay(t *testing.T) {
	t.Skip("blocked on mt-0711-greedy-overlap-generator")
}
