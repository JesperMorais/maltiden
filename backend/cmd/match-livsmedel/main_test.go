package main

import (
	"testing"

	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
)

func TestNeedsMatch(t *testing.T) {
	cases := []struct {
		name string
		ings []domain.Ingredient
		want bool
	}{
		{"no canonical", []domain.Ingredient{{Name: "Lök"}}, false},
		{"canonical unmatched", []domain.Ingredient{{Name: "Lök", CanonicalName: "lök"}}, true},
		{"canonical already matched", []domain.Ingredient{{Name: "Lök", CanonicalName: "lök", Livsmedelsnummer: 5}}, false},
		{"mixed: one unmatched", []domain.Ingredient{
			{Name: "Lök", CanonicalName: "lök", Livsmedelsnummer: 5},
			{Name: "Ris", CanonicalName: "ris"},
		}, true},
		{"all matched", []domain.Ingredient{
			{CanonicalName: "lök", Livsmedelsnummer: 5},
			{CanonicalName: "ris", Livsmedelsnummer: 6},
		}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := &domain.Recipe{Ingredients: c.ings}
			if got := needsMatch(r); got != c.want {
				t.Errorf("needsMatch = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDistinctCanonicals(t *testing.T) {
	r := &domain.Recipe{Ingredients: []domain.Ingredient{
		{CanonicalName: "Lök"},                          // normalized to "lök"
		{CanonicalName: "lök"},                          // dup, dropped
		{CanonicalName: "ris", Livsmedelsnummer: 7},     // already matched, dropped
		{CanonicalName: " ", Livsmedelsnummer: 0},       // empty, dropped
		{CanonicalName: "vitlök"},
	}}
	got := distinctCanonicals(r)
	if len(got) != 2 || got[0] != "lök" || got[1] != "vitlök" {
		t.Errorf("distinctCanonicals = %v", got)
	}
}

func setupMatchTestDB(t *testing.T) (*sqlite.RecipeStorage, *sqlite.LivsmedelStorage) {
	t.Helper()
	db, err := sqlite.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(`INSERT INTO households (id, name) VALUES (?, ?)`, "hh_match", "Match HH"); err != nil {
		t.Fatalf("seed household: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO livsmedel (livsmedelsnummer, namn, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		900100, "Kyckling, rå", 110, 23, 0, 1.5); err != nil {
		t.Fatalf("seed livsmedel: %v", err)
	}
	return sqlite.NewRecipeStorage(db), sqlite.NewLivsmedelStorage(db)
}

func TestApplyMatch_AppliesValidatesAndComputes(t *testing.T) {
	recipeStore, ls := setupMatchTestDB(t)

	r := &domain.Recipe{
		ID:          "rec_match",
		Name:        "Kyckling",
		Servings:    2,
		HouseholdID: "hh_match",
		DietClass:   domain.DietClassOmnivore,
		Ingredients: []domain.Ingredient{
			{Name: "Kycklingfilé", CanonicalName: "kyckling", GramsEquiv: 200},
		},
		Instructions: []string{"Stek"},
		Tags:         []string{},
	}
	if err := recipeStore.Create(r); err != nil {
		t.Fatalf("Create: %v", err)
	}

	offered := map[string]map[int]bool{"kyckling": {900100: true}}
	res := &matchResult{}
	res.Matches = append(res.Matches, struct {
		CanonicalName    string `json:"canonicalName"`
		Livsmedelsnummer int    `json:"livsmedelsnummer"`
	}{CanonicalName: "kyckling", Livsmedelsnummer: 900100})

	if !applyMatch(r, res, offered, ls, recipeStore) {
		t.Fatal("expected applyMatch to update the recipe")
	}
	if r.Ingredients[0].Livsmedelsnummer != 900100 {
		t.Errorf("expected livsmedelsnummer set, got %d", r.Ingredients[0].Livsmedelsnummer)
	}
	// 200g chicken at 23g protein/100g over 2 servings = 23g protein/serving.
	if r.Macros == nil || r.Macros.Protein != 23 {
		t.Errorf("expected per-serving protein 23, got %+v", r.Macros)
	}

	// Persisted.
	got, err := recipeStore.GetByID(r.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Macros == nil || got.Macros.Protein != 23 {
		t.Errorf("macros not persisted: %+v", got.Macros)
	}
	if got.Ingredients[0].Livsmedelsnummer != 900100 {
		t.Errorf("livsmedelsnummer not persisted: %d", got.Ingredients[0].Livsmedelsnummer)
	}
}

func TestApplyMatch_DropsNumbersNotOffered(t *testing.T) {
	recipeStore, ls := setupMatchTestDB(t)
	r := &domain.Recipe{
		Servings:    1,
		Ingredients: []domain.Ingredient{{CanonicalName: "kyckling", GramsEquiv: 100}},
	}
	// Model returns a number that was never in the offered shortlist.
	res := &matchResult{}
	res.Matches = append(res.Matches, struct {
		CanonicalName    string `json:"canonicalName"`
		Livsmedelsnummer int    `json:"livsmedelsnummer"`
	}{CanonicalName: "kyckling", Livsmedelsnummer: 555555})

	offered := map[string]map[int]bool{"kyckling": {900100: true}} // 555555 not offered
	if applyMatch(r, res, offered, ls, recipeStore) {
		t.Error("expected applyMatch to reject a non-offered number (no update)")
	}
	if r.Ingredients[0].Livsmedelsnummer != 0 {
		t.Errorf("expected ingredient left unmatched, got %d", r.Ingredients[0].Livsmedelsnummer)
	}
}

func TestApplyMatch_DropsNumberNotInTable(t *testing.T) {
	recipeStore, ls := setupMatchTestDB(t)
	r := &domain.Recipe{
		Servings:    1,
		Ingredients: []domain.Ingredient{{CanonicalName: "kyckling", GramsEquiv: 100}},
	}
	// Offered (passed the shortlist guard) but does not exist in the table.
	res := &matchResult{}
	res.Matches = append(res.Matches, struct {
		CanonicalName    string `json:"canonicalName"`
		Livsmedelsnummer int    `json:"livsmedelsnummer"`
	}{CanonicalName: "kyckling", Livsmedelsnummer: 777777})

	offered := map[string]map[int]bool{"kyckling": {777777: true}}
	if applyMatch(r, res, offered, ls, recipeStore) {
		t.Error("expected applyMatch to reject a number missing from the table")
	}
	if r.Ingredients[0].Livsmedelsnummer != 0 {
		t.Errorf("expected ingredient left unmatched, got %d", r.Ingredients[0].Livsmedelsnummer)
	}
}

// TestApplyMatch_DropsNumberOfferedForDifferentName guards the per-canonical
// shortlist: a number that was offered for ingredient A must not be accepted
// when the model assigns it to ingredient B (otherwise B gets A's nutrition).
func TestApplyMatch_DropsNumberOfferedForDifferentName(t *testing.T) {
	recipeStore, ls := setupMatchTestDB(t)
	r := &domain.Recipe{
		Servings: 1,
		Ingredients: []domain.Ingredient{
			{CanonicalName: "lök", GramsEquiv: 100},
			{CanonicalName: "lax", GramsEquiv: 100},
		},
	}
	// 900100 (chicken) was only offered for "lök"; the model wrongly assigns it
	// to "lax". It must be dropped, not borrowed onto the salmon ingredient.
	res := &matchResult{}
	res.Matches = append(res.Matches, struct {
		CanonicalName    string `json:"canonicalName"`
		Livsmedelsnummer int    `json:"livsmedelsnummer"`
	}{CanonicalName: "lax", Livsmedelsnummer: 900100})

	offered := map[string]map[int]bool{
		"lök": {900100: true},
		"lax": {}, // no candidates offered for lax
	}
	if applyMatch(r, res, offered, ls, recipeStore) {
		t.Error("expected applyMatch to reject a number offered for a different name")
	}
	if r.Ingredients[1].Livsmedelsnummer != 0 {
		t.Errorf("expected lax left unmatched, got %d", r.Ingredients[1].Livsmedelsnummer)
	}
}
