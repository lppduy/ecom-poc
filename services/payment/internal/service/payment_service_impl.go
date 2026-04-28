package service

import (
	"context"
	"fmt"
	"log"

	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"github.com/lppduy/ecom-poc/services/payment/internal/event"
	"github.com/lppduy/ecom-poc/services/payment/internal/repository"
)

type DefaultPaymentService struct {
	repo      repository.PaymentRepository
	publisher *event.KafkaPublisher
}

func NewPaymentService(repo repository.PaymentRepository, publisher *event.KafkaPublisher) *DefaultPaymentService {
	return &DefaultPaymentService{repo: repo, publisher: publisher}
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

	if err := s.repo.UpdateStatus(id, newStatus); err != nil {
		return domain.Payment{}, err
	}
	p.Status = newStatus

	// Publish payment event to Kafka; order service consumes and confirms/fails the order.
	go func() {
		if pubErr := s.publisher.PublishPaymentEvent(context.Background(), event.PaymentEvent{
			PaymentID: p.ID,
			OrderID:   p.OrderID,
			Status:    eventStatus,
		}); pubErr != nil {
			log.Printf("[payment] warn: failed to publish event for order %s: %v", p.OrderID, pubErr)
		} else {
			log.Printf("[payment] published %s event for order %s", eventStatus, p.OrderID)
		}
	}()

	return p, nil
}
