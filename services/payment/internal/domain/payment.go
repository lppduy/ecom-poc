package domain

import (
	"errors"
	"time"
)

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrAlreadyProcessed   = errors.New("payment already processed")
	ErrInvalidOrderID     = errors.New("order id is required")
)

type Payment struct {
	ID        string    `json:"id"         gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID   string    `json:"order_id"   gorm:"uniqueIndex;not null"`
	Amount    float64   `json:"amount"     gorm:"not null"`
	Status    Status    `json:"status"     gorm:"not null;default:'PENDING'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
