package services

import (
	"context"
	"encoding/json"
	"fmt"
	"maltiden/internal/domain"
	"maltiden/pkg/claude"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

// claudeSender is satisfied by *claude.Client in production and by stubs in tests.
type claudeSender interface {
	SendMessage(ctx context.Context, req claude.Request) (*claude.Response, error)
}

var recipeSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"name":     map[string]interface{}{"type": "string"},
		"servings": map[string]interface{}{"type": "integer"},
		"emoji":    map[string]interface{}{"type": "string"},
		"tags": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
		"ingredients": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":           map[string]interface{}{"type": "string"},
					"amount":         map[string]interface{}{"type": "number"},
					"unit":           map[string]interface{}{"type": "string"},
					"canonicalName":  map[string]interface{}{"type": "string"},
					"gramsEquiv":     map[string]interface{}{"type": "number"},
					"isPantryStaple": map[string]interface{}{"type": "boolean"},
					"isPerishable":   map[string]interface{}{"type": "boolean"},
				},
				"required":             []string{"name", "amount", "unit"},
				"additionalProperties": false,
			},
		},
		"instructions": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
		"mainProtein": map[string]interface{}{"type": "string"},
		"dietClass": map[string]interface{}{
			"type": "string",
			"enum": []string{"omnivore", "vegetarian", "vegan", "pescetarian"},
		},
		"batchable":   map[string]interface{}{"type": "boolean"},
		"cookMinutes": map[string]interface{}{"type": "integer"},
		"confidence":  map[string]interface{}{"type": "number"},
		"warnings": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
	},
	"required":             []string{"name", "servings", "ingredients", "instructions", "confidence"},
	"additionalProperties": false,
}

const systemPrompt = `You are a recipe parsing assistant for Måltiden, a Swedish meal planning app.

Your task is to extract structured recipe data from unstructured text input.

IMPORTANT RULES:
1. All output must be valid JSON matching the provided schema
2. Recipe names should be in Swedish if the input is Swedish
3. For "servings", extract the number of portions (default to 4 if not specified)
4. For "emoji", suggest ONE relevant food emoji that represents the dish
5. For "tags", suggest 2-5 relevant Swedish tags from: vardag, helg, snabb, vegetarisk, vegan, fisk, kyckling, kött, pasta, soppa, sallad, barn, fest, billigt, hälsosam
6. For "ingredients":
   - Parse amount as a number (use 0 for "efter smak" / "to taste")
   - Use standard Swedish units: g, kg, dl, l, msk, tsk, st, krm
   - Keep ingredient names in Swedish
7. For "instructions":
   - Keep as an array of strings, one step per item
   - Keep in Swedish
   - Clean up numbering (remove "1.", "2." etc.)
8. For each ingredient, ALSO emit metadata fields:
   - "canonicalName": the normalized Swedish lemma of the ingredient, stripped of
     preparation and form words (e.g. "riven ost" → "ost", "kycklingfilé" → "kyckling",
     "hackad gul lök" → "lök"). Keep it in Swedish, lowercase, singular.
   - "gramsEquiv": best-effort weight in grams for the listed amount (number). Use 0 if unknown.
   - "isPantryStaple": true for shelf-stable staples (salt, peppar, olja, mjöl, socker, etc.)
   - "isPerishable": true for fresh items that spoil quickly (kött, fisk, mejeri, färska grönsaker)
9. Emit recipe-level metadata:
   - "mainProtein": the dominant protein source in Swedish (e.g. "kyckling", "nötkött",
     "lax", "linser"); empty string if none.
   - "dietClass": one of "omnivore", "vegetarian", "vegan", "pescetarian".
   - "batchable": true if the dish reheats/freezes well for batch cooking.
   - "cookMinutes": total active cooking time in minutes (integer). Use 0 if unknown.
10. Set "confidence" (0.0-1.0) based on how well-structured the input was
11. Add "warnings" array for any issues (missing info, ambiguous amounts, etc.)

NEVER include anything except the JSON object in your response.`

type RecipeParserService struct {
	claudeClient claudeSender
}

func NewRecipeParserService(claudeClient *claude.Client) *RecipeParserService {
	return &RecipeParserService{
		claudeClient: claudeClient,
	}
}

// newRecipeParserServiceWithSender builds a parser service around any
// claudeSender. Used by tests to inject a stubbed Claude response without a
// live client.
func newRecipeParserServiceWithSender(sender claudeSender) *RecipeParserService {
	return &RecipeParserService{claudeClient: sender}
}

type parsedRecipeResponse struct {
	Name         string              `json:"name"`
	Servings     int                 `json:"servings"`
	Emoji        string              `json:"emoji"`
	Tags         []string            `json:"tags"`
	Ingredients  []domain.Ingredient `json:"ingredients"`
	Instructions []string            `json:"instructions"`
	MainProtein  string              `json:"mainProtein"`
	DietClass    string              `json:"dietClass"`
	Batchable    bool                `json:"batchable"`
	CookMinutes  int                 `json:"cookMinutes"`
	Confidence   float64             `json:"confidence"`
	Warnings     []string            `json:"warnings"`
}

func (s *RecipeParserService) parseOnce(ctx context.Context, rawText, extraGuidance string) (*domain.ParseRecipeResponse, error) {
	userContent := fmt.Sprintf("Parse this recipe:\n\n%s", rawText)
	if extraGuidance != "" {
		userContent += "\n\n" + extraGuidance
	}

	req := claude.Request{
		Model:     claude.HaikuModel,
		MaxTokens: 2000,
		System:    systemPrompt,
		Messages: []claude.Message{
			{
				Role:    "user",
				Content: userContent,
			},
		},
	}

	resp, err := s.claudeClient.SendMessage(ctx, req)
	if err != nil {
		sentry.CaptureException(err)
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	if len(resp.Content) == 0 || resp.Content[0].Type != "text" {
		return nil, fmt.Errorf("unexpected response format from Claude")
	}

	// Strip markdown code fences if Claude wraps the JSON
	text := strings.TrimSpace(resp.Content[0].Text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	var parsed parsedRecipeResponse
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		sentry.CaptureException(err)
		return nil, fmt.Errorf("failed to parse Claude response: %w", err)
	}

	// Drop a dietClass the model hallucinated outside the enum; empty = unknown.
	dietClass := parsed.DietClass
	if !domain.IsValidDietClass(dietClass) {
		dietClass = ""
	}

	return &domain.ParseRecipeResponse{
		Recipe: domain.CreateRecipeRequest{
			Name:         parsed.Name,
			Servings:     parsed.Servings,
			Emoji:        parsed.Emoji,
			Tags:         parsed.Tags,
			Ingredients:  parsed.Ingredients,
			Instructions: parsed.Instructions,
			MainProtein:  parsed.MainProtein,
			DietClass:    dietClass,
			Batchable:    parsed.Batchable,
			CookMinutes:  parsed.CookMinutes,
		},
		Confidence: parsed.Confidence,
		Warnings:   parsed.Warnings,
		RawText:    rawText,
	}, nil
}

func validateParsedContent(r *domain.ParseRecipeResponse) error {
	if err := domain.ValidateContent(r.Recipe.Name); err != nil {
		return err
	}
	for _, tag := range r.Recipe.Tags {
		if err := domain.ValidateContent(tag); err != nil {
			return err
		}
	}
	for _, ing := range r.Recipe.Ingredients {
		if err := domain.ValidateContent(ing.Name); err != nil {
			return err
		}
		if err := domain.ValidateContent(ing.Unit); err != nil {
			return err
		}
		if err := domain.ValidateContent(ing.CanonicalName); err != nil {
			return err
		}
	}
	for _, step := range r.Recipe.Instructions {
		if err := domain.ValidateContent(step); err != nil {
			return err
		}
	}
	if err := domain.ValidateContent(r.Recipe.MainProtein); err != nil {
		return err
	}
	return nil
}

func stripParsedEmoji(r *domain.ParseRecipeResponse) {
	r.Recipe.Name = domain.StripEmoji(r.Recipe.Name)
	for i, tag := range r.Recipe.Tags {
		r.Recipe.Tags[i] = domain.StripEmoji(tag)
	}
	for i := range r.Recipe.Ingredients {
		r.Recipe.Ingredients[i].Name = domain.StripEmoji(r.Recipe.Ingredients[i].Name)
		r.Recipe.Ingredients[i].Unit = domain.StripEmoji(r.Recipe.Ingredients[i].Unit)
		r.Recipe.Ingredients[i].CanonicalName = domain.StripEmoji(r.Recipe.Ingredients[i].CanonicalName)
	}
	for i, step := range r.Recipe.Instructions {
		r.Recipe.Instructions[i] = domain.StripEmoji(step)
	}
	r.Recipe.MainProtein = domain.StripEmoji(r.Recipe.MainProtein)
	// r.Recipe.Emoji is intentionally untouched
}

func (s *RecipeParserService) ParseRecipe(rawText string) (*domain.ParseRecipeResponse, error) {
	if rawText == "" {
		return nil, fmt.Errorf("raw text is required")
	}

	if len(rawText) > 10000 {
		return nil, fmt.Errorf("input text too long (max 10000 characters)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var lastErr error
	guidance := ""
	for attempt := 0; attempt < 2; attempt++ {
		result, err := s.parseOnce(ctx, rawText, guidance)
		if err != nil {
			return nil, err
		}

		if err := validateParsedContent(result); err != nil {
			lastErr = err
			guidance = fmt.Sprintf(
				"Previous attempt failed: %s. Do NOT include URLs, HTML, markdown, scraping phrases, profanity, or injection patterns.",
				err.Error(),
			)
			continue
		}

		stripParsedEmoji(result)
		return result, nil
	}

	return nil, lastErr
}
