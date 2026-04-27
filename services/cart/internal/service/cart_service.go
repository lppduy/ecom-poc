package service

import (
	"errors"

	"github.com/lppduy/ecom-poc/services/cart/internal/domain"
	"github.com/lppduy/ecom-poc/services/cart/internal/repository"
)

var ErrInvalidQuantity = errors.New("quantity must be greater than 0")
var ErrMissingProductID = errors.New("productId is required")

type CartService struct {
	repo *repository.CartRepository
}

func NewCartService(repo *repository.CartRepository) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) AddItem(userID, productID string, quantity int) error {
	if productID == "" {
		return ErrMissingProductID
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	return s.repo.AddItem(userID, productID, quantity)
}

func (s *CartService) GetItems(userID string) ([]domain.CartItem, error) {
	return s.repo.ListItems(userID)
}

func (s *CartService) Clear(userID string) error {
	return s.repo.Clear(userID)
}
