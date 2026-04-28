package repository

import "github.com/lppduy/ecom-poc/services/payment/internal/domain"

type PaymentRepository interface {
	Create(p *domain.Payment) error
	FindByID(id string) (domain.Payment, error)
	FindByOrderID(orderID string) (domain.Payment, error)
	UpdateStatus(id string, status domain.Status) error
}
