package repository

import (
	"time"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"gorm.io/gorm"
)

type GormOutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *GormOutboxRepository {
	return &GormOutboxRepository{db: db}
}

func (r *GormOutboxRepository) FetchPending(limit int) ([]domain.OutboxEvent, error) {
	var events []domain.OutboxEvent
	err := r.db.Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *GormOutboxRepository) MarkPublished(id int64) error {
	now := time.Now()
	return r.db.Model(&domain.OutboxEvent{}).
		Where("id = ?", id).
		Update("published_at", now).Error
}
