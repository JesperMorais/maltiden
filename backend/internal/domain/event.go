package domain

import "time"

// Allow-listed event types for session_events. Track() rejects anything else.
const (
	EventMenuGenerated      = "menu_generated"
	EventMenuSaved          = "menu_saved"
	EventShoppingListViewed = "shopping_list_viewed"
	EventRecipeParsed       = "recipe_parsed"
)

// Event is a no-PII analytics record. HouseholdIDHash must be a hash, never
// the raw household ID, and Props must not contain free text.
type Event struct {
	ID              string    `json:"id"`
	EventType       string    `json:"eventType"`
	HouseholdIDHash string    `json:"householdIdHash"`
	OccurredAt      time.Time `json:"occurredAt"`
	RequestID       string    `json:"requestId,omitempty"`
	Props           string    `json:"props,omitempty"`
}
