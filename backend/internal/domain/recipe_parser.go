package domain

// ParseRecipeRequest represents a request to parse unstructured recipe text.
type ParseRecipeRequest struct {
	RawText string `json:"rawText"`
	Source  string `json:"source,omitempty"`
}

// ParseRecipeResponse contains the parsed recipe data with confidence metadata.
type ParseRecipeResponse struct {
	Recipe     CreateRecipeRequest `json:"recipe"`
	Confidence float64             `json:"confidence"`
	Warnings   []string            `json:"warnings,omitempty"`
	RawText    string              `json:"rawText"`
	Nutrition  *Nutrition          `json:"nutrition,omitempty"`
}

// Nutrition holds estimated per-serving macros. Fields are pointers so an
// unknown value can be represented as null rather than a misleading zero.
type Nutrition struct {
	Calories *int     `json:"calories,omitempty"`
	ProteinG *float64 `json:"proteinG,omitempty"`
	CarbsG   *float64 `json:"carbsG,omitempty"`
	FatG     *float64 `json:"fatG,omitempty"`
}
