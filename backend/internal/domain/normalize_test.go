package domain

import "testing"

func TestNormalizeIngredientName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase and trim", "  Lök  ", "lök"},
		{"strips -filé suffix", "Kyckling-filé", "kyckling"},
		{"strips -lår suffix", "Kyckling-lår", "kyckling"},
		{"strips -bringa suffix", "Kyckling-bringa", "kyckling"},
		{"strips -strimlor suffix", "Fläsk-strimlor", "fläsk"},
		{"strips -färs suffix", "Nöt-färs", "nöt"},
		{"collapses whitespace", "gul   lök", "gul lök"},
		{"handles swedish runes", "ÅTELÖK", "åtelök"},
		{"no change for plain name", "morot", "morot"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeIngredientName(tc.in)
			if got != tc.want {
				t.Errorf("NormalizeIngredientName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsPantryStapleName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"salt is staple", "salt", true},
		{"olivolja is staple", "olivolja", true},
		{"vinäger is staple", "vinäger", true},
		{"lök is not staple", "lök", false},
		{"empty is not staple", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPantryStapleName(tc.in)
			if got != tc.want {
				t.Errorf("IsPantryStapleName(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
