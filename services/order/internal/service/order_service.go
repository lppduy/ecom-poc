package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"github.com/lppduy/ecom-poc/services/order/internal/repository"
)

var ErrEmptyCart = errors.New("cart is empty")

type OrderService struct {
	repo        *repository.OrderRepository
	cartBaseURL string
}

func NewOrderService(repo *repository.OrderRepository, cartBaseURL string) *OrderService {
	return &OrderService{repo: repo, cartBaseURL: cartBaseURL}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID, idempotencyKey string) (domain.Order, bool, error) {
	if existing, found, err := s.repo.FindByIdempotencyKey(idempotencyKey); err != nil {
		return domain.Order{}, false, err
	} else if found {
		return existing, true, nil
	}

	cartItems, err := s.fetchCartItems(userID)
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

	_ = s.clearCart(userID)
	return created, false, nil
}

func (s *OrderService) GetOrder(id string) (domain.Order, bool, error) {
	return s.repo.FindByID(id)
}

func (s *OrderService) fetchCartItems(userID string) ([]domain.OrderItem, error) {
	req, err := http.NewRequest(http.MethodGet, s.cartBaseURL+"/cart?userId="+url.QueryEscape(userID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-User-Id", userID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cart service returned %d", resp.StatusCode)
	}

	var payload struct {
		Items []domain.OrderItem `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if len(payload.Items) == 0 {
		return nil, ErrEmptyCart
	}
	return payload.Items, nil
}

func (s *OrderService) clearCart(userID string) error {
	req, err := http.NewRequest(http.MethodPost, s.cartBaseURL+"/cart/clear", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-User-Id", userID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cart clear failed with status %d", resp.StatusCode)
	}
	return nil
}
