package domain

import "testing"

func TestNormalizeCanonicalizeKycklingFamilyCollapses(t *testing.T) {
	want := "kyckling"
	inputs := []string{
		"Kycklingfilé",
		"kycklinglår",
		"Kycklingbröst",
		"  Kyckling  ",
		"kyckling, strimlad",
	}
	for _, in := range inputs {
		got := Canonicalize(in)
		if got != want {
			t.Errorf("Canonicalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeIsPantryStapleFlagsSalt(t *testing.T) {
	if !IsPantryStaple(Canonicalize("Salt")) {
		t.Errorf("expected salt to be flagged as pantry staple")
	}
	if IsPantryStaple(Canonicalize("Kycklingfilé")) {
		t.Errorf("expected kyckling to not be flagged as pantry staple")
	}
}
