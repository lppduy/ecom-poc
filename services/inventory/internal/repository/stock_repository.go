package repository

import "github.com/lppduy/ecom-poc/services/inventory/internal/domain"

type StockRepository interface {
	// Reserve atomically checks and holds stock for all items in an order.
	// Idempotent: returns nil if reservation for orderId already exists.
	Reserve(orderID int64, items []domain.ReserveItem) error

	// Release undoes reserved stock for a FAILED order.
	Release(orderID int64) error

	// Confirm deducts stock permanently for a CONFIRMED order.
	Confirm(orderID int64) error

	// GetStock returns current stock for a product.
	GetStock(productID string) (domain.Stock, bool, error)

	// SeedDefaultsIfEmpty inserts initial stock rows if table is empty.
	SeedDefaultsIfEmpty() error
}
