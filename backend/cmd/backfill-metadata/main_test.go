package main

import (
	"testing"

	"maltiden/internal/domain"
	"maltiden/pkg/claude"
)

func TestMergeMetadata_AppliesRecipeAndIngredientFields(t *testing.T) {
	r := &domain.Recipe{
		ID:       "rec_1",
		Name:     "Kycklinggryta",
		Servings: 4,
		Ingredients: []domain.Ingredient{
			{Name: "kycklingfilé", Amount: 500, Unit: "g"},
			{Name: "salt", Amount: 1, Unit: "tsk"},
		},
	}
	meta := &recipeMetadata{
		MainProtein: "kyckling",
		DietClass:   domain.DietClassOmnivore,
		Batchable:   true,
		CookMinutes: 40,
		Ingredients: []ingredientMetadata{
			{Name: "kycklingfilé", CanonicalName: "kyckling", GramsEquiv: 500, IsPerishable: true},
			{Name: "salt", CanonicalName: "salt", IsPantryStaple: true},
		},
	}

	mergeMetadata(r, meta)

	if r.MainProtein != "kyckling" || r.DietClass != domain.DietClassOmnivore || !r.Batchable || r.CookMinutes != 40 {
		t.Errorf("recipe metadata not merged: %+v", r)
	}
	if r.Ingredients[0].CanonicalName != "kyckling" || r.Ingredients[0].GramsEquiv != 500 || !r.Ingredients[0].IsPerishable {
		t.Errorf("first ingredient metadata not merged: %+v", r.Ingredients[0])
	}
	if r.Ingredients[1].CanonicalName != "salt" || !r.Ingredients[1].IsPantryStaple {
		t.Errorf("second ingredient metadata not merged: %+v", r.Ingredients[1])
	}
}

func TestMergeMetadata_DropsInvalidDietClass(t *testing.T) {
	r := &domain.Recipe{ID: "rec_1", DietClass: ""}
	mergeMetadata(r, &recipeMetadata{DietClass: "keto"})
	if r.DietClass != "" {
		t.Errorf("expected invalid diet class dropped, got %q", r.DietClass)
	}
}

func TestMergeMetadata_NameFallbackLeavesUnmatchedUntouched(t *testing.T) {
	// Unequal counts force the name-matching fallback. A metadata entry whose
	// name matches no ingredient must leave every ingredient untouched.
	r := &domain.Recipe{
		ID: "rec_1",
		Ingredients: []domain.Ingredient{
			{Name: "lök", Amount: 1, Unit: "st"},
			{Name: "salt", Amount: 1, Unit: "tsk"},
		},
	}
	mergeMetadata(r, &recipeMetadata{
		Ingredients: []ingredientMetadata{{Name: "vitlök", CanonicalName: "vitlök"}},
	})
	if r.Ingredients[0].CanonicalName != "" || r.Ingredients[1].CanonicalName != "" {
		t.Errorf("expected unmatched ingredients untouched, got %+v", r.Ingredients)
	}
}

func TestMergeMetadata_NameFallbackMatchesCaseInsensitively(t *testing.T) {
	// Unequal counts → name fallback, which normalizes case/whitespace.
	r := &domain.Recipe{
		ID: "rec_1",
		Ingredients: []domain.Ingredient{
			{Name: "Riven Ost", Amount: 100, Unit: "g"},
			{Name: "salt", Amount: 1, Unit: "tsk"},
		},
	}
	mergeMetadata(r, &recipeMetadata{
		Ingredients: []ingredientMetadata{{Name: "riven ost", CanonicalName: "ost"}},
	})
	if r.Ingredients[0].CanonicalName != "ost" {
		t.Errorf("expected case-insensitive name match, got %+v", r.Ingredients[0])
	}
}

func TestMergeMetadata_PositionalMatchWhenCountsEqual(t *testing.T) {
	// When counts match, metadata is applied positionally — robust to the model
	// echoing a reworded/normalized ingredient name.
	r := &domain.Recipe{
		ID: "rec_1",
		Ingredients: []domain.Ingredient{
			{Name: "kycklingfilé", Amount: 500, Unit: "g"},
		},
	}
	mergeMetadata(r, &recipeMetadata{
		Ingredients: []ingredientMetadata{{Name: "kyckling (filé)", CanonicalName: "kyckling", GramsEquiv: 500}},
	})
	if r.Ingredients[0].CanonicalName != "kyckling" || r.Ingredients[0].GramsEquiv != 500 {
		t.Errorf("expected positional match, got %+v", r.Ingredients[0])
	}
}

func TestExtractMetadata_ParsesStructuredJSON(t *testing.T) {
	msg := &claude.Response{
		Content: []claude.ContentBlock{
			{Type: "text", Text: `{"mainProtein":"lax","dietClass":"pescetarian","batchable":false,"cookMinutes":20,"ingredients":[]}`},
		},
	}
	meta, err := extractMetadata(msg)
	if err != nil {
		t.Fatalf("extractMetadata: %v", err)
	}
	if meta.MainProtein != "lax" || meta.DietClass != domain.DietClassPescetarian || meta.CookMinutes != 20 {
		t.Errorf("unexpected metadata: %+v", meta)
	}
}

func TestExtractMetadata_NonTextContentErrors(t *testing.T) {
	if _, err := extractMetadata(&claude.Response{}); err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestBuildBatchRequest_UsesHaikuAndStructuredOutput(t *testing.T) {
	r := &domain.Recipe{
		ID:           "rec_1",
		Name:         "Test",
		Servings:     4,
		Ingredients:  []domain.Ingredient{{Name: "x", Amount: 1, Unit: "st"}},
		Instructions: []string{"do it"},
	}
	req := buildBatchRequest(r)

	if req.CustomID != "rec_1" {
		t.Errorf("expected custom_id = recipe ID, got %q", req.CustomID)
	}
	if req.Params.Model != claude.HaikuModel {
		t.Errorf("expected Haiku model, got %q", req.Params.Model)
	}
	if req.Params.OutputConfig == nil || req.Params.OutputConfig.Format == nil ||
		req.Params.OutputConfig.Format.Type != "json_schema" {
		t.Errorf("expected json_schema structured output, got %+v", req.Params.OutputConfig)
	}
}
