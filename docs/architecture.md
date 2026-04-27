# Architecture

## Context

Educational e-commerce POC focused on core ordering flow and system design trade-offs. Go-only, monorepo structure.

## Services

| Service | Port | Responsibility |
|---------|------|----------------|
| `catalog` | 8081 | Product listing |
| `cart` | 8082 | Cart item management |
| `order` | 8083 | Order lifecycle + payment state |

## Internal Structure (per service)

```
cmd/api/          entry point
internal/
  domain/         entities, errors, constants
  repository/     DB interface + GORM impl
  service/        business logic interface + impl
  client/         inter-service HTTP clients
  api/
    controller/   request handling
    routes/       route registration
    dto/          request + response types
    httpx/        response helpers
  config/         env-based config
```

## Infrastructure

- **PostgreSQL** — transactional data (orders, cart items)
- **Redis** — available for cache/session (not yet wired)
- **Kafka (KRaft)** — available for event-driven workflows (not yet wired)
- **Kafka UI** — local topic inspection at `http://localhost:8088`

## Current Flow

`catalog` → `cart` → `order` (sync HTTP between services)

Order service calls cart service via `CartClient` interface to fetch + clear items.

## Next Milestones

1. Kafka outbox: publish `order.created` event on order creation
2. Inventory service: reserve stock, rollback on failure
3. Redis cart: replace Postgres cart with Redis for durability + TTL
4. Auth middleware: JWT validation across services
