package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	cache := NewCache(time.Minute)

	key := "https://example.com"
	val := []byte("testdata")

	cache.Add(key, val)

	got, exists := cache.Get(key)
	if !exists {
		t.Errorf("expected to find key %s", key)
	}
	if string(got) != string(val) {
		t.Errorf("expected %s, got %s", string(val), string(got))
	}
}

func TestAddGetMissing(t *testing.T) {
	cache := NewCache(time.Minute)

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Errorf("expected key not to exist")
	}
}

func TestReaping(t *testing.T) {
	interval := 5 * time.Millisecond
	cache := NewCache(interval)

	key := "https://example.com"
	cache.Add(key, []byte("testdata"))

	// Allow time for the ticker to tick and reap the item
	time.Sleep(interval + 10*time.Millisecond)

	_, exists := cache.Get(key)
	if exists {
		t.Errorf("expected key to be reaped and not exist")
	}
}
