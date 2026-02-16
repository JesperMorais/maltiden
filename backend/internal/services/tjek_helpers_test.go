package services

import "testing"

func TestMatchesWord_ExactMatch(t *testing.T) {
	if !matchesWord("Kyckling", "kyckling") {
		t.Error("expected exact match")
	}
}

func TestMatchesWord_PrefixMatch(t *testing.T) {
	if !matchesWord("Kycklingfilé", "kyckling") {
		t.Error("expected prefix match")
	}
}

func TestMatchesWord_SuffixMatch(t *testing.T) {
	if !matchesWord("Fläskfilé", "filé") {
		t.Error("expected suffix match")
	}
}

func TestMatchesWord_CaseInsensitive(t *testing.T) {
	if !matchesWord("KYCKLING", "kyckling") {
		t.Error("expected case-insensitive match")
	}
	if !matchesWord("kyckling", "KYCKLING") {
		t.Error("expected case-insensitive match (reverse)")
	}
}

func TestMatchesWord_NoMatch(t *testing.T) {
	if matchesWord("Bacon", "kyckling") {
		t.Error("expected no match")
	}
}

func TestMatchesWord_SpecialChars(t *testing.T) {
	// Words are split on non-letter/non-number boundaries
	if !matchesWord("100% nötkött", "nötkött") {
		t.Error("expected match after special char split")
	}
}

func TestMatchesWord_MultipleWords(t *testing.T) {
	if !matchesWord("Grillad kycklingfilé", "kyckling") {
		t.Error("expected match in multi-word text")
	}
}

func TestMatchesWord_EmptyQuery(t *testing.T) {
	// Empty query matches any word as prefix/suffix
	if !matchesWord("Something", "") {
		t.Error("empty query should match (empty string is prefix of everything)")
	}
}

// isComboOffer tests

func TestIsComboOffer_BelowThreshold(t *testing.T) {
	if isComboOffer("Kyckling / Ris") {
		t.Error("1 slash should not be combo")
	}
	if isComboOffer("Mjölk, Smör, Ost") {
		t.Error("2 commas should not be combo")
	}
}

func TestIsComboOffer_AtSlashThreshold(t *testing.T) {
	// >2 slashes means combo
	if isComboOffer("A / B / C") {
		t.Error("exactly 2 slashes should not be combo")
	}
	if !isComboOffer("A / B / C / D") {
		t.Error("3 slashes should be combo")
	}
}

func TestIsComboOffer_AtCommaThreshold(t *testing.T) {
	// >3 commas means combo
	if isComboOffer("A, B, C, D") {
		t.Error("3 commas should not be combo")
	}
	if !isComboOffer("A, B, C, D, E") {
		t.Error("4 commas should be combo")
	}
}

// isGroceryStore tests

func TestIsGroceryStore_KnownGrocery(t *testing.T) {
	stores := []string{"Willys", "ICA Maxi", "Coop Extra", "Lidl", "Hemköp", "City Gross"}
	for _, s := range stores {
		if !isGroceryStore(s) {
			t.Errorf("expected %q to be a grocery store", s)
		}
	}
}

func TestIsGroceryStore_KnownExclude(t *testing.T) {
	stores := []string{"Bauhaus", "Biltema", "JYSK", "Elgiganten", "Apoteket", "XXL"}
	for _, s := range stores {
		if isGroceryStore(s) {
			t.Errorf("expected %q to NOT be a grocery store", s)
		}
	}
}

func TestIsGroceryStore_UnknownStore(t *testing.T) {
	if isGroceryStore("Random Shop") {
		t.Error("unknown store should return false")
	}
}

// buildDescription tests

func TestBuildDescription_EmptyUnit(t *testing.T) {
	if got := buildDescription("", 100, 100, 1, 1); got != "" {
		t.Errorf("expected empty for empty unit, got %q", got)
	}
}

func TestBuildDescription_ZeroSize(t *testing.T) {
	if got := buildDescription("ml", 0, 0, 1, 1); got != "" {
		t.Errorf("expected empty for zero size, got %q", got)
	}
}

func TestBuildDescription_SingleDrinkMl(t *testing.T) {
	got := buildDescription("ml", 330, 330, 1, 1)
	if got != "1 burk (330 ml)" {
		t.Errorf("expected '1 burk (330 ml)', got %q", got)
	}
}

func TestBuildDescription_MultiPackDrink(t *testing.T) {
	got := buildDescription("cl", 33, 33, 6, 6)
	if got != "6-pack burk (33 cl)" {
		t.Errorf("expected '6-pack burk (33 cl)', got %q", got)
	}
}

func TestBuildDescription_LargeBottle(t *testing.T) {
	got := buildDescription("l", 1.5, 1.5, 1, 1)
	if got != "1 flaska (1.5 l)" {
		t.Errorf("expected '1 flaska (1.5 l)', got %q", got)
	}
}

func TestBuildDescription_FoodByWeight(t *testing.T) {
	got := buildDescription("g", 500, 500, 1, 1)
	if got != "500 g" {
		t.Errorf("expected '500 g', got %q", got)
	}
}

func TestBuildDescription_FoodByKg(t *testing.T) {
	got := buildDescription("kg", 2, 2, 1, 1)
	if got != "2 kg" {
		t.Errorf("expected '2 kg', got %q", got)
	}
}

func TestBuildDescription_MultipleItems(t *testing.T) {
	got := buildDescription("g", 200, 200, 3, 3)
	if got != "3 st (200 g)" {
		t.Errorf("expected '3 st (200 g)', got %q", got)
	}
}

func TestBuildDescription_Pieces(t *testing.T) {
	got := buildDescription("st", 1, 1, 6, 6)
	if got != "6-pack" {
		t.Errorf("expected '6-pack', got %q", got)
	}
}

func TestBuildDescription_PiecesMultipleSize(t *testing.T) {
	got := buildDescription("st", 4, 4, 1, 1)
	if got != "4 st" {
		t.Errorf("expected '4 st', got %q", got)
	}
}

func TestBuildDescription_SizeRange(t *testing.T) {
	got := buildDescription("g", 200, 300, 1, 1)
	if got != "200-300 g" {
		t.Errorf("expected '200-300 g', got %q", got)
	}
}

func TestBuildDescription_MultiStBottle(t *testing.T) {
	// 750ml drink, 3 pieces - should be "3 st (750 ml)" since >500ml = flaska
	got := buildDescription("ml", 750, 750, 3, 3)
	if got != "3 st (750 ml)" {
		t.Errorf("expected '3 st (750 ml)', got %q", got)
	}
}

// determineDrinkPackaging tests

func TestDetermineDrinkPackaging_SmallCan(t *testing.T) {
	if got := determineDrinkPackaging(250); got != "burk" {
		t.Errorf("250ml: expected 'burk', got %q", got)
	}
}

func TestDetermineDrinkPackaging_StandardCan(t *testing.T) {
	if got := determineDrinkPackaging(330); got != "burk" {
		t.Errorf("330ml: expected 'burk', got %q", got)
	}
}

func TestDetermineDrinkPackaging_LargeCan(t *testing.T) {
	if got := determineDrinkPackaging(500); got != "burk" {
		t.Errorf("500ml: expected 'burk', got %q", got)
	}
}

func TestDetermineDrinkPackaging_MediumBottle(t *testing.T) {
	if got := determineDrinkPackaging(750); got != "flaska" {
		t.Errorf("750ml: expected 'flaska', got %q", got)
	}
}

func TestDetermineDrinkPackaging_LargeBottle(t *testing.T) {
	if got := determineDrinkPackaging(1500); got != "flaska" {
		t.Errorf("1500ml: expected 'flaska', got %q", got)
	}
}

func TestDetermineDrinkPackaging_VeryLargeBottle(t *testing.T) {
	if got := determineDrinkPackaging(2000); got != "flaska" {
		t.Errorf("2000ml: expected 'flaska', got %q", got)
	}
}

// formatSize tests

func TestFormatSize_Integer(t *testing.T) {
	if got := formatSize(500, "g"); got != "500 g" {
		t.Errorf("expected '500 g', got %q", got)
	}
}

func TestFormatSize_Decimal(t *testing.T) {
	if got := formatSize(1.5, "l"); got != "1.5 l" {
		t.Errorf("expected '1.5 l', got %q", got)
	}
}

func TestFormatSize_WholeFloat(t *testing.T) {
	// 2.0 should format as integer
	if got := formatSize(2.0, "kg"); got != "2 kg" {
		t.Errorf("expected '2 kg', got %q", got)
	}
}
