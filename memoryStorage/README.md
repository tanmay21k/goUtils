# Memory Store

A reusable Go package for in-memory key/value storage with optional behaviors such as TTL, conditional writes, key inspection, and key management.

This project is intended to be imported by other Go projects. The public usage surface is intentionally small, and the runnable example is kept separate from the library code.

---

## Package layout

```text
memoryStorage/
├── README.md
├── stores/
│   ├── store.go
│   ├── kv.go
│   └── ttl.go
├── helpers/
├── storeService/
├── examples/
│   └── demo/
│       ├── main.go
│       └── commands.go
└── ...
```

The demo package is kept in `examples/demo` so the library itself remains clean and importable by other projects.

---

## Core API

```go
package stores

type Value string

type Store interface {
    Get(ctx context.Context, key string) (Value, error)
    Set(ctx context.Context, key string, value Value) error
    Del(ctx context.Context, key string) error
}
```

The main contract is intentionally small. Additional capabilities are exposed through optional interfaces rather than being forced into the base contract.

---

## Optional interfaces

```go
type Exister interface {
    Exists(ctx context.Context, key string) (bool, error)
}

type KeyStore interface {
    Rename(ctx context.Context, oldKey, newKey string) error
    Pop(ctx context.Context, key string) (Value, error)
}

type KeyLister interface {
    Keys() []string
}

type Scanner interface {
    Scan(ctx context.Context, pattern string, fn func(key string) error) error
}

type ConditionalStore interface {
    SetNX(ctx context.Context, key string, value Value) (bool, error)
    SetXX(ctx context.Context, key string, value Value) (bool, error)
}
```

---

## Example usage

The example program is intentionally minimal and only exposes the end-user function `Do`.

```go
func Do(ctx context.Context, store stores.Store, out chan<- error, logger *log.Logger, ops [][]string)
```

The command list is kept in a separate file, while the internal execution logic stays private.

Example command payload:

```go
var example = [][]string{
    {"SET", "user:1", "alice"},
    {"GET", "user:1"},
    {"EXISTS", "user:1"},
}
```

The user calls `Do` with:

- a context
- a concrete store instance
- an error channel
- a logger
- a slice of command arguments

Everything else in the demo package is private.

---

## Demo run

From the repository root:

```bash
go run ./memoryStorage/examples/demo
```

This runs the example without exposing private helper functions or command variables as part of the public API.

---

## Design goals

- small core interface
- optional capability interfaces
- multiple implementations behind the same API
- reusable library, not a standalone executable
- separate demo layer for examples and testing
