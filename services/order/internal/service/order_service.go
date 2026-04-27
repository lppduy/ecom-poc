package service

import (
	"context"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, idempotencyKey string) (domain.Order, bool, error)
	GetOrder(id string) (domain.Order, bool, error)
	ConfirmOrder(id string) (domain.Order, error)
	FailOrder(id string) (domain.Order, error)
}
