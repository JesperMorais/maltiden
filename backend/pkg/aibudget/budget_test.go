package aibudget

import (
	"sync"
	"testing"
	"time"
)

func fixedClock(t *time.Time) func() time.Time {
	return func() time.Time { return *t }
}

func TestAllow_PerKeyCap(t *testing.T) {
	b := New(2, 100)
	for i := 0; i < 2; i++ {
		if !b.Allow("hh_a") {
			t.Fatalf("request %d denied, want allowed", i+1)
		}
	}
	if b.Allow("hh_a") {
		t.Fatal("third request allowed, want denied by per-key cap")
	}
	if !b.Allow("hh_b") {
		t.Fatal("other key denied, want its own quota")
	}
}

func TestAllow_GlobalCap(t *testing.T) {
	b := New(10, 3)
	for _, k := range []string{"a", "b", "c"} {
		if !b.Allow(k) {
			t.Fatalf("key %s denied under global cap", k)
		}
	}
	if b.Allow("d") {
		t.Fatal("fourth request allowed, want denied by global cap")
	}
}

func TestAllow_ZeroDisables(t *testing.T) {
	if New(0, 100).Allow("a") {
		t.Fatal("perKey=0 allowed a request")
	}
	if New(100, 0).Allow("a") {
		t.Fatal("global=0 allowed a request")
	}
}

func TestAllow_ResetsAtUTCMidnight(t *testing.T) {
	now := time.Date(2026, 9, 24, 23, 59, 0, 0, time.UTC)
	b := New(1, 1)
	b.now = fixedClock(&now)

	if !b.Allow("a") {
		t.Fatal("first request denied")
	}
	if b.Allow("a") {
		t.Fatal("second request same day allowed")
	}
	now = now.Add(2 * time.Minute) // 00:01 next day
	if !b.Allow("a") {
		t.Fatal("request after UTC midnight denied, want fresh quota")
	}
}

func TestAllow_ConcurrentNeverExceedsGlobal(t *testing.T) {
	b := New(1000, 50)
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow("same") {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 50 {
		t.Fatalf("allowed %d, want exactly 50", allowed)
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv("TEST_AI_DAILY_PER_HOUSEHOLD", "7")
	t.Setenv("TEST_AI_DAILY_GLOBAL", "not-a-number")
	per, global := FromEnv("TEST_AI", 1, 99).Limits()
	if per != 7 {
		t.Errorf("perKey = %d, want 7 from env", per)
	}
	if global != 99 {
		t.Errorf("global = %d, want default 99 for invalid value", global)
	}

	t.Setenv("TEST_AI_DAILY_GLOBAL", "-5")
	if _, g := FromEnv("TEST_AI", 1, 99).Limits(); g != 99 {
		t.Errorf("negative global = %d, want default 99", g)
	}
}
