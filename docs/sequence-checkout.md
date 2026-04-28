# Sequence Diagrams

## Full Checkout Flow

```
User          auth     catalog    cart      order     inventory   payment    Kafka
 |              |         |         |          |           |          |        |
 |-- register ->|         |         |          |           |          |        |
 |-- login ---->|         |         |          |           |          |        |
 |<-- JWT ------|         |         |          |           |          |        |
 |                        |         |          |           |          |        |
 |-- GET /products ------>|         |          |           |          |        |
 |<-- [list] -------------|         |          |           |          |        |
 |                                  |          |           |          |        |
 |-- POST /cart/items [JWT] ------->|          |           |          |        |
 |<-- { items } --------------------|          |           |          |        |
 |                                             |           |          |        |
 |-- POST /orders [JWT + Idem-Key] ----------->|           |          |        |
 |                                 JWT forward |           |          |        |
 |                                  |<-- GET /cart --------|          |        |
 |                                  |-- [items] ---------->|          |        |
 |                                             |           |          |        |
 |                                             |-- Reserve(gRPC) ---->|        |
 |                                             |           |  BEGIN TX|        |
 |                                             |           |  SELECT FOR UPDATE|
 |                                             |           |  UPDATE stocks    |
 |                                             |           |  COMMIT  |        |
 |                                             |<-- OK --------------|         |
 |                                             |                               |
 |                                             |  BEGIN TX                     |
 |                                             |  INSERT orders                |
 |                                             |  INSERT order_items           |
 |                                             |  INSERT outbox_events         |
 |                                             |  COMMIT                       |
 |                                  |<-- POST /cart/clear  |          |        |
 |<-- 201 { id, PENDING } ----------|          |           |          |        |
 |                                             |                               |
 |                                             |  [relay goroutine, every 3s]  |
 |                                             |- - - - - - - - ->order.events |
 |                                                                             |
 |-- POST /payments [JWT] ------------------------------------------------->|  |
 |<-- { id, PENDING } ------------------------------------------------------|  |
 |                                                                             |
 |-- POST /payments/:id/callback [JWT] ------------------------------------>|  |
 |                                        BEGIN TX                          |  |
 |                                        UPDATE payments SET status=SUCCESS|  |
 |                                        INSERT payment_outbox_events       |  |
 |                                        COMMIT                            |  |
 |<-- { status: SUCCESS } --------------------------------------------------|  |
 |                                                                             |
 |                              [relay goroutine, every 3s]                   |
 |                              - - - - - - - - - - - ->payment.events        |
 |                                                                         |   |
 |             [order Kafka consumer]                                      |   |
 |             <-- payment.events ----------------------------------------|   |
 |             status=SUCCESS --> ConfirmOrder(orderID)                        |
 |                                             |-- Confirm(gRPC) ------->|     |
 |                                             |  UPDATE quantity -= qty  |     |
 |                                             |  UPDATE reserved -= qty  |     |
 |                                             |<-- OK ------------------|     |
 |                                             |  UPDATE order status=CONFIRMED|
```

## Idempotency

```
User                             order
 |                                  |
 |-- POST /orders (key: K1) ------->|
 |                                  |-- FindByIdempotencyKey(K1) -> not found
 |                                  |-- CreateWithItems(...)
 |<-- 201 { id:7, PENDING } --------|
 |                                  |
 |-- POST /orders (key: K1) ------->|   (retry / duplicate)
 |                                  |-- FindByIdempotencyKey(K1) -> found
 |<-- 200 { id:7, PENDING } --------|   (same order, no re-create)
```

## Order State Machine

```
         +---------+
         | PENDING |
         +----+----+
       +------+------+
       v             v
 +----------+   +--------+
 |CONFIRMED |   | FAILED |
 +----------+   +--------+
```

Transitions:
- `PENDING -> CONFIRMED` (via payment SUCCESS Kafka event)
- `PENDING -> FAILED` (via payment FAILED Kafka event)

All other transitions return `400 invalid status transition`.

## Payment -> Order via Kafka Outbox

```
payment service                 Kafka              order service
      |                           |                      |
      | callback(success)         |                      |
      |                           |                      |
      | BEGIN TX                  |                      |
      | UPDATE payments           |                      |
      | INSERT payment_outbox     |                      |
      | COMMIT                    |                      |
      |                           |                      |
      | [relay, every 3s]         |                      |
      |-- payment.events -------->|                      |
      |                           |-- ReadMessage() ---->|
      |                           |                      | ConfirmOrder(id)
      |                           |                      | UPDATE CONFIRMED
```

If relay fails or service crashes: event stays in `payment_outbox_events`, relay retries on next tick.

## gRPC Unary - order -> inventory

```
order service                       inventory gRPC :9084
      |                                     |
      |-- Reserve(orderID, items) --------->|
      |                                     | BEGIN TX
      |                                     | SELECT stocks FOR UPDATE
      |                                     | CHECK available >= qty
      |                                     | UPDATE reserved += qty
      |                                     | INSERT reservations
      |                                     | COMMIT
      |<-- { success: true } --------------|
```

## gRPC Server-Side Streaming - search -> catalog

```
search service                      catalog gRPC :9081
      |                                     |
      |-- StreamProducts({}) ------------->|
      |                                     | SELECT * FROM products
      |<-- Product{id:1, name:..., ...} ---|
      |<-- Product{id:2, name:..., ...} ---|
      |      ... (N messages)              |
      |<-- EOF -----------------------------|
      |                                     |
      | BulkIndex(products) -> Elasticsearch
```

Streaming lets catalog serve thousands of products without loading all into memory at once.

## Flash Sale Flow (Redis Atomic)

```
User A          User B          User C          Redis
  |               |               |               |
  |-- reserve(1) >|               |               |
  |               |-- reserve(1) >|               |
  |               |               |-- reserve(1) >|
  |               |               |               |
  |               |               | DECRBY flash:{id} 1  (atomic, sequential)
  |               |               | DECRBY flash:{id} 1
  |               |               | DECRBY flash:{id} 1
  |               |               |               |
  |               |               |  stock: 3->2->1->0
  |               |               |
  |-- reserve(1) >|               |               | DECRBY -> -1 (< 0)
  |               |               |               | INCRBY (rollback)
  |<-- sold out --|               |               |
```

No DB locks needed. Redis single-threaded model ensures exactly one winner per decrement.
