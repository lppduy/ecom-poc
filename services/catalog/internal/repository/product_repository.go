package repository

import "github.com/lppduy/ecom-poc/services/catalog/internal/domain"

type ProductRepository interface {
	List() ([]domain.Product, error)
}
