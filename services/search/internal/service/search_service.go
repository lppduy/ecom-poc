package service

import "github.com/lppduy/ecom-poc/services/search/internal/domain"

type SearchService interface {
	Search(query string, minPrice, maxPrice int) (domain.SearchResult, error)
	IndexProduct(product domain.Product) error
	BulkIndex(products []domain.Product) error
}
