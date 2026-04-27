package service

import (
	"github.com/lppduy/ecom-poc/services/search/internal/domain"
	"github.com/lppduy/ecom-poc/services/search/internal/repository"
)

type DefaultSearchService struct {
	repo repository.SearchRepository
}

func NewSearchService(repo repository.SearchRepository) *DefaultSearchService {
	return &DefaultSearchService{repo: repo}
}

func (s *DefaultSearchService) Search(query string, minPrice, maxPrice int) (domain.SearchResult, error) {
	return s.repo.Search(query, minPrice, maxPrice)
}

func (s *DefaultSearchService) IndexProduct(product domain.Product) error {
	return s.repo.Index(product)
}

func (s *DefaultSearchService) BulkIndex(products []domain.Product) error {
	return s.repo.BulkIndex(products)
}
