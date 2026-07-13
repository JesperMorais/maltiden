package services

import (
	"testing"
	"time"
)

func TestParseTjekTime_NonStandardTimezone(t *testing.T) {
	got := parseTjekTime("2025-07-01T09:00:00+0000")
	want := time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("parseTjekTime() = %v, want %v", got, want)
	}
}

func TestParseTjekTime_RFC3339(t *testing.T) {
	got := parseTjekTime("2025-07-01T09:00:00Z")
	want := time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("parseTjekTime() = %v, want %v", got, want)
	}
}

func TestParseTjekTime_InvalidInput(t *testing.T) {
	got := parseTjekTime("not-a-time")
	if !got.IsZero() {
		t.Errorf("parseTjekTime() = %v, want zero time", got)
	}
}
