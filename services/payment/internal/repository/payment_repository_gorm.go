package repository

import (
	"errors"

	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"gorm.io/gorm"
)

type GormPaymentRepository struct {
	db *gorm.DB
}

func NewGormPaymentRepository(db *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{db: db}
}

func (r *GormPaymentRepository) Create(p *domain.Payment) error {
	return r.db.Create(p).Error
}

func (r *GormPaymentRepository) FindByID(id string) (domain.Payment, error) {
	var p domain.Payment
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	return p, nil
}

func (r *GormPaymentRepository) FindByOrderID(orderID string) (domain.Payment, error) {
	var p domain.Payment
	if err := r.db.Where("order_id = ?", orderID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Payment{}, domain.ErrPaymentNotFound
		}
		return domain.Payment{}, err
	}
	return p, nil
}

func (r *GormPaymentRepository) UpdateStatus(id string, status domain.Status) error {
	return r.db.Model(&domain.Payment{}).Where("id = ?", id).Update("status", status).Error
}
