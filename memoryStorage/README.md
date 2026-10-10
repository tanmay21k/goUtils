# In-memory key/value store

This directory contains the reusable store package in `stores` and its runnable demo. `NewKVStore` creates a non-expiring store; `NewTTLStore` creates a store with a caller-specified `time.Duration` expiration, and `NewDefaultTTLStore` applies the package defaults. Both implementations support optional key-management capabilities.

For installation, copyable usage examples, API capabilities, defaults, and test instructions, see the [repository README](../README.md).

The package import path is `github.com/tanmay21k/goUtils/memoryStorage/stores`. The deprecated `NewkvStore` and `NewStore` constructors remain available for compatibility. Data is process-local and is not persisted; this package is not a Redis server or Redis protocol client.

Run the demo from the repository root with `go run ./memoryStorage/examples/demo`.
