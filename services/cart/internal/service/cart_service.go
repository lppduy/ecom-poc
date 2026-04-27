package service

import "github.com/lppduy/ecom-poc/services/cart/internal/domain"

type CartService interface {
	AddItem(userID, productID string, quantity int) error
	GetItems(userID string) ([]domain.CartItem, error)
	Clear(userID string) error
}
