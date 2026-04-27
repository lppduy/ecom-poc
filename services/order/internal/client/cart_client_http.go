package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type CartHTTPClient struct {
	baseURL string
}

func NewCartHTTPClient(baseURL string) *CartHTTPClient {
	return &CartHTTPClient{baseURL: baseURL}
}

func (c *CartHTTPClient) FetchCartItems(userID string) ([]domain.OrderItem, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/cart?userId="+url.QueryEscape(userID), nil)
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
		return nil, domain.ErrEmptyCart
	}
	return payload.Items, nil
}

func (c *CartHTTPClient) ClearCart(userID string) error {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/cart/clear", nil)
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
