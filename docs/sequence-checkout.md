# Checkout Sequence

## Happy Path (current + target)

1. Client fetches products from `catalog`
2. Client adds item to `cart`
3. Client submits order request to `order`
4. `order` creates `PENDING` order
5. (Next) payment success updates order to `CONFIRMED`

## Failure Path (target)

1. Payment fails or times out
2. Order transitions to `FAILED` or `CANCELLED`
3. Reserved inventory is released
4. Failure event is logged and emitted
