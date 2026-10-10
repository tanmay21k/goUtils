package stores

import (
	"context"
	"fmt"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/tanmay21k/goUtils/memoryStorage/helpers"
)

// Entry is a value and its expiration timestamp in a TTL store.
type Entry struct {
	// Value is the stored value.
	Value Value
	// ExpireAt is the time after which the entry is expired.
	ExpireAt time.Time
}

// DefaultTTL is the expiration duration used by NewDefaultTTLStore.
const DefaultTTL = 3 * time.Hour

// TTLStore is a concurrency-safe in-memory key/value store with lazy TTL expiration.
type TTLStore struct {
	mu           sync.RWMutex
	data         map[string]Entry
	maxValidSize int
	expireTime   time.Duration
}

var _ Store = (*TTLStore)(nil)
var _ Exister = (*TTLStore)(nil)
var _ KeyStore = (*TTLStore)(nil)
var _ KeyLister = (*TTLStore)(nil)
var _ Scanner = (*TTLStore)(nil)
var _ ConditionalStore = (*TTLStore)(nil)

// NewTTLStore creates a TTL store with the specified capacity and expiration
// duration. Capacity must be greater than 1 and ttl must be positive.
func NewTTLStore(capacity int, ttl time.Duration) (*TTLStore, error) {
	if capacity <= 1 {
		return nil, fmt.Errorf("size must be greater than 1")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("expire time must be greater than 0")
	}
	return newTTLStore(capacity, ttl), nil
}

// NewDefaultTTLStore creates a TTL store with capacity DefaultLimit and
// expiration duration DefaultTTL.
func NewDefaultTTLStore() *TTLStore {
	return newTTLStore(helpers.DefaultLimit, DefaultTTL)
}

// NewStore creates a TTL store using the legacy integer-based constructor.
// With no arguments it uses the defaults; with one argument it sets capacity;
// with two arguments it sets capacity and expiration duration in nanoseconds.
//
// Deprecated: use NewDefaultTTLStore or NewTTLStore.
func NewStore(args ...int) (*TTLStore, error) {
	size := helpers.DefaultLimit
	expireTime := helpers.DefaultExpireTime

	switch len(args) {
	case 0:
		// Use defaults.

	case 1:
		if args[0] <= 1 {
			return nil, fmt.Errorf("size must be greater than 1")
		}

		size = args[0]

	case 2:
		if args[0] <= 1 {
			return nil, fmt.Errorf("size must be greater than 1")
		}

		if args[1] <= 0 {
			return nil, fmt.Errorf("expire time must be greater than 0")
		}

		size = args[0]
		expireTime = time.Duration(args[1])

	default:
		return nil, helpers.ErrInvalidSize
	}

	return newTTLStore(size, expireTime), nil
}

func newTTLStore(size int, expireTime time.Duration) *TTLStore {
	return &TTLStore{
		data:         make(map[string]Entry, size),
		maxValidSize: size,
		expireTime:   expireTime,
	}
}

// Get returns the value for key, or an error if the key is missing or expired.
func (s *TTLStore) Get(ctx context.Context, key string) (Value, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[key]
	if !ok {
		return "", helpers.ErrKeyDoesNotExist
	}

	if time.Now().After(entry.ExpireAt) {
		delete(s.data, key)
		return "", helpers.ErrKeyExpired
	}

	return entry.Value, nil
}

// Set adds or replaces key with value and starts a fresh expiration period.
func (s *TTLStore) Set(ctx context.Context, key string, value Value) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key == "" {
		return helpers.ErrEmptyStringNotAllow
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpired(time.Now())
	return s.set(key, value)
}

func (s *TTLStore) set(key string, value Value) error {
	if _, exists := s.data[key]; !exists && s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = Entry{
		Value:    value,
		ExpireAt: time.Now().Add(s.expireTime),
	}

	return nil
}

// Del removes key. Deleting a missing key is a no-op.
func (s *TTLStore) Del(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

// Keys returns unexpired keys in alphabetical order.
func (s *TTLStore) Keys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	keys := make([]string, 0, len(s.data))

	for key, entry := range s.data {
		if now.After(entry.ExpireAt) {
			delete(s.data, key)
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// Rename changes old to new, replacing new if it exists and preserving expiry.
func (s *TTLStore) Rename(ctx context.Context, old, new string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if new == "" {
		return helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.data[old]
	if !ok {
		return helpers.ErrKeyDoesNotExist
	}

	if time.Now().After(entry.ExpireAt) {
		delete(s.data, old)
		return helpers.ErrKeyExpired
	}

	if old == new {
		return nil
	}

	delete(s.data, old)
	s.data[new] = entry

	return nil
}

// Pop returns the value for key and removes it from the store.
func (s *TTLStore) Pop(ctx context.Context, key string) (Value, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[key]
	if !ok {
		return "", helpers.ErrKeyDoesNotExist
	}

	if time.Now().After(entry.ExpireAt) {
		delete(s.data, key)
		return "", helpers.ErrKeyExpired
	}

	delete(s.data, key)

	return entry.Value, nil
}

// Exists reports whether key is present and unexpired.
func (s *TTLStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.data[key]
	if !exists {
		return false, nil
	}
	if time.Now().After(entry.ExpireAt) {
		delete(s.data, key)
		return false, nil
	}
	return true, nil
}

// Scan visits unexpired keys matching pattern in alphabetical order and stops on error.
func (s *TTLStore) Scan(ctx context.Context, pattern string, fn func(string) error) error {
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

// SetNX writes value only if key does not exist or is expired; the bool reports success.
func (s *TTLStore) SetNX(ctx context.Context, key string, value Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return false, helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.purgeExpired(now)
	if _, exists := s.data[key]; exists {
		return false, nil
	}
	if err := s.set(key, value); err != nil {
		return false, err
	}
	return true, nil
}

// SetXX writes value only if key exists and is unexpired; the bool reports success.
func (s *TTLStore) SetXX(ctx context.Context, key string, value Value) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if key == "" {
		return false, helpers.ErrEmptyStringNotAllow
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	entry, exists := s.data[key]
	if !exists {
		return false, nil
	}
	if now.After(entry.ExpireAt) {
		delete(s.data, key)
		return false, nil
	}
	s.data[key] = Entry{Value: value, ExpireAt: now.Add(s.expireTime)}
	return true, nil
}

func (s *TTLStore) purgeExpired(now time.Time) {
	for key, entry := range s.data {
		if now.After(entry.ExpireAt) {
			delete(s.data, key)
		}
	}
}
