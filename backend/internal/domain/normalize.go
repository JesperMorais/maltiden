package domain

import "strings"

var ingredientSuffixes = []string{
	"-filé",
	"-lår",
	"-bringa",
	"-strimlor",
	"-färs",
}

var pantryStaples = map[string]bool{
	"salt":     true,
	"peppar":   true,
	"olja":     true,
	"olivolja": true,
	"smör":     true,
	"mjöl":     true,
	"socker":   true,
	"vatten":   true,
	"vinäger":  true,
}

// NormalizeIngredientName lowercases, trims, strips known Swedish suffix
// qualifiers, and collapses whitespace so ingredient names can be compared
// across recipes.
func NormalizeIngredientName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, suffix := range ingredientSuffixes {
		normalized = strings.TrimSuffix(normalized, suffix)
	}

	normalized = strings.Join(strings.Fields(normalized), " ")

	return normalized
}

// IsPantryStapleName reports whether a normalized ingredient name matches
// the static pantry-staple set.
func IsPantryStapleName(normalizedName string) bool {
	return pantryStaples[normalizedName]
}
