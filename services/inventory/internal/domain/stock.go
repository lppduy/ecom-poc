package domain

import (
	"errors"
	"time"
)

var (
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrReservationNotFound  = errors.New("reservation not found")
	ErrProductNotFound      = errors.New("product not found in inventory")
)

const (
	ReservationStatusReserved  = "RESERVED"
	ReservationStatusConfirmed = "CONFIRMED"
	ReservationStatusReleased  = "RELEASED"
)

// Stock tracks total quantity and how much is reserved for pending orders.
// Available = Quantity - Reserved.
type Stock struct {
	ProductID string    `gorm:"primaryKey;type:text"`
	Quantity  int       `gorm:"not null;default:0"`
	Reserved  int       `gorm:"not null;default:0"`
	UpdatedAt time.Time
}

func (s Stock) Available() int {
	return s.Quantity - s.Reserved
}

// Reservation records a stock hold for a specific order + product.
// One row per (order_id, product_id) pair.
type Reservation struct {
	ID        int64     `gorm:"primaryKey"`
	OrderID   int64     `gorm:"uniqueIndex:idx_order_product;not null"`
	ProductID string    `gorm:"uniqueIndex:idx_order_product;type:text;not null"`
	Quantity  int       `gorm:"not null"`
	Status    string    `gorm:"type:text;not null"`
	CreatedAt time.Time
}

type ReserveItem struct {
	ProductID string
	Quantity  int
}
