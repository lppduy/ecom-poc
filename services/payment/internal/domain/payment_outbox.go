package domain

import "time"

// PaymentOutboxEvent is written in the same transaction as the payment status update,
// then relayed to Kafka by a background goroutine. This guarantees at-least-once delivery.
type PaymentOutboxEvent struct {
	ID          int64      `gorm:"primaryKey"`
	EventType   string     `gorm:"type:text;not null"`
	Payload     string     `gorm:"type:text;not null"`
	PublishedAt *time.Time `gorm:"index"`
	CreatedAt   time.Time
}
