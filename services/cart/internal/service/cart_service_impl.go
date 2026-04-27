package service

import (
	"errors"

	"github.com/lppduy/ecom-poc/services/cart/internal/domain"
	"github.com/lppduy/ecom-poc/services/cart/internal/repository"
)

var ErrInvalidQuantity = errors.New("quantity must be greater than 0")
var ErrMissingProductID = errors.New("productId is required")

type DefaultCartService struct {
	repo repository.CartRepository
}

func NewCartService(repo repository.CartRepository) *DefaultCartService {
	return &DefaultCartService{repo: repo}
}

func (s *DefaultCartService) AddItem(userID, productID string, quantity int) error {
	if productID == "" {
		return ErrMissingProductID
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return s.repo.AddItem(userID, productID, quantity)
}

func (s *DefaultCartService) GetItems(userID string) ([]domain.CartItem, error) {
	return s.repo.ListItems(userID)
}

func (s *DefaultCartService) Clear(userID string) error {
	return s.repo.Clear(userID)
}
