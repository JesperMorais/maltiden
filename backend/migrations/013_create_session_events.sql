CREATE TABLE IF NOT EXISTS session_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    household_id_hash TEXT NOT NULL,
    occurred_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    request_id TEXT,
    props TEXT
);
CREATE INDEX IF NOT EXISTS idx_session_events_occurred_at ON session_events(occurred_at);
CREATE INDEX IF NOT EXISTS idx_session_events_event_type ON session_events(event_type);
