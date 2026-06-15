package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestIngredientMetadata_RoundTrip verifies the new metadata fields marshal and
// unmarshal correctly inside the ingredients JSON.
func TestIngredientMetadata_RoundTrip(t *testing.T) {
	in := Ingredient{
		Name:           "riven ost",
		Amount:         100,
		Unit:           "g",
		CanonicalName:  "ost",
		GramsEquiv:     100,
		IsPantryStaple: false,
		IsPerishable:   true,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Ingredient
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != in {
		t.Errorf("round-trip mismatch: got %+v want %+v", out, in)
	}
}

// TestIngredient_OldJSON_BackwardCompat is the key backward-compat case: JSON
// written before Phase 0 (no metadata keys) must unmarshal to zero-value
// metadata, never an error.
func TestIngredient_OldJSON_BackwardCompat(t *testing.T) {
	const oldJSON = `{"name":"köttfärs","amount":400,"unit":"g"}`
	var ing Ingredient
	if err := json.Unmarshal([]byte(oldJSON), &ing); err != nil {
		t.Fatalf("unmarshal old JSON: %v", err)
	}
	if ing.Name != "köttfärs" || ing.Amount != 400 || ing.Unit != "g" {
		t.Errorf("core fields not preserved: %+v", ing)
	}
	if ing.CanonicalName != "" || ing.GramsEquiv != 0 || ing.IsPantryStaple || ing.IsPerishable {
		t.Errorf("expected zero-value metadata for old JSON, got %+v", ing)
	}
}

// TestIngredient_OmitemptyOmitsAbsentFields ensures unknown metadata is omitted
// from serialized JSON (so the wire stays clean for unknown values).
func TestIngredient_OmitemptyOmitsAbsentFields(t *testing.T) {
	data, err := json.Marshal(Ingredient{Name: "lök", Amount: 1, Unit: "st"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	for _, key := range []string{"canonicalName", "gramsEquiv", "isPantryStaple", "isPerishable"} {
		if strings.Contains(s, key) {
			t.Errorf("expected %q omitted from %s", key, s)
		}
	}
}

// TestRecipe_MetadataRoundTrip checks recipe-level metadata round-trips.
func TestRecipe_MetadataRoundTrip(t *testing.T) {
	in := Recipe{
		ID:          "rec_1",
		Name:        "Kycklinggryta",
		Servings:    4,
		Tags:        []string{"vardag"},
		Ingredients: []Ingredient{{Name: "kyckling", Amount: 500, Unit: "g"}},
		MainProtein: "kyckling",
		DietClass:   DietClassOmnivore,
		Batchable:   true,
		CookMinutes: 45,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Recipe
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.MainProtein != "kyckling" || out.DietClass != DietClassOmnivore || !out.Batchable || out.CookMinutes != 45 {
		t.Errorf("recipe metadata mismatch: %+v", out)
	}
}

// TestRecipeSummary_MetadataRoundTrip checks summary metadata round-trips.
func TestRecipeSummary_MetadataRoundTrip(t *testing.T) {
	in := RecipeSummary{
		ID:          "rec_1",
		Name:        "Linsgryta",
		Servings:    4,
		Tags:        []string{"vegan"},
		MainProtein: "linser",
		DietClass:   DietClassVegan,
		Batchable:   true,
		CookMinutes: 30,
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out RecipeSummary
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.MainProtein != "linser" || out.DietClass != DietClassVegan || !out.Batchable || out.CookMinutes != 30 {
		t.Errorf("summary metadata mismatch: %+v", out)
	}
}

// TestRecipe_OmitemptyOmitsUnknownMetadata ensures unknown recipe metadata is
// omitted from the wire.
func TestRecipe_OmitemptyOmitsUnknownMetadata(t *testing.T) {
	data, err := json.Marshal(Recipe{ID: "rec_1", Name: "X", Servings: 4})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	for _, key := range []string{"mainProtein", "dietClass", "batchable", "cookMinutes"} {
		if strings.Contains(s, key) {
			t.Errorf("expected %q omitted from %s", key, s)
		}
	}
}

func TestIsValidDietClass(t *testing.T) {
	valid := []string{"", DietClassOmnivore, DietClassVegetarian, DietClassVegan, DietClassPescetarian}
	for _, v := range valid {
		if !IsValidDietClass(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	invalid := []string{"keto", "paleo", "OMNIVORE", "veg", "0", "false"}
	for _, v := range invalid {
		if IsValidDietClass(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

// TestValidate_DietClass exercises the CreateRecipeRequest.Validate diet-class
// path and the CanonicalName length bound.
func TestValidate_DietClass(t *testing.T) {
	base := func() CreateRecipeRequest {
		return CreateRecipeRequest{
			Name:         "Test",
			Servings:     4,
			Tags:         []string{},
			Ingredients:  []Ingredient{{Name: "x", Amount: 1, Unit: "st"}},
			Instructions: []string{"do it"},
		}
	}

	t.Run("empty diet class accepted", func(t *testing.T) {
		r := base()
		if err := r.Validate(); err != nil {
			t.Errorf("expected empty diet class accepted, got %v", err)
		}
	})

	t.Run("valid diet class accepted", func(t *testing.T) {
		r := base()
		r.DietClass = DietClassVegan
		if err := r.Validate(); err != nil {
			t.Errorf("expected vegan accepted, got %v", err)
		}
	})

	t.Run("garbage diet class rejected", func(t *testing.T) {
		r := base()
		r.DietClass = "keto"
		if err := r.Validate(); err != ErrInvalidDietClass {
			t.Errorf("expected ErrInvalidDietClass, got %v", err)
		}
	})

	t.Run("over-long canonical name rejected", func(t *testing.T) {
		r := base()
		r.Ingredients[0].CanonicalName = strings.Repeat("a", 81)
		if err := r.Validate(); err == nil {
			t.Error("expected error for over-long canonical name")
		}
	})
}
