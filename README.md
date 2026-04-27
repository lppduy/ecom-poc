# E-Commerce POC Blueprint (Java/Go Friendly)

Muc tieu cua tai lieu nay la gom toan bo nhung thu can build cho mot POC e-commerce "toan dien", de ban nho ro system design qua thuc hanh.

## 1) Muc tieu POC

- Xay duoc luong mua hang hoan chinh: browse -> cart -> checkout -> payment -> order -> shipment update.
- Co event-driven flow toi thieu (de hoc eventual consistency, retry, idempotency).
- Co cache, search, auth, observability o muc du de hoc.
- Co tai lieu trade-off va backlog de nang cap tiep.

## 2) Chon stack de build nhanh

Ban dang can nhac Java hoac Go, tai lieu nay giu neutral:

- Backend:
  - Go: `Gin/Fiber` + `sqlc/GORM` + `kafka-go` (goi y neu uu tien toc do POC)
  - Java: `Spring Boot` + `Spring Data JPA` + `Spring Kafka` (goi y neu uu tien enterprise pattern)
- Frontend:
  - Option nhanh: `Next.js` (UI + BFF)
  - Option toi gian: 1 web app bat ky + Swagger/Postman de test
- Database:
  - `PostgreSQL` (OLTP)
  - `Redis` (cart/cache/session/rate limit)
  - `OpenSearch/Elasticsearch` (search)
- Event bus:
  - `Kafka` (hoac RabbitMQ neu muon de van hanh hon cho POC)
- Infra local:
  - `Docker Compose`
- Observability:
  - Bat dau: structured logs + health/readiness
  - Nang cao: Prometheus + Grafana + OpenTelemetry

## 3) Kien truc tong quan can co

- API Gateway/BFF
- Service layer:
  - User/Auth Service
  - Catalog Service
  - Search Service
  - Cart Service
  - Pricing/Promotion Service
  - Inventory Service
  - Order Service
  - Payment Service
  - Shipping Service (mock)
  - Notification Service
- Data layer:
  - Postgres cho order/payment/user/catalog core
  - Redis cho cart/cache
  - Search index cho full-text + faceting
- Async layer:
  - Message broker + DLQ + retry policy

## 4) Domain va data model toi thieu

- `users`
- `products`
- `skus`
- `inventory` (`available`, `reserved`, `sold`, `version`)
- `carts`, `cart_items`
- `orders`, `order_items`
- `payments`
- `shipments`
- `promotions`, `coupons`
- `outbox_events` (de dam bao phat event an toan)

## 5) API phai co (toi thieu)

### Auth
- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`

### Catalog/Search
- `GET /products`
- `GET /products/{id}`
- `GET /search?q=...&filters=...`

### Cart
- `GET /cart`
- `POST /cart/items`
- `PATCH /cart/items/{itemId}`
- `DELETE /cart/items/{itemId}`

### Checkout/Order
- `POST /checkout/preview` (re-price + validate)
- `POST /orders` (tao order pending + reserve stock)
- `GET /orders/{id}`

### Payment
- `POST /payments/create-intent`
- `POST /payments/webhook` (provider callback)

### Shipping/Tracking
- `POST /shipments`
- `GET /shipments/{orderId}`

## 6) Luong nghiep vu quan trong

### Luong A: Place order (critical)
1. Client goi checkout preview -> tinh gia cuoi + apply promo.
2. Tao order `PENDING`.
3. Reserve inventory.
4. Tao payment intent.
5. Nhan webhook thanh cong -> order `CONFIRMED`.
6. Tru ton kho chinh thuc, day event `order.confirmed`.
7. Tao shipment (mock) va gui notification.

### Luong B: Payment fail
1. Payment fail/timeout.
2. Order ve `FAILED`/`CANCELLED`.
3. Release reservation.
4. Ghi log + metric.

### Luong C: Rebuild search index
1. Catalog update.
2. Phat event `product.updated`.
3. Search consumer cap nhat index.

## 7) Tinh nang ky thuat bat buoc de hoc dung system design

- Idempotency key cho payment/order create.
- Optimistic locking cho inventory (`version` field).
- Outbox pattern cho event publishing.
- Retry + exponential backoff + DLQ.
- Correlation ID xuyen request/event.
- Structured logging (JSON).
- Basic rate limiting cho auth/checkout/payment.

## 8) Reliability, consistency, security

- Strong consistency:
  - payment status, order status, inventory reserve/commit.
- Eventual consistency:
  - search index, analytics, recommendation.
- Security:
  - JWT + RBAC
  - hash password (Argon2/bcrypt)
  - TLS khi deploy that
  - khong luu card data raw (dung provider tokenization)

## 9) Roadmap build de xong POC

## Phase 0 - Setup project (1-2 ngay)
- Tao mono-repo hoac multi-repo.
- Docker compose: Postgres, Redis, Kafka, Kafka UI, OpenSearch.
- Bootstrap service skeleton + shared libs (`errors`, `logger`, `config`, `middleware`).

## Phase 1 - Core order flow (4-6 ngay)
- Auth, Catalog, Cart, Inventory, Order, Payment (mock).
- Hoan thanh luong place order thanh cong + fail.
- Viet test integration cho checkout/payment webhook.

## Phase 2 - Search + Promotion + Shipping (3-5 ngay)
- Dong bo catalog -> search.
- Facet/filter/sort.
- Promotion rule co ban.
- Shipping mock + tracking.

## Phase 3 - Hardening (3-5 ngay)
- Outbox, retry, DLQ.
- Metrics + dashboard + tracing co ban.
- Load test nhe cho browse/search/checkout.

## 10) Definition of Done (POC)

POC duoc coi la xong khi:

- Co demo duoc 3 luong:
  - mua hang thanh cong
  - mua hang fail va rollback
  - cap nhat san pham va tim kiem thay doi
- Co tai lieu:
  - architecture diagram
  - sequence diagram checkout
  - list trade-off da chon
- Co test:
  - unit test cho domain rules
  - integration test cho order-payment-inventory
- Co observability co ban:
  - log co correlation id
  - dashboard metric chinh

## 11) Thu tu uu tien neu ban bi qua tai

Neu khong du thoi gian, uu tien theo thu tu:

1. Auth + Catalog + Cart + Order + Payment webhook
2. Inventory reservation + idempotency
3. Outbox + retry + DLQ
4. Search index
5. Promotion + shipping + notification

## 12) De xuat cau truc folder (goi y)

```txt
ecom-poc/
  docs/
    architecture.md
    sequence-checkout.md
    tradeoffs.md
    runbook.md
  services/
    api-gateway/
    auth/
    catalog/
    cart/
    inventory/
    order/
    payment/
    shipping/
    search/
    notification/
  infra/
    docker-compose.yml
    monitoring/
  scripts/
```

## 13) Checklist tiep theo ngay bay gio

- Chot stack cu the (Go hoac Java).
- Tao skeleton folder theo muc 12.
- Khoi dong infra local bang docker compose.
- Implement luong place order happy path truoc.
- Them payment webhook + rollback flow.
- Sau cung moi them search/promotion/observability.

---

Neu can, ban co the tao them file `PLAN-4-WEEKS.md` de theo doi tien do theo tuan va checklist tung ngay.

## 14) Tracking files (da tao)

- `PLAN-4-WEEKS.md`: Roadmap 4 tuan theo deliverables + learning goals.
- `DAY-1-CHECKLIST.md`: Checklist chia theo session de bat dau ngay.
- `BUILD-LOG.md`: Nhat ky build/hoc de tu review moi ngay.
