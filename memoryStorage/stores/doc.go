// Package stores provides interfaces and concurrency-safe in-memory key/value
// store implementations. Use NewKVStore for a store without expiration, or
// NewTTLStore and NewDefaultTTLStore for stores with lazy TTL expiration. The
// implementations support the Store contract and optional interfaces for key
// inspection, scanning, conditional writes, and key management.
//
// Data exists only for the lifetime of the process and is not persisted. See
// the repository README for installation and usage examples.
package stores
