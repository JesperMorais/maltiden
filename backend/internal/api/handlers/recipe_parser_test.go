package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRecipeParserHandler() *RecipeParserHandler {
	return NewRecipeParserHandler(nil, nil)
}

func TestRecipeParserHandler_ParseRecipe_MissingRawText(t *testing.T) {
	h := newTestRecipeParserHandler()

	body := `{"rawText":""}`
	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "raw_text_required" {
		t.Errorf("expected error 'raw_text_required', got %q", errResp["error"])
	}
}

func TestRecipeParserHandler_ParseRecipe_TooLong(t *testing.T) {
	h := newTestRecipeParserHandler()

	longText := strings.Repeat("a", 10001)
	payload, _ := json.Marshal(map[string]string{"rawText": longText})
	req := httptest.NewRequest("POST", "/recipes/parse", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != 400 {
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

	rr := httptest.NewRecorder()
	h.ParseRecipe(rr, req)

	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}
