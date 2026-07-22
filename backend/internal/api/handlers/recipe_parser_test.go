package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubParser is a test double for the recipeParser interface.
type stubParser struct {
	result *domain.ParseRecipeResponse
	err    error
}

func (s *stubParser) ParseRecipe(_ string) (*domain.ParseRecipeResponse, error) {
	return s.result, s.err
}

func newTestRecipeParserHandler() *RecipeParserHandler {
	parserSvc := services.NewRecipeParserService(nil)
	recipeSvc := services.NewRecipeService(newMockRecipeStorage())
	return NewRecipeParserHandler(parserSvc, recipeSvc)
}

// ParseRecipe validation tests

func TestRecipeParserHandler_ParseRecipe_EmptyRawText(t *testing.T) {
	h := newTestRecipeParserHandler()

	body := `{"rawText":""}`
	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "raw_text_required" {
		t.Errorf("expected error 'raw_text_required', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseRecipe_MissingRawText(t *testing.T) {
	h := newTestRecipeParserHandler()

	body := `{}`
	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "raw_text_required" {
		t.Errorf("expected error 'raw_text_required', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseRecipe_InputTooLong(t *testing.T) {
	h := newTestRecipeParserHandler()

	longText := strings.Repeat("a", 10001)
	payload := map[string]string{"rawText": longText}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "input_too_long" {
		t.Errorf("expected error 'input_too_long', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseRecipe_InvalidJSON(t *testing.T) {
	h := newTestRecipeParserHandler()

	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ParseAndSave validation tests

func TestRecipeParserHandler_ParseAndSave_EmptyRawText(t *testing.T) {
	h := newTestRecipeParserHandler()

	body := `{"rawText":""}`
	req := httptest.NewRequest("POST", "/recipes/parse-and-save", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseAndSave(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "raw_text_required" {
		t.Errorf("expected error 'raw_text_required', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseAndSave_MissingRawText(t *testing.T) {
	h := newTestRecipeParserHandler()

	body := `{}`
	req := httptest.NewRequest("POST", "/recipes/parse-and-save", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseAndSave(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "raw_text_required" {
		t.Errorf("expected error 'raw_text_required', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseAndSave_InputTooLong(t *testing.T) {
	h := newTestRecipeParserHandler()

	longText := strings.Repeat("a", 10001)
	payload := map[string]string{"rawText": longText}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/recipes/parse-and-save", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseAndSave(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "input_too_long" {
		t.Errorf("expected error 'input_too_long', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseAndSave_InvalidJSON(t *testing.T) {
	h := newTestRecipeParserHandler()

	req := httptest.NewRequest("POST", "/recipes/parse-and-save", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

	rr := httptest.NewRecorder()
	h.ParseAndSave(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// Regression tests: validation sentinels from recipeService.Create must map to 400, not 500.

func newTestRecipeParserHandlerWithStub(parser recipeParser, store *mockRecipeStorage) *RecipeParserHandler {
	return &RecipeParserHandler{
		parserService: parser,
		recipeService: services.NewRecipeService(store),
	}
}

func validParsedRecipe() *domain.ParseRecipeResponse {
	return &domain.ParseRecipeResponse{
		Recipe: domain.CreateRecipeRequest{
			Name:         "Köttbullar",
			Servings:     4,
			Ingredients:  []domain.Ingredient{{Name: "nötkött", Amount: 500, Unit: "g"}},
			Instructions: []string{"Blanda", "Rulla", "Stek"},
		},
		Confidence: 0.9,
	}
}

func TestRecipeParserHandler_ParseAndSave_SaveValidationSentinel_Returns400(t *testing.T) {
	cases := []struct {
		sentinel    error
		wantErrCode string
	}{
		{domain.ErrNameRequired, "name_required"},
		{domain.ErrInvalidServings, "invalid_servings"},
		{domain.ErrIngredientsRequired, "ingredients_required"},
		{domain.ErrInstructionsRequired, "instructions_required"},
		{domain.ErrNameTooLong, "name_too_long"},
		{domain.ErrTooManyIngredients, "too_many_ingredients"},
		{domain.ErrTooManyTags, "too_many_tags"},
		{domain.ErrTagTooLong, "tag_too_long"},
	}

	for _, tc := range cases {
		t.Run(tc.wantErrCode, func(t *testing.T) {
			store := newMockRecipeStorage()
			store.createErr = tc.sentinel
			parser := &stubParser{result: validParsedRecipe()}
			h := newTestRecipeParserHandlerWithStub(parser, store)

			body := `{"rawText":"some recipe text"}`
			req := httptest.NewRequest("POST", "/recipes/parse-and-save", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req = setAuthContext(req, "usr_test-1234-5678-9abc-def012345678", "hh_test-1234-5678-9abc-def012345678")

			rr := httptest.NewRecorder()
			h.ParseAndSave(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("sentinel %q: expected 400, got %d: %s", tc.sentinel, rr.Code, rr.Body.String())
			}
			var errResp map[string]string
			json.NewDecoder(rr.Body).Decode(&errResp)
			if errResp["error"] != tc.wantErrCode {
				t.Errorf("sentinel %q: expected error %q, got %q", tc.sentinel, tc.wantErrCode, errResp["error"])
			}
		})
	}
}
