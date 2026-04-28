package event

import (
	"context"
	"encoding/json"
	"log"

	kafka "github.com/segmentio/kafka-go"
	"github.com/lppduy/ecom-poc/services/order/internal/service"
)

const paymentEventsTopic = "payment.events"

type paymentEvent struct {
	PaymentID string `json:"paymentId"`
	OrderID   string `json:"orderId"`
	Status    string `json:"status"` // "SUCCESS" | "FAILED"
}

// StartPaymentConsumer subscribes to "payment.events" and drives the order
// state machine (PENDING -> CONFIRMED or FAILED) based on payment outcome.
// Runs as a background goroutine until ctx is cancelled.
func StartPaymentConsumer(ctx context.Context, brokers []string, orderSvc service.OrderService) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    paymentEventsTopic,
		GroupID:  "order-payment-consumer",
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	go func() {
		defer r.Close()
		log.Println("[payment-consumer] started")
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Println("[payment-consumer] stopped")
					return
				}
				log.Printf("[payment-consumer] read error: %v", err)
				continue
			}

			var e paymentEvent
			if err := json.Unmarshal(m.Value, &e); err != nil {
				log.Printf("[payment-consumer] unmarshal error: %v", err)
				continue
			}

			log.Printf("[payment-consumer] received: orderID=%s status=%s", e.OrderID, e.Status)

			switch e.Status {
			case "SUCCESS":
				if _, err := orderSvc.ConfirmOrder(e.OrderID); err != nil {
					log.Printf("[payment-consumer] confirm order %s error: %v", e.OrderID, err)
				} else {
					log.Printf("[payment-consumer] order %s confirmed", e.OrderID)
				}
			case "FAILED":
				if _, err := orderSvc.FailOrder(e.OrderID); err != nil {
					log.Printf("[payment-consumer] fail order %s error: %v", e.OrderID, err)
				} else {
					log.Printf("[payment-consumer] order %s failed", e.OrderID)
				}
			default:
				log.Printf("[payment-consumer] unknown status %q for order %s", e.Status, e.OrderID)
			}
		}
	}()
}
