package repository

import "github.com/lppduy/ecom-poc/services/order/internal/domain"

type OutboxRepository interface {
	FetchPending(limit int) ([]domain.OutboxEvent, error)
	MarkPublished(id int64) error
}
