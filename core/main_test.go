package core

import (
	"errors"
	"testing"
)

// CreateStore is a helper function that creates a
// key-value store for testing
func CreateStore(t *testing.T) Store {
	t.Helper() // Marks the function as a test helper
	// This code should be changed depending on the method
	// that the store is implemented in the code
	return make(map[string]string)
}

// TestPut uses table-driven testing to test a variety
// of test cases for the Store's put functionality
func TestPut(t *testing.T) {
	testCases := []struct {
		name        string // Test case name
		key         string // Keys to write to
		value       string // Values to associate with key
		attempts    int    // Number of put attempts to perform
		expectedErr error  // Error expected from PUT attempt
	}{
		{
			// SinglePut attempts to add a single key-value pair
			// once without error
			name:        "SinglePut",
			key:         "key1",
			value:       "value1",
			attempts:    1,
			expectedErr: nil,
		},
		{
			// MultiPut attempts to add a single key-value pair
			// 10 times without error
			name:        "MultiPut",
			key:         "key1",
			value:       "value1",
			attempts:    10,
			expectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt the put operation as many
			// times as specified
			for i := 0; i < tc.attempts; i++ {
				err := Put(tc.key, tc.value)
				// Check if we expected an error
				if tc.expectedErr != nil {
					// Check that we didn't get nil instead of
					// the expected error
					if err == nil {
						t.Fatalf("expected error, instead got nil")
					}
					// Check that we got the expected error
					if !errors.Is(err, tc.expectedErr) {
						t.Fatalf("expected %q, instead got %q", tc.expectedErr, err)
					}
					return
				}
				// Check that we didn't get an error
				if err != nil {
					t.Fatalf("unexpected error: %q", tc.expectedErr)
				}
			}
		})
	}
}

// TestIntegration tests that a user can put, get, delete and then
// not get a value from the store
func TestIntegration(t *testing.T) {
	key := "key1"
	val := "val1"
	// First we store an arbitrary value which should
	// not cause an error
	if putErr := Put(key, val); putErr != nil {
		t.Fatalf("unexpected error while PUTting object: %q", putErr)
	}
	// Next we try to get the stored object without error
	getRes, getErr := Get(key)
	if getErr != nil {
		t.Fatalf("unexpected error while GETting object: %q", getErr)
	}
	if getRes != val {
		t.Fatalf("expected get result of %s, instead got %s", val, getRes)
	}
	// Next we try to delete the object without error
	delErr := Delete(key)
	if delErr != nil {
		t.Fatalf("unexpected error while DELETEing object: %q", delErr)
	}
	// Finally we try to get the deleted object and expect an error
	if _, badGetErr := Get(key); !errors.Is(badGetErr, ErrNoSuchKey) {
		t.Fatalf("expected %q while trying to get deleted object, instead got %q", ErrNoSuchKey, badGetErr)
	}

}
