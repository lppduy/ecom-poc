package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/lppduy/ecom-poc/services/order/internal/client"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"github.com/lppduy/ecom-poc/services/order/internal/repository"
)

type DefaultOrderService struct {
	repo            repository.OrderRepository
	cartClient      client.CartClient
	inventoryClient client.InventoryClient
}

func NewOrderService(
	repo repository.OrderRepository,
	cartClient client.CartClient,
	inventoryClient client.InventoryClient,
) *DefaultOrderService {
	return &DefaultOrderService{
		repo:            repo,
		cartClient:      cartClient,
		inventoryClient: inventoryClient,
	}
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
	if len(cartItems) == 0 {
		return domain.Order{}, false, domain.ErrEmptyCart
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

	// Reserve stock — if this fails, transition the order to FAILED immediately.
	if err := s.inventoryClient.Reserve(created.ID, cartItems); err != nil {
		_ = s.repo.UpdateStatus(created.ID, domain.StatusFailed)
		return domain.Order{}, false, fmt.Errorf("stock reservation failed: %w", err)
	}

	_ = s.cartClient.ClearCart(userID)
	return created, false, nil
}

func (s *DefaultOrderService) GetOrder(id string) (domain.Order, bool, error) {
	return s.repo.FindByID(id)
}

func (s *DefaultOrderService) ConfirmOrder(id string) (domain.Order, error) {
	order, err := s.transitionOrder(id, domain.StatusConfirmed)
	if err != nil {
		return domain.Order{}, err
	}
	_ = s.inventoryClient.Confirm(order.ID)
	return order, nil
}

func (s *DefaultOrderService) FailOrder(id string) (domain.Order, error) {
	order, err := s.transitionOrder(id, domain.StatusFailed)
	if err != nil {
		return domain.Order{}, err
	}
	_ = s.inventoryClient.Release(order.ID)
	return order, nil
}

func (s *DefaultOrderService) transitionOrder(id, newStatus string) (domain.Order, error) {
	order, found, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Order{}, err
	}
	if !found {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if !domain.CanTransition(order.Status, newStatus) {
		return domain.Order{}, domain.ErrInvalidTransition
	}
	if err := s.repo.UpdateStatus(order.ID, newStatus); err != nil {
		return domain.Order{}, err
	}
	order.Status = newStatus
	return order, nil
}
