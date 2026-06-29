package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"maltiden/internal/domain"
	"maltiden/pkg/gemini"
)

// geminiJSONGenerator is the slice of the Gemini client the experience layer
// needs. It is an interface so tests can inject a stub and never touch the
// network (mirrors the claudeSender seam in recipe_parser_service.go).
type geminiJSONGenerator interface {
	GenerateJSON(ctx context.Context, systemPrompt, userContent string, schema map[string]interface{}) ([]byte, error)
}

// MenuExperienceService is the optional, provider-neutral "experience layer"
// (Phase 4). Backed by Gemini Flash-Lite, it (1) parses a household's free-text
// Swedish wishes into structured per-run constraints and (2) arranges the
// already-chosen recipes across weekdays and writes a short Swedish rationale.
// It is constructed only when an API key is present; MenuService treats a nil
// service as "feature unavailable" and degrades gracefully.
type MenuExperienceService struct {
	client geminiJSONGenerator
}

// NewMenuExperienceService wraps a real Gemini client.
func NewMenuExperienceService(client *gemini.Client) *MenuExperienceService {
	return &MenuExperienceService{client: client}
}

// newMenuExperienceServiceWith is the test seam: it accepts any
// geminiJSONGenerator (e.g. a stub) so unit tests run offline.
func newMenuExperienceServiceWith(g geminiJSONGenerator) *MenuExperienceService {
	return &MenuExperienceService{client: g}
}

// wishSchema constrains the wish-parsing output to known constraint fields only
// — never recipe choices. Gemini's responseSchema is an OpenAPI 3.0 subset, so
// `additionalProperties` is omitted and no field is required (the model sets
// only what the wish clearly implies; everything else stays null/omitted).
var wishSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"vegetarianDays": map[string]interface{}{"type": "integer"},
		"prepMode":       map[string]interface{}{"type": "boolean"},
		"days":           map[string]interface{}{"type": "integer"},
		"servings":       map[string]interface{}{"type": "integer"},
		"extraExcludedTags": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
		"extraDislikedIngredients": map[string]interface{}{
			"type":  "array",
			"items": map[string]interface{}{"type": "string"},
		},
	},
}

const wishSystemPrompt = `You convert a Swedish household's free-text wish for their weekly meal plan into structured generation constraints for Måltiden. Output JSON only.

Set a field ONLY when the wish clearly implies it; otherwise leave it null/omit. You do NOT choose or name recipes, and you do NOT invent tags the wish never mentions.

Fields:
- "vegetarianDays": number of vegetarian/meatless days requested (e.g. "två vegetariska dagar" → 2).
- "prepMode": true if they want batch cooking / leftovers ("matlådor", "laga en gång ät flera", "förbered").
- "days": number of days only if explicitly stated.
- "servings": portions per meal only if explicitly stated.
- "extraExcludedTags": Swedish recipe tags/categories to AVOID (e.g. "ingen fisk" → ["fisk"]).
- "extraDislikedIngredients": specific ingredients to avoid (e.g. "inget koriander" → ["koriander"]).

Cost wishes ("billig", "spara pengar") cannot be honored structurally — ignore them.`

// ParseWishes turns free-text Swedish wishes into structured constraints. The
// caller is responsible for clamping (domain.ParsedWishConstraints.Clamp) and
// merging the result. Returns an error on an LLM failure or malformed output;
// the caller then ignores the wishes and proceeds deterministically.
func (s *MenuExperienceService) ParseWishes(ctx context.Context, wishes string) (*domain.ParsedWishConstraints, error) {
	wishes = strings.TrimSpace(wishes)
	if wishes == "" {
		return nil, nil
	}
	raw, err := s.client.GenerateJSON(ctx, wishSystemPrompt, "Wish: "+wishes, wishSchema)
	if err != nil {
		return nil, fmt.Errorf("parse wishes: %w", err)
	}
	var pc domain.ParsedWishConstraints
	if err := json.Unmarshal(raw, &pc); err != nil {
		return nil, fmt.Errorf("parse wishes: decode: %w", err)
	}
	return &pc, nil
}

// ArrangeSlot is one freely-movable menu slot handed to the arranger: the
// recipe currently placed there, its display name, and the slot's weekday
// (Go time.Weekday, 0=Sun..6=Sat).
type ArrangeSlot struct {
	Slot     int
	Weekday  int
	RecipeID string
	Name     string
}

var arrangeSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"assignments": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"recipeId": map[string]interface{}{"type": "string"},
					"slot":     map[string]interface{}{"type": "integer"},
				},
				"required": []string{"recipeId", "slot"},
			},
		},
		"rationale": map[string]interface{}{"type": "string"},
	},
	"required": []string{"assignments", "rationale"},
}

const arrangeSystemPrompt = `You arrange a week of already-chosen recipes onto weekdays for Måltiden, and explain the week in Swedish. Output JSON only.

You are given a list of slots, each with a recipe (id + name) and a weekday. Reassign the recipes across EXACTLY these slots — a permutation: every listed recipe goes to exactly one of the listed slots, and every listed slot gets exactly one recipe. Reference only the given recipe ids and slot numbers; never invent or drop any.

Guidance: put heavier or slower dishes on weekend days (fredag, lördag, söndag) and quick/simple dishes on busy weekdays (måndag–torsdag).

Also write "rationale": one or two short Swedish sentences explaining why the week works (e.g. shared ingredients, a quick weekday, a weekend dish). Keep it warm and concrete; do not mention slot numbers or ids.`

type arrangeResult struct {
	Assignments []struct {
		RecipeID string `json:"recipeId"`
		Slot     int    `json:"slot"`
	} `json:"assignments"`
	Rationale string `json:"rationale"`
}

// ArrangeWeek asks the model to permute the given recipes across the given
// slots and produce a Swedish rationale. It enforces the guardrail strictly:
// the returned assignment must be a permutation of EXACTLY the input slots and
// reference EXACTLY the input recipe-id multiset. On any violation (or LLM/JSON
// failure) it returns an error and the caller keeps the deterministic order.
func (s *MenuExperienceService) ArrangeWeek(ctx context.Context, slots []ArrangeSlot) (map[int]string, string, error) {
	if len(slots) == 0 {
		return nil, "", fmt.Errorf("arrange: no slots")
	}

	var b strings.Builder
	b.WriteString("Slots to arrange:\n")
	for _, sl := range slots {
		fmt.Fprintf(&b, "- slot %d (%s): %q [id: %s]\n", sl.Slot, weekdayNameSv(sl.Weekday), sl.Name, sl.RecipeID)
	}

	raw, err := s.client.GenerateJSON(ctx, arrangeSystemPrompt, b.String(), arrangeSchema)
	if err != nil {
		return nil, "", fmt.Errorf("arrange: %w", err)
	}
	var res arrangeResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, "", fmt.Errorf("arrange: decode: %w", err)
	}

	// GUARDRAIL: validate the assignment is a permutation of exactly the input
	// slots and recipe-id multiset.
	expectedSlots := make(map[int]bool, len(slots))
	wantCount := make(map[string]int, len(slots))
	for _, sl := range slots {
		expectedSlots[sl.Slot] = true
		wantCount[sl.RecipeID]++
	}
	if len(res.Assignments) != len(slots) {
		return nil, "", fmt.Errorf("arrange: got %d assignments, want %d", len(res.Assignments), len(slots))
	}
	assignment := make(map[int]string, len(slots))
	gotCount := make(map[string]int, len(slots))
	for _, a := range res.Assignments {
		if !expectedSlots[a.Slot] {
			return nil, "", fmt.Errorf("arrange: slot %d not in the offered set", a.Slot)
		}
		if _, dup := assignment[a.Slot]; dup {
			return nil, "", fmt.Errorf("arrange: slot %d assigned twice", a.Slot)
		}
		assignment[a.Slot] = a.RecipeID
		gotCount[a.RecipeID]++
	}
	for id, want := range wantCount {
		if gotCount[id] != want {
			return nil, "", fmt.Errorf("arrange: recipe %s assigned %d times, want %d", id, gotCount[id], want)
		}
	}
	if len(gotCount) != len(wantCount) {
		return nil, "", fmt.Errorf("arrange: assignment references unexpected recipe ids")
	}

	return assignment, sanitizeRationale(res.Rationale), nil
}

// sanitizeRationale strips control characters (collapsing them to spaces),
// trims, and caps the result to domain.MaxRationaleLen runes.
func sanitizeRationale(s string) string {
	s = strings.Map(func(r rune) rune {
		if r != '\n' && r != '\t' && unicode.IsControl(r) {
			return -1
		}
		if r == '\n' || r == '\t' {
			return ' '
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > domain.MaxRationaleLen {
		runes := []rune(s)
		s = strings.TrimSpace(string(runes[:domain.MaxRationaleLen]))
	}
	return s
}

// weekdayNameSv maps a Go time.Weekday int (0=Sun..6=Sat) to its Swedish name.
func weekdayNameSv(w int) string {
	names := []string{"söndag", "måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag"}
	if w < 0 || w >= len(names) {
		return "okänd dag"
	}
	return names[w]
}
