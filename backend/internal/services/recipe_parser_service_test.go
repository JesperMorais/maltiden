package services

import (
	"encoding/json"
	"testing"
)

func TestParserResponse_NutritionNull(t *testing.T) {
	raw := `{
		"name": "Pannkakor",
		"servings": 4,
		"ingredients": [],
		"instructions": [],
		"confidence": 0.9,
		"nutrition": null
	}`

	var parsed parsedRecipeResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Nutrition != nil {
		t.Fatalf("expected nil Nutrition, got %+v", parsed.Nutrition)
	}
}

func TestParserResponse_NutritionPopulated(t *testing.T) {
	raw := `{
		"name": "Pannkakor",
		"servings": 4,
		"ingredients": [],
		"instructions": [],
		"confidence": 0.9,
		"nutrition": {
			"calories": 350,
			"proteinG": 12.5,
			"carbsG": 40.2,
			"fatG": 10.1
		}
	}`

	var parsed parsedRecipeResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if parsed.Nutrition == nil {
		t.Fatal("expected non-nil Nutrition")
	}
	if parsed.Nutrition.Calories == nil || *parsed.Nutrition.Calories != 350 {
		t.Errorf("Calories = %v, want 350", parsed.Nutrition.Calories)
	}
	if parsed.Nutrition.ProteinG == nil || *parsed.Nutrition.ProteinG != 12.5 {
		t.Errorf("ProteinG = %v, want 12.5", parsed.Nutrition.ProteinG)
	}
	if parsed.Nutrition.CarbsG == nil || *parsed.Nutrition.CarbsG != 40.2 {
		t.Errorf("CarbsG = %v, want 40.2", parsed.Nutrition.CarbsG)
	}
	if parsed.Nutrition.FatG == nil || *parsed.Nutrition.FatG != 10.1 {
		t.Errorf("FatG = %v, want 10.1", parsed.Nutrition.FatG)
	}
}
