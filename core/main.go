package core

import "errors"

// store is the most simple key-value store abstraction
// used for the initial stages of development
var store = make(map[string]string)

// Store type is defined for test consistency
type Store map[string]string

var ErrNoSuchKey = errors.New("no such key")

// Put adds the provided key value pair into the store
func Put(key string, value string) error {
	store[key] = value
	return nil
}

// Get returns the value associated with the provided key
// from the store or an error if the key was invalid
func Get(key string) (string, error) {
	// Attempt to get the value from the store
	value, ok := store[key]
	if !ok {
		return "", ErrNoSuchKey
	}
	return value, nil
}

// Delete removes the value associated with the provided key
// and returns an error if the deletion was unsuccessful
func Delete(key string) error {
	delete(store, key)
	return nil
}
