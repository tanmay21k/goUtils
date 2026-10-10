package helpers

import (
	"errors"
	"time"
)

// ErrKeyDoesNotExist is returned when a requested key is absent.
var ErrKeyDoesNotExist = errors.New("key does not exist")

// ErrEmptyStringNotAllow is returned when an operation receives an empty key.
var ErrEmptyStringNotAllow = errors.New("empty string not allowed")

// ErrInvalidSize is returned when a constructor receives an unsupported number
// of configuration arguments.
var ErrInvalidSize = errors.New("multiple parameters not allowed")

// ErrSizeExceed is returned when inserting a key would exceed store capacity.
var ErrSizeExceed = errors.New("max size exceeded")

// ErrKeyExpired is returned when an operation encounters an expired key.
var ErrKeyExpired error = errors.New("key expired")

// DefaultLimit is the default maximum number of keys in a store.
const DefaultLimit int = 5

// DefaultExpireTime is the default TTL used by the legacy TTL constructor.
//
// Deprecated: use stores.DefaultTTL and stores.NewDefaultTTLStore.
var DefaultExpireTime time.Duration = 3 * time.Hour
