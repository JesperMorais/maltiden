package domain

import "strings"

var cutQualifierReplacer = strings.NewReplacer(
	"filé", "",
	"file", "",
	"lår", "",
	"lar", "",
	"bröst", "",
	"brost", "",
	"hackad", "",
	"hackat", "",
	"hackade", "",
	"strimlad", "",
	"strimlat", "",
	"strimlade", "",
	"riven", "",
	"rivet", "",
	"rivna", "",
	"skivad", "",
	"skivat", "",
	"skivade", "",
	"färsk", "",
	"farsk", "",
	"färska", "",
	"farska", "",
	"torkad", "",
	"torkat", "",
	"torkade", "",
)

var canonicalPantryStaples = map[string]bool{
	"salt":    true,
	"peppar":  true,
	"olja":    true,
	"smör":    true,
	"smor":    true,
	"mjöl":    true,
	"mjol":    true,
	"socker":  true,
	"vatten":  true,
	"vinäger": true,
	"vinager": true,
}

// Canonicalize maps an ingredient name to a stable stem for overlap
// scoring: lowercase, strip cut/prep qualifiers, trim whitespace.
func Canonicalize(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, ",", " ")
	s = cutQualifierReplacer.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimSpace(s)
	return s
}

// IsPantryStaple reports whether a canonicalized ingredient name is a
// common Swedish pantry staple that should be excluded from shopping lists.
func IsPantryStaple(canonical string) bool {
	return canonicalPantryStaples[strings.TrimSpace(canonical)]
}
