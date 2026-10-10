# goUtils

Reusable Go utilities, currently including a concurrency-safe in-memory key/value store with optional TTL expiration.

> This store keeps data in the current process. It is not Redis, does not speak the Redis protocol, and does not persist data across restarts.

## Requirements

- Go 1.27.1 or newer (as declared in `go.mod`)

## Install

```sh
go get github.com/tanmay21k/goUtils/memoryStorage/stores
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tanmay21k/goUtils/memoryStorage/stores"
)

func main() {
	ctx := context.Background()
	store, err := stores.NewKVStore() // default capacity: 5 keys
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Set(ctx, "greeting", stores.Value("hello, world")); err != nil {
		log.Fatal(err)
	}

	value, err := store.Get(ctx, "greeting")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(value))
}
```

The constructor returns a `*stores.KVStore`, which satisfies `stores.Store` and the optional capability interfaces. You can pass it to functions that accept the base interface, keeping application code independent of the implementation:

```go
func readGreeting(ctx context.Context, store stores.Store) (stores.Value, error) {
	return store.Get(ctx, "greeting")
}
```

## Store implementations

### In-memory store

```go
defaultStore, err := stores.NewKVStore() // capacity defaults to 5
customStore, err := stores.NewKVStore(100) // capacity of 100 keys
```

`NewKVStore` accepts no argument or exactly one capacity. Capacity must be greater than 1. The store has no expiration policy. `NewkvStore` remains as a deprecated compatibility wrapper; use `NewKVStore` in new code.

### TTL store

```go
defaultTTLStore := stores.NewDefaultTTLStore() // capacity 5; TTL 3 hours
customTTLStore, err := stores.NewTTLStore(100, 10*time.Minute)
```

`NewTTLStore` accepts a capacity and a `time.Duration`; both are validated (capacity greater than 1 and positive TTL). `NewDefaultTTLStore` uses the defaults. `NewStore` remains as a deprecated compatibility constructor, including its legacy integer duration in nanoseconds. Each successful `Set` gives the key a fresh expiration time; overwriting with `SetXX` also refreshes its TTL. Expired entries are cleaned up lazily during store operations rather than by a background goroutine.

## API capabilities

The base `stores.Store` interface provides:

| Method | Behavior |
| --- | --- |
| `Get(ctx, key)` | Return a value or an error if the key is absent or expired. |
| `Set(ctx, key, value)` | Add or replace a key, subject to capacity. |
| `Del(ctx, key)` | Delete a key; deleting an absent key is a no-op. |

Implementations also provide optional capabilities. Use a type assertion when working with an implementation through `stores.Store`:

```go
if scanner, ok := store.(stores.Scanner); ok {
	err := scanner.Scan(ctx, "user:*", func(key string) error {
		fmt.Println(key)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

| Interface | Methods |
| --- | --- |
| `stores.Exister` | `Exists(ctx, key)` checks whether a key is present. |
| `stores.KeyStore` | `Rename(ctx, oldKey, newKey)` and `Pop(ctx, key)` (get and delete). |
| `stores.KeyLister` | `Keys()` returns current keys in alphabetical order. |
| `stores.Scanner` | `Scan(ctx, pattern, callback)` visits matching keys in alphabetical order. Patterns use Go's `path.Match` syntax; an empty pattern matches all keys. |
| `stores.ConditionalStore` | `SetNX(ctx, key, value)` writes only if absent; `SetXX(ctx, key, value)` writes only if present. The boolean reports whether the condition matched. |

Both built-in implementations are safe for concurrent use. Context cancellation is checked by context-aware operations. The TTL store considers expired keys absent for `Exists` and key listing; `Get`, `Pop`, and `Rename` return `helpers.ErrKeyExpired` when they encounter an expired key. `Get` on a missing key returns `helpers.ErrKeyDoesNotExist`.

## Run the demo

From the repository root:

```sh
go run ./memoryStorage/examples/demo
```

The demo is an executable example, not part of the importable library API. It runs operations concurrently, so output order is not guaranteed.

## Test

```sh
go test ./...
```

## Project layout

- `memoryStorage/stores` — store interfaces and in-memory implementations.
- `memoryStorage/helpers` — shared errors and defaults.
- `memoryStorage/examples/demo` — runnable example program.
- `memoryStorage/docs` — design notes.

For package-specific notes, see [memoryStorage/README.md](memoryStorage/README.md).

## License

This project is licensed under the [MIT License](LICENSE).
