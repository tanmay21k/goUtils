package ttl

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/tanmay21k/goUtils/internal/helpers"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name       string
		args       []int
		wantLimit  int
		wantTTL    time.Duration
		wantErr    error
		wantErrMsg string
	}{
		{name: "defaults", wantLimit: helpers.DefaultLimit, wantTTL: helpers.DefaultExpireTime},
		{name: "custom size", args: []int{3}, wantLimit: 3, wantTTL: helpers.DefaultExpireTime},
		{name: "custom size and ttl", args: []int{3, 10}, wantLimit: 3, wantTTL: 10},
		{name: "size too small", args: []int{1}, wantErrMsg: "size must be greater than 1"},
		{name: "ttl not positive", args: []int{2, 0}, wantErrMsg: "expire time must be greater than 0"},
		{name: "too many arguments", args: []int{2, 3, 4}, wantErr: helpers.ErrInvalidSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStore(tt.args...)
			if tt.wantErr != nil || tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("NewStore() error = nil, want %v", tt.wantErrMsg)
				}
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Errorf("NewStore() error = %v, want %v", err, tt.wantErr)
				}
				if tt.wantErrMsg != "" && err.Error() != tt.wantErrMsg {
					t.Errorf("NewStore() error = %v, want %q", err, tt.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewStore() error = %v", err)
			}

			created := got.(*store)
			if created.maxValidSize != tt.wantLimit {
				t.Errorf("maxValidSize = %d, want %d", created.maxValidSize, tt.wantLimit)
			}
			if created.expireTime != tt.wantTTL {
				t.Errorf("expireTime = %v, want %v", created.expireTime, tt.wantTTL)
			}
		})
	}
}

func newTestStore(t *testing.T) *store {
	t.Helper()
	created, err := NewStore(3, int(time.Hour))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return created.(*store)
}

func TestStoreGet(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Get("missing"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Fatalf("Get(missing) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if err := s.Set("key", "value"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Get("key"); err != nil || got != "value" {
		t.Errorf("Get(key) = %q, %v; want %q, nil", got, err, "value")
	}

	s.data["expired"] = Entry{Value: "stale", ExpireAt: time.Now().Add(-time.Second)}
	if got, err := s.Get("expired"); got != "" || !errors.Is(err, helpers.ErrKeyExpired) {
		t.Errorf("Get(expired) = %q, %v; want empty value and %v", got, err, helpers.ErrKeyExpired)
	}
	if _, ok := s.data["expired"]; ok {
		t.Error("Get(expired) should remove the expired entry")
	}
}

func TestStoreSet(t *testing.T) {
	s := newTestStore(t)
	if err := s.Set("", "value"); !errors.Is(err, helpers.ErrEmptyStringNotAllow) {
		t.Fatalf("Set(empty key) error = %v, want %v", err, helpers.ErrEmptyStringNotAllow)
	}
	for _, key := range []string{"first", "second", "third"} {
		if err := s.Set(key, key+" value"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Set("first", "updated"); err != nil {
		t.Fatalf("updating a key at capacity: %v", err)
	}
	if err := s.Set("fourth", "value"); !errors.Is(err, helpers.ErrSizeExceed) {
		t.Errorf("Set() over capacity error = %v, want %v", err, helpers.ErrSizeExceed)
	}
	if got, err := s.Get("first"); err != nil || got != "updated" {
		t.Errorf("Get(first) = %q, %v; want %q, nil", got, err, "updated")
	}
}

func TestStoreDel(t *testing.T) {
	s := newTestStore(t)
	if err := s.Set("key", "value"); err != nil {
		t.Fatal(err)
	}
	s.Del("key")
	if _, err := s.Get("key"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Get(deleted key) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	s.Del("missing")
}

func TestStoreKeys(t *testing.T) {
	s := newTestStore(t)
	for key, value := range map[string]string{"gamma": "3", "alpha": "1", "beta": "2"} {
		if err := s.Set(key, value); err != nil {
			t.Fatal(err)
		}
	}
	s.data["expired"] = Entry{Value: "stale", ExpireAt: time.Now().Add(-time.Second)}

	want := []string{"alpha", "beta", "gamma"}
	if got := s.Keys(); !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
	if _, ok := s.data["expired"]; ok {
		t.Error("Keys() should remove expired entries")
	}
}

func TestStoreRename(t *testing.T) {
	s := newTestStore(t)
	if err := s.Set("old", "value"); err != nil {
		t.Fatal(err)
	}
	originalExpiry := s.data["old"].ExpireAt

	if err := s.Rename("missing", "new"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Fatalf("Rename(missing) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if err := s.Rename("old", ""); !errors.Is(err, helpers.ErrEmptyStringNotAllow) {
		t.Fatalf("Rename(empty destination) error = %v, want %v", err, helpers.ErrEmptyStringNotAllow)
	}
	if err := s.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if got, err := s.Get("new"); err != nil || got != "value" {
		t.Errorf("Get(new) = %q, %v; want %q, nil", got, err, "value")
	}
	if s.data["new"].ExpireAt != originalExpiry {
		t.Error("Rename() should preserve the original expiration time")
	}
	if _, ok := s.data["old"]; ok {
		t.Error("Rename() should remove the old key")
	}

	if err := s.Rename("new", "new"); err != nil {
		t.Errorf("renaming a key to itself: %v", err)
	}
}

func TestStoreRenameExpiredKey(t *testing.T) {
	s := newTestStore(t)
	s.data["expired"] = Entry{Value: "stale", ExpireAt: time.Now().Add(-time.Second)}

	if err := s.Rename("expired", "new"); !errors.Is(err, helpers.ErrKeyExpired) {
		t.Errorf("Rename(expired) error = %v, want %v", err, helpers.ErrKeyExpired)
	}
	if _, ok := s.data["expired"]; ok {
		t.Error("Rename(expired) should remove the expired entry")
	}
}

func TestStorePop(t *testing.T) {
	s := newTestStore(t)
	if err := s.Set("key", "value"); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Pop("key"); err != nil || got != "value" {
		t.Fatalf("Pop(key) = %q, %v; want %q, nil", got, err, "value")
	}
	if _, err := s.Get("key"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Get(popped key) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if got, err := s.Pop("missing"); got != "" || !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Pop(missing) = %q, %v; want empty value and %v", got, err, helpers.ErrKeyDoesNotExist)
	}

	s.data["expired"] = Entry{Value: "stale", ExpireAt: time.Now().Add(-time.Second)}
	if got, err := s.Pop("expired"); got != "" || !errors.Is(err, helpers.ErrKeyExpired) {
		t.Errorf("Pop(expired) = %q, %v; want empty value and %v", got, err, helpers.ErrKeyExpired)
	}
	if _, ok := s.data["expired"]; ok {
		t.Error("Pop(expired) should remove the expired entry")
	}
}
