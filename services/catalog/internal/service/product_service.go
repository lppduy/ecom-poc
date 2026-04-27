package service

import (
	"github.com/lppduy/ecom-poc/services/catalog/internal/domain"
	"github.com/lppduy/ecom-poc/services/catalog/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) ListProducts() ([]domain.Product, error) {
	return s.repo.List()
}
