package stores

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/tanmay21k/goUtils/memoryStorage/helpers"
)

func TestNewKVStore(t *testing.T) {
	store, err := NewKVStore()
	if err != nil {
		t.Fatalf("NewKVStore() error = %v", err)
	}
	if store.maxValidSize != DefaultLimit {
		t.Fatalf("default capacity = %d, want %d", store.maxValidSize, DefaultLimit)
	}

	if _, err := NewKVStore(1); err == nil {
		t.Fatal("NewKVStore(1) expected an error")
	}
	if _, err := NewKVStore(2, 3); !errors.Is(err, helpers.ErrInvalidSize) {
		t.Fatalf("NewKVStore with multiple arguments error = %v, want %v", err, helpers.ErrInvalidSize)
	}
	//lint:ignore SA1019 Verify the backward-compatible constructor remains available.
	if _, err := NewkvStore(2); err != nil {
		t.Fatalf("deprecated NewkvStore(2) error = %v", err)
	}
}

func TestKVStoreOperations(t *testing.T) {
	ctx := context.Background()
	store, err := NewKVStore(2)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.Get(ctx, "missing"); !errors.Is(err, helpers.ErrKeyDoesNotExist) {
		t.Fatalf("Get(missing) error = %v", err)
	}
	if err := store.Set(ctx, "", "value"); !errors.Is(err, helpers.ErrEmptyStringNotAllow) {
		t.Fatalf("Set(empty key) error = %v", err)
	}
	if err := store.Set(ctx, "b", "two"); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "a", "one"); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "a", "updated"); err != nil {
		t.Fatalf("replacing an existing key failed: %v", err)
	}
	if err := store.Set(ctx, "c", "three"); !errors.Is(err, helpers.ErrSizeExceed) {
		t.Fatalf("Set beyond capacity error = %v, want %v", err, helpers.ErrSizeExceed)
	}

	value, err := store.Get(ctx, "a")
	if err != nil || value != "updated" {
		t.Fatalf("Get(a) = %q, %v; want %q, nil", value, err, "updated")
	}
	if exists, err := store.Exists(ctx, "a"); err != nil || !exists {
		t.Fatalf("Exists(a) = %t, %v; want true, nil", exists, err)
	}
	if exists, err := store.Exists(ctx, "missing"); err != nil || exists {
		t.Fatalf("Exists(missing) = %t, %v; want false, nil", exists, err)
	}

	if created, err := store.SetNX(ctx, "a", "ignored"); err != nil || created {
		t.Fatalf("SetNX(existing) = %t, %v; want false, nil", created, err)
	}
	if updated, err := store.SetXX(ctx, "missing", "ignored"); err != nil || updated {
		t.Fatalf("SetXX(missing) = %t, %v; want false, nil", updated, err)
	}
	if updated, err := store.SetXX(ctx, "a", "conditional"); err != nil || !updated {
		t.Fatalf("SetXX(existing) = %t, %v; want true, nil", updated, err)
	}

	if err := store.Rename(ctx, "a", "b"); err != nil {
		t.Fatalf("Rename(a, b) error = %v", err)
	}
	value, err = store.Get(ctx, "b")
	if err != nil || value != "conditional" {
		t.Fatalf("Get(b) after rename = %q, %v", value, err)
	}
	if got := store.Keys(); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("Keys() = %v, want [b]", got)
	}

	var scanned []string
	if err := store.Scan(ctx, "*", func(key string) error {
		scanned = append(scanned, key)
		return nil
	}); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if !reflect.DeepEqual(scanned, []string{"b"}) {
		t.Fatalf("Scan() keys = %v, want [b]", scanned)
	}
	if err := store.Scan(ctx, "[", func(string) error { return nil }); err == nil {
		t.Fatal("Scan with an invalid pattern expected an error")
	}
	if err := store.Scan(ctx, "*", nil); err == nil {
		t.Fatal("Scan with a nil callback expected an error")
	}

	popped, err := store.Pop(ctx, "b")
	if err != nil || popped != "conditional" {
		t.Fatalf("Pop(b) = %q, %v", popped, err)
	}
	if err := store.Del(ctx, "absent"); err != nil {
		t.Fatalf("Del(absent) error = %v", err)
	}
}

func TestNewTTLStore(t *testing.T) {
	store, err := NewTTLStore(12, 2*time.Minute)
	if err != nil {
		t.Fatalf("NewTTLStore() error = %v", err)
	}
	if store.maxValidSize != 12 || store.expireTime != 2*time.Minute {
		t.Fatalf("constructor settings = (%d, %s), want (12, 2m)", store.maxValidSize, store.expireTime)
	}

	defaults := NewDefaultTTLStore()
	if defaults.maxValidSize != DefaultLimit || defaults.expireTime != DefaultTTL {
		t.Fatalf("default settings = (%d, %s), want (%d, %s)", defaults.maxValidSize, defaults.expireTime, DefaultLimit, DefaultTTL)
	}
	if _, err := NewTTLStore(1, time.Minute); err == nil {
		t.Fatal("NewTTLStore with capacity 1 expected an error")
	}
	if _, err := NewTTLStore(2, 0); err == nil {
		t.Fatal("NewTTLStore with zero TTL expected an error")
	}
	//lint:ignore SA1019 Verify the backward-compatible constructor remains available.
	if _, err := NewStore(3, int(time.Minute)); err != nil {
		t.Fatalf("legacy NewStore(3, duration) error = %v", err)
	}
}

func TestTTLStoreExpirationAndRefresh(t *testing.T) {
	ctx := context.Background()
	store, err := NewTTLStore(3, time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Set(ctx, "expired", "old"); err != nil {
		t.Fatal(err)
	}
	store.data["expired"] = Entry{Value: "old", ExpireAt: time.Now().Add(-time.Second)}
	if _, err := store.Get(ctx, "expired"); !errors.Is(err, helpers.ErrKeyExpired) {
		t.Fatalf("Get(expired) error = %v, want %v", err, helpers.ErrKeyExpired)
	}
	if exists, err := store.Exists(ctx, "expired"); err != nil || exists {
		t.Fatalf("Exists(expired) = %t, %v; want false, nil", exists, err)
	}

	store.data["conditional"] = Entry{Value: "old", ExpireAt: time.Now().Add(-time.Second)}
	if updated, err := store.SetXX(ctx, "conditional", "new"); err != nil || updated {
		t.Fatalf("SetXX(expired) = %t, %v; want false, nil", updated, err)
	}
	if created, err := store.SetNX(ctx, "conditional", "new"); err != nil || !created {
		t.Fatalf("SetNX(expired) = %t, %v; want true, nil", created, err)
	}
	if entry := store.data["conditional"]; entry.Value != "new" || time.Until(entry.ExpireAt) <= 0 {
		t.Fatalf("SetNX did not create a live entry: %+v", entry)
	}
}

func TestStoreContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store, err := NewKVStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Set(ctx, "key", "value"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set with canceled context error = %v, want context.Canceled", err)
	}
}

func TestStoresConcurrentAccess(t *testing.T) {
	const workers = 8
	const perWorker = 20
	ctx := context.Background()
	stores := []Store{
		mustNewKVStore(t, workers*perWorker),
		mustNewTTLStore(t, workers*perWorker, time.Minute),
	}

	for storeIndex, store := range stores {
		var wg sync.WaitGroup
		for worker := 0; worker < workers; worker++ {
			wg.Add(1)
			go func(worker int) {
				defer wg.Done()
				for item := 0; item < perWorker; item++ {
					key := fmt.Sprintf("%d:%d", worker, item)
					if err := store.Set(ctx, key, Value(key)); err != nil {
						t.Errorf("store %d Set(%q): %v", storeIndex, key, err)
						return
					}
					if _, err := store.Get(ctx, key); err != nil {
						t.Errorf("store %d Get(%q): %v", storeIndex, key, err)
						return
					}
				}
			}(worker)
		}
		wg.Wait()
		if got, want := len(store.(KeyLister).Keys()), workers*perWorker; got != want {
			t.Errorf("store %d key count = %d, want %d", storeIndex, got, want)
		}
	}
}

func mustNewKVStore(t *testing.T, capacity int) *KVStore {
	t.Helper()
	store, err := NewKVStore(capacity)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func mustNewTTLStore(t *testing.T, capacity int, ttl time.Duration) *TTLStore {
	t.Helper()
	store, err := NewTTLStore(capacity, ttl)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
