package repository

import "github.com/lppduy/ecom-poc/services/search/internal/domain"

type SearchRepository interface {
	// Index upserts a product document into Elasticsearch.
	Index(product domain.Product) error

	// BulkIndex upserts multiple product documents.
	BulkIndex(products []domain.Product) error

	// Search performs a full-text search over product name.
	// Supports optional minPrice / maxPrice filters (0 = no filter).
	Search(query string, minPrice, maxPrice int) (domain.SearchResult, error)
}
