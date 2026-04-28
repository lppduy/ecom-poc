# ecom-poc (Go)

Hands-on e-commerce system built with Go — focused on production-minded patterns: outbox, state machine, inter-service HTTP, Redis, Kafka, Elasticsearch.

## Services

| Service | Port | Stack | Responsibility |
|---------|------|-------|----------------|
| `catalog` | 8081 | Gin + GORM + Postgres | Product listing |
| `cart` | 8082 | Gin + Redis | Cart (Hash, TTL 24h) |
| `order` | 8083 | Gin + GORM + Kafka | Order lifecycle + outbox relay |
| `inventory` | 8084 | Gin + GORM + Postgres | Stock reserve / confirm / release |
| `search` | 8085 | Gin + Elasticsearch | Full-text product search |

## Architecture Highlights

- **Outbox pattern** — order events written to DB in same transaction, relayed to Kafka every 3s
- **State machine** — `PENDING → CONFIRMED / FAILED`, enforced at domain layer
- **Idempotency** — `Idempotency-Key` header prevents duplicate orders on retry
- **SELECT FOR UPDATE** — prevents oversell when multiple orders reserve same stock
- **Redis Hash** — cart stored as `cart:{userId}` Hash with 24h TTL
- **Elasticsearch** — fuzzy full-text search with price range filter, auto-indexed from catalog on startup

## Infrastructure

```
PostgreSQL   → transactional data (orders, inventory)
Redis        → cart session storage
Kafka        → async event bus (order.events topic)
Kafka UI     → http://localhost:8088
Elasticsearch → product search index
```

## Quick Start

**1. Start infrastructure:**

```bash
cd infra
docker compose up -d
```

**2. Run all services with hot-reload:**

```bash
./scripts/dev-air.sh
```

Or run individually:

```bash
cd services/catalog   && PORT=8081 go run ./cmd/api/main.go
cd services/cart      && PORT=8082 go run ./cmd/api/main.go
cd services/order     && PORT=8083 go run ./cmd/api/main.go
cd services/inventory && PORT=8084 go run ./cmd/api/main.go
cd services/search    && PORT=8085 go run ./cmd/api/main.go
```

**3. Run smoke tests:**

```bash
./scripts/smoke-flow.sh
```

## Key Flows

### Checkout flow

```
POST /cart/items       → add product to Redis cart
POST /orders           → fetch cart, create order + outbox in TX, reserve inventory, clear cart
PATCH /orders/:id/confirm → PENDING → CONFIRMED, deduct stock permanently
PATCH /orders/:id/fail    → PENDING → FAILED, release reserved stock
```

### Search

```bash
curl "http://localhost:8085/search?q=iphone"
curl "http://localhost:8085/search?q=airpods&maxPrice=7000000"
curl "http://localhost:8085/search?q=iphoen"   # fuzzy — typo-tolerant
```

## Project Structure

```
ecom-poc/
  docs/        technical docs
  infra/       docker-compose + infra configs
  services/    Go microservices
  scripts/     dev helpers and smoke tests
  local/       local-only notes (gitignored)
```

Each service follows the same internal layout:

```
cmd/api/        entry point
internal/
  domain/       entities, errors, constants
  repository/   interface + GORM / Redis / ES impl
  service/      business logic
  client/       inter-service HTTP clients
  api/          controller, routes, dto, httpx
  config/       env-based config
```
