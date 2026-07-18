package store

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	s := New()

	s.Set("key1", "value1")

	got, ok := s.Get("key1")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got.Data != "value1" {
		t.Fatalf("expected value1 got %s", got.Data)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	s := New()

	got, ok := s.Get("missing")
	if ok {
		t.Fatal("expected key to not exist")
	}
	if got.Data != "" {
		t.Fatalf("expected empty data got %s", got.Data)
	}
}

func TestSetOverwritesExistingKey(t *testing.T) {
	s := New()

	s.Set("key", "first")
	s.Set("key", "second")

	got, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got.Data != "second" {
		t.Fatalf("expected second got %s", got.Data)
	}
}

func TestDeleteExistingKey(t *testing.T) {
	s := New()

	s.Set("key", "value")
	s.Delete("key")

	_, ok := s.Get("key")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	s := New()

	s.Delete("ghost")

	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected key to not exist")
	}
}

func TestSetWithExpirationNotExpired(t *testing.T) {
	s := New()

	s.SetWithExpiration("ttl", "val", 2*time.Second)

	got, ok := s.Get("ttl")
	if !ok {
		t.Fatal("expected key to exist before expiration")
	}
	if got.Data != "val" {
		t.Fatalf("expected val got %s", got.Data)
	}
}

func TestSetWithExpirationExpired(t *testing.T) {
	s := New()

	s.SetWithExpiration("ttl", "val", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	s.clearExpired()

	_, ok := s.Get("ttl")
	if ok {
		t.Fatal("expected key to be removed after expiration")
	}
}

func TestSetWithoutExpirationNoExpiry(t *testing.T) {
	s := New()

	s.Set("key", "val")

	got, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if !got.expiration.IsZero() {
		t.Fatal("expected no expiration on plain Set")
	}
}

func TestSetWithExpirationHasExpiry(t *testing.T) {
	s := New()

	s.SetWithExpiration("key", "val", 5*time.Second)

	got, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got.expiration.IsZero() {
		t.Fatal("expected expiration to be set")
	}
}

func TestClearExpiredOnlyRemovesExpired(t *testing.T) {
	s := New()

	s.SetWithExpiration("expired", "gone", 1*time.Millisecond)
	s.Set("alive", "here")
	time.Sleep(10 * time.Millisecond)

	s.clearExpired()

	_, ok1 := s.Get("expired")
	if ok1 {
		t.Fatal("expected expired key to be removed")
	}

	got, ok2 := s.Get("alive")
	if !ok2 {
		t.Fatal("expected alive key to still exist")
	}
	if got.Data != "here" {
		t.Fatalf("expected here got %s", got.Data)
	}
}

func TestConcurrentSetAndGet(t *testing.T) {
	s := New()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key"
			s.Set(key, "value")
			s.Get(key)
			s.Delete(key)
		}(i)
	}
	wg.Wait()
}
