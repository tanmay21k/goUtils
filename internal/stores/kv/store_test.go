package kv

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tanmay21k/goUtils/internal/helpers"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name       string
		args       []int
		wantLimit  int
		wantErr    error
		wantErrMsg string
	}{
		{name: "default size", wantLimit: DefaultLimit},
		{name: "custom size", args: []int{3}, wantLimit: 3},
		{name: "multiple sizes", args: []int{2, 3}, wantErr: helpers.ErrInvalidSize},
		{name: "size too small", args: []int{1}, wantErrMsg: "size must be greater than 1"},
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

			for i := 0; i < tt.wantLimit; i++ {
				if err := got.Set(string(rune('a'+i)), "value"); err != nil {
					t.Fatalf("Set() within capacity: %v", err)
				}
			}
			if err := got.Set("overflow", "value"); !errors.Is(err, helpers.ErrSizeExceed) {
				t.Errorf("Set() over capacity error = %v, want %v", err, helpers.ErrSizeExceed)
			}
		})
	}
}

func TestStoreGet(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Get("missing"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Fatalf("Get(missing) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if err := store.Set("empty", ""); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("empty")
	if err != nil {
		t.Fatalf("Get(empty) error = %v", err)
	}
	if got != "" {
		t.Errorf("Get(empty) = %q, want empty string", got)
	}
}

func TestStoreSet(t *testing.T) {
	store, err := NewStore(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set("", "value"); !errors.Is(err, helpers.ErrEmptyStringNotAllow) {
		t.Fatalf("Set(empty key) error = %v, want %v", err, helpers.ErrEmptyStringNotAllow)
	}
	if err := store.Set("first", "1"); err != nil {
		t.Fatal(err)
	}
	if err := store.Set("second", "2"); err != nil {
		t.Fatal(err)
	}
	if err := store.Set("first", "updated"); err != nil {
		t.Fatalf("updating a key at capacity: %v", err)
	}
	if err := store.Set("third", "3"); !errors.Is(err, helpers.ErrSizeExceed) {
		t.Errorf("Set() over capacity error = %v, want %v", err, helpers.ErrSizeExceed)
	}
	got, err := store.Get("first")
	if err != nil || got != "updated" {
		t.Errorf("Get(first) = %q, %v; want %q, nil", got, err, "updated")
	}
}

func TestStoreDel(t *testing.T) {
	store, _ := NewStore()
	if err := store.Set("key", "value"); err != nil {
		t.Fatal(err)
	}

	store.Del("key")
	if _, err := store.Get("key"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Get(deleted key) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	store.Del("missing")
}

func TestStoreKeys(t *testing.T) {
	store, _ := NewStore()
	for key, value := range map[string]string{"gamma": "3", "alpha": "1", "beta": "2"} {
		if err := store.Set(key, value); err != nil {
			t.Fatal(err)
		}
	}

	want := []string{"alpha", "beta", "gamma"}
	if got := store.Keys(); !reflect.DeepEqual(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestStoreRename(t *testing.T) {
	store, _ := NewStore()
	if err := store.Set("old", "value"); err != nil {
		t.Fatal(err)
	}

	if err := store.Rename("missing", "new"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Fatalf("Rename(missing) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if err := store.Rename("old", "new"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if got, err := store.Get("new"); err != nil || got != "value" {
		t.Errorf("Get(new) = %q, %v; want %q, nil", got, err, "value")
	}
	if _, err := store.Get("old"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Get(old) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
}

func TestStorePop(t *testing.T) {
	store, _ := NewStore()
	if err := store.Set("key", "value"); err != nil {
		t.Fatal(err)
	}

	got, err := store.Pop("key")
	if err != nil || got != "value" {
		t.Fatalf("Pop(key) = %q, %v; want %q, nil", got, err, "value")
	}
	if _, err := store.Get("key"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Get(popped key) error = %v, want %v", err, helpers.ErrKeyDoesNotExist)
	}
	if got, err := store.Pop("missing"); got != "" || !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Errorf("Pop(missing) = %q, %v; want empty value and %v", got, err, helpers.ErrKeyDoesNotExist)
	}
}
