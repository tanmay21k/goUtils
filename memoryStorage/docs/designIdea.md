Start
 │
 ├── 1. Identify the core behavior
 │      └── What operations must every store support?
 │          ├── Get
 │          ├── Set
 │          ├── Del
 │          ├── Keys
 │          ├── Rename
 │          └── Pop
 │
 ├── 2. Define the interface
 │      └── stores.Store
 │          └── Describe behavior, not implementation
 │
 ├── 3. Create the first implementation
 │      └── stores.KVStore
 │          ├── Keep internal state private
 │          ├── Implement stores.Store
 │          └── Expose NewKVStore()
 │
 ├── 4. Identify variations in behavior
 │      └── Need expiration?
 │          └── Create stores.TTLStore
 │              ├── Implement stores.Store
 │              └── Expose NewTTLStore() and NewDefaultTTLStore()
 │
 ├── 5. Add cross-cutting business logic
 │      └── Middleware
 │          ├── Accept stores.Store
 │          ├── Logging
 │          ├── Metrics
 │          └── Other policies
 │
 ├── 6. Compose the implementation
 │      └── main
 │          └── Middleware
 │              └── stores.Store
 │                  ├── stores.KVStore
 │                  └── stores.TTLStore
 │
 └── 7. Keep dependencies pointing inward
        ├── Business logic → interface
        ├── Middleware → interface
        └── Application → public constructors