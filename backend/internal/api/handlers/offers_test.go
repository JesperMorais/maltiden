package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- parseLocationParams ---

func TestParseLocationParams_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search", nil)
	lat, lng, radius, err := parseLocationParams(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lat != defaultLat {
		t.Errorf("lat: want %f got %f", defaultLat, lat)
	}
	if lng != defaultLng {
		t.Errorf("lng: want %f got %f", defaultLng, lng)
	}
	if radius != defaultRadius {
		t.Errorf("radius: want %d got %d", defaultRadius, radius)
	}
}

func TestParseLocationParams_ValidOverrides(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?lat=59.33&lng=18.06&radius=5000", nil)
	lat, lng, radius, err := parseLocationParams(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lat != 59.33 {
		t.Errorf("lat: want 59.33 got %f", lat)
	}
	if lng != 18.06 {
		t.Errorf("lng: want 18.06 got %f", lng)
	}
	if radius != 5000 {
		t.Errorf("radius: want 5000 got %d", radius)
	}
}

func TestParseLocationParams_InvalidLat(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?lat=notanumber", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for invalid lat, got nil")
	}
}

func TestParseLocationParams_InvalidLng(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?lng=bad", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for invalid lng, got nil")
	}
}

func TestParseLocationParams_InvalidRadius(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?radius=abc", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for invalid radius, got nil")
	}
}

func TestParseLocationParams_LatOutOfRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?lat=999", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for lat=999, got nil")
	}
}

func TestParseLocationParams_LngOutOfRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?lng=500", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for lng=500, got nil")
	}
}

func TestParseLocationParams_RadiusZero(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?radius=0", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for radius=0, got nil")
	}
}

func TestParseLocationParams_RadiusNegative(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?radius=-5", nil)
	_, _, _, err := parseLocationParams(req)
	if err == nil {
		t.Fatal("expected error for radius=-5, got nil")
	}
}

// --- parseExcludeStores ---

func TestParseExcludeStores_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search", nil)
	stores := parseExcludeStores(req)
	if stores != nil {
		t.Errorf("expected nil slice, got %v", stores)
	}
}

func TestParseExcludeStores_Single(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?exclude=ICA", nil)
	stores := parseExcludeStores(req)
	if len(stores) != 1 || stores[0] != "ICA" {
		t.Errorf("expected [ICA], got %v", stores)
	}
}

func TestParseExcludeStores_Multiple(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?exclude=ICA,Coop,Lidl", nil)
	stores := parseExcludeStores(req)
	if len(stores) != 3 {
		t.Fatalf("expected 3 stores, got %d: %v", len(stores), stores)
	}
	want := []string{"ICA", "Coop", "Lidl"}
	for i, s := range want {
		if stores[i] != s {
			t.Errorf("stores[%d]: want %q got %q", i, s, stores[i])
		}
	}
}

func TestParseExcludeStores_TrimsWhitespace(t *testing.T) {
	req := httptest.NewRequest("GET", "/offers/search?exclude=+ICA+,+Coop+", nil)
	stores := parseExcludeStores(req)
	if len(stores) != 2 {
		t.Fatalf("expected 2 stores, got %d: %v", len(stores), stores)
	}
	if stores[0] != "ICA" {
		t.Errorf("stores[0]: want %q got %q", "ICA", stores[0])
	}
	if stores[1] != "Coop" {
		t.Errorf("stores[1]: want %q got %q", "Coop", stores[1])
	}
}

// --- Handler validation paths (bad coordinates → 400, no service call) ---

func TestOffersHandler_SearchOffers_InvalidLat(t *testing.T) {
	h := &OffersHandler{tjekService: nil} // service never reached
	req := httptest.NewRequest("GET", "/offers/search?lat=bad", nil)
	rr := httptest.NewRecorder()
	h.SearchOffers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid_coordinates" {
		t.Errorf("error: want %q got %q", "invalid_coordinates", body["error"])
	}
}

func TestOffersHandler_GetDiscounts_InvalidRadius(t *testing.T) {
	h := &OffersHandler{tjekService: nil}
	req := httptest.NewRequest("GET", "/offers/discounts?radius=notanint", nil)
	rr := httptest.NewRecorder()
	h.GetDiscounts(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid_coordinates" {
		t.Errorf("error: want %q got %q", "invalid_coordinates", body["error"])
	}
}

func TestOffersHandler_GetStores_InvalidLng(t *testing.T) {
	h := &OffersHandler{tjekService: nil}
	req := httptest.NewRequest("GET", "/offers/stores?lng=xyz", nil)
	rr := httptest.NewRecorder()
	h.GetStores(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid_coordinates" {
		t.Errorf("error: want %q got %q", "invalid_coordinates", body["error"])
	}
}

// TestOffersHandler_SearchOffers_QueryTooLong asserts the >256-char query
// guard in SearchOffers returns 400 invalid_query. (Carried from main during
// the dev->main sync; dev's suite lacked this length-guard case.)
func TestOffersHandler_SearchOffers_QueryTooLong(t *testing.T) {
	h := &OffersHandler{tjekService: nil} // service never reached
	longQuery := ""
	for i := 0; i < 257; i++ {
		longQuery += "a"
	}
	req := httptest.NewRequest("GET", "/offers/search?q="+longQuery, nil)
	rr := httptest.NewRecorder()
	h.SearchOffers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["error"] != "invalid_query" {
		t.Errorf("error: want %q got %q", "invalid_query", body["error"])
	}
}
