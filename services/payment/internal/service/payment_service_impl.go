package service

import (
	"encoding/json"
	"fmt"

	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"github.com/lppduy/ecom-poc/services/payment/internal/event"
	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
)

type DefaultPaymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) *DefaultPaymentService {
	return &DefaultPaymentService{repo: repo}
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

	var (
		newStatus   domain.Status
		eventStatus string
	)
	switch result {
	case "success":
		newStatus = domain.StatusSuccess
		eventStatus = "SUCCESS"
	case "fail":
		newStatus = domain.StatusFailed
		eventStatus = "FAILED"
	default:
		return domain.Payment{}, fmt.Errorf("invalid result: must be 'success' or 'fail'")
	}

	// Build the outbox payload before writing to DB
	payload, err := json.Marshal(event.PaymentEvent{
		PaymentID: p.ID,
		OrderID:   p.OrderID,
		Status:    eventStatus,
	})
	if err != nil {
		return domain.Payment{}, fmt.Errorf("marshal event: %w", err)
	}

	// Atomic: UPDATE payment status + INSERT outbox event in one transaction.
	// The relay goroutine will read the outbox and publish to Kafka.
	if err := s.repo.UpdateStatusWithOutbox(id, newStatus, string(payload)); err != nil {
		return domain.Payment{}, err
	}

	p.Status = newStatus
	return p, nil
}
