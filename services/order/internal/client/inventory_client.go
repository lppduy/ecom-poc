package client

import "github.com/lppduy/ecom-poc/services/order/internal/domain"

type InventoryClient interface {
	Reserve(orderID int64, items []domain.OrderItem) error
	Release(orderID int64) error
	Confirm(orderID int64) error
}
