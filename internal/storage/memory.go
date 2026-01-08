package storage

import "sync"

type listStruct struct {
	key string
	
}

type Entry struct {
	value     string
	expiresAt int64 // unix timestamp in ms, 0 = no expiry
}

type Store struct {
	mu   sync.RWMutex
	data map[string]Entry
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]Entry),
	}
}

func (s *Store) Set(key, value string, expiresAt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = Entry{
		value:     value,
		expiresAt: expiresAt,
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.data[key]
	if !ok {
		return "", false
	}

	return entry.value, true
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}
