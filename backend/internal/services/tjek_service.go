package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maltiden/internal/domain"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/getsentry/sentry-go"
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

// cacheEntry represents a cached API response with TTL
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

const maxCacheEntries = 500

type TjekService struct {
	baseURL     string
	httpClient  *http.Client
	cache       map[string]cacheEntry
	cacheMu     sync.Mutex
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64
}

func NewTjekService() *TjekService {
	s := &TjekService{
		baseURL: "https://api.etilbudsavis.dk/v2",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]cacheEntry),
	}
	go s.logCacheStats()
	return s
}

// getFromCache retrieves a cached entry if it exists and hasn't expired
func (s *TjekService) getFromCache(key string) (interface{}, bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	entry, exists := s.cache[key]
	if !exists {
		s.cacheMisses.Add(1)
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		delete(s.cache, key)
		s.cacheMisses.Add(1)
		return nil, false
	}

	s.cacheHits.Add(1)
	return entry.data, true
}

// CacheStats returns current cache hit/miss counts
func (s *TjekService) CacheStats() (hits, misses int64) {
	return s.cacheHits.Load(), s.cacheMisses.Load()
}

// logCacheStats periodically logs cache statistics
func (s *TjekService) logCacheStats() {
	for {
		time.Sleep(5 * time.Minute)
		hits, misses := s.CacheStats()
		total := hits + misses
		if total == 0 {
			continue
		}
		rate := float64(hits) / float64(total) * 100
		s.cacheMu.Lock()
		size := len(s.cache)
		s.cacheMu.Unlock()
		slog.Info("tjek cache stats", "hits", hits, "misses", misses, "hit_rate", fmt.Sprintf("%.1f%%", rate), "entries", size)
	}
}

// setCache stores an entry in cache with TTL, evicting expired entries if cache is full
func (s *TjekService) setCache(key string, data interface{}, ttl time.Duration) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Evict expired entries if at capacity
	if len(s.cache) >= maxCacheEntries {
		now := time.Now()
		for k, v := range s.cache {
			if now.After(v.expiresAt) {
				delete(s.cache, k)
			}
		}
		// If still at capacity after evicting expired, drop oldest entry
		if len(s.cache) >= maxCacheEntries {
			var oldestKey string
			var oldestTime time.Time
			for k, v := range s.cache {
				if oldestKey == "" || v.expiresAt.Before(oldestTime) {
					oldestKey = k
					oldestTime = v.expiresAt
				}
			}
			delete(s.cache, oldestKey)
		}
	}

	s.cache[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

// doWithRetry retries HTTP requests on transient failures with exponential backoff
func (s *TjekService) doWithRetry(fn func() (*http.Response, error)) (*http.Response, error) {
	maxAttempts := 3
	backoffs := []time.Duration{0, 100 * time.Millisecond, 200 * time.Millisecond}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(backoffs[attempt])
		}

		resp, err := fn()

		// Success case
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Determine if we should retry
		shouldRetry := false

		// Network errors - always retry
		if err != nil {
			lastErr = err
			shouldRetry = true
		} else {
			// HTTP errors
			if resp.StatusCode == 429 || resp.StatusCode >= 500 {
				// Rate limit or server error - retry
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				shouldRetry = true
				if resp.Body != nil {
					resp.Body.Close()
				}
			} else {
				// 4xx client error (except 429) - don't retry
				return resp, nil
			}
		}

		// Last attempt or shouldn't retry - return error
		if !shouldRetry || attempt == maxAttempts-1 {
			if resp != nil {
				return resp, lastErr
			}
			return nil, lastErr
		}
	}

	return nil, lastErr
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
	sort.Slice(stores, func(i, j int) bool {
		return stores[i] < stores[j]
	})

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

	// Step 3: Filter catalogs before concurrent fetch
	query := strings.ToLower(req.Query)

	// Build exclude map for fast lookup
	excludeMap := make(map[string]bool)
	for _, store := range req.ExcludeStores {
		excludeMap[strings.ToLower(store)] = true
	}

	// Filter to relevant catalogs
	var relevantCatalogs []catalogResponse
	for _, catalog := range catalogs {
		if !isGroceryStore(catalog.Branding.Name) {
			continue
		}
		if excludeMap[strings.ToLower(catalog.Branding.Name)] {
			continue
		}
		relevantCatalogs = append(relevantCatalogs, catalog)
	}

	// Step 4: Fetch offers concurrently (max 5 goroutines)
	type catalogOffers struct {
		catalog catalogResponse
		offers  []domain.TjekOffer
	}

	resultsCh := make(chan catalogOffers, len(relevantCatalogs))
	sem := make(chan struct{}, 5) // Limit to 5 concurrent requests
	var wg sync.WaitGroup

	for _, catalog := range relevantCatalogs {
		wg.Add(1)
		go func(cat catalogResponse) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Get store address
			store := storesByDealer[cat.DealerID]

			// Fetch offers
			offers, err := s.getCatalogOffers(cat, store)
			if err != nil {
				return
			}

			resultsCh <- catalogOffers{catalog: cat, offers: offers}
		}(catalog)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Step 5: Collect and filter results
	var allOffers []domain.TjekOffer
	for result := range resultsCh {
		for _, offer := range result.offers {
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

	// Step 3: Filter catalogs before concurrent fetch
	excludeMap := make(map[string]bool)
	for _, store := range excludeStores {
		excludeMap[strings.ToLower(store)] = true
	}

	// Filter to relevant catalogs
	var relevantCatalogs []catalogResponse
	for _, catalog := range catalogs {
		if !isGroceryStore(catalog.Branding.Name) {
			continue
		}
		if excludeMap[strings.ToLower(catalog.Branding.Name)] {
			continue
		}
		relevantCatalogs = append(relevantCatalogs, catalog)
	}

	// Step 4: Fetch offers concurrently (max 5 goroutines)
	type catalogOffers struct {
		catalog catalogResponse
		offers  []domain.TjekOffer
	}

	resultsCh := make(chan catalogOffers, len(relevantCatalogs))
	sem := make(chan struct{}, 5) // Limit to 5 concurrent requests
	var wg sync.WaitGroup

	for _, catalog := range relevantCatalogs {
		wg.Add(1)
		go func(cat catalogResponse) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			store := storesByDealer[cat.DealerID]
			offers, err := s.getCatalogOffers(cat, store)
			if err != nil {
				return
			}

			resultsCh <- catalogOffers{catalog: cat, offers: offers}
		}(catalog)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Step 5: Collect offers with discounts
	type offerWithDiscount struct {
		offer    domain.TjekOffer
		discount float64
	}
	var discountedOffers []offerWithDiscount

	for result := range resultsCh {
		for _, offer := range result.offers {
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
	sort.Slice(discountedOffers, func(i, j int) bool {
		return discountedOffers[i].discount > discountedOffers[j].discount
	})

	// Extract sorted offers (top 50)
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
	// Check cache first
	cacheKey := fmt.Sprintf("catalogs:%f:%f:%d", lat, lng, radius)
	if cached, ok := s.getFromCache(cacheKey); ok {
		return cached.([]catalogResponse), nil
	}

	endpoint := fmt.Sprintf("%s/catalogs", s.baseURL)
	params := url.Values{}
	params.Set("r_lat", fmt.Sprintf("%f", lat))
	params.Set("r_lng", fmt.Sprintf("%f", lng))
	params.Set("r_radius", fmt.Sprintf("%d", radius))
	params.Set("types", "paged")

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := s.doWithRetry(func() (*http.Response, error) {
		return s.httpClient.Get(fullURL)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalogs API returned status %d", resp.StatusCode)
	}

	var catalogs []catalogResponse
	if err := json.NewDecoder(resp.Body).Decode(&catalogs); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	// Cache for 1 hour
	s.setCache(cacheKey, catalogs, time.Hour)

	return catalogs, nil
}

// getCatalogOffers fetches all offers from a specific catalog
func (s *TjekService) getCatalogOffers(catalog catalogResponse, store storeResponse) ([]domain.TjekOffer, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("hotspots:%s", catalog.ID)
	if cached, ok := s.getFromCache(cacheKey); ok {
		return cached.([]domain.TjekOffer), nil
	}

	endpoint := fmt.Sprintf("%s/catalogs/%s/hotspots", s.baseURL, catalog.ID)

	resp, err := s.doWithRetry(func() (*http.Response, error) {
		return s.httpClient.Get(endpoint)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hotspots API returned status %d", resp.StatusCode)
	}

	var hotspots []hotspotResponse
	if err := json.NewDecoder(resp.Body).Decode(&hotspots); err != nil {
		sentry.CaptureException(err)
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

	// Cache for 1 hour
	s.setCache(cacheKey, offers, time.Hour)

	return offers, nil
}

// getStores fetches physical store locations near a location
func (s *TjekService) getStores(lat, lng float64, radius int) ([]storeResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("stores:%f:%f:%d", lat, lng, radius)
	if cached, ok := s.getFromCache(cacheKey); ok {
		return cached.([]storeResponse), nil
	}

	endpoint := fmt.Sprintf("%s/stores", s.baseURL)
	params := url.Values{}
	params.Set("r_lat", fmt.Sprintf("%f", lat))
	params.Set("r_lng", fmt.Sprintf("%f", lng))
	params.Set("r_radius", fmt.Sprintf("%d", radius))

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	resp, err := s.doWithRetry(func() (*http.Response, error) {
		return s.httpClient.Get(fullURL)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stores API returned status %d", resp.StatusCode)
	}

	var stores []storeResponse
	if err := json.NewDecoder(resp.Body).Decode(&stores); err != nil {
		sentry.CaptureException(err)
		return nil, err
	}

	// Cache for 1 hour
	s.setCache(cacheKey, stores, time.Hour)

	return stores, nil
}
