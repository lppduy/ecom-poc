package domain

import "errors"

var ErrEmptyCart = errors.New("cart is empty")

type Order struct {
	ID             int64  `json:"id" gorm:"primaryKey"`
	UserID         string `json:"userId" gorm:"type:text;not null"`
	Status         string `json:"status" gorm:"type:text;not null"`
	IdempotencyKey string `json:"-" gorm:"type:text;uniqueIndex;not null"`
}

type OrderItem struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	OrderID   int64  `json:"orderId" gorm:"index;not null"`
	ProductID string `json:"productId" gorm:"type:text;not null"`
	Quantity  int    `json:"quantity" gorm:"not null"`
}
