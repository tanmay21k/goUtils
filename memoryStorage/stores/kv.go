package stores

import (
	"context"
	"fmt"
	"path"
	"sort"
	"sync"

	"github.com/tanmay21k/goUtils/memoryStorage/helpers"
)

const DefaultLimit = 5

type kvStore struct {
	mu           sync.RWMutex
	data         map[string]string
	maxValidSize int
}

var _ Store = (*kvStore)(nil)
var _ Exister = (*kvStore)(nil)
var _ KeyStore = (*kvStore)(nil)
var _ KeyLister = (*kvStore)(nil)
var _ Scanner = (*kvStore)(nil)
var _ ConditionalStore = (*kvStore)(nil)

func NewkvStore(limit ...int) (*kvStore, error) {
	if len(limit) == 0 {
		return &kvStore{
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

	return &kvStore{
		data:         make(map[string]string, limit[0]),
		maxValidSize: limit[0],
	}, nil
}

func (s *kvStore) Get(ctx context.Context, key string) (Value, error) {
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

func (s *kvStore) Set(ctx context.Context, key string, value Value) error {
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

func (s *kvStore) set(key string, value Value) error {
	if _, exists := s.data[key]; !exists &&
		s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = string(value)
	return nil
}

func (s *kvStore) Del(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *kvStore) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))

	for key := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (s *kvStore) Rename(ctx context.Context, old, new string) error {
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

func (s *kvStore) Pop(ctx context.Context, key string) (Value, error) {
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

func (s *kvStore) Exists(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists, nil
}

func (s *kvStore) Scan(ctx context.Context, pattern string, fn func(string) error) error {
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

func (s *kvStore) SetNX(ctx context.Context, key string, value Value) (bool, error) {
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

func (s *kvStore) SetXX(ctx context.Context, key string, value Value) (bool, error) {
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
