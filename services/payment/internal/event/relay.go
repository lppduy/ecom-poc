package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
)

const (
	relayBatchSize = 10
	relayInterval  = 3 * time.Second
)

// StartRelay reads unpublished payment outbox events and forwards them to Kafka.
// Runs as a background goroutine until ctx is cancelled.
func StartRelay(ctx context.Context, outbox repository.PaymentOutboxRepository, pub *KafkaPublisher) {
	go func() {
		ticker := time.NewTicker(relayInterval)
		defer ticker.Stop()
		log.Println("[payment-relay] started")
		for {
			select {
			case <-ctx.Done():
				log.Println("[payment-relay] stopped")
				return
			case <-ticker.C:
				if err := relayOnce(ctx, outbox, pub); err != nil {
					log.Printf("[payment-relay] error: %v", err)
				}
			}
		}
	}()
}

func relayOnce(ctx context.Context, outbox repository.PaymentOutboxRepository, pub *KafkaPublisher) error {
	events, err := outbox.FetchPending(relayBatchSize)
	if err != nil {
		return err
	}
	for _, e := range events {
		var pe PaymentEvent
		if err := json.Unmarshal([]byte(e.Payload), &pe); err != nil {
			log.Printf("[payment-relay] bad payload for event %d: %v", e.ID, err)
			continue
		}
		if err := pub.PublishPaymentEvent(ctx, pe); err != nil {
			return err
		}
		if err := outbox.MarkPublished(e.ID); err != nil {
			log.Printf("[payment-relay] warn: failed to mark event %d published: %v", e.ID, err)
		}
		log.Printf("[payment-relay] published event %d orderID=%s status=%s", e.ID, pe.OrderID, pe.Status)
	}
	return nil
}
