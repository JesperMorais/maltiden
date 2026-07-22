package services

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"testing"

	"maltiden/internal/domain"
)

// routingGemini returns a wish payload or a dynamically-built arrange payload
// depending on which system prompt it sees, so one stub serves both calls.
type routingGemini struct {
	wish []byte
	// arrangeValid, when true, echoes an identity permutation parsed from the
	// prompt (guardrail passes); when false it returns a bogus assignment.
	arrangeValid bool
}

var slotLineRe = regexp.MustCompile(`slot (\d+) .*\[id: ([^\]]+)\]`)

func (g *routingGemini) GenerateJSON(_ context.Context, system, user string, _ map[string]interface{}) ([]byte, error) {
	if system == arrangeSystemPrompt {
		if !g.arrangeValid {
			return []byte(`{"assignments":[{"recipeId":"rec_DOES_NOT_EXIST","slot":0}],"rationale":"nope"}`), nil
		}
		type a struct {
			RecipeID string `json:"recipeId"`
			Slot     int    `json:"slot"`
		}
		var out struct {
			Assignments []a    `json:"assignments"`
			Rationale   string `json:"rationale"`
		}
		for _, m := range slotLineRe.FindAllStringSubmatch(user, -1) {
			slot, _ := strconv.Atoi(m[1])
			out.Assignments = append(out.Assignments, a{RecipeID: m[2], Slot: slot})
		}
		out.Rationale = "En balanserad vecka med delade råvaror."
		b, _ := json.Marshal(out)
		return b, nil
	}
	return g.wish, nil
}

func TestMergeWishConstraints(t *testing.T) {
	prefs := domain.MenuPreferences{
		ExcludedTags:        []string{"fisk"},
		DislikedIngredients: []string{},
		VegetarianDays:      0,
	}
	req := domain.GenerateMenuRequest{}
	pc := &domain.ParsedWishConstraints{
		VegetarianDays:           intpSvc(2),
		PrepMode:                 boolpSvc(true),
		ExtraExcludedTags:        []string{"Fisk", "skaldjur"}, // "Fisk" dedupes against "fisk"
		ExtraDislikedIngredients: []string{"koriander"},
	}
	mergeWishConstraints(&prefs, &req, pc)

	if prefs.VegetarianDays != 2 {
		t.Errorf("VegetarianDays = %d, want 2", prefs.VegetarianDays)
	}
	// A prep-mode wish enables batch cooking via the request flag (dev stores no
	// prep-mode preference).
	if !req.PrepMode {
		t.Error("req.PrepMode = false, want true after a prep-mode wish")
	}
	if len(prefs.ExcludedTags) != 2 { // fisk + skaldjur (Fisk deduped)
		t.Errorf("ExcludedTags = %v, want [fisk, skaldjur]", prefs.ExcludedTags)
	}
	if len(prefs.DislikedIngredients) != 1 || prefs.DislikedIngredients[0] != "koriander" {
		t.Errorf("DislikedIngredients = %v", prefs.DislikedIngredients)
	}
}

func TestMenuGenerate_WishesIgnoredWithoutExperience(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)
	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days: 5, Servings: 4, Wishes: "två vegetariska dagar",
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !resp.WishesIgnored {
		t.Error("expected WishesIgnored=true when no experience layer is configured")
	}
	if resp.Rationale != "" {
		t.Errorf("expected empty rationale, got %q", resp.Rationale)
	}
}

func TestMenuGenerate_ArrangeSetsRationale(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)
	env.menuService.WithExperience(newMenuExperienceServiceWith(&routingGemini{arrangeValid: true}))

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days: 5, Servings: 4, Arrange: true,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if resp.Rationale == "" {
		t.Error("expected a rationale when arrange succeeds")
	}
	// Identity permutation ⇒ every day still has a recipe (none dropped).
	for _, d := range resp.Days {
		if d.RecipeID == "" && !d.Skip {
			t.Error("a non-skip day lost its recipe after arrangement")
		}
	}
}

func TestMenuGenerate_ArrangeGuardrailFallback(t *testing.T) {
	env := newMenuTestEnv(t)
	env.seedRecipes(t, 5)
	env.menuService.WithExperience(newMenuExperienceServiceWith(&routingGemini{arrangeValid: false}))

	resp, err := env.menuService.Generate(env.householdID, domain.GenerateMenuRequest{
		Days: 5, Servings: 4, Arrange: true,
	})
	if err != nil {
		t.Fatalf("Generate should not fail when arrange is rejected: %v", err)
	}
	if resp.Rationale != "" {
		t.Errorf("expected empty rationale on guardrail rejection, got %q", resp.Rationale)
	}
	// The deterministic week is still complete.
	filled := 0
	for _, d := range resp.Days {
		if d.RecipeID != "" {
			filled++
		}
	}
	if filled != 5 {
		t.Errorf("expected 5 filled days after fallback, got %d", filled)
	}
}

func intpSvc(v int) *int    { return &v }
func boolpSvc(v bool) *bool { return &v }
