package store

import (
	"context"
	"log"
	"sync"
	"time"
)

type Entry struct {
	Data       string
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
func (s *Store) StartStoreGC(ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Second)

	wg.Go(func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Println("Running GC...")
				s.clearExpired()

			case <-ctx.Done():
				return
			}
		}
	})
}

func (s *Store) clearExpired() {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	for key, entry := range s.enteries {
		if !entry.expiration.IsZero() && now.After(entry.expiration) {
			log.Printf("Deleting expired entry: %s\n", key)
			delete(s.enteries, key)
		}
	}
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enteries[key] = Entry{
		Data:       value,
		expiration: time.Time{},
	}
}

func (s *Store) Get(key string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.enteries[key]
	return entry, exists
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
		Data:       value,
		expiration: time.Now().Add(ttl),
	}
}
