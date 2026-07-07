package services

import (
	"context"
	"encoding/json"
	"fmt"
	"maltiden/pkg/claude"
	"testing"
)

type mockClaudeSender struct {
	resp *claude.Response
	err  error
}

func (m *mockClaudeSender) SendMessage(_ context.Context, _ claude.Request) (*claude.Response, error) {
	return m.resp, m.err
}

func validRecipeJSON(t *testing.T) string {
	t.Helper()
	data := map[string]interface{}{
		"name":         "Pastasoppa",
		"servings":     4,
		"emoji":        "🍝",
		"tags":         []string{"vardag", "pasta"},
		"ingredients":  []map[string]interface{}{{"name": "pasta", "amount": 300, "unit": "g"}},
		"instructions": []string{"Koka pastan i saltat vatten tills den ar al dente"},
		"confidence":   0.9,
		"warnings":     []string{},
	}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRecipeParser_ParseRecipe_HappyPath(t *testing.T) {
	stub := &mockClaudeSender{
		resp: &claude.Response{
			Content: []claude.ContentBlock{
				{Type: "text", Text: validRecipeJSON(t)},
			},
		},
	}
	svc := &RecipeParserService{claudeClient: stub}

	result, err := svc.ParseRecipe("Pastasoppa med 300g pasta. Koka i saltat vatten.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Recipe.Name != "Pastasoppa" {
		t.Errorf("expected name 'Pastasoppa', got %q", result.Recipe.Name)
	}
	if len(result.Recipe.Ingredients) == 0 {
		t.Error("expected non-empty ingredients")
	}
	if result.Recipe.Ingredients[0].Name != "pasta" {
		t.Errorf("expected ingredient name 'pasta', got %q", result.Recipe.Ingredients[0].Name)
	}
}

func TestRecipeParser_ParseRecipe_EmptyInput(t *testing.T) {
	svc := &RecipeParserService{claudeClient: &mockClaudeSender{}}

	_, err := svc.ParseRecipe("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
	if err.Error() != "raw text is required" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestRecipeParser_ParseRecipe_APIError(t *testing.T) {
	stub := &mockClaudeSender{
		err: fmt.Errorf("network failure"),
	}
	svc := &RecipeParserService{claudeClient: stub}

	_, err := svc.ParseRecipe("Kycklingsoppa med gronsaker")
	if err == nil {
		t.Fatal("expected error on API failure")
	}
}

func TestRecipeParser_ParseRecipe_MalformedJSON(t *testing.T) {
	stub := &mockClaudeSender{
		resp: &claude.Response{
			Content: []claude.ContentBlock{
				{Type: "text", Text: "not valid json at all"},
			},
		},
	}
	svc := &RecipeParserService{claudeClient: stub}

	_, err := svc.ParseRecipe("Tomatsoppa med gronsaker och kryddor")
	if err == nil {
		t.Fatal("expected error for malformed JSON response")
	}
}
