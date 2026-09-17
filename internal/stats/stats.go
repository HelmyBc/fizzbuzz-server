// Package stats tracks how many times each distinct fizz-buzz request has
// been made, and can report the most frequent one.
package stats

import "sync"

// Entry is a snapshot of the most frequent request and its hit count.
type Entry struct {
	Key   string
	Count int64
}

// Store is a thread-safe in-memory counter keyed by request signature.
//
// Implementation note: Most() scans the whole map, which is O(n) in the
// number of distinct requests seen so far.
//
// A RWMutex is used so concurrent /statistics reads don't block each other
// only writes (Increment) need exclusive access.
type Store struct {
	mu      sync.RWMutex
	records map[string]int64
}

// New creates an empty stats store ready for use.
func New() *Store {
	return &Store{
		records: make(map[string]int64),
	}
}

// Increment records one hit for a given key.
// It uses a RWMutex to ensure thread safety.
func (s *Store) Increment(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key]++
}

// Most returns the request key with the highest hit count.
// ok is false if the store is empty (no requests have been recorded yet).
// Note: In case of a tie, the winner is arbitrary.
func (s *Store) Most() (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.records) == 0 {
		return Entry{}, false
	}

	var best Entry
	first := true // We haven't examined the first map entry yet
	for key, count := range s.records {
		if first || count > best.Count {
			best = Entry{key, count}
			first = false
		}
	}
	return best, true
}
