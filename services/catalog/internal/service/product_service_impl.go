package service

import (
	"github.com/lppduy/ecom-poc/services/catalog/internal/domain"
	"github.com/lppduy/ecom-poc/services/catalog/internal/repository"
)

type DefaultProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *DefaultProductService {
	return &DefaultProductService{repo: repo}
}

func (s *DefaultProductService) ListProducts() ([]domain.Product, error) {
	return s.repo.List()
}
