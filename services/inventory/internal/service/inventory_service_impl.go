package service

import (
	"github.com/lppduy/ecom-poc/services/inventory/internal/domain"
	"github.com/lppduy/ecom-poc/services/inventory/internal/repository"
)

type DefaultInventoryService struct {
	repo repository.StockRepository
}

func NewInventoryService(repo repository.StockRepository) *DefaultInventoryService {
	return &DefaultInventoryService{repo: repo}
}

func (s *DefaultInventoryService) Reserve(orderID int64, items []domain.ReserveItem) error {
	return s.repo.Reserve(orderID, items)
}

func (s *DefaultInventoryService) Release(orderID int64) error {
	return s.repo.Release(orderID)
}

func (s *DefaultInventoryService) Confirm(orderID int64) error {
	return s.repo.Confirm(orderID)
}

func (s *DefaultInventoryService) GetStock(productID string) (domain.Stock, bool, error) {
	return s.repo.GetStock(productID)
}
