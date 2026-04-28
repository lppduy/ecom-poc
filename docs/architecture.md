# Architecture

## Overview

Educational e-commerce POC in Go - monorepo, multiple services communicating over HTTP, with async event publishing via Kafka.

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| `catalog` | 8081 | Product listing, Postgres-backed, seed on startup |
| `cart` | 8082 | Cart management via Redis Hash (TTL 24h) |
| `order` | 8083 | Order lifecycle, Kafka outbox relay |
| `inventory` | 8084 | Stock reservation, confirm, release |
| `search` | 8085 | Full-text search via Elasticsearch |

## Internal Structure (per service)

```
cmd/api/          entry point - wire deps, start HTTP server
internal/
  domain/         entities, domain errors, constants
  repository/     DB interface + impl (GORM / Redis / ES)
  service/        business logic interface + impl
  client/         inter-service HTTP clients
  api/
    controller/   Gin request handlers
    routes/       route registration
    dto/          request + response types
    httpx/        response helpers (OK, BadRequest, etc.)
  config/         env-based config with defaults
```

## Infrastructure

| Component | Role |
|-----------|------|
| PostgreSQL 16 | Transactional data: orders, order_items, outbox_events, stocks, reservations |
| Redis 7 | Cart session storage: Hash per user, 24h TTL |
| Kafka (KRaft) | Async event bus: `order.events` topic |
| Kafka UI | Local topic inspector at `http://localhost:8088` |
| Elasticsearch 8 | Product search index |

## Key Patterns

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

This guarantees no event is lost even if the process crashes after commit.

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

### Idempotency

`POST /orders` requires `Idempotency-Key` header. If a key is seen again, the existing order is returned - safe to retry on timeout or network error.

## Service Communication

```
[client]
    |
    +- GET /products              -> catalog:8081
    +- POST /cart/items           -> cart:8082
    +- GET /cart                  -> cart:8082
    |
    +- POST /orders               -> order:8083
    |      +- GET /cart           -> cart:8082   (fetch items)
    |      +- POST /inventory/reserve -> inventory:8084
    |      +- POST /cart/clear    -> cart:8082
    |
    +- PATCH /orders/:id/confirm  -> order:8083
    |      +- POST /inventory/confirm -> inventory:8084
    |
    +- PATCH /orders/:id/fail     -> order:8083
    |      +- POST /inventory/release -> inventory:8084
    |
    +- GET /search?q=...          -> search:8085
```
