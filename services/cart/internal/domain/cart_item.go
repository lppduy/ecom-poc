package domain

import "time"

type CartItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"userId" gorm:"index;type:text;not null"`
	ProductID string    `json:"productId" gorm:"type:text;not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
