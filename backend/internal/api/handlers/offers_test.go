package handlers

import (
	"encoding/json"
	"maltiden/internal/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestOffersHandler creates an OffersHandler wired to a real TjekService.
// Only pre-service 400 validation paths are exercised in these tests, so no
// network call to the tjek API is ever made.
func newTestOffersHandler() *OffersHandler {
	return NewOffersHandler(services.NewTjekService())
}

func TestOffersHandler_SearchOffers_InvalidCoordinates(t *testing.T) {
	h := newTestOffersHandler()

	req := httptest.NewRequest("GET", "/offers/search?lat=notanumber", nil)
	rr := httptest.NewRecorder()
	h.SearchOffers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_coordinates" {
		t.Errorf("expected error 'invalid_coordinates', got %q", errResp["error"])
	}
}

func TestOffersHandler_SearchOffers_QueryTooLong(t *testing.T) {
	h := newTestOffersHandler()

	longQuery := strings.Repeat("a", 257)
	req := httptest.NewRequest("GET", "/offers/search?q="+longQuery, nil)
	rr := httptest.NewRecorder()
	h.SearchOffers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}

	var errResp map[string]string
	json.NewDecoder(rr.Body).Decode(&errResp)
	if errResp["error"] != "invalid_query" {
		t.Errorf("expected error 'invalid_query', got %q", errResp["error"])
	}
}
