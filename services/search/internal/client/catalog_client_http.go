package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lppduy/ecom-poc/services/search/internal/domain"
)

type CatalogHTTPClient struct {
	baseURL string
}

func NewCatalogHTTPClient(baseURL string) *CatalogHTTPClient {
	return &CatalogHTTPClient{baseURL: baseURL}
}

func (c *CatalogHTTPClient) FetchAllProducts() ([]domain.Product, error) {
	resp, err := http.Get(c.baseURL + "/products")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog returned %d", resp.StatusCode)
	}

	var payload []domain.Product
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}
