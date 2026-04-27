# E-Commerce POC - Plan 4 Weeks (Go Track)

Tai lieu nay de ban theo doi tien do theo tuan, vua build vua hoc system design.

## Muc tieu 4 tuan

- Xong luong mua hang end-to-end.
- Co event flow co ban va xu ly fail an toan.
- Co search, observability, tai lieu trade-off.
- Co demo va test can ban.

## Cach dung file nay

- Moi ngay, cap nhat trang thai task: `[ ]` -> `[x]`.
- Ghi ket qua vao `BUILD-LOG.md`.
- Neu bi block > 1 ngay, ghi vao `Risks/Blockers`.

## Week 1 - Foundation + Core Happy Path

### Deliverables
- [ ] Khoi tao workspace POC va folder structure.
- [ ] Docker compose chay duoc `postgres`, `redis`, `kafka`, `kafka-ui`.
- [ ] Service skeleton: `auth`, `catalog`, `cart`, `inventory`, `order`, `payment`.
- [ ] API happy path:
  - [ ] `GET /products`
  - [ ] `POST /cart/items`
  - [ ] `POST /orders` -> tao order `PENDING`
- [ ] Payment mock webhook success -> order `CONFIRMED`.
- [ ] Demo script happy path bang curl/Postman.

### Learning goals
- Hieu boundary giua service.
- Hieu vi sao order + payment + inventory la critical path.
- Hieu idempotency can dat o dau.

## Week 2 - Failure Handling + Consistency

### Deliverables
- [ ] Inventory reservation (`available`, `reserved`, `sold`).
- [ ] Optimistic locking cho inventory.
- [ ] Payment fail flow:
  - [ ] Order `FAILED/CANCELLED`
  - [ ] Release reservation
- [ ] Idempotency key cho `POST /orders` va payment webhook.
- [ ] Outbox table + publisher worker.
- [ ] Retry + DLQ cho consumer.

### Learning goals
- Strong consistency vs eventual consistency trong checkout.
- Trade-off outbox pattern.
- Du lieu nao bat buoc "exactly-once effect" (khong phai exactly-once delivery).

## Week 3 - Search + Promotion + Shipping

### Deliverables
- [ ] Dong bo product vao OpenSearch.
- [ ] `GET /search` co query + filter + sort.
- [ ] Promotion rule co ban (percent/fixed coupon).
- [ ] Shipping mock:
  - [ ] Tao shipment
  - [ ] Tracking status
- [ ] Notification async (email mock/log).

### Learning goals
- Tai sao search nen eventual consistency.
- Duong di cua event `product.updated`.
- Chien luoc cache cho read-heavy APIs.

## Week 4 - Hardening + Demo + Docs

### Deliverables
- [ ] Correlation ID xuyen HTTP + events.
- [ ] Structured logging JSON.
- [ ] Metrics: request latency, error rate, order success rate.
- [ ] Dashboard co ban (hoac it nhat metric endpoint + screenshot).
- [ ] Integration tests:
  - [ ] order-payment-inventory happy path
  - [ ] payment fail rollback
- [ ] Docs:
  - [ ] `docs/architecture.md`
  - [ ] `docs/sequence-checkout.md`
  - [ ] `docs/tradeoffs.md`
  - [ ] `docs/runbook.md`
- [ ] Final demo script.

### Learning goals
- Observability giai quyet su co nhu the nao.
- Cac trade-off da chon co hop ly khong.
- Tu danh gia POC da "production-thinking" den muc nao.

## Risks / Blockers

- [ ] Chua quen Kafka/OpenSearch tooling.
- [ ] Pham vi qua rong, de "sa da" vao optimize som.
- [ ] Co the thieu thoi gian cho test va docs.

## Scope guard (de tranh over-engineering)

- Khong can tach qua nhieu service ngay tu dau.
- Khong can auth provider that, dung local JWT.
- Khong can UI dep, uu tien flow dung va quan sat duoc.
- Khong can Kubernetes cho POC local.

## Definition of Done

- [ ] Demo 3 luong:
  - [ ] Happy path checkout
  - [ ] Payment fail rollback
  - [ ] Product update -> search update
- [ ] Test pass cho core integration flows.
- [ ] Docs du de nguoi khac clone va chay.
