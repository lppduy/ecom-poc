package service

import (
	"context"
	"errors"

	"github.com/lppduy/ecom-poc/services/order/internal/client"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"github.com/lppduy/ecom-poc/services/order/internal/repository"
)

type DefaultOrderService struct {
	repo       repository.OrderRepository
	cartClient client.CartClient
}

func NewOrderService(repo repository.OrderRepository, cartClient client.CartClient) *DefaultOrderService {
	return &DefaultOrderService{repo: repo, cartClient: cartClient}
}

func (s *DefaultOrderService) CreateOrder(ctx context.Context, userID, idempotencyKey string) (domain.Order, bool, error) {
	if existing, found, err := s.repo.FindByIdempotencyKey(idempotencyKey); err != nil {
		return domain.Order{}, false, err
	} else if found {
		return existing, true, nil
	}

	cartItems, err := s.cartClient.FetchCartItems(userID)
	if err != nil {
		return domain.Order{}, false, err
	}

	created, err := s.repo.CreateWithItems(ctx, userID, idempotencyKey, cartItems)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateIdempotencyKey) {
			existing, found, findErr := s.repo.FindByIdempotencyKey(idempotencyKey)
			if findErr == nil && found {
				return existing, true, nil
			}
		}
		return domain.Order{}, false, err
	}

	_ = s.cartClient.ClearCart(userID)
	return created, false, nil
}

func (s *DefaultOrderService) GetOrder(id string) (domain.Order, bool, error) {
	return s.repo.FindByID(id)
}
