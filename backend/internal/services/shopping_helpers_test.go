package services

import "testing"

func TestCategorizeIngredient_KnownCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"köttfärs", "Kött & Fisk"},
		{"bacon", "Kött & Fisk"},
		{"lax", "Kött & Fisk"},
		{"ägg", "Mejeri"},
		{"smör", "Mejeri"},
		{"tomat", "Frukt & Grönt"},
		{"potatis", "Frukt & Grönt"},
		{"spaghetti", "Skafferi"},
		{"ris", "Skafferi"},
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
