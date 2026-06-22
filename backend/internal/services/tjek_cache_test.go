package services

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTjekCache_StatsAccumulate(t *testing.T) {
	svc := NewTjekService()

	svc.setCache("k1", "v1", time.Hour)
	svc.setCache("k2", "v2", time.Hour)

	svc.getFromCache("k1")
	svc.getFromCache("k2")
	svc.getFromCache("missing1")
	svc.getFromCache("missing2")

	hits, misses := svc.CacheStats()
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
	if misses != 2 {
		t.Errorf("expected 2 misses, got %d", misses)
	}
}

func TestTjekCache_OverwriteEntry(t *testing.T) {
	svc := NewTjekService()

	svc.setCache("key", "first", time.Hour)
	svc.setCache("key", "second", time.Hour)

	val, ok := svc.getFromCache("key")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val.(string) != "second" {
		t.Errorf("expected 'second', got %v", val)
	}
}

func TestTjekCache_ExpiredEntryCountsAsMiss(t *testing.T) {
	svc := NewTjekService()

	svc.setCache("short", "data", time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	_, ok := svc.getFromCache("short")
	if ok {
		t.Error("expected miss for expired entry")
	}

	_, misses := svc.CacheStats()
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}

func TestTjekCache_ConcurrentAccess(t *testing.T) {
	svc := NewTjekService()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc.setCache(fmt.Sprintf("concurrent%d", i), i, time.Hour)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			svc.getFromCache(fmt.Sprintf("concurrent%d", i))
		}(i)
	}

	wg.Wait()

	svc.cacheMu.Lock()
	size := len(svc.cache)
	svc.cacheMu.Unlock()

	if size > maxCacheEntries {
		t.Errorf("cache size %d exceeds max %d", size, maxCacheEntries)
	}
}
