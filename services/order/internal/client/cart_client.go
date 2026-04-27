package client

import "github.com/lppduy/ecom-poc/services/order/internal/domain"

type CartClient interface {
	FetchCartItems(userID string) ([]domain.OrderItem, error)
	ClearCart(userID string) error
}
