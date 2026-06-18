package domain

import "testing"

func TestComputeRecipeMacros_KnownFixture(t *testing.T) {
	table := map[int]Livsmedel{
		1: {Livsmedelsnummer: 1, Namn: "kyckling", KcalPer100g: 100, ProteinPer100g: 20, CarbsPer100g: 0, FatPer100g: 2},
		2: {Livsmedelsnummer: 2, Namn: "ris", KcalPer100g: 350, ProteinPer100g: 7, CarbsPer100g: 78, FatPer100g: 1},
	}
	r := Recipe{
		Servings: 2,
		Ingredients: []Ingredient{
			{Name: "Kyckling", Livsmedelsnummer: 1, GramsEquiv: 200}, // 200g chicken
			{Name: "Ris", Livsmedelsnummer: 2, GramsEquiv: 100},      // 100g rice
		},
	}
	// Totals: kcal 200+350=550, protein 40+7=47, carbs 0+78=78, fat 4+1=5.
	// Per serving (÷2): kcal 275, protein 23.5, carbs 39, fat 2.5.
	m := ComputeRecipeMacros(r, table)
	if m == nil {
		t.Fatal("expected non-nil macros")
	}
	if m.Kcal != 275 || m.Protein != 23.5 || m.Carbs != 39 || m.Fat != 2.5 {
		t.Errorf("unexpected macros: %+v", m)
	}
}

func TestComputeRecipeMacros_IgnoresUnmatchedAndZeroGrams(t *testing.T) {
	table := map[int]Livsmedel{
		1: {Livsmedelsnummer: 1, ProteinPer100g: 20},
		3: {Livsmedelsnummer: 3, ProteinPer100g: 99},
	}
	r := Recipe{
		Servings: 1,
		Ingredients: []Ingredient{
			{Name: "matched", Livsmedelsnummer: 1, GramsEquiv: 100}, // contributes 20g protein
			{Name: "no-number", Livsmedelsnummer: 0, GramsEquiv: 100},
			{Name: "zero-grams", Livsmedelsnummer: 3, GramsEquiv: 0},   // ignored: no grams
			{Name: "not-in-table", Livsmedelsnummer: 99, GramsEquiv: 100}, // ignored: missing
		},
	}
	m := ComputeRecipeMacros(r, table)
	if m == nil {
		t.Fatal("expected non-nil macros")
	}
	if m.Protein != 20 {
		t.Errorf("expected only the matched ingredient to contribute, got protein=%v", m.Protein)
	}
}

func TestComputeRecipeMacros_ServingsGuard(t *testing.T) {
	table := map[int]Livsmedel{1: {Livsmedelsnummer: 1, KcalPer100g: 100}}
	r := Recipe{
		Servings:    0, // guarded to 1
		Ingredients: []Ingredient{{Livsmedelsnummer: 1, GramsEquiv: 100}},
	}
	m := ComputeRecipeMacros(r, table)
	if m == nil || m.Kcal != 100 {
		t.Errorf("expected Servings<1 treated as 1 (kcal=100), got %+v", m)
	}
}

func TestComputeRecipeMacros_NilWhenNoMatch(t *testing.T) {
	table := map[int]Livsmedel{1: {Livsmedelsnummer: 1, KcalPer100g: 100}}
	cases := []struct {
		name string
		r    Recipe
	}{
		{"no ingredients", Recipe{Servings: 2}},
		{"all unmatched", Recipe{Servings: 2, Ingredients: []Ingredient{{Livsmedelsnummer: 0, GramsEquiv: 100}}}},
		{"empty table", Recipe{Servings: 2, Ingredients: []Ingredient{{Livsmedelsnummer: 1, GramsEquiv: 100}}}},
	}
	for _, c := range cases {
		tbl := table
		if c.name == "empty table" {
			tbl = map[int]Livsmedel{}
		}
		if m := ComputeRecipeMacros(c.r, tbl); m != nil {
			t.Errorf("%s: expected nil macros, got %+v", c.name, m)
		}
	}
}

func TestIsValidProteinTarget(t *testing.T) {
	cases := []struct {
		v    float64
		want bool
	}{
		{0, true},
		{120, true},
		{MaxProteinTargetPerDay, true},
		{-1, false},
		{MaxProteinTargetPerDay + 1, false},
	}
	for _, c := range cases {
		if got := IsValidProteinTarget(c.v); got != c.want {
			t.Errorf("IsValidProteinTarget(%v) = %v, want %v", c.v, got, c.want)
		}
	}
}
