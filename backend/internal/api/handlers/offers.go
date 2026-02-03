package handlers

import (
	"encoding/json"
	"maltiden/internal/domain"
	"maltiden/internal/services"
	"net/http"
	"strconv"
	"strings"
)

// Haninge coordinates (hardcoded for POC)
const (
	defaultLat    = 59.1739
	defaultLng    = 18.1489
	defaultRadius = 15000 // 15km
)

type OffersHandler struct {
	tjekService *services.TjekService
}

func NewOffersHandler(tjekService *services.TjekService) *OffersHandler {
	return &OffersHandler{tjekService: tjekService}
}

func (h *OffersHandler) SearchOffers(w http.ResponseWriter, r *http.Request) {
	// Get query from URL params
	query := r.URL.Query().Get("q")

	// Optional: allow overriding coordinates
	lat := defaultLat
	lng := defaultLng
	radius := defaultRadius

	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = parsed
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		if parsed, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = parsed
		}
	}
	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		if parsed, err := strconv.Atoi(radiusStr); err == nil {
			radius = parsed
		}
	}

	// Get excluded stores (comma-separated)
	var excludeStores []string
	if excludeStr := r.URL.Query().Get("exclude"); excludeStr != "" {
		for _, s := range strings.Split(excludeStr, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				excludeStores = append(excludeStores, trimmed)
			}
		}
	}

	// Build request
	req := domain.OfferSearchRequest{
		Query:         query,
		Latitude:      lat,
		Longitude:     lng,
		Radius:        radius,
		ExcludeStores: excludeStores,
	}

	// Call Tjek API
	result, err := h.tjekService.SearchOffers(req)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch offers","details":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *OffersHandler) GetDiscounts(w http.ResponseWriter, r *http.Request) {
	lat := defaultLat
	lng := defaultLng
	radius := defaultRadius

	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = parsed
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		if parsed, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = parsed
		}
	}
	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		if parsed, err := strconv.Atoi(radiusStr); err == nil {
			radius = parsed
		}
	}

	// Get excluded stores (comma-separated)
	var excludeStores []string
	if excludeStr := r.URL.Query().Get("exclude"); excludeStr != "" {
		for _, s := range strings.Split(excludeStr, ",") {
			if trimmed := strings.TrimSpace(s); trimmed != "" {
				excludeStores = append(excludeStores, trimmed)
			}
		}
	}

	result, err := h.tjekService.GetTopDiscounts(lat, lng, radius, excludeStores)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch discounts","details":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *OffersHandler) GetStores(w http.ResponseWriter, r *http.Request) {
	lat := defaultLat
	lng := defaultLng
	radius := defaultRadius

	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if parsed, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = parsed
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		if parsed, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = parsed
		}
	}
	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		if parsed, err := strconv.Atoi(radiusStr); err == nil {
			radius = parsed
		}
	}

	stores, err := h.tjekService.GetAvailableStores(lat, lng, radius)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch stores","details":"`+err.Error()+`"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stores": stores,
	})
}
