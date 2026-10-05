package helpers

import (
	"errors"
	"time"
)

var ErrKeyDoesNotExist error = errors.New("key does not exist")
var ErrEmptyStringNotAllow error = errors.New("empty string not allowed")
var ErrInvalidSize error = errors.New("multiple parameters not allowed")
var ErrSizeExceed error = errors.New("max size exceeded")

var ErrKeyExpired error = errors.New("key expired")

const DefaultLimit int = 5

var DefaultExpireTime time.Duration = 3 * time.Hour
