package service

import "github.com/lppduy/ecom-poc/services/catalog/internal/domain"

type ProductService interface {
	ListProducts() ([]domain.Product, error)
}
