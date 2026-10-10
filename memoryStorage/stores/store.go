package stores

import (
	"context"
	"errors"
)

// Value is the value stored for a key. It is currently represented as a string.
type Value string

// ErrUnsupportedOperation is returned when a requested optional operation is
// not supported by a store implementation.
var ErrUnsupportedOperation = errors.New("store does not support this operation")

// Store is the core interface implemented by key/value stores.
type Store interface {
	// Get returns the value for key, or an error if the key does not exist.
	Get(ctx context.Context, key string) (Value, error)
	// Set adds or replaces key with value.
	Set(ctx context.Context, key string, value Value) error
	// Del removes key. Deleting a missing key is a no-op.
	Del(ctx context.Context, key string) error
}

// Exister reports whether a key is present in a store.
type Exister interface {
	// Exists reports whether key is present.
	Exists(ctx context.Context, key string) (bool, error)
}

// KeyStore provides operations for managing individual keys.
type KeyStore interface {
	// Rename changes oldKey to newKey.
	Rename(ctx context.Context, oldKey, newKey string) error
	// Pop returns the value for key and removes it from the store.
	Pop(ctx context.Context, key string) (Value, error)
}

// KeyLister returns all keys in alphabetical order.
type KeyLister interface {
	// Keys returns all current keys in alphabetical order.
	Keys() []string
}

// Scanner visits keys matching a glob pattern in alphabetical order.
type Scanner interface {
	// Scan visits keys matching pattern in alphabetical order and stops on error.
	Scan(ctx context.Context, pattern string, fn func(key string) error) error
}

// ConditionalStore performs writes only when the key's current state permits it.
type ConditionalStore interface {
	// SetNX writes value only if key does not exist; the bool reports success.
	SetNX(ctx context.Context, key string, value Value) (bool, error)
	// SetXX writes value only if key exists; the bool reports whether it existed.
	SetXX(ctx context.Context, key string, value Value) (bool, error)
}
