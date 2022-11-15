package core

import (
	"errors"
	"sync"
)

// Store type is a simple concurrent-safe map
type Store struct {
	sync.RWMutex
	m map[string]string
}

var store = Store{m: make(map[string]string)}

var ErrNoSuchKey = errors.New("no such key")

// Put adds the provided key value pair into the store
func Put(key string, value string) error {
	// Ensure operation is concurrent-safe
	store.Lock()
	defer store.Unlock()
	// Write value to the store
	store.m[key] = value
	return nil
}

// Get returns the value associated with the provided key
// from the store or an error if the key was invalid
func Get(key string) (string, error) {
	// Ensure operation is concurrent-safe
	store.RLock()
	defer store.RUnlock()
	// Attempt to get the value from the store
	value, ok := store.m[key]
	if !ok {
		return "", ErrNoSuchKey
	}
	return value, nil
}

// Delete removes the value associated with the provided key
// and returns an error if the deletion was unsuccessful
func Delete(key string) error {
	// Ensure operation is concurrent-safe
	store.Lock()
	defer store.Unlock()
	delete(store.m, key)
	return nil
}
