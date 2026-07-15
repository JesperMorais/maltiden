package services

import (
	"encoding/json"
	"maltiden/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestTjekService(handler http.HandlerFunc) (*TjekService, *httptest.Server) {
	srv := httptest.NewServer(handler)
	s := &TjekService{
		baseURL:    srv.URL,
		httpClient: srv.Client(),
		cache:      make(map[string]cacheEntry),
	}
	return s, srv
}

func TestGetAvailableStores_FiltersGroceryOnly(t *testing.T) {
	s, srv := newTestTjekService(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/catalogs":
			catalogs := []catalogResponse{
				{ID: "c1", DealerID: "d1"},
				{ID: "c2", DealerID: "d2"},
			}
			catalogs[0].Branding.Name = "Willys"
			catalogs[1].Branding.Name = "Bauhaus"
			json.NewEncoder(w).Encode(catalogs)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()

	stores, err := s.GetAvailableStores(59.3, 18.0, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stores) != 1 || stores[0] != "Willys" {
		t.Errorf("expected [Willys], got %v", stores)
	}
}

func TestGetCatalogs_CacheHit(t *testing.T) {
	calls := 0
	s, srv := newTestTjekService(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode([]catalogResponse{})
	})
	defer srv.Close()

	if _, err := s.getCatalogs(59.3, 18.0, 5000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := s.getCatalogs(59.3, 18.0, 5000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 upstream call (second served from cache), got %d", calls)
	}
}

func TestGetCatalogs_ErrorStatus(t *testing.T) {
	s, srv := newTestTjekService(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()
	s.httpClient.Timeout = 0

	if _, err := s.getCatalogs(59.3, 18.0, 5000); err == nil {
		t.Error("expected error on 500 status")
	}
}

func TestSearchOffers_MatchesQueryAndSkipsCombo(t *testing.T) {
	s, srv := newTestTjekService(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/catalogs":
			c := catalogResponse{ID: "c1", DealerID: "d1"}
			c.Branding.Name = "Willys"
			json.NewEncoder(w).Encode([]catalogResponse{c})
		case "/stores":
			json.NewEncoder(w).Encode([]storeResponse{})
		case "/catalogs/c1/hotspots":
			hs := hotspotResponse{Type: "offer"}
			hs.Offer.ID = "o1"
			hs.Offer.Heading = "Kycklingfilé"
			hs.Offer.Pricing.Price = 49
			combo := hotspotResponse{Type: "offer"}
			combo.Offer.ID = "o2"
			combo.Offer.Heading = "A / B / C / D"
			json.NewEncoder(w).Encode([]hotspotResponse{hs, combo})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()

	resp, err := s.SearchOffers(domain.OfferSearchRequest{Latitude: 59.3, Longitude: 18.0, Radius: 5000, Query: "kyckling"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 1 || resp.Offers[0].ID != "o1" {
		t.Errorf("expected only o1 to match, got %+v", resp.Offers)
	}
}

func TestGetTopDiscounts_SortsByDiscountDescending(t *testing.T) {
	s, srv := newTestTjekService(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/catalogs":
			c := catalogResponse{ID: "c1", DealerID: "d1"}
			c.Branding.Name = "ICA"
			json.NewEncoder(w).Encode([]catalogResponse{c})
		case "/stores":
			json.NewEncoder(w).Encode([]storeResponse{})
		case "/catalogs/c1/hotspots":
			low := hotspotResponse{Type: "offer"}
			low.Offer.ID = "low"
			low.Offer.Heading = "Lite rabatt"
			low.Offer.Pricing.Price = 90
			low.Offer.Pricing.PrePrice = 100

			high := hotspotResponse{Type: "offer"}
			high.Offer.ID = "high"
			high.Offer.Heading = "Stor rabatt"
			high.Offer.Pricing.Price = 50
			high.Offer.Pricing.PrePrice = 100

			json.NewEncoder(w).Encode([]hotspotResponse{low, high})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	})
	defer srv.Close()

	resp, err := s.GetTopDiscounts(59.3, 18.0, 5000, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Count != 2 || resp.Offers[0].ID != "high" || resp.Offers[1].ID != "low" {
		t.Errorf("expected [high, low] sorted by discount, got %+v", resp.Offers)
	}
}
