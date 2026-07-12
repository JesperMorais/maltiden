package services

import "testing"

func TestParseTjekTime_OffsetNoColon(t *testing.T) {
	got := parseTjekTime("2024-01-15T10:30:00+0000")
	if got.IsZero() {
		t.Error("expected parsed time, got zero value")
	}
}

func TestParseTjekTime_RFC3339(t *testing.T) {
	got := parseTjekTime("2024-01-15T10:30:00Z")
	if got.IsZero() {
		t.Error("expected parsed time, got zero value")
	}
}

func TestParseTjekTime_Invalid(t *testing.T) {
	got := parseTjekTime("not-a-time")
	if !got.IsZero() {
		t.Errorf("expected zero value for invalid input, got %v", got)
	}
}
