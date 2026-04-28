package event

import (
	"context"
	"encoding/json"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

const PaymentEventsTopic = "payment.events"

type PaymentEvent struct {
	PaymentID string `json:"paymentId"`
	OrderID   string `json:"orderId"`
	// Status: "SUCCESS" | "FAILED"
	Status string `json:"status"`
}

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) PublishPaymentEvent(ctx context.Context, e PaymentEvent) error {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal payment event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: PaymentEventsTopic,
		Key:   []byte(e.OrderID), // partition by order so events for same order land on same partition
		Value: b,
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
