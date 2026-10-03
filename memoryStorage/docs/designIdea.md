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
 │      └── kv.store
 │          ├── Keep concrete type private
 │          ├── Implement stores.Store
 │          └── Expose NewStore()
 │
 ├── 4. Identify variations in behavior
 │      └── Need expiration?
 │          └── Create ttl.store
 │              ├── Implement stores.Store
 │              └── Expose NewStore()
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
 │                  ├── kv.store
 │                  └── ttl.store
 │
 └── 7. Keep dependencies pointing inward
        ├── Business logic → interface
        ├── Middleware → interface
        └── Application → public constructors