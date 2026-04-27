package repository

import (
	"github.com/lppduy/ecom-poc/services/inventory/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func InitSchema(db *gorm.DB) error {
	return db.AutoMigrate(&domain.Stock{}, &domain.Reservation{})
}
