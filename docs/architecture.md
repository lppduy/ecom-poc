# Architecture

## Context

The project is an educational e-commerce POC focused on core ordering flow and system design trade-offs.

## Core Components

- `catalog`: product listing and lookup
- `cart`: cart item operations
- `order`: order creation lifecycle
- `auth`, `inventory`, `payment`: skeleton services for next phases

## Infrastructure

- PostgreSQL for transactional data
- Redis for cache/session/cart evolution
- Kafka (KRaft) for event-driven workflows
- Kafka UI for local topic inspection

## Next Architecture Milestones

1. Persist orders to PostgreSQL
2. Add inventory reservation and rollback flow
3. Introduce outbox event publishing
4. Add payment webhook + state transitions
