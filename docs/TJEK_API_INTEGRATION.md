# Tjek API Integration Guide

## Overview

This document describes how to integrate the Tjek API (etilbudsavis.dk) to fetch grocery offers from Swedish stores. The goal is to use real-time offer data to optimize meal planning by suggesting recipes based on what's currently on sale.

## API Details

### Base URL
```
https://api.etilbudsavis.dk/v2
```

### Authentication
**No API key required** for the endpoints we use. The public endpoints for catalogs and hotspots are freely accessible.

### Key Endpoints

#### 1. Get Catalogs (Weekly Flyers)
```
GET /catalogs?r_lat={lat}&r_lng={lng}&r_radius={radius}&types=paged
```

Returns weekly advertising flyers from stores near a location.

**Parameters:**
- `r_lat` - Latitude (e.g., 59.1739 for Haninge)
- `r_lng` - Longitude (e.g., 18.1489 for Haninge)
- `r_radius` - Search radius in meters (e.g., 15000 for 15km)
- `types` - Catalog type, use `paged` for traditional flyers

**Response:**
```json
{
  "id": "LiCzFMAh",
  "dealer_id": "abc123",
  "run_from": "2026-01-04T23:00:00+0000",
  "run_till": "2026-01-11T22:59:59+0000",
  "branding": {
    "name": "Coop",
    "logo": "https://..."
  }
}
```

#### 2. Get Catalog Offers (Hotspots)
```
GET /catalogs/{catalog_id}/hotspots
```

Returns individual offers from a specific catalog.

**Response:**
```json
{
  "type": "offer",
  "offer": {
    "id": "abc123",
    "heading": "KYCKLINGFILÉ",
    "pricing": {
      "price": 79.9,
      "pre_price": 99.9,    // Original price (may be null!)
      "currency": "SEK"
    },
    "quantity": {
      "unit": { "symbol": "kg" },
      "size": { "from": 1, "to": 1 },
      "pieces": { "from": 1, "to": 1 }
    },
    "run_from": "2026-01-04T23:00:00+0000",
    "run_till": "2026-01-11T22:59:59+0000"
  }
}
```

#### 3. Get Stores
```
GET /stores?r_lat={lat}&r_lng={lng}&r_radius={radius}
```

Returns physical store locations for address information.

---

## Important Limitations

### 1. No Product Images in Hotspots
The `/catalogs/{id}/hotspots` endpoint does NOT include product images. Images are only available when fetching whole catalog pages, which show the entire flyer page rather than individual items.

### 2. Pre-Price Availability Varies by Store
Different retailers report pricing differently:

| Store | Reports pre_price? | Notes |
|-------|-------------------|-------|
| Willys | Yes (some items) | Can calculate discount % |
| Coop | No | Only shows offer price |
| Stora Coop | No | Only shows offer price |
| ICA Maxi | No | Only shows offer price |
| Lidl | No | Only shows offer price |

**All items in catalogs ARE discounts** - they're weekly offers. We just can't always calculate the percentage.

### 3. Date Format
Tjek uses non-standard timezone format: `+0000` instead of `+00:00`. Parse with multiple format attempts:
```go
formats := []string{
    "2006-01-02T15:04:05-0700",
    "2006-01-02T15:04:05Z0700",
    time.RFC3339,
}
```

### 4. Swedish Product Names
All product names are in Swedish. Search must handle Swedish characters (å, ä, ö) and compound words.

---

## Grocery Store Filter

Not all catalogs are from grocery stores. Filter using store name:

**Include (grocery stores):**
- Willys, Willys Hemma
- ICA (Maxi, Kvantum, Supermarket, Nära)
- Coop, Stora Coop
- Lidl
- Hemköp
- City Gross
- ÖoB (Öob)
- DollarStore
- Nära Dej, Din Mat, Tempo

**Exclude (non-food):**
- Bauhaus, Biltema, Jula, Rusta
- Elgiganten, NetOnNet, MediaMarkt, Power
- Apoteket, Clas Ohlson
- XXL, Stadium, Intersport
- Jysk, Mio

---

## Integration Strategy for Meal Planning

### Concept
When generating a weekly meal plan, the system should:

1. **Fetch current offers** from nearby stores
2. **Match ingredients** from recipes against available offers
3. **Prioritize recipes** that use discounted ingredients
4. **Calculate potential savings** for the shopping list

### Data Flow
```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Recipe Engine  │────▶│  Offer Matcher   │────▶│  Meal Planner   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │                        │
        ▼                        ▼                        ▼
   Recipe DB              Tjek API Cache           Weekly Plan
   - Ingredients          - Current offers         - Optimized meals
   - Quantities           - Store prices           - Shopping list
```

### Matching Algorithm

#### Step 1: Normalize Ingredient Names
```go
// Recipe ingredient: "500g kycklingfilé"
// Offer heading: "KYCKLINGFILÉ" or "Kyckling filé"

func normalizeForMatch(text string) string {
    text = strings.ToLower(text)
    text = removeQuantities(text)  // Remove "500g", "1 kg", etc.
    return text
}
```

#### Step 2: Word-Boundary Matching
Don't use simple substring matching - it causes false positives:
- "ägg" should NOT match "pålägg"
- "bröd" should match "fullkornsbröd"

```go
func matchesIngredient(offerHeading, ingredient string) bool {
    words := splitIntoWords(offerHeading)
    for _, word := range words {
        if strings.HasPrefix(word, ingredient) ||
           strings.HasSuffix(word, ingredient) {
            return true
        }
    }
    return false
}
```

#### Step 3: Category Matching
For flexible matching, map ingredients to categories:

```go
var ingredientCategories = map[string][]string{
    "protein": {"kyckling", "nötkött", "fläsk", "lax", "torsk", "ägg"},
    "dairy":   {"mjölk", "ost", "yoghurt", "grädde", "smör"},
    "carbs":   {"pasta", "ris", "potatis", "bröd", "nudlar"},
    "veggies": {"tomat", "lök", "morot", "paprika", "gurka"},
}
```

### Caching Strategy

Offers change weekly. Implement caching:

```go
type OfferCache struct {
    offers    []TjekOffer
    fetchedAt time.Time
    expiresAt time.Time  // Usually Sunday night when new flyers start
}

func (c *OfferCache) IsValid() bool {
    return time.Now().Before(c.expiresAt)
}
```

**Cache invalidation:** Most Swedish flyers run Monday-Sunday. Refresh cache Sunday evening or Monday morning.

### Example: Recipe Optimization

```go
type RecipeWithOffers struct {
    Recipe        Recipe
    MatchedOffers []OfferMatch
    TotalSavings  float64
    MatchScore    float64  // 0-1, how many ingredients are on offer
}

type OfferMatch struct {
    Ingredient string
    Offer      TjekOffer
    Savings    float64  // pre_price - price (if available)
}

func OptimizeWeeklyPlan(recipes []Recipe, offers []TjekOffer) []RecipeWithOffers {
    var scored []RecipeWithOffers

    for _, recipe := range recipes {
        matches := findOfferMatches(recipe.Ingredients, offers)
        scored = append(scored, RecipeWithOffers{
            Recipe:        recipe,
            MatchedOffers: matches,
            TotalSavings:  calculateSavings(matches),
            MatchScore:    float64(len(matches)) / float64(len(recipe.Ingredients)),
        })
    }

    // Sort by match score (most ingredients on offer first)
    sort.Slice(scored, func(i, j int) bool {
        return scored[i].MatchScore > scored[j].MatchScore
    })

    return scored
}
```

### Shopping List Enhancement

When generating the shopping list, include offer information:

```go
type ShoppingItem struct {
    Ingredient  string
    Quantity    string
    OnOffer     bool
    OfferPrice  float64
    StoreName   string
    ValidUntil  time.Time
}
```

Display example:
```
Inköpslista vecka 2:
─────────────────────────────────────
✓ Kycklingfilé (500g)     79.90 kr  [Willys - ord. 99.90]
✓ Pasta (500g)            15.90 kr  [Willys]
  Lök (3 st)              ~10 kr
✓ Krossade tomater        12.90 kr  [ICA Maxi]
─────────────────────────────────────
Beräknad besparing: ~30 kr
```

---

## Backend Implementation

### Current POC Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/offers/search?q={query}` | Search offers by product name |
| GET | `/offers/discounts` | Get offers with known discounts |
| GET | `/offers/stores` | Get available store names |

### Query Parameters

- `q` - Search query (Swedish product name)
- `lat` - Override latitude (default: 59.1739)
- `lng` - Override longitude (default: 18.1489)
- `radius` - Override radius in meters (default: 15000)
- `exclude` - Comma-separated store names to exclude

### Response Format

```json
{
  "offers": [
    {
      "id": "abc123",
      "heading": "KYCKLINGFILÉ",
      "description": "1 kg",
      "price": 79.9,
      "prePrice": 99.9,
      "currency": "SEK",
      "validFrom": "2026-01-04T23:00:00Z",
      "validTo": "2026-01-11T22:59:59Z",
      "storeName": "Willys",
      "storeLogo": "https://...",
      "storeAddress": "Storgatan 1",
      "storeCity": "Haninge"
    }
  ],
  "count": 1
}
```

---

## Future Improvements

1. **User Location** - Use browser geolocation instead of hardcoded Haninge coordinates

2. **Favorite Stores** - Let users select preferred stores for personalized results

3. **Price History** - Track prices over time to identify real deals vs. fake discounts

4. **Ingredient Synonyms** - Map recipe ingredients to common Swedish product names
   - "chicken breast" → "kycklingfilé", "kycklingbröst"
   - "ground beef" → "nötfärs", "köttfärs"

5. **Quantity Normalization** - Parse offer quantities to compare prices per unit
   - "2 för 30 kr" → 15 kr/st
   - "79.90 kr/kg" vs "39.90 kr/500g"

---

## Drink Packaging Logic

The API returns size in various units (ml, cl, l) but doesn't specify packaging type. We determine can vs bottle based on size:

```go
func determineDrinkPackaging(sizeInMl float64) string {
    switch {
    case sizeInMl <= 250:   return "burk"   // Energy drinks (Red Bull 250ml)
    case sizeInMl <= 355:   return "burk"   // Standard cans (33cl, 35.5cl)
    case sizeInMl <= 500:   return "burk"   // Large cans (Monster 500ml)
    case sizeInMl < 1000:   return "flaska" // Small bottles (750ml)
    default:                return "flaska" // Large bottles (1L, 1.5L, 2L)
    }
}
```

**Examples:**
| Input | Output |
|-------|--------|
| 25cl, 1 piece | 1 burk (25 cl) |
| 33cl, 1 piece | 1 burk (33 cl) |
| 50cl, 1 piece | 1 burk (50 cl) |
| 33cl, 6 pieces | 6-pack burk (33 cl) |
| 150cl, 1 piece | 1 flaska (150 cl) |
| 1.5l, 1 piece | 1 flaska (1.5 l) |
| 2l, 1 piece | 1 flaska (2 l) |

**Note:** 500ml is ambiguous (could be large can or small bottle). We default to "burk" since energy drinks at this size are common in Swedish stores.

6. **Store Preference Learning** - Track which stores user actually shops at

---

## Testing

### Backend
```bash
cd backend
go run cmd/server/main.go

# Test search
curl "http://localhost:8080/offers/search?q=kyckling"

# Test discounts
curl "http://localhost:8080/offers/discounts"

# Test stores
curl "http://localhost:8080/offers/stores"
```

### Direct API Testing
```bash
# Get catalogs near Haninge
curl "https://api.etilbudsavis.dk/v2/catalogs?r_lat=59.1739&r_lng=18.1489&r_radius=15000&types=paged"

# Get offers from a specific catalog
curl "https://api.etilbudsavis.dk/v2/catalogs/{catalog_id}/hotspots"
```

---

## Files Reference

| File | Purpose |
|------|---------|
| `backend/internal/services/tjek_service.go` | Core API integration |
| `backend/internal/domain/offers.go` | Data types |
| `backend/internal/api/handlers/offers.go` | HTTP handlers |
| `frontend/src/api/offers.api.ts` | Frontend API client |
| `frontend/src/components/poc/OfferSearch.vue` | POC search UI |
