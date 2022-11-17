package core

import (
	"errors"
	"os"
	"sync"
)

// Store type is a simple concurrent-safe map
type Store struct {
	sync.RWMutex
	key []byte
	m   map[string][]byte
}

var store = Store{
	m:   make(map[string][]byte),
	key: []byte(os.Getenv("VILE_SECRET_KEY")),
}

var ErrNoSuchKey = errors.New("no such key")

// func encrypt(secret, key string) ([]byte, error) {
// 	// Create a md5 hash from the private key
// 	md5Hash := md5.Sum([]byte(key))
// 	hash := hex.EncodeToString(md5Hash[:])
// 	// Create a cypte using aes
// 	aesBlock, err := aes.NewCipher([]byte(hash))
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Create a Galois Counter MOde instance to create the nonce
// 	gcmInstance, err := cipher.NewGCM(aesBlock)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Create a nonce
// 	nonce := make([]byte, gcmInstance.NonceSize())
// 	_, _ = io.ReadFull(rand.Reader, nonce)
// 	return []byte(gcmInstance.Seal(nonce, nonce, []byte(secret), nil)), nil
// }

// func decrypt(cipheredValue []byte, key string) (string, error) {
// 	// Create a md5 hash from the private key
// 	md5Hash := md5.Sum([]byte(key))
// 	hash := hex.EncodeToString(md5Hash[:])
// 	// Create a cypte using aes
// 	aesBlock, err := aes.NewCipher([]byte(hash))
// 	if err != nil {
// 		return "", err
// 	}
// 	// Create a Galois Counter MOde instance to create the nonce
// 	gcmInstance, err := cipher.NewGCM(aesBlock)
// 	if err != nil {
// 		return "", err
// 	}
// 	// Get the size of the nonce
// 	nonceSize := gcmInstance.NonceSize()
// 	nonce, cipheredText := cipheredValue[:nonceSize], cipheredValue[nonceSize:]
// 	value, err := gcmInstance.Open(nil, nonce, cipheredText, nil)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(value), nil
// }

// Put adds the provided key value pair into the store
func Put(key string, value string) error {
	// Ensure operation is concurrent-safe
	store.Lock()
	defer store.Unlock()
	// Encrypt the data
	// cipheredValue, err := encrypt(value, string(store.key))
	// if err != nil {
	// 	return err
	// }
	// Write value to the store
	store.m[key] = []byte(value)
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
	// value, err := decrypt(chipheredValue, string(store.key))
	// if err != nil {
	// 	return "", err
	// }
	if !ok {
		return "", ErrNoSuchKey
	}
	return string(value), nil
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
