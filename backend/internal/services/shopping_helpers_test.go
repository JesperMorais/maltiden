package services

import (
	"math"
	"testing"
)

func TestCategorizeIngredient_KnownCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"köttfärs", "Kött & Fisk"},
		{"bacon", "Kött & Fisk"},
		{"lax", "Kött & Fisk"},
		{"ägg", "Mejeri & Ägg"},
		{"smör", "Mejeri & Ägg"},
		{"tomat", "Grönsaker"},
		{"potatis", "Grönsaker"},
		{"spaghetti", "Pasta, ris & spannmål"},
		{"ris", "Pasta, ris & spannmål"},
		{"salt", "Kryddor"},
		{"svartpeppar", "Kryddor"},
		{"oregano", "Kryddor"},
		{"paprikapulver", "Kryddor"},
		{"kanel", "Kryddor"},
	}
	for _, tt := range tests {
		got := categorizeIngredient(tt.input)
		if got != tt.expected {
			t.Errorf("categorizeIngredient(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// TestCategorizeIngredient_HandlesNewCategories covers the post-#168 category
// split — fruit vs veg vs fresh herbs vs bread vs pantry vs sauces vs frozen —
// to make sure each new bucket has at least one representative ingredient that
// routes to the right place.
func TestCategorizeIngredient_HandlesNewCategories(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"äpple", "Frukt"},
		{"banan", "Frukt"},
		{"avokado", "Frukt"},
		{"vitlök", "Grönsaker"},
		{"morot", "Grönsaker"},
		{"persilja", "Färska örter"},
		{"färsk basilika", "Färska örter"},
		{"knäckebröd", "Bröd"},
		{"tortilla", "Bröd"},
		{"havregryn", "Pasta, ris & spannmål"},
		{"linser", "Pasta, ris & spannmål"},
		{"krossade tomater", "Konserver"},
		{"kikärtor", "Konserver"},
		{"olivolja", "Såser & olja"},
		{"sojasås", "Såser & olja"},
		{"frysta ärtor", "Frys"},
		{"glass", "Frys"},
		{"halloumi", "Mejeri & Ägg"},
		{"yoghurt", "Mejeri & Ägg"},
		{"räkor", "Kött & Fisk"},
	}
	for _, tt := range tests {
		got := categorizeIngredient(tt.input)
		if got != tt.expected {
			t.Errorf("categorizeIngredient(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCategorizeIngredient_Unknown(t *testing.T) {
	got := categorizeIngredient("dragon fruit")
	if got != "Övrigt" {
		t.Errorf("expected 'Övrigt' for unknown ingredient, got %q", got)
	}
}

func TestCategorizeIngredient_CaseSensitive(t *testing.T) {
	// The function lowercases the name before lookup
	got := categorizeIngredient("Bacon")
	if got != "Kött & Fisk" {
		t.Errorf("expected case-insensitive match for 'Bacon', got %q", got)
	}

	got = categorizeIngredient("BACON")
	if got != "Kött & Fisk" {
		t.Errorf("expected case-insensitive match for 'BACON', got %q", got)
	}
}

func TestGenerateItemID_Deterministic(t *testing.T) {
	id1 := generateItemID("menu_1", "Tomato", "st")
	id2 := generateItemID("menu_1", "Tomato", "st")
	if id1 != id2 {
		t.Errorf("expected deterministic output, got %q and %q", id1, id2)
	}
}

func TestGenerateItemID_DifferentInputs(t *testing.T) {
	id1 := generateItemID("menu_1", "Tomato", "st")
	id2 := generateItemID("menu_1", "Potato", "st")
	id3 := generateItemID("menu_2", "Tomato", "st")
	id4 := generateItemID("menu_1", "Tomato", "kg")

	ids := []string{id1, id2, id3, id4}
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID found: %q", id)
		}
		seen[id] = true
	}
}

func TestGenerateItemID_HasPrefix(t *testing.T) {
	id := generateItemID("menu_1", "Test", "st")
	if len(id) < 5 || id[:5] != "item_" {
		t.Errorf("expected item_ prefix, got %q", id)
	}
}

func TestGenerateItemID_CaseInsensitiveName(t *testing.T) {
	// generateItemID lowercases the name
	id1 := generateItemID("menu_1", "Tomato", "st")
	id2 := generateItemID("menu_1", "tomato", "st")
	if id1 != id2 {
		t.Errorf("expected same ID for different case, got %q and %q", id1, id2)
	}
}

func TestRoundAmount(t *testing.T) {
	tests := []struct {
		amount float64
		unit   string
		want   float64
	}{
		{3.6666666666666665, "st", 4},   // countable rounds to integer
		{3.4, "st", 3},                  // countable rounds down
		{2.5, "stycken", 3},             // alternate countable spelling
		{3.6666666666666665, "g", 3.67}, // weight keeps 2 decimals
		{100.125, "ml", 100.13},         // volume rounds to 2
		{1, "tsk", 1},                   // exact integer stays exact
		{1.0 / 3.0, "kg", 0.33},         // pathological float
	}
	for _, tt := range tests {
		got := roundAmount(tt.amount, tt.unit)
		if math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("roundAmount(%v, %q) = %v, want %v", tt.amount, tt.unit, got, tt.want)
		}
	}
}
