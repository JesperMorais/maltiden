package handlers

import (
	"bytes"
	"encoding/json"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestRecipeParserHandler() *RecipeParserHandler {
	parserSvc := services.NewRecipeParserService(nil)
	recipeSvc := services.NewRecipeService(newMockRecipeStorage(), nil)
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
