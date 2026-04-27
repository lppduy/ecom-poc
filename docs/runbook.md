# Runbook

## Start Infrastructure

```bash
cd infra
docker compose up -d
docker compose ps
```

## Run Services

```bash
cd services/catalog && PORT=8081 go run ./cmd/main.go
cd services/cart && PORT=8082 go run ./cmd/main.go
cd services/order && PORT=8083 go run ./cmd/main.go
```

## Smoke Tests

```bash
curl -s http://localhost:8081/health
curl -s http://localhost:8081/products | jq
curl -i -X POST http://localhost:8082/cart/items -H "Content-Type: application/json" -d '{"productId":"sku_iphone_15_128","quantity":1}'
curl -i -X POST http://localhost:8083/orders -H "Content-Type: application/json" -d '{"userId":"u_001"}'
```

## Troubleshooting

- Port conflict: change `PORT` env and retry
- Docker not running: start Docker Desktop before compose
- Kafka UI empty: check broker status with `docker compose ps`
