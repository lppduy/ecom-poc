# Trade-offs

## Decisions Made

### Auth: HS256 JWT with shared secret

- Simple to implement and verify across services
- Secret distributed via environment variable (acceptable for POC)
- Production: use asymmetric keys (RS256) or a dedicated auth server (Keycloak, Auth0)

### Cart: Redis Hash with TTL

- Fast reads/writes, natural expiry via TTL
- No persistence after Redis restart
- Production: Redis AOF/RDB persistence, or hybrid with DB for checkout

### Token Forwarding (OBO) for order -> cart

- Order forwards user JWT to cart so cart authenticates normally
- Avoids unauthenticated internal endpoints
- Production: add service-to-service auth (mTLS, service tokens) on top

### gRPC for order -> inventory (Unary)

- Type-safe, schema-enforced contract via Protobuf
- More setup than HTTP but better for critical internal calls
- Interface-driven: `InventoryClient` interface lets you swap HTTP/gRPC without touching business logic

### gRPC Server-Side Streaming for search -> catalog

- Catalog streams products one-by-one instead of returning a giant array
- Lower memory pressure on both sides at scale
- For this POC (2 products): no practical difference vs HTTP; demonstrates the pattern

### Kafka Outbox Pattern (order + payment)

- Guarantees at-least-once Kafka delivery even on crash
- DB write + event insertion are atomic in one TX
- Relay goroutine retries automatically on failure
- Trade-off: extra `outbox_events` / `payment_outbox_events` tables, relay latency (~3s)

### Kafka Consumer: simple partition reader (no GroupID)

- kafka-go consumer group protocol does not work reliably with this Docker Kafka listener setup
- Simple reader with `StartOffset: FirstOffset` re-processes all historical events on restart
- State machine rejects invalid transitions idempotently (safe re-processing)
- Production: fix Kafka listener config for proper consumer groups, use committed offsets

### Flash Sale: Redis Atomic Counter

- Redis `DECRBY` is atomic; no DB locks needed
- Handles thousands of concurrent requests without contention
- Pattern name: "Redis Atomic Counter" / "Lock-Free Stock Deduction"
- Trade-off: separate from DB stock, requires sync if orders are cancelled

### Rate Limiting: Sliding Window via Redis Sorted Set

- More accurate than fixed window (no thundering herd at window boundary)
- Applied per-IP on auth endpoints and per-user on order creation
- Trade-off: Redis dependency for middleware; adds latency per request

### SELECT FOR UPDATE for regular stock reservation

- Pessimistic locking - safe for moderate concurrency
- Clear, auditable reservation trail in `reservations` table
- Trade-off: row-level lock contention at very high concurrency (use flash sale counter instead)

### Single PostgreSQL instance (shared schema)

- Each service uses its own tables in the same `ecom` database
- Simple for local dev; no per-service DB setup
- Production: separate DB per service (true isolation), use migrations tool (Goose, Atlas)

### No service discovery / load balancing

- Services call each other via hardcoded `localhost:<port>`
- Acceptable for POC; production needs Consul, Kubernetes DNS, or API gateway

## Known Limitations

| Limitation | Impact | Production Fix |
|---|---|---|
| Kafka consumer re-reads all history on restart | Duplicate event processing (safe but wasteful) | Consumer groups with committed offsets |
| JWT secret shared via env var | All services can forge tokens | Asymmetric keys, short TTLs, refresh tokens |
| No service-to-service auth | Any process can call internal routes | mTLS, service mesh, API gateway |
| Single Kafka partition (per topic) | Limited parallelism | Multiple partitions + consumer group |
| No circuit breaker | Cascading failures on dependency outage | Hystrix, resilience4j, or custom retry |
| ES indexed on startup only | New products not searchable until reindex | CDC (Debezium) or event-driven indexing |
