package stores

import (
	"context"
	"fmt"
	"path"
	"sort"
	"sync"

	"github.com/tanmay21k/goUtils/memoryStorage/helpers"
)

// DefaultLimit is the default maximum number of keys for NewKVStore.
const DefaultLimit = 5

// KVStore is a concurrency-safe, non-expiring in-memory key/value store.
type KVStore struct {
	mu           sync.RWMutex
	data         map[string]string
	maxValidSize int
}

var _ Store = (*KVStore)(nil)
var _ Exister = (*KVStore)(nil)
var _ KeyStore = (*KVStore)(nil)
var _ KeyLister = (*KVStore)(nil)
var _ Scanner = (*KVStore)(nil)
var _ ConditionalStore = (*KVStore)(nil)

// NewKVStore creates a non-expiring in-memory store. With no argument it uses
// DefaultLimit; otherwise exactly one capacity greater than 1 is required.
func NewKVStore(limit ...int) (*KVStore, error) {
	if len(limit) == 0 {
		return &KVStore{
			data:         make(map[string]string, DefaultLimit),
			maxValidSize: DefaultLimit,
		}, nil
	}

	if len(limit) > 1 {
		return nil, helpers.ErrInvalidSize
	}

	if limit[0] <= 1 {
		return nil, fmt.Errorf("size must be greater than 1")
	}

	return &KVStore{
		data:         make(map[string]string, limit[0]),
		maxValidSize: limit[0],
	}, nil
}

// NewkvStore is an outdated spelling of NewKVStore.
//
// Deprecated: use NewKVStore.
func NewkvStore(limit ...int) (*KVStore, error) {
	return NewKVStore(limit...)
}

// Get returns the value for key, or an error if the key does not exist.
func (s *KVStore) Get(ctx context.Context, key string) (Value, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return "", helpers.ErrKeyDoesNotExist
	}

	return Value(value), nil
}

// Set adds or replaces key with value, subject to the store's capacity.
func (s *KVStore) Set(ctx context.Context, key string, value Value) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return helpers.ErrEmptyStringNotAllow
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.set(key, value)
}

func (s *KVStore) set(key string, value Value) error {
	if _, exists := s.data[key]; !exists &&
		s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = string(value)
	return nil
}

// Del removes key. Deleting a missing key is a no-op.
func (s *KVStore) Del(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// Keys returns all current keys in alphabetical order.
func (s *KVStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))

	for key := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// Rename changes old to new, replacing new if it already exists.
func (s *KVStore) Rename(ctx context.Context, old, new string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if new == "" {
		return helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[old]
	if !ok {
		return helpers.ErrKeyDoesNotExist
	}
	if old != new {
		delete(s.data, old)
		s.data[new] = value
	}
	return nil
}

// Pop returns the value for key and removes it from the store.
func (s *KVStore) Pop(ctx context.Context, key string) (Value, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[key]
	if !ok {
		return "", helpers.ErrKeyDoesNotExist
	}
	delete(s.data, key)
	return Value(value), nil
}

// Exists reports whether key is present.
func (s *KVStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists, nil
}

// Scan visits keys matching pattern in alphabetical order and stops on error.
func (s *KVStore) Scan(ctx context.Context, pattern string, fn func(string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("scan callback must not be nil")
	}
	if pattern == "" {
		pattern = "*"
	}
	if _, err := path.Match(pattern, ""); err != nil {
		return err
	}
	for _, key := range s.Keys() {
		if err := ctx.Err(); err != nil {
			return err
		}
		matched, err := path.Match(pattern, key)
		if err != nil {
			return err
		}
		if matched {
			if err := fn(key); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetNX writes value only if key does not exist; the bool reports success.
func (s *KVStore) SetNX(ctx context.Context, key string, value Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return false, helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; exists {
		return false, nil
	}
	if err := s.set(key, value); err != nil {
		return false, err
	}
	return true, nil
}

// SetXX writes value only if key exists; the bool reports whether it existed.
func (s *KVStore) SetXX(ctx context.Context, key string, value Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return false, helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.data[key]; !exists {
		return false, nil
	}
	return true, s.set(key, value)
}
