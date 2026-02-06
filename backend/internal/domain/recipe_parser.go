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
}
