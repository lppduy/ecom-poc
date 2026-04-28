# Runbook

## Start Infrastructure

```bash
cd infra
docker compose up -d
docker compose ps
```

Services started: `postgres`, `redis`, `kafka`, `kafka-ui`, `elasticsearch`

> Elasticsearch takes ~30-40s to become healthy on first start.

## Run Services

All services with hot-reload (Air):

```bash
./scripts/dev-air.sh
```

Or individually:

```bash
cd services/catalog   && PORT=8081 go run ./cmd/api/main.go
cd services/cart      && PORT=8082 go run ./cmd/api/main.go
cd services/order     && PORT=8083 go run ./cmd/api/main.go
cd services/inventory && PORT=8084 go run ./cmd/api/main.go
cd services/search    && PORT=8085 go run ./cmd/api/main.go
```

## Health Checks

```bash
curl -s http://localhost:8081/health
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health
curl -s http://localhost:8084/health
curl -s http://localhost:8085/health
```

## Manual Flow

### 1. Browse products

```bash
curl -s http://localhost:8081/products | jq
```

### 2. Add item to cart

```bash
curl -X POST http://localhost:8082/cart/items \
  -H "Content-Type: application/json" \
  -d '{"userId":"u_001","productId":"sku_iphone_15_128","quantity":2}'
```

### 3. Create order

```bash
curl -X POST http://localhost:8083/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: order-001" \
  -d '{"userId":"u_001"}'
```

Save the `id` from the response for next steps.

### 4. Confirm or fail the order

```bash
curl -X PATCH http://localhost:8083/orders/{id}/confirm
curl -X PATCH http://localhost:8083/orders/{id}/fail
```

### 5. Check inventory stock

```bash
curl -s http://localhost:8084/inventory/stock/sku_iphone_15_128 | jq
```

Returns `quantity`, `reserved`, `available`.

### 6. Search products

```bash
curl "http://localhost:8085/search?q=iphone"
curl "http://localhost:8085/search?q=airpods&maxPrice=7000000"
curl "http://localhost:8085/search?q=iphoen"
```

### 7. Reindex search (sync from catalog)

```bash
curl -X POST http://localhost:8085/search/reindex
```

## Automated Smoke Test

```bash
./scripts/smoke-flow.sh
```

Runs assertions covering the full checkout flow end-to-end.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Port conflict | Change `PORT` in `.air.toml` for that service |
| Docker not running | Start Docker Desktop before `docker compose up` |
| Kafka UI empty | `docker compose ps` - check broker health |
| `cart is empty` on create order | Add item via `POST /cart/items` first |
| `Idempotency-Key header is required` | Include `-H "Idempotency-Key: <unique>"` |
| ES search returns empty | Call `POST /search/reindex` to index products |
| ES not ready | Wait ~40s, check `GET localhost:9200/_cluster/health` |
