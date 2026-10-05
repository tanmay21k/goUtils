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

type Entry struct {
	Value    Value
	ExpireAt time.Time
}

type ttlStore struct {
	mu           sync.RWMutex
	data         map[string]Entry
	maxValidSize int
	expireTime   time.Duration
}

var _ Store = (*ttlStore)(nil)
var _ Exister = (*ttlStore)(nil)
var _ KeyStore = (*ttlStore)(nil)
var _ KeyLister = (*ttlStore)(nil)
var _ Scanner = (*ttlStore)(nil)
var _ ConditionalStore = (*ttlStore)(nil)

func NewStore(args ...int) (*ttlStore, error) {
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

	return &ttlStore{
		data:         make(map[string]Entry, size),
		maxValidSize: size,
		expireTime:   expireTime,
	}, nil
}

func (s *ttlStore) Get(ctx context.Context, key string) (Value, error) {
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

func (s *ttlStore) Set(ctx context.Context, key string, value Value) error {
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

func (s *ttlStore) set(key string, value Value) error {
	if _, exists := s.data[key]; !exists && s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = Entry{
		Value:    value,
		ExpireAt: time.Now().Add(s.expireTime),
	}

	return nil
}

func (s *ttlStore) Del(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *ttlStore) Keys() []string {
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

func (s *ttlStore) Rename(ctx context.Context, old, new string) error {
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

func (s *ttlStore) Pop(ctx context.Context, key string) (Value, error) {
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

func (s *ttlStore) Exists(ctx context.Context, key string) (bool, error) {
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

func (s *ttlStore) Scan(ctx context.Context, pattern string, fn func(string) error) error {
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

func (s *ttlStore) SetNX(ctx context.Context, key string, value Value) (bool, error) {
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

func (s *ttlStore) SetXX(ctx context.Context, key string, value Value) (bool, error) {
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

func (s *ttlStore) purgeExpired(now time.Time) {
	for key, entry := range s.data {
		if now.After(entry.ExpireAt) {
			delete(s.data, key)
		}
	}
}
