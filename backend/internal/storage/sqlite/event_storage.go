package sqlite

import (
	"context"
	"database/sql"
	"maltiden/internal/domain"
	"time"
)

type EventStorage struct {
	db *sql.DB
}

func NewEventStorage(db *sql.DB) *EventStorage {
	return &EventStorage{db: db}
}

func (s *EventStorage) Insert(ctx context.Context, event *domain.Event) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO session_events (id, event_type, household_id_hash, occurred_at, request_id, props)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.EventType, event.HouseholdIDHash, event.OccurredAt, event.RequestID, event.Props)
	return err
}
