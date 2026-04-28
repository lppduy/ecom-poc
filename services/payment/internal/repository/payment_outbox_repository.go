package repository

import "github.com/lppduy/ecom-poc/services/payment/internal/domain"

type PaymentOutboxRepository interface {
	FetchPending(limit int) ([]domain.PaymentOutboxEvent, error)
	MarkPublished(id int64) error
}
