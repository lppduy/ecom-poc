# Ecom PoC

E-commerce backend POC: Go, Gin, JWT auth, Kafka, Redis, Postgres, Elasticsearch, gRPC.

## Services

| Service | Port | Stack | Responsibility |
|---------|------|-------|----------------|
| `auth` | 8086 | Gin + GORM + bcrypt + JWT | Register, login, JWT issuance |
| `catalog` | 8081 | Gin + GORM + Postgres | Product listing |
| `cart` | 8082 | Gin + Redis | Cart (Hash, TTL 24h), JWT-protected |
| `order` | 8083 | Gin + GORM + Kafka | Order lifecycle, outbox relay, gRPC client |
| `inventory` | 8084 | Gin + GORM + Redis + gRPC | Stock reserve / confirm / release, flash sale |
| `search` | 8085 | Gin + Elasticsearch | Full-text product search |
| `payment` | 8087 | Gin + GORM + Postgres | Mock payment, webhook callback |

## Architecture Highlights

- **JWT auth**: HS256 tokens issued by `auth`, verified by shared `pkg/jwtutil` middleware across all services
- **Token forwarding (OBO)**: when `order` calls `cart` internally, it forwards the user's JWT so `cart` authenticates normally without extra internal routes
- **Rate limiting**: sliding window algorithm via Redis Sorted Sets on `POST /auth/register`, `POST /auth/login`, `POST /orders`
- **Outbox pattern**: order events written to DB in same transaction, relayed to Kafka every 3s
- **State machine**: `PENDING -> CONFIRMED / FAILED`, enforced at domain layer
- **Idempotency**: `Idempotency-Key` header prevents duplicate orders on retry
- **SELECT FOR UPDATE**: prevents oversell when multiple orders reserve same stock simultaneously
- **gRPC (order -> inventory)**: internal stock operations use gRPC instead of HTTP; demonstrates interface-driven swap
- **Redis atomic counter**: flash sale stock uses Redis `DECRBY` for lock-free decrement under high concurrency
- **Redis Hash**: cart stored as `cart:{userId}` Hash with 24h TTL
- **Elasticsearch**: fuzzy full-text search with price range filter, auto-indexed from catalog on startup

## Infrastructure

```
PostgreSQL    transactional data (users, orders, inventory, payments)
Redis         cart sessions + flash sale counter + rate limiter
Kafka         async event bus (order.events, payment.events topics)
Kafka UI      http://localhost:8088
Elasticsearch product search index
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
cd services/auth      && PORT=8086 go run ./cmd/api/main.go
cd services/catalog   && PORT=8081 GRPC_PORT=9081 go run ./cmd/api/main.go
cd services/cart      && PORT=8082 go run ./cmd/api/main.go
cd services/order     && PORT=8083 go run ./cmd/api/main.go
cd services/inventory && PORT=8084 GRPC_PORT=9084 go run ./cmd/api/main.go
cd services/search    && PORT=8085 go run ./cmd/api/main.go
cd services/payment   && PORT=8087 go run ./cmd/api/main.go
```

**3. Run smoke tests:**

```bash
./scripts/smoke-flow.sh
```

## Key Flows

### Auth flow

```
POST /auth/register    create user (bcrypt password)
POST /auth/login       verify password, return JWT
GET  /auth/me          decode JWT, return user info
```

### Checkout flow

```
POST /cart/items       add product to Redis cart (JWT required)
POST /orders           order service fetches cart (forwards JWT), creates order + outbox in TX,
                       reserves inventory via gRPC, clears cart
PATCH /orders/:id/confirm  PENDING -> CONFIRMED, deduct stock permanently
PATCH /orders/:id/fail     PENDING -> FAILED, release reserved stock
```

### Payment flow

```
POST /payments                   create payment record (PENDING)
POST /payments/:id/callback      mock webhook: updates status + writes outbox event
                                 Kafka relay publishes to payment.events
                                 order consumer drives PENDING -> CONFIRMED / FAILED
```

### Flash sale flow

```
POST /inventory/flash-sale/init       initialize Redis counter with stock qty
POST /inventory/flash-sale/reserve    atomic DECRBY, rollback if negative (no oversell)
GET  /inventory/flash-sale/stock/:id  current remaining counter
```

### Search

```bash
curl "http://localhost:8085/search?q=iphone"
curl "http://localhost:8085/search?q=airpods&maxPrice=7000000"
curl "http://localhost:8085/search?q=iphoen"   # fuzzy, typo-tolerant
```

## Project Structure

```
ecom-poc/
  docs/        technical docs
  gen/         protobuf generated code
  infra/       docker-compose + infra configs
  pkg/         shared packages (jwtutil, ratelimit)
  proto/       .proto definitions
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
  client/       inter-service HTTP / gRPC clients
  grpc/         gRPC server implementation
  api/          controller, routes, dto, response
  config/       env-based config
```
