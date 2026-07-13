package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"maltiden/internal/domain"
	"maltiden/internal/storage/sqlite"
	"os"
	"time"

	"github.com/google/uuid"
)

// allowedEventTypes is the allow-list for Track(). Anything else is rejected.
var allowedEventTypes = map[string]bool{
	domain.EventMenuGenerated:      true,
	domain.EventMenuSaved:          true,
	domain.EventShoppingListViewed: true,
	domain.EventRecipeParsed:       true,
}

// ErrEventTypeNotAllowed is returned by Track for event types outside the allow-list.
var ErrEventTypeNotAllowed = errors.New("event_type_not_allowed")

type EventService struct {
	storage   *sqlite.EventStorage
	requestID string
	enabled   bool
}

// NewEventService reads EVENTS_ENABLED once at construction. If unset/false,
// Track becomes a permanent no-op that never touches the database.
func NewEventService(storage *sqlite.EventStorage, requestID string) *EventService {
	return &EventService{
		storage:   storage,
		requestID: requestID,
		enabled:   os.Getenv("EVENTS_ENABLED") == "true",
	}
}

// Track records an allow-listed event. householdID is hashed before storage;
// the raw value is never persisted. No-op when the feature flag is off.
func (s *EventService) Track(ctx context.Context, eventType, householdID, props string) error {
	if s == nil || !s.enabled {
		return nil
	}

	if !allowedEventTypes[eventType] {
		return ErrEventTypeNotAllowed
	}

	sum := sha256.Sum256([]byte(householdID))
	event := &domain.Event{
		ID:              "evt_" + uuid.New().String(),
		EventType:       eventType,
		HouseholdIDHash: hex.EncodeToString(sum[:]),
		OccurredAt:      time.Now(),
		RequestID:       s.requestID,
		Props:           props,
	}

	return s.storage.Insert(ctx, event)
}
