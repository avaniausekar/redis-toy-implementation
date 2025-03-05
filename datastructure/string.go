package datastructures

import (
	"sync"
	"time"
	"fmt"
)

// StringStore -> string-specific operations
type StringStore struct {
	data     map[string]stringEntry
	mu       sync.RWMutex
}

// stringEntry -> string value with metadata
type stringEntry struct {
	value     string
	createdAt time.Time
	expiresAt time.Time
}

// NewStringStore creates a new string store
func NewStringStore() *StringStore {
	store := &StringStore{
		data: make(map[string]stringEntry),
	}
	
	// Start background cleanup
	go store.backgroundCleanup()
	
	return store
}

// Set adds or updates a string value with optional expiration
func (s *StringStore) Set(key string, value string, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	entry := stringEntry{
		value:     value,
		createdAt: time.Now(),
		expiresAt: time.Now().Add(expiration),
	}
	
	s.data[key] = entry
	return nil
}

// Get retrieves a string value
func (s *StringStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	entry, exists := s.data[key]
	if !exists {
		return "", false
	}
	
	// Check for expiration
	if time.Now().After(entry.expiresAt) {
		return "", false
	}
	
	return entry.value, true
}

// Delete removes a key
func (s *StringStore) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		return true
	}
	return false
}

// Increment increments a numeric string value
func (s *StringStore) Increment(key string, delta int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	entry, exists := s.data[key]
	if !exists {
		// If key doesn't exist, start from 0
		newValue := delta
		s.data[key] = stringEntry{
			value:     fmt.Sprintf("%d", newValue),
			createdAt: time.Now(),
			expiresAt: time.Time{}, // No expiration
		}
		return newValue, nil
	}
	
	// Parse existing value
	var currentValue int
	_, err := fmt.Sscanf(entry.value, "%d", &currentValue)
	if err != nil {
		return 0, fmt.Errorf("cannot increment non-numeric value")
	}
	
	// Calculate new value
	newValue := currentValue + delta
	
	// Update store
	s.data[key] = stringEntry{
		value:     fmt.Sprintf("%d", newValue),
		createdAt: entry.createdAt,
		expiresAt: entry.expiresAt,
	}
	
	return newValue, nil
}

// backgroundCleanup removes expired entries
func (s *StringStore) backgroundCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		
		for key, entry := range s.data {
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				delete(s.data, key)
			}
		}
		
		s.mu.Unlock()
	}
}

// Keys returns all current keys
func (s *StringStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}