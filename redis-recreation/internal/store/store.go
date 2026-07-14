package store

import (
	"context"
	"sync"
	"time"
)

type Entry struct {
	data       string
	expiration time.Time
}

type Store struct {
	mu       sync.RWMutex
	enteries map[string]Entry
}

const DEFAULT_CACHE_CAPACITY = 10

func New() *Store {
	return &Store{
		enteries: make(map[string]Entry, DEFAULT_CACHE_CAPACITY),
	}
}

/*
* The GC goroutine removes the expired enteries
* from the store
**/
func (s *Store) StartStoreGC(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.clearExpired()

			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *Store) clearExpired() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.enteries {
		if !entry.expiration.IsZero() && now.After(entry.expiration) {
			delete(s.enteries, key)
		}
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enteries[key] = Entry{
		data:       value,
		expiration: time.Time{},
	}
}

func (s *Store) Get(key string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.enteries[key]
	return value, exists
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.enteries, key)
}

func (s *Store) SetWithExpiration(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enteries[key] = Entry{
		data:       value,
		expiration: time.Now().Add(ttl),
	}
}
