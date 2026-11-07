package pokecache

import (
	"testing"
)

func TestAddGet(t *testing.T) {
	cache := NewCache()

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
	cache := NewCache()

	_, exists := cache.Get("nonexistent")
	if exists {
		t.Errorf("expected key not to exist")
	}
}

func TestReassign(t *testing.T) {
	cache := NewCache()

	key := "https://example.com"
	cache.Add(key, []byte("first"))
	cache.Add(key, []byte("second"))

	got, exists := cache.Get(key)
	if !exists {
		t.Errorf("expected to find key %s", key)
	}
	if string(got) != "second" {
		t.Errorf("expected 'second', got %s", string(got))
	}
}
