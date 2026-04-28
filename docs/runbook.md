# Runbook

## Start Infrastructure

```bash
cd infra
docker compose up -d
docker compose ps
```

Services: `postgres`, `redis`, `kafka`, `kafka-ui`, `elasticsearch`

> Elasticsearch takes ~30-40s to become healthy on first start.

## Run Services

All 7 services with hot-reload (Air):

```bash
./scripts/dev-air.sh
```

Or individually:

```bash
cd services/auth      && PORT=8086 go run ./cmd/api/main.go
cd services/catalog   && PORT=8081 GRPC_PORT=9081 go run ./cmd/api/main.go
cd services/cart      && PORT=8082 go run ./cmd/api/main.go
cd services/order     && PORT=8083 go run ./cmd/api/main.go
cd services/inventory && PORT=8084 GRPC_PORT=9084 go run ./cmd/api/main.go
cd services/search    && PORT=8085 go run ./cmd/api/main.go
cd services/payment   && PORT=8087 go run ./cmd/api/main.go
```

## Health Checks

```bash
for port in 8086 8081 8082 8083 8084 8085 8087; do
  echo -n "port $port: "; curl -s http://localhost:$port/health
done
```

## Automated Smoke Test

Runs assertions covering the full flow end-to-end (33 checks):

```bash
./scripts/smoke-flow.sh
```

Covers: auth JWT, catalog gRPC stream, cart token forwarding, order idempotency, payment Kafka outbox, flash sale atomic counter, auth guard (401).

## Manual Flow

### 1. Register and login

```bash
curl -X POST http://localhost:8086/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'

TOKEN=$(curl -s -X POST http://localhost:8086/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")
```

All subsequent requests require `-H "Authorization: Bearer $TOKEN"`.

### 2. Browse products

```bash
curl -s http://localhost:8081/products | jq
```

### 3. Add item to cart

```bash
curl -X POST http://localhost:8082/cart/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"productId":"sku_iphone_15_128","quantity":2}'
```

### 4. Create order

```bash
ORDER=$(curl -s -X POST http://localhost:8083/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: order-$(date +%s)" \
  -d '{}')

ORDER_ID=$(echo $ORDER | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
```

Order fetches cart via JWT forwarding, reserves stock via gRPC, clears cart.

### 5. Create payment

```bash
PAY=$(curl -s -X POST http://localhost:8087/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"orderId\":\"$ORDER_ID\",\"amount\":999000}")

PAY_ID=$(echo $PAY | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
```

### 6. Trigger payment callback (mock webhook)

```bash
# success -> Kafka event -> order becomes CONFIRMED
curl -X POST http://localhost:8087/payments/$PAY_ID/callback \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"result":"success"}'

# wait ~5s for outbox relay + Kafka consumer
sleep 5 && curl -s http://localhost:8083/orders/$ORDER_ID \
  -H "Authorization: Bearer $TOKEN" | jq .status
```

### 7. Check inventory stock

```bash
curl -s http://localhost:8084/inventory/stock/sku_iphone_15_128 | jq
# returns: quantity, reserved, available
```

### 8. Flash sale

```bash
# initialize stock in Redis
curl -X POST http://localhost:8084/inventory/flash-sale/init \
  -H "Content-Type: application/json" \
  -d '{"productId":"flash-001","quantity":100}'

# atomic reserve (Black Friday scenario)
curl -X POST http://localhost:8084/inventory/flash-sale/reserve \
  -H "Content-Type: application/json" \
  -d '{"productId":"flash-001","quantity":1}'

# check remaining
curl -s http://localhost:8084/inventory/flash-sale/stock/flash-001 | jq
```

### 9. Search products

```bash
curl "http://localhost:8085/search?q=iphone"
curl "http://localhost:8085/search?q=airpods&maxPrice=7000000"
curl "http://localhost:8085/search?q=iphoen"   # fuzzy, typo-tolerant
```

### 10. Reindex search (re-fetch from catalog via gRPC stream)

```bash
curl -X POST http://localhost:8085/search/reindex
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Port conflict on startup | `lsof -ti :<port> \| xargs kill -9` |
| Docker not running | Start Docker Desktop before `docker compose up` |
| Kafka UI empty | `docker compose ps` - check kafka broker health |
| `401 Unauthorized` | Include `-H "Authorization: Bearer $TOKEN"` |
| `cart is empty` on create order | Add item via `POST /cart/items` first |
| `Idempotency-Key header is required` | Include `-H "Idempotency-Key: <unique>"` |
| ES search returns empty | Call `POST /search/reindex` or wait for startup reindex |
| ES not ready | Wait ~40s, check `GET localhost:9200/_cluster/health` |
| Order stuck at PENDING after payment | Wait ~5s for Kafka relay + consumer; check `docker compose ps kafka` |
| gRPC connection refused | Check inventory `:9084` and catalog `:9081` are up |
