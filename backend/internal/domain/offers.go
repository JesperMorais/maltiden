package domain

import "time"

// TjekOffer represents a grocery offer from the Tjek API
type TjekOffer struct {
	ID           string    `json:"id"`
	Heading      string    `json:"heading"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	PrePrice     float64   `json:"prePrice,omitempty"`
	Currency     string    `json:"currency"`
	ValidFrom    time.Time `json:"validFrom"`
	ValidTo      time.Time `json:"validTo"`
	StoreName    string    `json:"storeName"`
	StoreLogo    string    `json:"storeLogo,omitempty"`
	StoreAddress string    `json:"storeAddress,omitempty"`
	StoreCity    string    `json:"storeCity,omitempty"`
	ImageURL     string    `json:"imageUrl,omitempty"`
}

// OfferSearchRequest represents search parameters
type OfferSearchRequest struct {
	Query         string   `json:"query"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	Radius        int      `json:"radius"` // meters
	ExcludeStores []string `json:"excludeStores,omitempty"`
}

// OfferSearchResponse wraps the API response
type OfferSearchResponse struct {
	Offers []TjekOffer `json:"offers"`
	Count  int         `json:"count"`
}
