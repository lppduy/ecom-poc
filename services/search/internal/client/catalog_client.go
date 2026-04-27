package client

import "github.com/lppduy/ecom-poc/services/search/internal/domain"

type CatalogClient interface {
	FetchAllProducts() ([]domain.Product, error)
}
