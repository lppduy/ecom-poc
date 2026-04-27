package domain

import "time"

type OutboxEvent struct {
	ID          int64      `gorm:"primaryKey"`
	EventType   string     `gorm:"type:text;not null"`
	Payload     string     `gorm:"type:text;not null"`
	PublishedAt *time.Time `gorm:"index"`
	CreatedAt   time.Time
}
