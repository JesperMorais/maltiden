package handlers

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"log"
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

// parseLocationParams extracts lat, lng, radius from query parameters.
// Returns an error if any present parameter fails to parse (VALID-06).
func parseLocationParams(r *http.Request) (lat float64, lng float64, radius int, err error) {
	lat, lng, radius = defaultLat, defaultLng, defaultRadius
	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid lat: %w", err)
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		lng, err = strconv.ParseFloat(lngStr, 64)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid lng: %w", err)
		}
	}
	if radiusStr := r.URL.Query().Get("radius"); radiusStr != "" {
		radius, err = strconv.Atoi(radiusStr)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid radius: %w", err)
		}
	}
	return lat, lng, radius, nil
}

// parseExcludeStores extracts the comma-separated exclude query parameter into a string slice.
func parseExcludeStores(r *http.Request) []string {
	excludeStr := r.URL.Query().Get("exclude")
	if excludeStr == "" {
		return nil
	}
	var stores []string
	for _, s := range strings.Split(excludeStr, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			stores = append(stores, trimmed)
		}
	}
	return stores
}

func (h *OffersHandler) SearchOffers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	lat, lng, radius, err := parseLocationParams(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_coordinates")
		return
	}

	excludeStores := parseExcludeStores(r)

	req := domain.OfferSearchRequest{
		Query:         query,
		Latitude:      lat,
		Longitude:     lng,
		Radius:        radius,
		ExcludeStores: excludeStores,
	}

	result, err := h.tjekService.SearchOffers(req)
	if err != nil {
		log.Printf("ERROR [SearchOffers] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusBadGateway, "service_unavailable")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *OffersHandler) GetDiscounts(w http.ResponseWriter, r *http.Request) {
	lat, lng, radius, err := parseLocationParams(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_coordinates")
		return
	}

	excludeStores := parseExcludeStores(r)

	result, err := h.tjekService.GetTopDiscounts(lat, lng, radius, excludeStores)
	if err != nil {
		log.Printf("ERROR [GetDiscounts] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusBadGateway, "service_unavailable")
		return
	}

	WriteJSON(w, http.StatusOK, result)
}

func (h *OffersHandler) GetStores(w http.ResponseWriter, r *http.Request) {
	lat, lng, radius, err := parseLocationParams(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_coordinates")
		return
	}

	stores, err := h.tjekService.GetAvailableStores(lat, lng, radius)
	if err != nil {
		log.Printf("ERROR [GetStores] %v", err)
		sentry.CaptureException(err)
		WriteError(w, http.StatusBadGateway, "service_unavailable")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"stores": stores,
	})
}
