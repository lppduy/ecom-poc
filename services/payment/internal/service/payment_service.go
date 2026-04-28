package service

import "github.com/lppduy/ecom-poc/services/payment/internal/domain"

type PaymentService interface {
	CreatePayment(orderID string, amount float64) (domain.Payment, error)
	GetPayment(id string) (domain.Payment, error)
	// Callback simulates the payment gateway webhook: result is "success" or "fail"
	Callback(id, result string) (domain.Payment, error)
}
