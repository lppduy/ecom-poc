package service

import "github.com/lppduy/ecom-poc/services/inventory/internal/domain"

type InventoryService interface {
	Reserve(orderID int64, items []domain.ReserveItem) error
	Release(orderID int64) error
	Confirm(orderID int64) error
	GetStock(productID string) (domain.Stock, bool, error)
}
