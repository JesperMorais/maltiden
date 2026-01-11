package services

import (
	"encoding/json"
	"fmt"
	"maltiden/internal/domain"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// Tjek API uses non-standard timezone format (+0000 instead of +00:00)
var tjekTimeFormats = []string{
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05Z0700",
	time.RFC3339,
}

func parseTjekTime(s string) time.Time {
	for _, format := range tjekTimeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

type TjekService struct {
	baseURL    string
	httpClient *http.Client
}

func NewTjekService() *TjekService {
	return &TjekService{
		baseURL: "https://api.etilbudsavis.dk/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// catalogResponse represents a weekly flyer/catalog
type catalogResponse struct {
	ID       string `json:"id"`
	DealerID string `json:"dealer_id"`
	RunFrom  string `json:"run_from"`
	RunTill  string `json:"run_till"`
	Branding struct {
		Name string `json:"name"`
		Logo string `json:"logo"`
	} `json:"branding"`
}

// storeResponse represents a physical store location
type storeResponse struct {
	ID       string `json:"id"`
	Street   string `json:"street"`
	City     string `json:"city"`
	DealerID string `json:"dealer_id"`
}

// hotspotResponse represents an offer from a catalog
type hotspotResponse struct {
	Type  string `json:"type"`
	Offer struct {
		ID      string `json:"id"`
		Heading string `json:"heading"`
		Pricing struct {
			Price    float64 `json:"price"`
			PrePrice float64 `json:"pre_price"`
			Currency string  `json:"currency"`
		} `json:"pricing"`
		Quantity struct {
			Unit struct {
				Symbol string `json:"symbol"`
			} `json:"unit"`
			Size struct {
				From float64 `json:"from"`
				To   float64 `json:"to"`
			} `json:"size"`
			Pieces struct {
				From float64 `json:"from"`
				To   float64 `json:"to"`
			} `json:"pieces"`
		} `json:"quantity"`
		RunFrom string `json:"run_from"`
		RunTill string `json:"run_till"`
	} `json:"offer"`
}

// matchesWord checks if the query matches as a word component
func matchesWord(text, query string) bool {
	text = strings.ToLower(text)
	query = strings.ToLower(query)

	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for _, word := range words {
		if strings.HasPrefix(word, query) || strings.HasSuffix(word, query) {
			return true
		}
	}
	return false
}

// isComboOffer checks if an offer contains multiple products (combo deal)
func isComboOffer(heading string) bool {
	// Count separators that indicate multiple products
	slashCount := strings.Count(heading, "/")
	commaCount := strings.Count(heading, ",")

	// If more than 2 separators, it's likely a combo offer
	return slashCount > 2 || commaCount > 3
}

// buildDescription creates a human-readable description from quantity info
func buildDescription(unit string, sizeFrom, sizeTo, piecesFrom, piecesTo float64) string {
	if unit == "" || sizeFrom <= 0 {
		return ""
	}

	unit = strings.ToLower(unit)
	pieces := int(piecesFrom)
	size := sizeFrom

	// Convert to common unit for comparison (ml)
	sizeInMl := size
	if unit == "cl" {
		sizeInMl = size * 10
	} else if unit == "l" {
		sizeInMl = size * 1000
	}

	// Format the size value
	var sizeStr string
	if sizeFrom == sizeTo {
		sizeStr = formatSize(size, unit)
	} else {
		sizeStr = fmt.Sprintf("%.0f-%.0f %s", sizeFrom, sizeTo, unit)
	}

	// Determine packaging type based on unit and size
	switch unit {
	case "cl", "ml", "l":
		// Drinks - determine packaging based on size
		packaging := determineDrinkPackaging(sizeInMl)

		if pieces > 1 {
			// Multiple items
			if packaging == "burk" {
				return fmt.Sprintf("%d-pack %s (%s)", pieces, packaging, sizeStr)
			}
			return fmt.Sprintf("%d st (%s)", pieces, sizeStr)
		}
		// Single item
		return fmt.Sprintf("1 %s (%s)", packaging, sizeStr)

	case "g":
		// Food by weight
		if pieces > 1 {
			return fmt.Sprintf("%d st (%s)", pieces, sizeStr)
		}
		return sizeStr

	case "kg":
		// Heavier food
		if pieces > 1 {
			return fmt.Sprintf("%d st (%s)", pieces, sizeStr)
		}
		return sizeStr

	case "st", "pcs":
		// Pieces/items
		if pieces > 1 {
			return fmt.Sprintf("%d-pack", pieces)
		}
		if size > 1 {
			return fmt.Sprintf("%.0f st", size)
		}
		return ""

	default:
		if pieces > 1 {
			return fmt.Sprintf("%d st (%s)", pieces, sizeStr)
		}
		return sizeStr
	}
}

// determineDrinkPackaging determines if a drink is in a can or bottle based on size
func determineDrinkPackaging(sizeInMl float64) string {
	// Common can sizes: 250ml, 330ml, 355ml, 500ml
	// Common bottle sizes: 500ml, 1000ml, 1500ml, 2000ml
	// Energy drinks: 250ml, 500ml cans
	// Soda cans: 330ml, 355ml
	// Soda bottles: 500ml, 1.5L, 2L

	switch {
	case sizeInMl <= 250:
		// Small energy drink can (Monster, Red Bull, etc.)
		return "burk"
	case sizeInMl <= 355:
		// Standard soda can (33cl, 35.5cl)
		return "burk"
	case sizeInMl <= 500:
		// Could be large can (50cl energy drink) or small bottle
		// 500ml is ambiguous - common for both
		// Energy drinks at 500ml are usually cans, soda at 500ml usually bottles
		// Default to burk for 500ml since energy drinks are common
		return "burk"
	case sizeInMl < 1000:
		// 750ml etc - usually bottles
		return "flaska"
	default:
		// 1L and above - definitely bottles
		return "flaska"
	}
}

// formatSize formats size with appropriate precision
func formatSize(size float64, unit string) string {
	if size == float64(int(size)) {
		return fmt.Sprintf("%.0f %s", size, unit)
	}
	return fmt.Sprintf("%.1f %s", size, unit)
}

// isGroceryStore checks if the dealer is a grocery store
func isGroceryStore(name string) bool {
	name = strings.ToLower(name)

	// Exclude non-food stores
	excludeList := []string{
		"bauhaus", "biltema", "jula", "jysk", "mio", "rusta",
		"netonnet", "elgiganten", "mediamarkt", "kjell", "power",
		"apoteket", "apotek", "clas ohlson", "xxl", "stadium",
		"intersport", "sportamore", "zalando", "hm", "lindex",
	}
	for _, exclude := range excludeList {
		if strings.Contains(name, exclude) {
			return false
		}
	}

	// Include known grocery stores
	includeList := []string{
		"willys", "ica", "coop", "lidl", "hemköp", "city gross",
		"citygross", "öob", "ö&b", "matöppet", "nära dej", "näradej",
		"daglivs", "din mat", "tempo", "handlar'n", "dollarstore",
		"erlandsons", "kvantum", "maxi", "supermarket",
	}
	for _, include := range includeList {
		if strings.Contains(name, include) {
			return true
		}
	}

	return false
}

// GetAvailableStores returns list of grocery stores in the area
func (s *TjekService) GetAvailableStores(lat, lng float64, radius int) ([]string, error) {
	catalogs, err := s.getCatalogs(lat, lng, radius)
	if err != nil {
		return nil, err
	}

	// Get unique grocery store names
	storeMap := make(map[string]bool)
	for _, catalog := range catalogs {
		if isGroceryStore(catalog.Branding.Name) {
			storeMap[catalog.Branding.Name] = true
		}
	}

	var stores []string
	for name := range storeMap {
		stores = append(stores, name)
	}

	// Sort alphabetically
	for i := 0; i < len(stores)-1; i++ {
		for j := i + 1; j < len(stores); j++ {
			if stores[j] < stores[i] {
				stores[i], stores[j] = stores[j], stores[i]
			}
		}
	}

	return stores, nil
}

// SearchOffers fetches offers from nearby grocery store catalogs
func (s *TjekService) SearchOffers(req domain.OfferSearchRequest) (*domain.OfferSearchResponse, error) {
	// Step 1: Get catalogs near the location
	catalogs, err := s.getCatalogs(req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalogs: %w", err)
	}

	// Step 2: Get nearby stores for address info
	stores, _ := s.getStores(req.Latitude, req.Longitude, req.Radius)
	storesByDealer := make(map[string]storeResponse)
	for _, store := range stores {
		if _, exists := storesByDealer[store.DealerID]; !exists {
			storesByDealer[store.DealerID] = store
		}
	}

	// Step 3: Get offers from each catalog
	var allOffers []domain.TjekOffer
	query := strings.ToLower(req.Query)

	// Build exclude map for fast lookup
	excludeMap := make(map[string]bool)
	for _, store := range req.ExcludeStores {
		excludeMap[strings.ToLower(store)] = true
	}

	for _, catalog := range catalogs {
		// Filter to grocery stores
		if !isGroceryStore(catalog.Branding.Name) {
			continue
		}

		// Skip excluded stores
		if excludeMap[strings.ToLower(catalog.Branding.Name)] {
			continue
		}

		// Get store address
		store := storesByDealer[catalog.DealerID]

		// Get offers from this catalog
		offers, err := s.getCatalogOffers(catalog, store)
		if err != nil {
			continue
		}

		// Filter by search query and skip combo offers
		for _, offer := range offers {
			// Skip combo offers (multiple products in one deal)
			if isComboOffer(offer.Heading) {
				continue
			}

			if query == "" || matchesWord(offer.Heading, query) {
				allOffers = append(allOffers, offer)
			}
		}
	}

	// Limit results
	if len(allOffers) > 100 {
		allOffers = allOffers[:100]
	}

	return &domain.OfferSearchResponse{
		Offers: allOffers,
		Count:  len(allOffers),
	}, nil
}

// GetTopDiscounts fetches the best discounted offers sorted by discount percentage
func (s *TjekService) GetTopDiscounts(lat, lng float64, radius int, excludeStores []string) (*domain.OfferSearchResponse, error) {
	// Step 1: Get catalogs near the location
	catalogs, err := s.getCatalogs(lat, lng, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalogs: %w", err)
	}

	// Step 2: Get nearby stores for address info
	stores, _ := s.getStores(lat, lng, radius)
	storesByDealer := make(map[string]storeResponse)
	for _, store := range stores {
		if _, exists := storesByDealer[store.DealerID]; !exists {
			storesByDealer[store.DealerID] = store
		}
	}

	// Step 3: Get offers from each catalog, only with discounts
	type offerWithDiscount struct {
		offer    domain.TjekOffer
		discount float64
	}
	var discountedOffers []offerWithDiscount

	// Build exclude map for fast lookup
	excludeMap := make(map[string]bool)
	for _, store := range excludeStores {
		excludeMap[strings.ToLower(store)] = true
	}

	for _, catalog := range catalogs {
		if !isGroceryStore(catalog.Branding.Name) {
			continue
		}

		// Skip excluded stores
		if excludeMap[strings.ToLower(catalog.Branding.Name)] {
			continue
		}

		store := storesByDealer[catalog.DealerID]
		offers, err := s.getCatalogOffers(catalog, store)
		if err != nil {
			continue
		}

		for _, offer := range offers {
			// Skip combo offers
			if isComboOffer(offer.Heading) {
				continue
			}

			// Only include offers with actual discount
			if offer.PrePrice > 0 && offer.PrePrice > offer.Price {
				discount := (1 - offer.Price/offer.PrePrice) * 100
				discountedOffers = append(discountedOffers, offerWithDiscount{
					offer:    offer,
					discount: discount,
				})
			}
		}
	}

	// Sort by discount percentage (highest first)
	for i := 0; i < len(discountedOffers)-1; i++ {
		for j := i + 1; j < len(discountedOffers); j++ {
			if discountedOffers[j].discount > discountedOffers[i].discount {
				discountedOffers[i], discountedOffers[j] = discountedOffers[j], discountedOffers[i]
			}
		}
	}

	// Extract sorted offers
	var result []domain.TjekOffer
	for _, item := range discountedOffers {
		result = append(result, item.offer)
		if len(result) >= 50 {
			break
		}
	}

	return &domain.OfferSearchResponse{
		Offers: result,
		Count:  len(result),
	}, nil
}

// getCatalogs fetches weekly flyers near a location
func (s *TjekService) getCatalogs(lat, lng float64, radius int) ([]catalogResponse, error) {
	endpoint := fmt.Sprintf("%s/catalogs", s.baseURL)
	params := url.Values{}
	params.Set("r_lat", fmt.Sprintf("%f", lat))
	params.Set("r_lng", fmt.Sprintf("%f", lng))
	params.Set("r_radius", fmt.Sprintf("%d", radius))
	params.Set("types", "paged")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := s.httpClient.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalogs API returned status %d", resp.StatusCode)
	}

	var catalogs []catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalogs); err != nil {
		return nil, err
	}

	return catalogs, nil
}

// getCatalogOffers fetches all offers from a specific catalog
func (s *TjekService) getCatalogOffers(catalog catalogResponse, store storeResponse) ([]domain.TjekOffer, error) {
	endpoint := fmt.Sprintf("%s/catalogs/%s/hotspots", s.baseURL, catalog.ID)

	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hotspots API returned status %d", resp.StatusCode)
	}

	var hotspots []hotspotResponse
	if err := json.NewDecoder(resp.Body).Decode(&hotspots); err != nil {
		return nil, err
	}

	var offers []domain.TjekOffer
	for _, hs := range hotspots {
		if hs.Type != "offer" || hs.Offer.Heading == "" {
			continue
		}

		validFrom := parseTjekTime(hs.Offer.RunFrom)
		validTo := parseTjekTime(hs.Offer.RunTill)

		// Build description with quantity info
		description := buildDescription(
			hs.Offer.Quantity.Unit.Symbol,
			hs.Offer.Quantity.Size.From,
			hs.Offer.Quantity.Size.To,
			hs.Offer.Quantity.Pieces.From,
			hs.Offer.Quantity.Pieces.To,
		)

		offers = append(offers, domain.TjekOffer{
			ID:           hs.Offer.ID,
			Heading:      hs.Offer.Heading,
			Description:  description,
			Price:        hs.Offer.Pricing.Price,
			PrePrice:     hs.Offer.Pricing.PrePrice,
			Currency:     hs.Offer.Pricing.Currency,
			ValidFrom:    validFrom,
			ValidTo:      validTo,
			StoreName:    catalog.Branding.Name,
			StoreLogo:    catalog.Branding.Logo,
			StoreAddress: store.Street,
			StoreCity:    store.City,
		})
	}

	return offers, nil
}

// getStores fetches physical store locations near a location
func (s *TjekService) getStores(lat, lng float64, radius int) ([]storeResponse, error) {
	endpoint := fmt.Sprintf("%s/stores", s.baseURL)
	params := url.Values{}
	params.Set("r_lat", fmt.Sprintf("%f", lat))
	params.Set("r_lng", fmt.Sprintf("%f", lng))
	params.Set("r_radius", fmt.Sprintf("%d", radius))

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := s.httpClient.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stores API returned status %d", resp.StatusCode)
	}

	var stores []storeResponse
	if err := json.NewDecoder(resp.Body).Decode(&stores); err != nil {
		return nil, err
	}

	return stores, nil
}
