package service

import (
	"fmt"

	"github.com/lppduy/ecom-poc/services/payment/internal/client"
	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
)

type DefaultPaymentService struct {
	repo        repository.PaymentRepository
	orderClient client.OrderClient
}

func NewPaymentService(repo repository.PaymentRepository, orderClient client.OrderClient) *DefaultPaymentService {
	return &DefaultPaymentService{repo: repo, orderClient: orderClient}
}

func (s *DefaultPaymentService) CreatePayment(orderID string, amount float64) (domain.Payment, error) {
	if orderID == "" {
		return domain.Payment{}, domain.ErrInvalidOrderID
	}
	p := &domain.Payment{
		OrderID: orderID,
		Amount:  amount,
		Status:  domain.StatusPending,
	}
	if err := s.repo.Create(p); err != nil {
		return domain.Payment{}, fmt.Errorf("create payment: %w", err)
	}
	return *p, nil
}

func (s *DefaultPaymentService) GetPayment(id string) (domain.Payment, error) {
	return s.repo.FindByID(id)
}

func (s *DefaultPaymentService) Callback(id, result string) (domain.Payment, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Payment{}, err
	}
	if p.Status != domain.StatusPending {
		return domain.Payment{}, domain.ErrAlreadyProcessed
	}

	var newStatus domain.Status
	switch result {
	case "success":
		newStatus = domain.StatusSuccess
	case "fail":
		newStatus = domain.StatusFailed
	default:
		return domain.Payment{}, fmt.Errorf("invalid result: must be 'success' or 'fail'")
	}

	if err := s.repo.UpdateStatus(id, newStatus); err != nil {
		return domain.Payment{}, err
	}
	p.Status = newStatus

	// Notify order service asynchronously-style (best effort for mock)
	if newStatus == domain.StatusSuccess {
		_ = s.orderClient.ConfirmOrder(p.OrderID)
	} else {
		_ = s.orderClient.FailOrder(p.OrderID)
	}

	return p, nil
}
