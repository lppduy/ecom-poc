# Architecture

## Overview

Educational e-commerce POC in Go - monorepo, multiple services communicating over HTTP (with JWT token forwarding) and gRPC, with async event publishing via Kafka.

## Services

| Service | Port | gRPC Port | Responsibility |
|---------|------|-----------|----------------|
| `auth` | 8086 | - | User registration, login, JWT issuance |
| `catalog` | 8081 | - | Product listing, Postgres-backed, seed on startup |
| `cart` | 8082 | - | Cart management via Redis Hash (TTL 24h), JWT-protected |
| `order` | 8083 | - | Order lifecycle, Kafka outbox relay, gRPC client to inventory |
| `inventory` | 8084 | 9084 | Stock reservation, confirm, release; flash sale via Redis; gRPC server |
| `search` | 8085 | - | Full-text search via Elasticsearch |
| `payment` | 8087 | - | Mock payment, webhook callback to order |

## Internal Structure (per service)

```
cmd/api/          entry point - wire deps, start HTTP (+ optional gRPC) server
internal/
  domain/         entities, domain errors, constants
  repository/     DB interface + impl (GORM / Redis / ES)
  service/        business logic interface + impl
  client/         inter-service HTTP / gRPC clients
  grpc/           gRPC server implementation (inventory only)
  api/
    controller/   Gin request handlers
    routes/       route registration
    dto/          request + response types
    httpx/        response helpers (OK, BadRequest, etc.)
  config/         env-based config with defaults
pkg/
  jwtutil/        JWT sign, verify, RequireAuth middleware, token context helpers
  ratelimit/      sliding window rate limiter (Redis Sorted Set)
proto/            .proto service definitions
gen/              protobuf-generated Go code
```

## Infrastructure

| Component | Role |
|-----------|------|
| PostgreSQL 16 | Transactional data: users, orders, order_items, outbox_events, stocks, reservations, payments |
| Redis 7 | Cart sessions (Hash + TTL), flash sale counter (String), rate limiter (Sorted Set) |
| Kafka (KRaft) | Async event bus: `order.events` topic |
| Kafka UI | Local topic inspector at `http://localhost:8088` |
| Elasticsearch 8 | Product search index |

## Key Patterns

### JWT Authentication

All user-facing endpoints require `Authorization: Bearer <token>`. Tokens are issued by `auth` service (HS256, 24h TTL) and verified by `pkg/jwtutil.RequireAuth` middleware shared across services.

### Token Forwarding (On-Behalf-Of)

When `order` service calls `cart` internally during checkout, it forwards the user's original JWT in the `Authorization` header. This means:
- Cart authenticates normally using the same middleware
- No unauthenticated internal endpoints needed
- Token already encodes `userID`, no extra query params required

```
User -> POST /orders (Authorization: Bearer <token>)
  order: verify token -> store raw token in request.Context()
  order -> GET /cart (Authorization: Bearer <token>)   [forwarded]
    cart: verify token -> resolve userID from JWT claims
```

### Rate Limiting (Sliding Window)

Applied to sensitive endpoints (`/auth/register`, `/auth/login`, `POST /orders`). Uses Redis Sorted Set with score = Unix nanosecond timestamp:

```
ZADD key <now_ns> <request_id>
ZREMRANGEBYSCORE key 0 <now_ns - window>
ZCARD key  -> reject if > limit
EXPIRE key <window>
```

### Outbox Pattern

Order creation and event publishing are atomic - both happen inside the same DB transaction:

```
BEGIN TX
  INSERT orders
  INSERT order_items
  INSERT outbox_events (type: order.created)
COMMIT

[goroutine] every 3s:
  SELECT unpublished outbox events
  publish to Kafka "order.events"
  mark published
```

### State Machine

Order statuses are enforced at the domain layer:

```
PENDING -> CONFIRMED
PENDING -> FAILED
```

Any other transition returns `ErrInvalidTransition`.

### Stock Reservation (SELECT FOR UPDATE)

When an order is created, inventory is reserved inside a DB transaction with row-level locking to prevent oversell:

```
BEGIN TX
  SELECT * FROM stocks WHERE product_id = ? FOR UPDATE
  CHECK available >= requested
  UPDATE stocks SET reserved += qty
  INSERT reservations
COMMIT
```

On confirm: `quantity -= qty, reserved -= qty`

On fail: `reserved -= qty`

### Flash Sale (Redis Atomic Counter)

High-concurrency stock deduction for flash sales uses Redis atomic operations instead of DB transactions:

```
SETNX flash:{productId} <qty>   # initialize once
DECRBY flash:{productId} <qty>  # atomic decrement
  -> if result < 0: INCRBY (rollback), return ErrOutOfStock
```

Single-threaded Redis guarantees no oversell without DB locks.

### gRPC (order -> inventory)

Internal stock operations use gRPC for type-safe, efficient service-to-service calls:

```
order service        inventory service
  InventoryClient  ->  InventoryServiceServer (gRPC :9084)
  (grpc impl)          (inventory_server.go)
```

The `InventoryClient` interface is also implemented by an HTTP client (`inventory_client_http.go`) as a reference for interface-driven design - swapping transports requires zero changes to business logic.

### Idempotency

`POST /orders` requires `Idempotency-Key` header. If a key is seen again, the existing order is returned - safe to retry on timeout or network error.

## Service Communication

```
[user]
    |
    +- POST /auth/register|login       -> auth:8086
    |
    +- GET /products                   -> catalog:8081
    +- POST /cart/items                -> cart:8082   [JWT]
    +- GET  /cart                      -> cart:8082   [JWT]
    |
    +- POST /orders                    -> order:8083  [JWT + rate limit]
    |      +- GET /cart                -> cart:8082   [JWT forwarded]
    |      +- Reserve(orderID, items)  -> inventory:9084  [gRPC]
    |      +- POST /cart/clear         -> cart:8082   [JWT forwarded]
    |
    +- POST /payments                  -> payment:8087 [JWT]
    |      +- POST /payments/:id/callback
    |             +- POST /internal/orders/:id/confirm -> order:8083
    |
    +- GET /search?q=...               -> search:8085
