package domain

import "testing"

func intp(v int) *int { return &v }

func TestParsedWishConstraints_Clamp(t *testing.T) {
	c := &ParsedWishConstraints{
		Days:                     intp(99), // > 31
		Servings:                 intp(0),  // < 1
		VegetarianDays:           intp(-3), // < 0
		ExtraExcludedTags:        []string{"fisk", "", "  ", string(make([]byte, 80))},
		ExtraDislikedIngredients: nil,
	}
	c.Clamp()

	if *c.Days != 31 {
		t.Errorf("Days = %d, want clamped to 31", *c.Days)
	}
	if *c.Servings != 1 {
		t.Errorf("Servings = %d, want clamped to 1", *c.Servings)
	}
	if *c.VegetarianDays != 0 {
		t.Errorf("VegetarianDays = %d, want clamped to 0", *c.VegetarianDays)
	}
	// "" dropped (empty), the 80-byte string dropped (>50 chars); "  " is kept
	// (3 chars, non-empty) — only empty and oversized entries are removed.
	if len(c.ExtraExcludedTags) != 2 {
		t.Errorf("ExtraExcludedTags = %v, want [fisk, '  ']", c.ExtraExcludedTags)
	}
}

func TestParsedWishConstraints_Clamp_NilFieldsStayNil(t *testing.T) {
	c := &ParsedWishConstraints{}
	c.Clamp()
	if c.Days != nil || c.Servings != nil || c.VegetarianDays != nil || c.PrepMode != nil {
		t.Error("nil fields must stay nil after Clamp")
	}
}

func TestParsedWishConstraints_Clamp_NilReceiver(t *testing.T) {
	var c *ParsedWishConstraints
	c.Clamp() // must not panic
}
