package repository

import (
	"strings"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func InitSchema(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.Order{}, &domain.OrderItem{}); err != nil {
		return err
	}
	// Backfill legacy rows if idempotency key is still empty.
	if err := db.Exec(`
		UPDATE orders
		SET idempotency_key = CONCAT('legacy-', id::text)
		WHERE idempotency_key IS NULL OR idempotency_key = ''
	`).Error; err != nil {
		return err
	}
	return nil
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
