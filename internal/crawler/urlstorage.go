package crawler

import (
	"sync"
)

type URLStorage struct {
	seen map[string]bool
	lock sync.RWMutex
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		seen: make(map[string]bool),
		lock: sync.RWMutex{},
	}
}

// Returns true if a URL string has already been added.
func (s *URLStorage) Seen(u string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.seen[u]
}

// Adds an URL string to the slice.
func (s *URLStorage) Add(u string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.seen[u] {
		s.seen[u] = true
	}
}

// Iterate over the seen map, applying the provided function f to the iteration's current element.
func (s *URLStorage) Iterate(f func(string)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for u := range s.seen {
		f(u)
	}
}
