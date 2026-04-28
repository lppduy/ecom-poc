package repository

import (
	"time"

	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"gorm.io/gorm"
)

type GormPaymentOutboxRepository struct {
	db *gorm.DB
}

func NewPaymentOutboxRepository(db *gorm.DB) *GormPaymentOutboxRepository {
	return &GormPaymentOutboxRepository{db: db}
}

func (r *GormPaymentOutboxRepository) FetchPending(limit int) ([]domain.PaymentOutboxEvent, error) {
	var events []domain.PaymentOutboxEvent
	err := r.db.Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *GormPaymentOutboxRepository) MarkPublished(id int64) error {
	now := time.Now()
	return r.db.Model(&domain.PaymentOutboxEvent{}).
		Where("id = ?", id).
		Update("published_at", now).Error
}
