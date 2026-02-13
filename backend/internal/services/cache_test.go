package services

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_HitAndMiss(t *testing.T) {
	svc := NewTjekService()

	// Miss: key not present
	_, ok := svc.getFromCache("missing")
	if ok {
		t.Error("expected cache miss for missing key")
	}

	// Set and hit
	svc.setCache("key1", "value1", time.Hour)
	val, ok := svc.getFromCache("key1")
	if !ok {
		t.Error("expected cache hit for key1")
	}
	if val.(string) != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	hits, misses := svc.CacheStats()
	if hits != 1 {
		t.Errorf("expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}

func TestCache_Expiry(t *testing.T) {
	svc := NewTjekService()

	// Set with very short TTL
	svc.setCache("expiring", "data", time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, ok := svc.getFromCache("expiring")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestCache_MaxEntries_Eviction(t *testing.T) {
	svc := NewTjekService()

	// Fill cache to max
	for i := 0; i < maxCacheEntries; i++ {
		svc.setCache(fmt.Sprintf("key%d", i), i, time.Hour)
	}

	// Adding one more should evict one
	svc.setCache("overflow", "data", time.Hour)

	svc.cacheMu.Lock()
	size := len(svc.cache)
	svc.cacheMu.Unlock()

	if size > maxCacheEntries {
		t.Errorf("cache exceeded max entries: %d > %d", size, maxCacheEntries)
	}

	// The new entry should be present
	val, ok := svc.getFromCache("overflow")
	if !ok {
		t.Error("expected overflow entry to be present")
	}
	if val.(string) != "data" {
		t.Errorf("expected 'data', got %v", val)
	}
}

func TestCache_ExpiredEviction(t *testing.T) {
	svc := NewTjekService()

	// Fill cache with mostly expired entries
	for i := 0; i < maxCacheEntries; i++ {
		svc.setCache(fmt.Sprintf("expired%d", i), i, time.Millisecond)
	}
	time.Sleep(5 * time.Millisecond)

	// This should evict all expired entries first
	svc.setCache("fresh", "data", time.Hour)

	svc.cacheMu.Lock()
	size := len(svc.cache)
	svc.cacheMu.Unlock()

	// After evicting expired entries, should be just 1
	if size != 1 {
		t.Errorf("expected 1 entry after expired eviction, got %d", size)
	}
}
