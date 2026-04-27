# Day 1 Checklist - Start Fast (Go)

Muc tieu ngay 1: chay duoc he thong nen + tao order pending + co tracking ro rang.

## Session A (60-90 phut): Setup skeleton

- [x] Tao folder:
  - [x] `services/`
  - [x] `infra/`
  - [x] `docs/`
  - [x] `scripts/`
- [x] Tao service roots:
  - [x] `services/auth`
  - [x] `services/catalog`
  - [x] `services/cart`
  - [x] `services/inventory`
  - [x] `services/order`
  - [x] `services/payment`
- [ ] Moi service co:
  - [x] `cmd/main.go`
  - [x] `internal/`
  - [ ] `.env.example`
  - [x] endpoint `/health`

## Session B (60-90 phut): Infra local

- [x] Viet `infra/docker-compose.yml` voi:
  - [x] postgres
  - [x] redis
  - [x] kafka
  - [x] kafka-ui
- [x] Chay compose thanh cong.
- [ ] Verify ket noi tu service toi postgres/redis.

## Session C (90-120 phut): Core API minimum

- [x] `GET /products` (mock data duoc).
- [x] `POST /cart/items` (luu tam in-memory hoac redis).
- [x] `POST /orders`:
  - [ ] validate cart
  - [x] tao order `PENDING`
  - [ ] luu vao postgres

## Session D (30-45 phut): Tracking + learning

- [ ] Cap nhat `BUILD-LOG.md`:
  - [ ] Hom nay da xong gi
  - [ ] Gap van de gi
  - [ ] Bai hoc system design rut ra
- [ ] Ve sequence checkout ban dau (5-7 buoc) vao `docs/sequence-checkout.md`.

## Exit criteria Day 1

- [ ] Tat ca service len duoc va co `/health`.
- [ ] Co endpoint tao order `PENDING`.
- [ ] Co it nhat 1 request end-to-end duoc luu log day du.
- [ ] Da cap nhat log hoc tap trong ngay.

## Neu bi tac, xu ly theo thu tu

1. Giam scope: bo payment/shipping ngay 1.
2. Dung mock data cho catalog.
3. Dung in-memory cart neu redis loi.
4. Van giu muc tieu order `PENDING` la bat buoc.
