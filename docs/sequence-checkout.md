# Checkout Sequence

## Happy Path

```
Client          catalog         cart            order
  |                |               |               |
  |-- GET /products -->            |               |
  |<-- [ list ] ---|               |               |
  |                                |               |
  |-- POST /cart/items ----------->|               |
  |<-- { items } -----------------|               |
  |                                               |
  |-- POST /orders (Idempotency-Key) ------------>|
  |                                |-- FetchCartItems(userId) -->|
  |                                |<-- [ items ] --------------|
  |                                |               |
  |                                |   BEGIN TX    |
  |                                |   INSERT orders           |
  |                                |   INSERT order_items      |
  |                                |   COMMIT                  |
  |                                |-- ClearCart(userId) ------>|
  |<-- 201 { id, status:PENDING } -|               |
  |                                                |
  |-- PATCH /orders/{id}/confirm ----------------->|
  |                                |   UPDATE status=CONFIRMED |
  |<-- 200 { id, status:CONFIRMED }|               |
```

## Idempotency Flow

```
Client                          order
  |                               |
  |-- POST /orders (key: K1) ---->|
  |                               |-- FindByIdempotencyKey(K1) --> not found
  |                               |-- CreateWithItems(...)
  |<-- 201 { id:42, PENDING } ----|
  |                               |
  |-- POST /orders (key: K1) ---->|   (retry / duplicate)
  |                               |-- FindByIdempotencyKey(K1) --> found
  |<-- 200 { id:42, PENDING } ----|   (same order, no re-create)
```

## Fail Path

```
Client                          order
  |                               |
  |-- POST /orders (key: K2) ---->|
  |<-- 201 { id:43, PENDING } ----|
  |                               |
  |-- PATCH /orders/43/fail ------>|
  |                               |-- UpdateStatus(43, FAILED)
  |<-- 200 { id:43, FAILED } -----|
  |                               |
  |-- PATCH /orders/43/fail ------>|   (invalid: FAILED → FAILED)
  |<-- 400 invalid transition -----|
```

## Order State Machine

```
         ┌─────────┐
         │ PENDING │
         └────┬────┘
       ┌──────┴──────┐
       ▼             ▼
 ┌──────────┐   ┌────────┐
 │CONFIRMED │   │ FAILED │
 └──────────┘   └────────┘
```

Allowed transitions:
- `PENDING → CONFIRMED`
- `PENDING → FAILED`

All other transitions return `400 invalid status transition`.
