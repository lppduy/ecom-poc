package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type InventoryHTTPClient struct {
	baseURL string
}

func NewInventoryHTTPClient(baseURL string) *InventoryHTTPClient {
	return &InventoryHTTPClient{baseURL: baseURL}
}

func (c *InventoryHTTPClient) Reserve(orderID int64, items []domain.OrderItem) error {
	type item struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
	}
	type body struct {
		OrderID int64  `json:"orderId"`
		Items   []item `json:"items"`
	}

	mapped := make([]item, len(items))
	for i, it := range items {
		mapped[i] = item{ProductID: it.ProductID, Quantity: it.Quantity}
	}

	return c.post("/inventory/reserve", body{OrderID: orderID, Items: mapped})
}

func (c *InventoryHTTPClient) Release(orderID int64) error {
	return c.post("/inventory/release", map[string]int64{"orderId": orderID})
}

func (c *InventoryHTTPClient) Confirm(orderID int64) error {
	return c.post("/inventory/confirm", map[string]int64{"orderId": orderID})
}

func (c *InventoryHTTPClient) post(path string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("inventory service returned %d for %s", resp.StatusCode, path)
	}
	return nil
}
