package repository

import (
	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&domain.Payment{}, &domain.PaymentOutboxEvent{}); err != nil {
		return nil, err
	}
	return db, nil
}
