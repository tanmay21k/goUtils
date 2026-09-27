package kv

import (
	"fmt"
	"sort"

	"github.com/tanmay21k/goUtils/internal/helpers"
	"github.com/tanmay21k/goUtils/internal/stores"
)

const DefaultLimit = 5

type store struct {
	data         map[string]string
	maxValidSize int
}

func NewStore(limit ...int) (stores.Store, error) {
	if len(limit) == 0 {
		return &store{
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

	return &store{
		data:         make(map[string]string, limit[0]),
		maxValidSize: limit[0],
	}, nil
}

func (s *store) Get(key string) (string, error) {
	value, ok := s.data[key]
	if !ok {
		return "", helpers.ErrKeyDoesNotExist
	}

	return value, nil
}

func (s *store) Set(key, value string) error {
	if key == "" {
		return helpers.ErrEmptyStringNotAllow
	}

	if _, exists := s.data[key]; !exists &&
		s.maxValidSize <= len(s.data) {
		return helpers.ErrSizeExceed
	}

	s.data[key] = value
	return nil
}

func (s *store) Del(key string) {
	delete(s.data, key)
}

func (s *store) Keys() []string {
	keys := make([]string, 0, len(s.data))

	for key := range s.data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (s *store) Rename(old, new string) error {
	value, err := s.Get(old)
	if err != nil {
		return err
	}

	s.Del(old)

	return s.Set(new, value)
}

func (s *store) Pop(key string) (string, error) {
	value, err := s.Get(key)
	if err != nil {
		return "", err
	}

	s.Del(key)

	return value, nil
}
