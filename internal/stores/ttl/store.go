package ttl

import (
	"fmt"
	"sort"
	"time"

	"github.com/tanmay21k/goUtils/internal/helpers"
	"github.com/tanmay21k/goUtils/internal/stores"
)

type Entry struct {
	Value    string
	ExpireAt time.Time
}

type store struct {
	data         map[string]Entry
	maxValidSize int
	expireTime   time.Duration
}

func NewStore(args ...int) (stores.Store, error) {
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

	return &store{
		data:         make(map[string]Entry, size),
		maxValidSize: size,
		expireTime:   expireTime,
	}, nil
}

func (s *store) Get(key string) (string, error) {
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

func (s *store) Set(key, value string) error {
	if key == "" {
		return helpers.ErrEmptyStringNotAllow
	}

	if _, exists := s.data[key]; !exists &&
		s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = Entry{
		Value:    value,
		ExpireAt: time.Now().Add(s.expireTime),
	}

	return nil
}

func (s *store) Del(key string) {
	delete(s.data, key)
}

func (s *store) Keys() []string {
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

func (s *store) Rename(old, new string) error {
	if new == "" {
		return helpers.ErrEmptyStringNotAllow
	}

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

func (s *store) Pop(key string) (string, error) {
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
