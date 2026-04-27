# Trade-offs

## Current Decisions

- **In-memory cart for Day 1**
  - Fast iteration and easy debugging
  - Not durable, not multi-instance safe

- **Mock product list**
  - Enables API testing immediately
  - No real catalog persistence yet

- **KRaft Kafka setup**
  - Modern Kafka mode without Zookeeper
  - Slightly more setup detail than legacy mode

## Near-term Trade-offs

- Move from in-memory cart to Redis (durability vs complexity)
- Persist orders in PostgreSQL (consistency vs implementation speed)
- Add outbox pattern (reliability vs code overhead)
