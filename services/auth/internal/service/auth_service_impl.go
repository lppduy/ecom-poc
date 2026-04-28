package service

import (
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/auth/internal/domain"
	"github.com/lppduy/ecom-poc/services/auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type DefaultAuthService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewAuthService(repo repository.UserRepository, jwtSecret string) *DefaultAuthService {
	return &DefaultAuthService{repo: repo, jwtSecret: jwtSecret}
}

func (s *DefaultAuthService) Register(username, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(user); err != nil {
		return domain.User{}, err
	}
	return *user, nil
}

func (s *DefaultAuthService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", domain.ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidPassword
	}
	return jwtutil.Sign(user.ID, s.jwtSecret)
}

func (s *DefaultAuthService) Me(userID string) (domain.User, error) {
	return s.repo.FindByID(userID)
}
