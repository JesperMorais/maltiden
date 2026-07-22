package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"maltiden/internal/domain"
)

// stubGemini is an offline geminiJSONGenerator: it returns canned bytes/err and
// records what it was asked, so the experience layer can be tested without the
// network.
type stubGemini struct {
	resp      []byte
	err       error
	gotSystem string
	gotUser   string
	calls     int
}

func (s *stubGemini) GenerateJSON(_ context.Context, system, user string, _ map[string]interface{}) ([]byte, error) {
	s.calls++
	s.gotSystem = system
	s.gotUser = user
	return s.resp, s.err
}

func TestParseWishes_HappyPath(t *testing.T) {
	stub := &stubGemini{resp: []byte(`{"vegetarianDays":2,"prepMode":true,"extraDislikedIngredients":["koriander"]}`)}
	svc := newMenuExperienceServiceWith(stub)

	pc, err := svc.ParseWishes(context.Background(), "två vegetariska dagar, matlådor, inget koriander")
	if err != nil {
		t.Fatalf("ParseWishes: %v", err)
	}
	if pc.VegetarianDays == nil || *pc.VegetarianDays != 2 {
		t.Errorf("VegetarianDays = %v, want 2", pc.VegetarianDays)
	}
	if pc.PrepMode == nil || !*pc.PrepMode {
		t.Errorf("PrepMode = %v, want true", pc.PrepMode)
	}
	if len(pc.ExtraDislikedIngredients) != 1 || pc.ExtraDislikedIngredients[0] != "koriander" {
		t.Errorf("ExtraDislikedIngredients = %v", pc.ExtraDislikedIngredients)
	}
	if !strings.Contains(stub.gotUser, "koriander") {
		t.Errorf("user prompt missing the wish text: %q", stub.gotUser)
	}
}

func TestParseWishes_EmptyShortCircuits(t *testing.T) {
	stub := &stubGemini{resp: []byte(`{}`)}
	svc := newMenuExperienceServiceWith(stub)
	pc, err := svc.ParseWishes(context.Background(), "   ")
	if err != nil || pc != nil {
		t.Fatalf("expected (nil,nil) for blank wishes, got (%v,%v)", pc, err)
	}
	if stub.calls != 0 {
		t.Errorf("expected no LLM call for blank wishes, got %d", stub.calls)
	}
}

func TestParseWishes_ErrorPropagates(t *testing.T) {
	svc := newMenuExperienceServiceWith(&stubGemini{err: errors.New("boom")})
	if _, err := svc.ParseWishes(context.Background(), "snabb vecka"); err == nil {
		t.Fatal("expected an error to propagate")
	}
}

func TestParseWishes_MalformedJSON(t *testing.T) {
	svc := newMenuExperienceServiceWith(&stubGemini{resp: []byte(`not json`)})
	if _, err := svc.ParseWishes(context.Background(), "snabb vecka"); err == nil {
		t.Fatal("expected a decode error")
	}
}

func arrangeSlots() []ArrangeSlot {
	return []ArrangeSlot{
		{Slot: 0, Weekday: 1, RecipeID: "rec_a", Name: "Tacos"},
		{Slot: 1, Weekday: 2, RecipeID: "rec_b", Name: "Lasagne"},
	}
}

func TestArrangeWeek_ValidPermutation(t *testing.T) {
	// Model swaps the two recipes across the two slots.
	stub := &stubGemini{resp: []byte(`{"assignments":[{"recipeId":"rec_b","slot":0},{"recipeId":"rec_a","slot":1}],"rationale":"Lasagne på måndag, tacos på tisdag."}`)}
	svc := newMenuExperienceServiceWith(stub)

	assignment, rationale, err := svc.ArrangeWeek(context.Background(), arrangeSlots())
	if err != nil {
		t.Fatalf("ArrangeWeek: %v", err)
	}
	if assignment[0] != "rec_b" || assignment[1] != "rec_a" {
		t.Errorf("assignment = %v, want {0:rec_b, 1:rec_a}", assignment)
	}
	if !strings.Contains(rationale, "Lasagne") {
		t.Errorf("rationale = %q", rationale)
	}
}

func TestArrangeWeek_GuardrailRejections(t *testing.T) {
	cases := map[string]string{
		"invented id":    `{"assignments":[{"recipeId":"rec_b","slot":0},{"recipeId":"rec_X","slot":1}],"rationale":"x"}`,
		"unknown slot":   `{"assignments":[{"recipeId":"rec_a","slot":0},{"recipeId":"rec_b","slot":9}],"rationale":"x"}`,
		"duplicate slot": `{"assignments":[{"recipeId":"rec_a","slot":0},{"recipeId":"rec_b","slot":0}],"rationale":"x"}`,
		"wrong count":    `{"assignments":[{"recipeId":"rec_a","slot":0}],"rationale":"x"}`,
		"dup recipe":     `{"assignments":[{"recipeId":"rec_a","slot":0},{"recipeId":"rec_a","slot":1}],"rationale":"x"}`,
		"malformed":      `nope`,
	}
	for name, resp := range cases {
		t.Run(name, func(t *testing.T) {
			svc := newMenuExperienceServiceWith(&stubGemini{resp: []byte(resp)})
			if _, _, err := svc.ArrangeWeek(context.Background(), arrangeSlots()); err == nil {
				t.Errorf("expected guardrail to reject %q", name)
			}
		})
	}
}

func TestArrangeWeek_SanitizesRationale(t *testing.T) {
	// Build the payload with json.Marshal so the rationale carries a real NUL
	// control character (rune 0) and an over-long tail — both injected at
	// runtime, keeping the source file free of NUL bytes. sanitizeRationale must
	// strip the control char and truncate to the cap.
	dirty := "bad" + string(rune(0)) + "char " + strings.Repeat("å", 800)
	payload := map[string]any{
		"assignments": []map[string]any{
			{"recipeId": "rec_a", "slot": 0},
			{"recipeId": "rec_b", "slot": 1},
		},
		"rationale": dirty,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	svc := newMenuExperienceServiceWith(&stubGemini{resp: raw})

	_, rationale, err := svc.ArrangeWeek(context.Background(), arrangeSlots())
	if err != nil {
		t.Fatalf("ArrangeWeek: %v", err)
	}
	if strings.ContainsRune(rationale, rune(0)) {
		t.Error("rationale still contains a control character")
	}
	if n := len([]rune(rationale)); n > domain.MaxRationaleLen {
		t.Errorf("rationale length %d exceeds cap %d", n, domain.MaxRationaleLen)
	}
}
