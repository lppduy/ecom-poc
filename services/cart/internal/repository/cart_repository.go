package repository

import "github.com/lppduy/ecom-poc/services/cart/internal/domain"

type CartRepository interface {
	AddItem(userID, productID string, quantity int) error
	ListItems(userID string) ([]domain.CartItem, error)
	Clear(userID string) error
}
