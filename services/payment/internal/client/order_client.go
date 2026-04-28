// Package client previously held an OrderHTTPClient that called
// /internal/orders/:id/confirm|fail directly after payment callback.
//
// This has been replaced by Kafka event publishing (see event/kafka_publisher.go).
// Payment service publishes to "payment.events" topic; order service consumes
// and drives the state machine asynchronously.
//
// This file is kept as a reference for the synchronous HTTP alternative.
// To swap back: instantiate OrderHTTPClient, pass to NewPaymentService instead of KafkaPublisher.

package client

import (
	"fmt"
	"net/http"
	"time"
)

type OrderClient interface {
	ConfirmOrder(orderID string) error
	FailOrder(orderID string) error
}

type OrderHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOrderHTTPClient(baseURL string) *OrderHTTPClient {
	return &OrderHTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *OrderHTTPClient) ConfirmOrder(orderID string) error {
	return c.patch(fmt.Sprintf("%s/internal/orders/%s/confirm", c.baseURL, orderID))
}

func (c *OrderHTTPClient) FailOrder(orderID string) error {
	return c.patch(fmt.Sprintf("%s/internal/orders/%s/fail", c.baseURL, orderID))
}

func (c *OrderHTTPClient) patch(url string) error {
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("order client: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("order client: unexpected status %d", resp.StatusCode)
	}
	return nil
}
