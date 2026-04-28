package client

import (
	"context"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type CartClient interface {
	FetchCartItems(ctx context.Context, userID string) ([]domain.OrderItem, error)
	ClearCart(ctx context.Context, userID string) error
}
