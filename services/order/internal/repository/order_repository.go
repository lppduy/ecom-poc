package repository

import (
	"context"
	"errors"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")

type OrderRepository interface {
	FindByID(id string) (domain.Order, bool, error)
	FindByIdempotencyKey(key string) (domain.Order, bool, error)
	CreateWithItems(ctx context.Context, userID, idempotencyKey string, items []domain.OrderItem) (domain.Order, error)
	UpdateStatus(id int64, newStatus string) error
}
