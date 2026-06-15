package services

import (
	"maltiden/internal/domain"
	"testing"
)

func TestComputeMenuEconomy_SharedAndCounts(t *testing.T) {
	recipes := map[string]domain.Recipe{
		"a": recWith("a", "nöt", "", nil,
			ing("Gul lök", "lök", false),
			ing("Grädde", "grädde", false),
			ing("Salt", "salt", true), // staple → excluded
		),
		"b": recWith("b", "fågel", "", nil,
			ing("Lök", "lök", false), // shares lök with a (different raw name)
			ing("Kyckling", "kyckling", false),
		),
	}

	econ := computeMenuEconomy([]string{"a", "b"}, recipes)

	// Distinct non-staple canonicals: lök, grädde, kyckling = 3.
	if econ.DistinctItemsToBuy != 3 {
		t.Errorf("DistinctItemsToBuy = %d, want 3", econ.DistinctItemsToBuy)
	}
	// Total refs: a has 2 (lök, grädde; salt excluded), b has 2 (lök, kyckling) = 4.
	if econ.TotalIngredientRefs != 4 {
		t.Errorf("TotalIngredientRefs = %d, want 4", econ.TotalIngredientRefs)
	}
	if len(econ.SharedIngredients) != 1 {
		t.Fatalf("expected 1 shared ingredient, got %d (%+v)", len(econ.SharedIngredients), econ.SharedIngredients)
	}
	shared := econ.SharedIngredients[0]
	if shared.CanonicalName != "lök" {
		t.Errorf("shared canonical = %q, want lök", shared.CanonicalName)
	}
	if shared.RecipeCount != 2 {
		t.Errorf("shared recipeCount = %d, want 2", shared.RecipeCount)
	}
	// Name is the first raw name seen for the canonical ("Gul lök" from recipe a).
	if shared.Name != "Gul lök" {
		t.Errorf("shared name = %q, want 'Gul lök'", shared.Name)
	}
}

func TestComputeMenuEconomy_StaplesExcludedFromAll(t *testing.T) {
	recipes := map[string]domain.Recipe{
		"a": recWith("a", "", "", nil,
			ing("Salt", "salt", true),
			ing("Olivolja", "olivolja", true),
			ing("Lök", "lök", false),
		),
		"b": recWith("b", "", "", nil,
			ing("Salt", "salt", true),
			ing("Lök", "lök", false),
		),
	}
	econ := computeMenuEconomy([]string{"a", "b"}, recipes)
	// Only "lök" is non-staple; shared across both.
	if econ.DistinctItemsToBuy != 1 {
		t.Errorf("DistinctItemsToBuy = %d, want 1", econ.DistinctItemsToBuy)
	}
	if econ.TotalIngredientRefs != 2 {
		t.Errorf("TotalIngredientRefs = %d, want 2", econ.TotalIngredientRefs)
	}
	for _, s := range econ.SharedIngredients {
		if s.CanonicalName == "salt" || s.CanonicalName == "olivolja" {
			t.Errorf("staple %q leaked into shared ingredients", s.CanonicalName)
		}
	}
}

func TestComputeMenuEconomy_FallbackStaples(t *testing.T) {
	// No metadata: fallback staple list applies (salt, smör excluded).
	recipes := map[string]domain.Recipe{
		"a": {ID: "a", Name: "a", Ingredients: []domain.Ingredient{
			{Name: "Salt", Amount: 1, Unit: "tsk"},
			{Name: "Smör", Amount: 1, Unit: "msk"},
			{Name: "Morötter", Amount: 3, Unit: "st"},
		}},
	}
	econ := computeMenuEconomy([]string{"a"}, recipes)
	if econ.DistinctItemsToBuy != 1 {
		t.Errorf("DistinctItemsToBuy = %d, want 1 (only morötter)", econ.DistinctItemsToBuy)
	}
	if econ.TotalIngredientRefs != 1 {
		t.Errorf("TotalIngredientRefs = %d, want 1", econ.TotalIngredientRefs)
	}
}

func TestComputeMenuEconomy_CapsTopShared(t *testing.T) {
	// Build many shared ingredients (>8) and verify the cap + sort order.
	canon := []string{"i1", "i2", "i3", "i4", "i5", "i6", "i7", "i8", "i9", "i10"}
	a := domain.Recipe{ID: "a", Name: "a"}
	b := domain.Recipe{ID: "b", Name: "b"}
	for _, c := range canon {
		a.Ingredients = append(a.Ingredients, ing(c, c, false))
		b.Ingredients = append(b.Ingredients, ing(c, c, false))
	}
	econ := computeMenuEconomy([]string{"a", "b"}, map[string]domain.Recipe{"a": a, "b": b})
	if len(econ.SharedIngredients) != economyMaxShared {
		t.Errorf("expected shared capped to %d, got %d", economyMaxShared, len(econ.SharedIngredients))
	}
}

func TestComputeMenuEconomy_MissingRecipesSkipped(t *testing.T) {
	econ := computeMenuEconomy([]string{"ghost"}, map[string]domain.Recipe{})
	if econ.DistinctItemsToBuy != 0 || econ.TotalIngredientRefs != 0 {
		t.Errorf("missing recipe should contribute nothing, got %+v", econ)
	}
	if len(econ.SharedIngredients) != 0 {
		t.Errorf("expected no shared ingredients, got %+v", econ.SharedIngredients)
	}
}
