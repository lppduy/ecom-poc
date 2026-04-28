package service

import "github.com/lppduy/ecom-poc/services/auth/internal/domain"

type AuthService interface {
	Register(username, password string) (domain.User, error)
	Login(username, password string) (token string, err error)
	Me(userID string) (domain.User, error)
}
