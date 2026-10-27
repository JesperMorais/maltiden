// Package aibudget caps how many paid LLM requests the app makes per day.
//
// The per-IP rate limiters in pkg/middleware smooth bursts but do not bound
// spend: 2 req/s is still ~170k requests a day. A Budget adds a hard daily
// ceiling, both per key (a household) and globally, that resets at 00:00 UTC.
//
// Counters live in memory. On Fly the machine can auto-stop when idle, which
// resets them early; that only ever grants a fresh day's quota after a quiet
// period, so the cap still holds under sustained abuse.
package aibudget

import (
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
)

// Budget is a fixed-window daily counter. The zero value is not usable; build
// one with New or FromEnv. Safe for concurrent use.
type Budget struct {
	perKey int
	global int
	now    func() time.Time

	mu     sync.Mutex
	day    string
	total  int
	counts map[string]int
}

// New returns a Budget allowing perKey requests per key and global requests in
// total per UTC day. A limit of 0 denies every request (feature off).
func New(perKey, global int) *Budget {
	return &Budget{
		perKey: perKey,
		global: global,
		now:    time.Now,
		counts: make(map[string]int),
	}
}

// FromEnv builds a Budget from <prefix>_DAILY_PER_HOUSEHOLD and
// <prefix>_DAILY_GLOBAL. Unset, non-numeric or negative values fall back to
// the given defaults, so a typo can never lift the cap.
func FromEnv(prefix string, defPerKey, defGlobal int) *Budget {
	return New(
		envInt(prefix+"_DAILY_PER_HOUSEHOLD", defPerKey),
		envInt(prefix+"_DAILY_GLOBAL", defGlobal),
	)
}

func envInt(name string, def int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		slog.Warn("invalid AI budget value, using default", "var", name, "value", raw, "default", def)
		return def
	}
	return v
}

// Allow reports whether key may make one more request today and, if so,
// consumes it. Denied requests consume nothing.
func (b *Budget) Allow(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	today := b.now().UTC().Format("2006-01-02")
	if today != b.day {
		b.day = today
		b.total = 0
		b.counts = make(map[string]int)
	}

	if b.total >= b.global || b.counts[key] >= b.perKey {
		return false
	}
	b.total++
	b.counts[key]++
	return true
}

// Limits returns the configured per-key and global daily limits.
func (b *Budget) Limits() (perKey, global int) {
	return b.perKey, b.global
}
