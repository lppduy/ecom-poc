package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type CartHTTPClient struct {
	baseURL string
}

func NewCartHTTPClient(baseURL string) *CartHTTPClient {
	return &CartHTTPClient{baseURL: baseURL}
}

func (c *CartHTTPClient) FetchCartItems(ctx context.Context, userID string) ([]domain.OrderItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/cart", nil)
	if err != nil {
		return nil, err
	}
	// Forward the user's JWT so cart service authenticates and resolves userID on its own.
	// The token already encodes userID — no need to pass it as a query param.
	if token := jwtutil.TokenFromContext(ctx); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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
		return nil, domain.ErrEmptyCart
	}
	return payload.Items, nil
}

func (c *CartHTTPClient) ClearCart(ctx context.Context, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/cart/clear", nil)
	if err != nil {
		return err
	}
	if token := jwtutil.TokenFromContext(ctx); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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
