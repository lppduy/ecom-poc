# E-Commerce POC (Go) - Public Overview

This repository is a hands-on learning project for building an end-to-end e-commerce system with production-minded architecture.

## Current Scope

- Service skeletons: `auth`, `catalog`, `cart`, `inventory`, `order`, `payment`
- Local infrastructure: PostgreSQL, Redis, Kafka (KRaft), Kafka UI
- Day-1 APIs implemented:
  - `GET /products` (catalog mock data)
  - `POST /cart/items` (in-memory cart)
  - `POST /orders` (creates `PENDING` order)

## Project Structure

```txt
ecom-poc/
  docs/                  public technical docs
  docs-private/          local-only notes (gitignored)
  infra/                 docker compose and infra configs
  services/              Go microservice skeletons
  scripts/               helper scripts
```

## Quick Start

1. Start infrastructure:

```bash
cd infra
docker compose up -d
```

2. Run services (separate terminals):

```bash
cd services/catalog && PORT=8081 go run ./cmd/main.go
cd services/cart && PORT=8082 go run ./cmd/main.go
cd services/order && PORT=8083 go run ./cmd/main.go
```

3. Test endpoints:

```bash
curl -s http://localhost:8081/products | jq
curl -i -X POST http://localhost:8082/cart/items -H "Content-Type: application/json" -d '{"productId":"sku_iphone_15_128","quantity":1}'
curl -i -X POST http://localhost:8083/orders -H "Content-Type: application/json" -d '{"userId":"u_001"}'
```

## Public vs Private Docs

See `DOCS-VISIBILITY.md` for what should remain public and what should stay local.
