package domain

import (
	"encoding/json"
	"testing"
)

func ptr(f float64) *float64 { return &f }

func TestNutrition_JSONOmitEmpty(t *testing.T) {
	n := Nutrition{}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "{}" {
		t.Errorf("empty Nutrition: want {}, got %s", b)
	}
}

func TestNutrition_JSONRoundTrip(t *testing.T) {
	n := Nutrition{
		Calories: ptr(500),
		ProteinG: ptr(30),
		CarbsG:   ptr(60),
		FatG:     ptr(15),
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var got Nutrition
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Calories != 500 || *got.ProteinG != 30 || *got.CarbsG != 60 || *got.FatG != 15 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

func TestNutrition_PartialFields(t *testing.T) {
	n := Nutrition{Calories: ptr(350)}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatal(err)
	}
	var got Nutrition
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Calories != 350 {
		t.Errorf("calories: want 350, got %v", *got.Calories)
	}
	if got.ProteinG != nil || got.CarbsG != nil || got.FatG != nil {
		t.Error("unset fields should be nil after partial round-trip")
	}
}

func TestRecipe_NilNutritionOmitted(t *testing.T) {
	r := Recipe{
		ID:       "rec_1",
		Name:     "Test",
		Servings: 2,
		Tags:     []string{},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["nutrition"]; ok {
		t.Error("nutrition key should be absent when nil")
	}
}

func TestRecipe_NutritionPresent(t *testing.T) {
	r := Recipe{
		ID:        "rec_2",
		Name:      "Macro Bowl",
		Servings:  1,
		Tags:      []string{},
		Nutrition: &Nutrition{Calories: ptr(700), ProteinG: ptr(45)},
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["nutrition"]; !ok {
		t.Error("nutrition key should be present when set")
	}
}
