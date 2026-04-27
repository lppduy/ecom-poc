# Runbook

## Start Infrastructure

```bash
cd infra
docker compose up -d
docker compose ps   # postgres, redis, kafka, kafka-ui should be healthy
```

## Run Services (with autoreload)

```bash
./scripts/dev-air.sh
```

Ports:
- `catalog` → `http://localhost:8081`
- `cart`    → `http://localhost:8082`
- `order`   → `http://localhost:8083`

## Health Check

```bash
curl -s http://localhost:8081/health
curl -s http://localhost:8082/health
curl -s http://localhost:8083/health
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

### 3. Create order (requires Idempotency-Key)
```bash
curl -X POST http://localhost:8083/orders \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-001" \
  -d '{"userId":"u_001"}'
```

### 4. Confirm or fail the order
```bash
curl -X PATCH http://localhost:8083/orders/{id}/confirm
curl -X PATCH http://localhost:8083/orders/{id}/fail
```

### 5. Get order
```bash
curl -s http://localhost:8083/orders/{id}
```

## Automated Smoke Test

```bash
./scripts/smoke-flow.sh
```

Runs 13 assertions covering the full flow end-to-end.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Port conflict | Change `PORT` env in `.air.toml` |
| Docker not running | Start Docker Desktop before compose |
| Kafka UI empty | `docker compose ps` — check broker health |
| `cart is empty` on create order | Add item via `POST /cart/items` first |
| `Idempotency-Key header is required` | Include `-H "Idempotency-Key: <unique>"` |
