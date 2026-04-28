# Architecture

## Overview

Educational e-commerce backend POC in Go. Monorepo with 7 independent services communicating over HTTP (JWT token forwarding), gRPC (unary + server-side streaming), and Kafka (async event bus with outbox pattern).

## Services

| Service | HTTP Port | gRPC Port | Responsibility |
|---------|-----------|-----------|----------------|
| `auth` | 8086 | - | User registration, login, JWT issuance (HS256) |
| `catalog` | 8081 | 9081 | Product listing, Postgres-backed, seeded on startup |
| `cart` | 8082 | - | Cart management via Redis Hash (TTL 24h), JWT-protected |
| `order` | 8083 | - | Order lifecycle, Kafka outbox relay, gRPC client to inventory |
| `inventory` | 8084 | 9084 | Stock reservation, confirm, release; Redis flash sale; gRPC server |
| `search` | 8085 | - | Full-text search via Elasticsearch, gRPC streaming client to catalog |
| `payment` | 8087 | - | Mock payment with Kafka outbox, webhook callback |

## Internal Structure (per service)

```
cmd/api/          entry point - wire deps, start HTTP (+ optional gRPC) server
internal/
  domain/         entities, domain errors, constants
  repository/     DB interface + impl (GORM / Redis / ES)
  service/        business logic interface + impl
  client/         inter-service HTTP / gRPC clients
  grpc/           gRPC server implementation (catalog, inventory only)
  api/
    controller/   Gin request handlers
    routes/       route registration
    dto/          request + response types
    response/        response helpers (OK, BadRequest, etc.)
  event/          Kafka publisher, consumer, relay (order, payment only)
  config/         env-based config with defaults
pkg/
  jwtutil/        JWT sign, verify, RequireAuth middleware, TokenFromContext
  ratelimit/      sliding window rate limiter (Redis Sorted Set)
proto/            .proto service definitions (inventory, catalog)
gen/              protobuf-generated Go code
```

## Infrastructure

| Component | Role |
|-----------|------|
| PostgreSQL 16 | Transactional data: users, orders, order_items, outbox_events, stocks, reservations, payments, payment_outbox_events |
| Redis 7 | Cart sessions (Hash + TTL), flash sale counter (String), rate limiter (Sorted Set) |
| Kafka (KRaft, no Zookeeper) | Async event bus: `order.events`, `payment.events` topics |
| Kafka UI | Local topic inspector at `http://localhost:8088` |
| Elasticsearch 8 | Product search index |

## Key Patterns

### JWT Authentication

All user-facing endpoints require `Authorization: Bearer <token>`. Tokens are issued by `auth` service (HS256, 24h TTL) and verified by `pkg/jwtutil.RequireAuth` middleware. JWT secret is shared via environment variable across services.

### Token Forwarding (On-Behalf-Of)

When `order` service calls `cart` internally during checkout, it forwards the user's JWT from the request context. Cart authenticates normally — no unauthenticated internal routes needed.

```
User --> POST /orders (Authorization: Bearer <token>)
  order: verify token, store raw token in request.Context()
  order --> GET /cart   (Authorization: Bearer <token>)   [forwarded via TokenFromContext]
    cart: verify token, resolve userID from JWT claims
```

### Rate Limiting (Sliding Window)

Applied to sensitive endpoints (`POST /auth/register`, `POST /auth/login`, `POST /orders`). Uses Redis Sorted Set: score = Unix nanosecond timestamp, window = configurable duration.

```
ZADD key <now_ns> <req_id>
ZREMRANGEBYSCORE key 0 <now_ns - window>
ZCARD key  -> reject 429 if > limit
EXPIRE key <window>
```

### Outbox Pattern (order + payment)

Guarantees at-least-once Kafka delivery even on process crash. DB write and event insertion are atomic in a single transaction.

```
BEGIN TX
  -- order service:
  INSERT orders, order_items, outbox_events (type: order.created)
  -- payment service:
  UPDATE payments SET status=X  +  INSERT payment_outbox_events
COMMIT

[relay goroutine] every 3s:
  SELECT unpublished outbox events
  publish to Kafka
  mark published
```

### State Machine (order)

```
PENDING --> CONFIRMED
PENDING --> FAILED
```

Any other transition returns `ErrInvalidTransition`.

### Stock Reservation (SELECT FOR UPDATE)

When an order is created, inventory reserves stock inside a DB transaction with row-level locking:

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

Lock-free stock deduction for high-concurrency flash sales:

```
SETNX flash:{productId} <qty>      # init once
DECRBY flash:{productId} <qty>     # atomic decrement
  if result < 0: INCRBY (rollback) -> ErrSoldOut
```

Single-threaded Redis guarantees no oversell without DB locks.

### gRPC Unary (order -> inventory)

Internal stock operations use type-safe gRPC instead of HTTP:

```
order (InventoryGRPCClient)  -->  inventory gRPC server :9084
  Reserve(orderID, items)         check stock, lock rows, update
  Confirm(orderID)                deduct permanently
  Release(orderID)                release reservation
```

`InventoryClient` interface is also implemented by `InventoryHTTPClient` as a reference for interface-driven design — swapping transports requires zero business logic changes.

### gRPC Server-Side Streaming (search -> catalog)

Catalog streams all products one-by-one; search collects them for bulk indexing into Elasticsearch:

```
search (CatalogGRPCClient)  -->  catalog gRPC server :9081
  StreamProducts({})              for each product: stream.Send(p)
                                  return nil (EOF)
  stream.Recv() loop until EOF    collect -> BulkIndex to ES
```

Streaming avoids loading all products into memory at once on either side.

### Idempotency (order)

`POST /orders` requires `Idempotency-Key` header. Duplicate keys return the existing order without re-processing.

## Service Communication Map

```
[user]
    |
    +-- POST /auth/register|login         --> auth:8086
    |
    +-- GET  /products                    --> catalog:8081
    +-- POST /cart/items                  --> cart:8082    [JWT]
    +-- GET  /cart                        --> cart:8082    [JWT]
    |
    +-- POST /orders                      --> order:8083   [JWT + rate limit]
    |      +-- GET /cart                  --> cart:8082    [JWT forwarded]
    |      +-- Reserve(orderID, items)    --> inventory:9084  [gRPC unary]
    |      +-- POST /cart/clear           --> cart:8082    [JWT forwarded]
    |      [outbox relay]                 --> Kafka: order.events
    |
    +-- POST  /payments                   --> payment:8087 [JWT]
    +-- POST  /payments/:id/callback      --> payment:8087 [JWT]
    |      [outbox relay]                 --> Kafka: payment.events
    |
    [order Kafka consumer]
      <-- payment.events                  <-- Kafka
          -> ConfirmOrder / FailOrder (internal)
    |
    +-- GET /search?q=...                 --> search:8085
    |      [startup reindex via gRPC stream]
    |      StreamProducts({})             --> catalog:9081 [gRPC streaming]
```
