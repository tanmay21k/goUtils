package stores

import (
	"context"
	"errors"
)

// Value is the value stored for a key. It is currently represented as a string.
type Value string

var ErrUnsupportedOperation = errors.New("store does not support this operation")

type Store interface {
	Get(ctx context.Context, key string) (Value, error)
	Set(ctx context.Context, key string, value Value) error
	Del(ctx context.Context, key string) error
}

// Exister reports whether a key is present in a store.
type Exister interface {
	Exists(ctx context.Context, key string) (bool, error)
}

// KeyStore provides operations for managing individual keys.
type KeyStore interface {
	Rename(ctx context.Context, oldKey, newKey string) error
	Pop(ctx context.Context, key string) (Value, error)
}

// KeyLister returns all keys in alphabetical order.
type KeyLister interface {
	Keys() []string
}

// Scanner visits keys matching a glob pattern in alphabetical order.
type Scanner interface {
	Scan(ctx context.Context, pattern string, fn func(key string) error) error
}

// ConditionalStore performs writes only when the key's current state permits it.
type ConditionalStore interface {
	SetNX(ctx context.Context, key string, value Value) (bool, error)
	SetXX(ctx context.Context, key string, value Value) (bool, error)
}
