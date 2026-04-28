package repository

import "github.com/lppduy/ecom-poc/services/auth/internal/domain"

type UserRepository interface {
	Create(user *domain.User) error
	FindByUsername(username string) (domain.User, error)
	FindByID(id string) (domain.User, error)
}
