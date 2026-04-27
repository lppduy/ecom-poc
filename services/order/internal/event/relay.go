package event

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lppduy/ecom-poc/services/order/internal/repository"
)

const (
	orderEventsTopic = "order.events"
	relayBatchSize   = 10
	relayInterval    = 3 * time.Second
)

// Relay reads unpublished outbox events and forwards them to Kafka.
// Runs as a background goroutine until ctx is cancelled.
func StartRelay(ctx context.Context, outbox repository.OutboxRepository, pub Publisher) {
	go func() {
		ticker := time.NewTicker(relayInterval)
		defer ticker.Stop()
		log.Println("[relay] started")
		for {
			select {
			case <-ctx.Done():
				log.Println("[relay] stopped")
				return
			case <-ticker.C:
				if err := runOnce(ctx, outbox, pub); err != nil {
					log.Printf("[relay] error: %v", err)
				}
			}
		}
	}()
}

func runOnce(ctx context.Context, outbox repository.OutboxRepository, pub Publisher) error {
	events, err := outbox.FetchPending(relayBatchSize)
	if err != nil {
		return fmt.Errorf("fetch pending: %w", err)
	}
	for _, e := range events {
		if err := pub.Publish(ctx, orderEventsTopic, e.EventType, e.Payload); err != nil {
			return fmt.Errorf("publish event %d: %w", e.ID, err)
		}
		if err := outbox.MarkPublished(e.ID); err != nil {
			log.Printf("[relay] warn: failed to mark event %d published: %v", e.ID, err)
		}
		log.Printf("[relay] published event %d type=%s", e.ID, e.EventType)
	}
	return nil
}
